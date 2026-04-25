package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"

	"finance-app/migrations"
)

// RunMigrations executa todas as migrations SQL pendentes contra o PostgreSQL.
//
// Em Laravel, o equivalente é `php artisan migrate`. Aqui a execução acontece
// no startup da aplicação — sem comando manual. Cada migration é um par de
// arquivos: NNN_nome.up.sql (aplica) e NNN_nome.down.sql (reverte).
//
// O histórico de versões aplicadas é mantido automaticamente na tabela
// schema_migrations — equivalente à tabela migrations do Laravel.
//
// A biblioteca usa um advisory lock exclusivo no PostgreSQL durante a execução,
// garantindo que duas instâncias subindo em paralelo (deploy rolling) nunca
// executem a mesma migration ao mesmo tempo — sem race condition.
func RunMigrations(db *gorm.DB) error {
	// Recupera o *sql.DB subjacente ao GORM para repassar ao driver de migration.
	// Isso reutiliza o pool de conexões já configurado em NewPostgres(),
	// sem abrir uma segunda conexão com o banco.
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("getting sql.DB from gorm: %w", err)
	}

	// WithInstance cria o driver de migration usando a conexão existente.
	// É diferente de migrate.New("postgres://...") que abriria uma nova conexão.
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver for migrate: %w", err)
	}

	// iofs.New lê as migrations diretamente do embed.FS em memória.
	// Em desenvolvimento com Air (hot reload), as migrations são lidas do FS
	// embedado no momento do build — para aplicar uma nova migration basta
	// reiniciar o servidor (o Air faz isso automaticamente ao salvar o arquivo).
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("creating migrations source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Println("migrations: database is up to date")
	return nil
}

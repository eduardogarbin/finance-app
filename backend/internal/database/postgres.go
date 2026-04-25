// Package database centraliza a inicializacao das conexoes com infraestrutura.
//
// Este pacote e o unico responsavel por criar conexoes com banco de dados e cache.
// Nenhum outro pacote deve instanciar conexoes diretamente — elas sao criadas aqui
// e injetadas via construtor nos repositories (ver internal/repository/).
//
// Em Laravel, o equivalente e o config/database.php combinado com os Service Providers
// que registram as conexoes no container IoC. Aqui fazemos o mesmo de forma explicita,
// sem container — as conexoes sao construidas em main.go e passadas adiante.
package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres cria, configura e retorna uma conexao com o banco PostgreSQL via GORM.
//
// As configuracoes de conexao sao lidas de variaveis de ambiente, com valores
// padrao para desenvolvimento local. Em producao, essas variaveis devem ser
// fornecidas pelo orquestrador (Docker Compose, Kubernetes Secrets, etc.).
//
// Retorna (*gorm.DB, nil) em caso de sucesso.
// Retorna (nil, error) se a conexao falhar — o chamador (main.go) deve tratar
// este erro encerrando a aplicacao, pois subir sem banco causaria falhas em cascata.
func NewPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "finance_app"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger em modo Info exibe todas as queries SQL no terminal durante
		// o desenvolvimento, facilitando a depuracao. Em producao, considere
		// mudar para logger.Warn ou logger.Silent para reduzir o volume de logs.
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	// GORM usa database/sql internamente. Sem configurar o pool, o comportamento
	// padrao e: conexoes abertas ilimitadas e fechadas imediatamente apos cada uso.
	// Isso gera overhead de TCP handshake em cada query e pode esgotar o
	// max_connections do PostgreSQL (padrao: 100) sob carga concorrente.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting sql.DB from gorm: %w", err)
	}

	// MaxOpenConns limita o total de conexoes TCP abertas simultaneamente.
	// Escolhemos 25 como ponto de partida conservador: deixa margem para outros
	// clientes (migrations, admintools) sem chegar perto do limite do Postgres.
	sqlDB.SetMaxOpenConns(25)

	// MaxIdleConns define quantas conexoes ficam "em standby" no pool entre requests.
	// Conexoes idle evitam o overhead de reconexao em rajadas de trafego. Manter
	// metade do maximo aberto e uma heuristica segura para workloads web tipicos.
	sqlDB.SetMaxIdleConns(10)

	// ConnMaxLifetime forca o pool a reciclar conexoes apos 30 minutos, mesmo que
	// ainda estejam saudaveis. Isso evita problemas com conexoes "zumbis" que o
	// Postgres ou um firewall intermediario ja fechou silenciosamente do outro lado.
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

// getEnv retorna o valor da variavel de ambiente identificada por key.
// Se a variavel nao estiver definida ou estiver vazia, retorna fallback.
// Centralizar essa logica evita repeticao de os.Getenv + verificacao em todo o pacote.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Package repository implementa a camada de acesso a dados da aplicacao.
//
// Os repositories sao o equivalente aos Models/Repositories do Laravel: a unica
// camada que tem permissao de falar diretamente com o banco de dados ou cache.
// Nenhum outro pacote (handlers, services) deve importar drivers de banco diretamente.
//
// Essa separacao tem um beneficio pratico imediato nos testes: ao depender de
// interfaces, os services podem ser testados com repositorios falsos (mocks)
// sem precisar de um banco real rodando — o mesmo que o Laravel faz com
// DatabaseMigrations + Factories nos testes de feature.
//
// Dependencia de entrada: *gorm.DB e *redis.Client (infraestrutura).
// Retorno esperado: erros nativos do Go (error), nunca excecoes.
package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthRepository define o contrato de acesso a dados para verificacoes de saude.
// Qualquer struct que implemente PingDB e PingRedis satisfaz esta interface,
// incluindo mocks de teste — sem necessidade de heranca ou declaracao explicita.
type HealthRepository interface {
	// PingDB verifica se a conexao com o PostgreSQL esta ativa.
	// Retorna nil em caso de sucesso ou um error descritivo em caso de falha.
	PingDB(ctx context.Context) error

	// PingRedis verifica se a conexao com o Redis esta ativa.
	// Retorna nil em caso de sucesso ou um error descritivo em caso de falha.
	PingRedis(ctx context.Context) error
}

// healthRepository e a implementacao concreta de HealthRepository.
// Nome em minusculo = privado ao pacote, acessivel apenas pela interface publica.
type healthRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHealthRepository cria e retorna uma nova instancia de HealthRepository.
// Recebe as conexoes ja inicializadas (injetadas pelo main.go) em vez de
// cria-las internamente — isso mantem o repositorio testavel e sem efeitos colaterais
// na construcao, seguindo o principio de injecao de dependencias.
func NewHealthRepository(db *gorm.DB, redis *redis.Client) HealthRepository {
	return &healthRepository{db: db, redis: redis}
}

// PingDB verifica a conectividade ativa com o PostgreSQL.
// Usa PingContext (em vez de Ping) para respeitar o timeout e cancelamento
// propagados pelo context — se o handler cancelar a requisicao, esta operacao
// tambem e interrompida, liberando recursos do banco imediatamente.
//
// Retorna error se a conexao estiver perdida ou se o context expirar.
func (r *healthRepository) PingDB(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// PingRedis verifica a conectividade ativa com o Redis enviando um comando PING.
// O Redis responde com PONG em caso de sucesso — qualquer outro resultado e tratado
// como erro pelo cliente.
//
// Retorna error se o Redis estiver inacessivel ou se o context expirar.
func (r *healthRepository) PingRedis(ctx context.Context) error {
	return r.redis.Ping(ctx).Err()
}

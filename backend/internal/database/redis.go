package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis cria, configura e retorna um cliente Redis pronto para uso.
//
// Alem de criar o cliente, executa um Ping de validacao durante a inicializacao.
// Esse comportamento e intencional: preferimos falhar rapido (fail fast) no startup
// a descobrir que o Redis esta inacessivel apenas na primeira requisicao do usuario.
// E o mesmo principio do DB::connection()->getPdo() que o Laravel usa para validar
// conexoes em alguns ambientes de CI.
//
// Retorna (*redis.Client, nil) em caso de sucesso.
// Retorna (nil, error) se a conexao ou o Ping falharem — o chamador deve encerrar
// a aplicacao, pois cache indisponivel no startup indica problema de infraestrutura.
func NewRedis() (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s",
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,

		// PoolSize define quantas conexoes TCP abertas o cliente pode manter
		// simultaneamente com o Redis. Cada goroutine que precisar do Redis
		// pega uma conexao do pool e a devolve ao terminar — sem overhead
		// de abrir/fechar conexoes a cada operacao.
		PoolSize: 10,

		// Timeouts de conexao e de operacoes individuais. Valores distintos
		// permitem controle fino: DialTimeout afeta apenas o estabelecimento
		// da conexao TCP, enquanto Read/WriteTimeout afetam cada comando enviado.
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Timeout de 5 segundos para o Ping de validacao. Usar context com timeout
	// em vez de bloqueio indefinido e a pratica padrao em Go para qualquer
	// operacao de I/O — evita que um Redis lento trave o startup indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis at %s: %w", addr, err)
	}

	return client, nil
}

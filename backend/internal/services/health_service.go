// Package services implementa a camada de logica de negocio da aplicacao.
//
// Os services sao o equivalente aos Services do Laravel (ex: App\Services\HealthService):
// orquestram operacoes, aplicam regras de negocio e coordenam chamadas ao repositorio.
// Eles nao sabem nada sobre HTTP (isso e responsabilidade dos handlers) e nao
// falam diretamente com o banco (isso e responsabilidade dos repositories).
//
// Dependencia de entrada: repository (interfaces).
// Retorno esperado: structs de dominio com o resultado das operacoes.
package services

import (
	"context"
	"time"

	"finance-app/internal/repository"
)

// HealthStatus representa o resultado agregado de uma verificacao de saude.
// Cada campo corresponde a uma chave na resposta JSON enviada ao cliente.
// O campo Services mapeia o nome de cada dependencia ao seu status individual,
// permitindo identificar qual servico especifico esta com problema.
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// HealthService define o contrato da logica de health check.
// Em Go, interfaces sao definidas onde sao consumidas (no pacote que depende delas),
// mas por convencao de Clean Architecture as mantemos junto a implementacao.
// Qualquer struct que implemente o metodo Check satisfaz esta interface automaticamente
// — diferente do PHP/Laravel onde voce declara explicitamente "implements HealthService".
type HealthService interface {
	// Check executa a verificacao de todas as dependencias e retorna o status agregado.
	// Recebe um context.Context para suportar cancelamento e timeout propagados
	// do handler — se o cliente desconectar, as operacoes de banco sao interrompidas.
	Check(ctx context.Context) HealthStatus
}

// healthService e a implementacao concreta de HealthService.
// O nome em minusculo e intencional: torna a struct privada ao pacote, forcando
// que todo acesso externo ocorra pela interface — principio de encapsulamento.
type healthService struct {
	repo repository.HealthRepository
}

// NewHealthService cria e retorna uma nova instancia de HealthService.
// Retorna a interface, nao o tipo concreto (*healthService), o que garante
// que o chamador (main.go) so enxergue o contrato publico, nunca os detalhes internos.
func NewHealthService(repo repository.HealthRepository) HealthService {
	return &healthService{repo: repo}
}

// Check verifica a conectividade com cada dependencia e retorna um HealthStatus agregado.
//
// O status geral sera "ok" apenas se todas as dependencias responderem com sucesso.
// Qualquer falha individual muda o status para "degraded" e registra a mensagem
// de erro no campo Services — nunca oculta falhas parciais com um status generico.
//
// Nao retorna error: falhas de dependencia sao informacao de negocio (refletidas
// no status), nao erros de execucao do proprio servico.
func (s *healthService) Check(ctx context.Context) HealthStatus {
	svcStatus := make(map[string]string, 2)
	overallStatus := "ok"

	if err := s.repo.PingDB(ctx); err != nil {
		svcStatus["postgres"] = "unavailable: " + err.Error()
		overallStatus = "degraded"
	} else {
		svcStatus["postgres"] = "ok"
	}

	if err := s.repo.PingRedis(ctx); err != nil {
		svcStatus["redis"] = "unavailable: " + err.Error()
		overallStatus = "degraded"
	} else {
		svcStatus["redis"] = "ok"
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Services:  svcStatus,
	}
}

// Package handlers implementa a camada de interface HTTP da aplicacao.
//
// Os handlers sao o equivalente aos Controllers do Laravel: recebem a requisicao
// HTTP, extraem os dados necessarios, delegam o processamento para a camada de
// servico e devolvem a resposta ao cliente. Eles nao contem logica de negocio —
// apenas traduzem HTTP para chamadas de servico e vice-versa.
//
// Dependencia de entrada: services (interfaces, nunca implementacoes concretas).
// Retorno esperado: respostas HTTP com status code e corpo JSON adequados.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"finance-app/internal/services"
)

// HealthHandler agrupa os handlers relacionados a verificacao de saude da aplicacao.
// Ele depende de services.HealthService (interface), o que permite substituir a
// implementacao real por um mock nos testes sem alterar nenhum codigo deste arquivo.
type HealthHandler struct {
	service services.HealthService
}

// NewHealthHandler cria e retorna um novo HealthHandler com o servico injetado.
// Segue o padrao construtor de Go: uma funcao New* que recebe dependencias
// e retorna a struct pronta para uso — equivalente ao __construct() do Laravel.
func NewHealthHandler(service services.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// Check verifica a saude das dependencias da aplicacao e responde com o status agregado.
//
// Retorna HTTP 200 quando todos os servicos estao operacionais.
// Retorna HTTP 503 (Service Unavailable) quando qualquer dependencia falha.
//
// Usar 503 em vez de 200 com corpo de erro e uma convencao importante:
// load balancers (AWS ALB, GCP Load Balancer) e orquestradores (Kubernetes
// readiness probes) tomam decisoes baseadas no status code HTTP, nao no corpo
// JSON — o 503 faz a instancia ser retirada do pool automaticamente.
//
// O timeout de 5 segundos impede que uma dependencia lenta bloqueie o handler
// indefinidamente, liberando a goroutine e o worker do servidor.
func (h *HealthHandler) Check(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	status := h.service.Check(ctx)

	httpStatus := http.StatusOK
	if status.Status != "ok" {
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, status)
}

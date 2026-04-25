// Package main e o ponto de entrada da aplicacao finance-app.
//
// Responsabilidades deste arquivo:
//   - Inicializar as conexoes com infraestrutura (PostgreSQL, Redis)
//   - Montar o grafo de dependencias manualmente (Repository -> Service -> Handler)
//   - Configurar e iniciar o servidor HTTP com Echo
//   - Garantir encerramento gracioso ao receber sinal de interrupcao
//
// Em Laravel, esse papel e dividido entre public/index.php (bootstrap),
// app/Providers/AppServiceProvider.php (injecao de dependencias via container IoC)
// e server.php. Em Go, fazemos tudo isso explicitamente em um unico lugar,
// o que torna o fluxo de inicializacao completamente rastreavel sem "magia".
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"finance-app/internal/database"
	"finance-app/internal/handlers"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

// main inicializa a aplicacao na seguinte ordem:
//  1. Conexoes com dependencias externas (Postgres, Redis)
//  2. Grafo de dependencias: Repository -> Service -> Handler
//  3. Servidor HTTP com middlewares e rotas registradas
//  4. Graceful shutdown: aguarda requisicoes em andamento antes de encerrar
//
// Se qualquer conexao falhar na inicializacao, a aplicacao encerra imediatamente
// via log.Fatalf — subir sem banco ou cache causaria erros em cascata nos handlers.
func main() {
	db, err := database.NewPostgres()
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	redisClient, err := database.NewRedis()
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	// Migrations: descomentar o bloco abaixo quando o primeiro arquivo .sql
	// for criado em backend/migrations/ e a diretiva //go:embed *.sql for
	// adicionada em backend/migrations/embed.go.
	//
	// if err := database.RunMigrations(db); err != nil {
	// 	log.Fatalf("migrations: %v", err)
	// }

	// Injecao de dependencias manual: cada camada recebe apenas a interface
	// da camada abaixo, nunca a implementacao concreta. Isso e equivalente
	// ao que o container IoC do Laravel faz automaticamente — aqui fazemos
	// de forma explicita para ter controle total e sem reflexao em runtime.
	healthRepo := repository.NewHealthRepository(db, redisClient)
	healthSvc := services.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthSvc)

	e := echo.New()
	e.HideBanner = true

	// Logger: registra metodo, path, status e latencia de cada requisicao.
	// Recover: captura panics e os converte em HTTP 500, evitando crash do processo.
	// RequestID: injeta um ID unico por requisicao no header X-Request-Id,
	// util para rastrear logs de uma mesma requisicao em multiplos servicos.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.GET("/health", healthHandler.Check)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// O servidor sobe em uma goroutine separada para que o fluxo principal
	// possa bloquear aguardando o sinal de interrupcao (Ctrl+C ou SIGINT).
	go func() {
		log.Printf("server listening on :%s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Graceful shutdown: da ate 10 segundos para requisicoes em andamento
	// terminarem antes de fechar o processo. Sem isso, um deploy ou restart
	// do container poderia interromper respostas no meio da transmissao.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}

	log.Println("server stopped gracefully")
}

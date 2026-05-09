package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-api/internal/handler"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/internal/telemetry"
	"todo-api/internal/worker"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelApp, err := telemetry.New()

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := otelApp.Shutdown(shutdownCtx); err != nil {
			log.Println("otel shutdown error: ", err)
		}
	}()

	repo := repository.NewTaskRespository()

	auditevents := make(chan worker.AuditEvent, 10)
	auditWorker := worker.NewAuditWorker(auditevents)
	go auditWorker.Start(ctx)

	taskService := service.NewTaskService(repo, auditevents)
	taskHandler := handler.NewTaskHandlerGin(taskService)

	appNewrelic, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Todo ApI"),
		newrelic.ConfigLicense("f726ae39af11d7d728e2ede8ea10e26f87f0NRAL"),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)

	router := gin.Default()

	router.Use(otelgin.Middleware("todo-api"))
	router.Use(nrgin.Middleware(appNewrelic))

	taskHandler.RegisterRoutes(router)

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	router.GET("/metrics", gin.WrapH(otelApp.MetricsHandler()))

	add := os.Getenv("ADDR")

	if add == "" {
		add = ":7000"
	}

	server := &http.Server{
		Addr:              add,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("server running on http://localhost" + add)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Println("graceful shutdown failed:", err)
		return
	}

	log.Println("server stopped")
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("request received /health")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

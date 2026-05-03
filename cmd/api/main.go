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
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	otelApp, err := telemetry.New()

	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewTaskRespository()

	auditEvents := make(chan worker.AuditEvent, 10)
	auditWorker := worker.NewAuditWorker(auditEvents)
	go auditWorker.Start(ctx)

	taskService := service.NewTaskService(repo, auditEvents)
	taskHandler := handler.NewTaskHandler(taskService)

	mux := http.NewServeMux()
	taskHandler.RegisterRouters(mux)

	mux.Handle("/metrics", otelApp.MetricsHandler())

	add := os.Getenv("ADDR")

	if add == "" {
		add = ":8000"
	}

	server := &http.Server{
		Addr:              add,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server running on '%s'", add)

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

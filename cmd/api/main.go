package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-api/internal/config"
	"todo-api/internal/database"
	"todo-api/internal/handler"
	"todo-api/internal/repository/postgres"
	"todo-api/internal/service"
	"todo-api/internal/telemetry"
	"todo-api/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

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

	auditevents := make(chan worker.AuditEvent, 10)
	auditWorker := worker.NewAuditWorker(auditevents)
	go auditWorker.Start(ctx)

	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewGorm(database.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
		TimeZone: cfg.DB.TimeZone,
	})

	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&postgres.TaskRecord{}); err != nil {
		log.Fatal(err)
	}

	taskRepo := postgres.NewTaskGormRepository(db)

	taskService := service.NewTaskService(taskRepo, auditevents)
	taskHandler := handler.NewTaskHandlerGin(taskService)

	// appNewrelic, err := newrelic.NewApplication(
	// 	newrelic.ConfigAppName("Todo ApI"),
	// 	newrelic.ConfigLicense(""),
	// 	newrelic.ConfigAppLogForwardingEnabled(true),
	// )

	router := gin.Default()

	router.Use(otelgin.Middleware("todo-api"))
	// router.Use(nrgin.Middleware(appNewrelic))

	taskHandler.RegisterRoutes(router)

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	router.GET("/metrics", gin.WrapH(otelApp.MetricsHandler()))

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("server running on http://localhost:" + cfg.AppPort)

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

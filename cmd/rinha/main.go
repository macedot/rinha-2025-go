package main

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/handlers"
	"rinha-2025-go/internal/services"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewListenSocket(socketPath string) net.Listener {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on Unix socket: %v", err)
	}
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set socket permissions: %v", err)
	}
	return listener
}

func main() {
	runtime.GOMAXPROCS(1)

	cfg := config.ConfigInstance().Init()

	redis := database.NewRedisClient(cfg)
	defer redis.Close()

	client := services.NewHttpClient()

	health := services.NewHealth(cfg, redis, client)
	defer health.Close()
	go health.ProcessServicesHealth()

	worker := services.NewPaymentWorker(cfg, redis, client, health)
	defer worker.Close()
	for range cfg.NumWorkers {
		go worker.ProcessQueue()
	}

	e := echo.New()
	// pprof.Register(e)

	e.Use(middleware.Recover())
	// e.Use(echoprometheus.NewMiddleware("rinha"))

	e.POST("/payments", handlers.PostPayment(worker))
	e.GET("/payments-summary", handlers.GetSummary(worker))
	e.POST("/purge-payments", handlers.PostPurgePayments(worker))
	e.POST("/health", handlers.GetHealth(worker))
	//e.GET("/metrics", echoprometheus.NewHandler())

	log.Println("Starting Echo server")
	log.Printf("Listening on %s", cfg.ServerSocket)
	listener := NewListenSocket(cfg.ServerSocket)
	log.Fatal(e.Server.Serve(listener))
	// e.Logger.Fatal(e.Start(":9999"))
}

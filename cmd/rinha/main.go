package main

import (
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/server"
	"rinha-2025-go/internal/services"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(3)
	// debug.SetGCPercent(-1)
	// debug.SetMemoryLimit(-1)

	cfg := config.ConfigInstance().Init()

	redis := database.NewRedisClient(cfg)
	defer redis.Close()

	client := services.NewHttpClient()

	health := services.NewHealth(cfg, redis, client)
	defer health.Close()
	go health.ProcessServicesHealth()

	worker := services.NewPaymentWorker(cfg, redis, client, health)
	defer worker.Close()
	// worker.UpdateServicesFee()
	log.Println("Starting workers:", cfg.NumWorkers)
	for range cfg.NumWorkers {
		go worker.ProcessQueue()
	}

	log.Fatalln(server.RunServer(cfg, worker))
}

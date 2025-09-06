package main

import (
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/consumer"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/services"
	"rinha-2025-go/pkg/client"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(3)
	cfg := config.ConfigInstance().Init()
	redis := database.NewRedisClient(cfg.RedisSocket)
	defer redis.Close()
	client := services.NewHttpClient(client.NewHttpClient())
	health := consumer.NewHealth(cfg, redis, client)
	defer health.Close()
	go health.ProcessServicesHealth()
	worker := consumer.NewConsumer(cfg, redis, client, health)
	defer worker.Close()
	log.Println("Starting workers:", cfg.NumWorkers)
	for range cfg.NumWorkers {
		go worker.ProcessQueue()
	}
	select {}
}

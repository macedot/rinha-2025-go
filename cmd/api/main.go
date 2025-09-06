package main

import (
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/producer"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(3)
	cfg := config.ConfigInstance().Init()
	redis := database.NewRedisClient(cfg.RedisSocket)
	defer redis.Close()
	instance := producer.NewProducer(cfg, redis)
	defer instance.Close()
	log.Fatalln(producer.RunServer(cfg, instance))
}

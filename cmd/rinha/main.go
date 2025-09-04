package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/server"
	"rinha-2025-go/internal/services"
	"runtime"
)

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

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "OK")
		})
		log.Fatalln(http.ListenAndServe(":7777", nil))
	}()

	log.Fatal(server.RunSilverlining(cfg, worker))
}

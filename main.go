package main

import (
	"log"
	"math/rand"
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/server"
	"rinha-2025/services"
	"runtime"
	"runtime/debug"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(90 * 1024 * 1024)

	cfg := config.ConfigInstance().Init()
	services.HttpClientInstance().Init()
	database.RedisInstance().Connect(cfg)

	go func() {
		services.ResetHealthTimeout()
		sleep := time.Duration(rand.Intn(3))
		log.Printf("Sleep for %d seconds...", sleep)
		time.Sleep(sleep * time.Second)
		for {
			now := time.Now().UTC().UnixNano()
			expiration := services.GetHealthTimeout()
			if expiration < now {
				services.SetHealthTimeout(time.Hour)
				services.RefreshServiceStatus(cfg)
				services.SetHealthTimeout(cfg.ServiceRefreshInterval)
			}
			time.Sleep(cfg.ServiceRefreshInterval)
		}
	}()

	queue := services.NewQueue()
	go func() {
		for {
			payment := queue.Dequeue()
			if err := services.ProcessPayment(&payment); err != nil {
				queue.Enqueue(&payment)
			}
		}
	}()

	switch cfg.ServerType {
	case "fasthttp":
		log.Fatal(server.RunFastHTTP(cfg, queue))
	case "gin":
		log.Fatal(server.RunGin(cfg, queue))
	case "silverlining":
		log.Fatal(server.RunSilverlining(cfg, queue))
	case "gearbox":
		log.Fatal(server.RunGearbox(cfg, queue))
	}

	log.Fatal(server.RunFiber(cfg, queue))
}

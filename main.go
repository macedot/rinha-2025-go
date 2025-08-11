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

	log.Println("Server type:", cfg.ServerType)
	// switch cfg.ServerType {
	// case "echo":
	// 	log.Fatal(server.RunEcho(cfg, queue))
	// case "silverlining":
	// 	log.Fatal(server.RunSilverlining(cfg, queue))
	// }

	log.Fatal(server.RunEcho(cfg, queue))
}

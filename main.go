package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/models"
	"rinha-2025/services"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

func NewListenUnix(socketPath string) net.Listener {
	if socketPath == "" {
		return nil
	}
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
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

	cfg := config.ConfigInstance()
	cfg.Init()

	client := services.HttpClientInstance()
	client.Init()

	redis := database.RedisInstance()
	redis.Connect(cfg)

	listener := NewListenUnix(cfg.ServerSocket)

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
				log.Println("ProcessPayment:", err.Error())
				queue.Enqueue(&payment)
			}
		}
	}()

	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	if cfg.DebugMode {
		app.Use(logger.New())
		app.Use(pprof.New())
	}

	app.Post("/payments", func(c *fiber.Ctx) error {
		var payment models.Payment
		if err := c.BodyParser(&payment); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		services.EnqueuePayment(&payment, queue)
		return c.JSON(fiber.Map{})
	})

	app.Get("/payments-summary", func(c *fiber.Ctx) error {
		summaryReg := models.SummaryRequest{
			StartTime: c.Query("from"),
			EndTime:   c.Query("to"),
		}
		summaryRes, err := services.GetSummary(&summaryReg)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(summaryRes)
	})

	app.Post("/purge-payments", func(c *fiber.Ctx) error {
		if err := services.PurgePayments(); err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{})
	})

	if listener != nil {
		log.Printf("Server listening on Unix socket: %s", cfg.ServerSocket)
		log.Fatal(app.Listener(listener))
	}
	log.Fatal(app.Listen(cfg.ServerURL))
}

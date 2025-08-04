package main

import (
	"log"
	"math/rand"
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/models"
	"rinha-2025/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.ConfigInstance()
	cfg.Init()

	client := services.HttpClientInstance()
	client.Init()

	redis := database.RedisInstance()
	redis.Connect(cfg)

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

	app := fiber.New()
	if cfg.DebugMode {
		app.Use(logger.New())
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
			StartTime: c.Query("from", ""),
			EndTime:   c.Query("to", ""),
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

	log.Fatal(app.Listen(cfg.ServerURL))
}

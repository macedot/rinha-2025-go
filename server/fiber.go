package server

import (
	"log"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"
	"rinha-2025/utils"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

func RunFiber(cfg *config.Config, queue *services.Queue) error {
	app := fiber.New()
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
		var request models.SummaryRequest
		if err := c.QueryParser(&request); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		response, err := services.GetSummary(&request)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(response)
	})

	app.Post("/purge-payments", func(c *fiber.Ctx) error {
		if err := services.PurgePayments(); err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		runtime.GC()
		return c.JSON(fiber.Map{})
	})

	if cfg.ServerSocket != "" {
		log.Printf("Listening on %s", cfg.ServerSocket)
		listener := utils.NewListenUnix(cfg.ServerSocket)
		return app.Listener(listener)
	}

	return app.Listen(cfg.ServerURL)
}

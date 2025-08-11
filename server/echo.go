package server

import (
	"log"
	"net/http"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"
	"rinha-2025/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunEcho(cfg *config.Config, queue *services.Queue) error {
	log.Println("Starting Echo server")

	e := echo.New()
	e.Use(middleware.Recover())

	e.POST("/payments", func(c echo.Context) error {
		var payment models.Payment
		if err := c.Bind(&payment); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		services.EnqueuePayment(&payment, queue)
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/payments-summary", func(c echo.Context) error {
		var request models.SummaryRequest
		if err := c.Bind(&request); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		response, err := services.GetSummary(&request)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, response)
	})

	e.POST("/purge-payments", func(c echo.Context) error {
		if err := services.PurgePayments(); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "OK")
	})

	if cfg.ServerSocket != "" {
		log.Printf("Listening on %s", cfg.ServerSocket)
		listener := utils.NewListenUnix(cfg.ServerSocket)
		return e.Server.Serve(listener)
	}

	return e.Start(cfg.ServerURL)
}

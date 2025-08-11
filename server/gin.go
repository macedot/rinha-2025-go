package server

import (
	"log"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"
	"rinha-2025/utils"

	"github.com/gin-gonic/gin"
)

func RunGin(cfg *config.Config, queue *services.Queue) error {
	log.Println("Starting Gin server")

	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.New()
	app.Use(gin.Recovery())

	if cfg.DebugMode {
		app.Use(gin.Logger())
	}

	app.GET("/payments-summary", func(c *gin.Context) {
		var request models.SummaryRequest
		request.StartTime = c.Query("from")
		request.EndTime = c.Query("to")
		response, err := services.GetSummary(&request)
		if err != nil {
			c.JSON(500, err.Error())
		} else {
			c.JSON(200, response)
		}
	})

	app.POST("/payments", func(c *gin.Context) {
		var payment models.Payment
		if err := c.ShouldBindJSON(&payment); err != nil {
			c.JSON(400, err.Error())
		} else {
			services.EnqueuePayment(&payment, queue)
			c.JSON(200, nil)
		}
	})

	app.POST("/purge-payments", func(c *gin.Context) {
		if err := services.PurgePayments(); err != nil {
			c.JSON(500, err.Error())
		} else {
			c.JSON(200, nil)
		}
	})

	if cfg.ServerSocket != "" {
		log.Printf("Listening on %s", cfg.ServerSocket)
		listener := utils.NewListenUnix(cfg.ServerSocket)
		return app.RunListener(listener)
	}

	return app.Run(cfg.ServerURL)
}

package server

import (
	"log"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"

	"github.com/gogearbox/gearbox"
)

func RunGearbox(cfg *config.Config, queue *services.Queue) error {
	log.Println("Starting Gearbox server")

	gb := gearbox.New()

	gb.Post("/payments", func(ctx gearbox.Context) {
		var payment models.Payment
		if err := ctx.ParseBody(&payment); err != nil {
			ctx.Status(500).SendString(err.Error())
			return
		}
		services.EnqueuePayment(&payment, queue)
		ctx.SendString("")
	})

	gb.Get("/payments-summary", func(ctx gearbox.Context) {
		var request models.SummaryRequest
		request.StartTime = ctx.Query("from")
		request.EndTime = ctx.Query("to")
		response, err := services.GetSummary(&request)
		if err != nil {
			ctx.Status(500).SendString(err.Error())
			return
		}
		ctx.SendJSON(response)
	})

	gb.Post("/purge-payments", func(ctx gearbox.Context) {
		if err := services.PurgePayments(); err != nil {
			ctx.Status(500).SendString(err.Error())
			return
		}
		ctx.SendString("")
	})

	return gb.Start(cfg.ServerURL)
}

package server

// import (
// 	"context"
// 	"log"
// 	"rinha-2025/config"
// 	"rinha-2025/models"
// 	"rinha-2025/services"
// 	"rinha-2025/utils"

// 	"github.com/teambition/gear"
// )

// func RunGear(cfg *config.Config, queue *services.Queue) error {
// 	log.Println("Starting Gear server")

// 	app := gear.New()
// 	router := gear.NewRouter()

// 	router.Post("/payments", func(ctx *gear.Context) error {
// 		var payment models.Payment
// 		if err := ctx.ParseBody(&payment); err != nil {
// 			return ctx.HTML(400, err.Error())
// 		}
// 		services.EnqueuePayment(&payment, queue)
// 		return ctx.HTML(200, "OK")
// 	})

// 	router.Get("/payments-summary", func(ctx *gear.Context) error {
// 		var request models.SummaryRequest
// 		request.StartTime = ctx.Query("from")
// 		request.EndTime = ctx.Query("to")
// 		response, err := services.GetSummary(&request)
// 		if err != nil {
// 			return ctx.HTML(500, err.Error())
// 		}
// 		return ctx.JSON(200, response)
// 	})

// 	router.Post("/purge-payments", func(ctx *gear.Context) error {
// 		if err := services.PurgePayments(); err != nil {
// 			return ctx.HTML(500, err.Error())
// 		}
// 		return ctx.HTML(200, "OK")
// 	})

// 	app.UseHandler(router)

// 	if cfg.ServerSocket != "" {
// 		log.Println("Server socket:", cfg.ServerSocket)
// 		listener := utils.NewListenUnix(cfg.ServerSocket)
// 		return app.ServeWithContext(gear.ContextWithSignal(context.Background()), listener)
// 	}

// 	return app.Listen(cfg.ServerURL)
// }

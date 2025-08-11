package server

import (
	"log"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

func RunFastHTTP(cfg *config.Config, queue *services.Queue) error {
	log.Println("Starting FastHTTP server")

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/payments":
			var payment models.Payment
			if err := oj.Unmarshal(ctx.Request.Body(), &payment); err != nil {
				ctx.Error(err.Error(), fasthttp.StatusBadRequest)
				return
			}
			services.EnqueuePayment(&payment, queue)
			ctx.SetStatusCode(fasthttp.StatusOK)
			return

		case "/payments-summary":
			var request models.SummaryRequest
			request.StartTime = string(ctx.QueryArgs().Peek("from"))
			request.EndTime = string(ctx.QueryArgs().Peek("to"))
			response, err := services.GetSummary(&request)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
				return
			}
			body, err := oj.Marshal(response)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
				return
			}
			ctx.SetBody(body)
			ctx.SetStatusCode(fasthttp.StatusOK)
			return

		case "/purge-payments":
			if err := services.PurgePayments(); err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
				return
			}
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}
	}

	if cfg.ServerSocket != "" {
		log.Printf("Listening on %s", cfg.ServerSocket)
		log.Fatal(fasthttp.ListenAndServeUNIX(cfg.ServerSocket, 0666, requestHandler))
	}

	return fasthttp.ListenAndServe(cfg.ServerURL, requestHandler)
}

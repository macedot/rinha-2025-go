package server

import (
	"log"
	"net"
	"net/http"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/internal/services"
	"rinha-2025-go/pkg/utils"

	"github.com/go-www/silverlining"
)

func ServeListener(ln net.Listener, handler silverlining.Handler) error {
	srv := &silverlining.Server{
		Listener: ln,
		Handler:  handler,
	}
	return srv.Serve(ln)
}

func RunSilverlining(cfg *config.Config, worker *services.PaymentWorker) error {
	log.Println("Starting Silverlining server")

	listener := utils.NewListenUnix(cfg.ServerSocket)
	handler := func(c *silverlining.Context) {
		switch string(c.Path()) {
		case "/payments":
			if c.Method() != silverlining.MethodPOST {
				c.WriteFullBodyString(405, "Method not allowed")
				return
			}
			var payment models.Payment
			if err := c.ReadJSON(&payment); err != nil {
				c.WriteFullBodyString(http.StatusBadRequest, "Invalid JSON payload")
				return
			}
			c.WriteFullBodyString(http.StatusAccepted, "Accepted")
			go worker.EnqueuePayment(&payment)
		case "/payments-summary":
			if c.Method() != silverlining.MethodGET {
				c.WriteFullBodyString(405, "Method not allowed")
				return
			}
			from, err := c.GetQueryParamString("from")
			if err != nil {
				c.WriteFullBodyString(http.StatusBadRequest, err.Error())
				return
			}
			to, err := c.GetQueryParamString("to")
			if err != nil {
				c.WriteFullBodyString(http.StatusBadRequest, err.Error())
				return
			}
			response, err := worker.GetSummary(from, to)
			if err != nil {
				c.WriteFullBodyString(http.StatusInternalServerError, err.Error())
				return
			}
			c.WriteJSON(200, response)
		case "/purge-payments":
			if c.Method() != silverlining.MethodPOST {
				c.WriteFullBodyString(405, "Method not allowed")
				return
			}
			if err := worker.PurgePayments(); err != nil {
				c.WriteFullBodyString(http.StatusInternalServerError, err.Error())
				return
			}
			c.WriteFullBodyString(http.StatusOK, "OK")
		case "/health":
			c.WriteFullBodyString(http.StatusOK, "GOOD")
		default:
			c.WriteFullBodyString(http.StatusNotFound, "Not found")
		}
	}

	log.Printf("Listening on %s", cfg.ServerSocket)
	return ServeListener(listener, handler)
}

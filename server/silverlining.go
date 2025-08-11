package server

import (
	"log"
	"net"
	"rinha-2025/config"
	"rinha-2025/models"
	"rinha-2025/services"
	"rinha-2025/utils"

	"github.com/go-www/silverlining"
)

func ServeListener(ln net.Listener, handler silverlining.Handler) error {
	srv := &silverlining.Server{
		Listener: ln,
		Handler:  handler,
	}
	return srv.Serve(ln)
}

func RunSilverlining(cfg *config.Config, queue *services.Queue) error {
	log.Println("Starting Silverlining server")

	handler := func(c *silverlining.Context) {
		switch string(c.Path()) {
		case "/payments":
			// if c.Method() != silverlining.MethodPOST {
			// 	c.WriteFullBodyString(405, "Method not allowed")
			// 	return
			// }
			var payment models.Payment
			if err := c.ReadJSON(&payment); err != nil {
				c.WriteFullBodyString(400, "Invalid JSON payload")
				return
			}
			services.EnqueuePayment(&payment, queue)
			c.WriteFullBodyString(200, "OK")
			return

		case "/payments-summary":
			// if c.Method() != silverlining.MethodGET {
			// 	c.WriteFullBodyString(405, "Method not allowed")
			// 	return
			// }
			var request models.SummaryRequest
			if err := c.BindQuery(&request); err != nil {
				c.WriteFullBodyString(500, err.Error())
				return
			}
			response, err := services.GetSummary(&request)
			if err != nil {
				c.WriteFullBodyString(500, err.Error())
				return
			}
			c.WriteJSON(200, response)
			return

		case "/purge-payments":
			// if c.Method() != silverlining.MethodPOST {
			// 	c.WriteFullBodyString(405, "Method not allowed")
			// 	return
			// }
			if err := services.PurgePayments(); err != nil {
				c.WriteFullBodyString(500, err.Error())
				return
			}
			c.WriteFullBodyString(200, "OK")
			return

		default:
			c.WriteFullBodyString(404, "Not found")
		}
	}

	if cfg.ServerSocket != "" {
		log.Printf("Listening on %s", cfg.ServerSocket)
		listener := utils.NewListenUnix(cfg.ServerSocket)
		return ServeListener(listener, handler)
	}

	return silverlining.ListenAndServe(cfg.ServerURL, handler)
}

package handlers

import (
	"net/http"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/internal/services"

	"github.com/labstack/echo/v4"
)

func PostPayment(worker *services.PaymentWorker) func(c echo.Context) error {
	return func(c echo.Context) error {
		var payment models.Payment
		if err := c.Bind(&payment); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		go worker.EnqueuePayment(&payment)
		return c.JSON(http.StatusAccepted, nil)
	}
}

func GetSummary(worker *services.PaymentWorker) func(c echo.Context) error {
	return func(c echo.Context) error {
		from := c.QueryParam("from")
		to := c.QueryParam("to")
		summary, err := worker.GetSummary(from, to)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, summary)
	}
}

func PostPurgePayments(worker *services.PaymentWorker) func(c echo.Context) error {
	return func(c echo.Context) error {
		if err := worker.PurgePayments(); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, nil)
	}
}

func GetHealth(worker *services.PaymentWorker) func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, worker.GetHealth())
	}
}

// func ServeListener(ln net.Listener, handler silverlining.Handler) error {
// 	srv := &silverlining.Server{
// 		Listener: ln,
// 		Handler:  handler,
// 	}
// 	return srv.Serve(ln)
// }

// func RunSilverlining(cfg *config.Config, worker *services.PaymentWorker) error {
// 	log.Println("Starting Silverlining server")

// 	listener := utils.NewListenUnix(cfg.ServerSocket)
// 	handler := func(c *silverlining.Context) {
// 		switch string(c.Path()) {
// 		case "/payments":
// 			if c.Method() != silverlining.MethodPOST {
// 				c.WriteFullBodyString(405, "Method not allowed")
// 				return
// 			}
// 			var payment models.Payment
// 			if err := c.ReadJSON(&payment); err != nil {
// 				c.WriteFullBodyString(http.StatusBadRequest, "Invalid JSON payload")
// 				return
// 			}
// 			c.WriteFullBodyString(http.StatusAccepted, "Accepted")
// 			go worker.EnqueuePayment(&payment)
// 		case "/payments-summary":
// 			if c.Method() != silverlining.MethodGET {
// 				c.WriteFullBodyString(405, "Method not allowed")
// 				return
// 			}
// 			from, err := c.GetQueryParamString("from")
// 			if err != nil {
// 				c.WriteFullBodyString(http.StatusBadRequest, err.Error())
// 				return
// 			}
// 			to, err := c.GetQueryParamString("to")
// 			if err != nil {
// 				c.WriteFullBodyString(http.StatusBadRequest, err.Error())
// 				return
// 			}
// 			response, err := worker.GetSummary(from, to)
// 			if err != nil {
// 				c.WriteFullBodyString(http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			c.WriteJSON(200, response)
// 		case "/purge-payments":
// 			if c.Method() != silverlining.MethodPOST {
// 				c.WriteFullBodyString(405, "Method not allowed")
// 				return
// 			}
// 			if err := worker.PurgePayments(); err != nil {
// 				c.WriteFullBodyString(http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			c.WriteFullBodyString(http.StatusOK, "OK")
// 		case "/health":
// 			c.WriteFullBodyString(http.StatusOK, "GOOD")
// 		default:
// 			c.WriteFullBodyString(http.StatusNotFound, "Not found")
// 		}
// 	}

// 	log.Printf("Listening on %s", cfg.ServerSocket)
// 	return ServeListener(listener, handler)
// }

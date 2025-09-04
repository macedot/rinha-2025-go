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

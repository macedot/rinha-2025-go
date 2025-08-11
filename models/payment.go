package models

import (
	"time"
)

type Payment struct {
	PaymentID string    `json:"correlationId" binding:"required"`
	Amount    float64   `json:"amount" binding:"required,ge=0"` // Amount in dollars (e.g., 99.99)
	Timestamp time.Time `json:"requestedAt"`
}

// func (p *Payment) Validate() error {
// 	// if p.PaymentID == "" {
// 	// 	return errors.New("paymentId is required")
// 	// }
// 	// if p.Amount < 0 {
// 	// 	return errors.New("amount must be greater than zero")
// 	// }
// 	return nil
// }

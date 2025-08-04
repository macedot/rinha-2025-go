package models

import "time"

type Payment struct {
	PaymentID string    `json:"correlationId" binding:"required"`
	Amount    float64   `json:"amount" binding:"required,gt=0"` // Amount in dollars (e.g., 99.99)
	Timestamp time.Time `json:"requestedAt"`
}

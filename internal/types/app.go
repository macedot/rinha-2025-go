package types

import (
	"time"
)

type PaymentRequest struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}
type PaymentRecord struct {
	Amount      int64     `json:"amount"`
	RequestedAt time.Time `json:"requestedAt"`
}

type SummaryServer struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type SummaryResponse struct {
	Default  SummaryServer `json:"default"`
	Fallback SummaryServer `json:"fallback"`
}

type HealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type ProcessorHealth struct {
	Failing         bool      `json:"failing"`
	MinResponseTime int       `json:"minResponseTime"`
	LastChecked     time.Time `json:"lastChecked,omitempty"`
}

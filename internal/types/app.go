package types

import (
	"sync"
	"time"
)

type PaymentRequest struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

type PaymentRecord struct {
	Amount      int64
	RequestedAt time.Time
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

type HealthSummary struct {
	Default  *ProcessorHealth `json:"default"`
	Fallback *ProcessorHealth `json:"fallback"`
}

type HealthDB struct {
	MuDefault  sync.Mutex
	MuFallback sync.Mutex
	Default    ProcessorHealth
	Fallback   ProcessorHealth
}

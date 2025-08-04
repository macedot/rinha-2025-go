package models

type SummaryRequest struct {
	StartTime string `form:"from" binding:"datetime=2006-01-02T15:04:05Z"`
	EndTime   string `form:"to" binding:"datetime=2006-01-02T15:04:05Z"`
}

type SummaryParam struct {
	StartTime string
	EndTime   string
}

type ProcessorSummary struct {
	RequestCount int     `json:"totalRequests"`
	TotalAmount  float64 `json:"totalAmount"`
}

type SummaryResponse struct {
	Default  ProcessorSummary `json:"default"`
	Fallback ProcessorSummary `json:"fallback"`
}

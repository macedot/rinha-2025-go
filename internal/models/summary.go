package models

type SummaryRequest struct {
	StartTime string `query:"from"`
	EndTime   string `query:"to"`
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
	Default  *ProcessorSummary `json:"default"`
	Fallback *ProcessorSummary `json:"fallback"`
}

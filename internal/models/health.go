package models

type HealthResponse struct {
	Failing         bool   `json:"failing"`
	MinResponseTime uint32 `json:"minResponseTime"`
}

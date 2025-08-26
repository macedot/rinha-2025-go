package health

import (
	"sync"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

type HealthManager struct {
	mu        sync.Mutex
	Client    *client.HttpClient
	Processor string
	Endpoint  string
	Health    *types.ProcessorHealth
}

func NewHealthManager(client *client.HttpClient, processor, endpoint string) *HealthManager {
	return &HealthManager{
		Client:    client,
		Processor: processor,
		Endpoint:  endpoint,
		Health:    &types.ProcessorHealth{},
	}
}

func (h *HealthManager) CheckAndUpdateHealth() {
	ticker := time.NewTicker(5010 * time.Millisecond)
	for range ticker.C {
		h.updateHealth()
	}
}

func (h *HealthManager) updateHealth() {
	if !h.Expired(10 * time.Second) {
		return
	}
	status, body, err := h.Client.Get(h.Endpoint)
	if err != nil || status != fasthttp.StatusOK {
		h.SaveHealth(true, 9999)
		return
	}
	var health types.HealthResponse
	if err := health.UnmarshalJSON(body); err != nil {
		h.SaveHealth(true, 9999)
		return
	}
	h.SaveHealth(health.Failing, health.MinResponseTime)
}

func (h *HealthManager) SaveHealth(failing bool, minResp int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Health.Failing = failing
	h.Health.MinResponseTime = minResp
	h.Health.LastChecked = time.Now().UTC()
}

func (h *HealthManager) GetHealth() *types.ProcessorHealth {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.Health
}

func (h *HealthManager) Expired(duration time.Duration) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return time.Since(h.Health.LastChecked) > time.Duration(duration)
}

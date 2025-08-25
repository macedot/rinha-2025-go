package health

import (
	"context"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type HealthManager struct {
	Redis     *redis.Client
	Client    *client.HttpClient
	Processor string
	Endpoint  string
}

func (h *HealthManager) CheckAndUpdateHealth() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		h.updateHealth()
	}
}

func (h *HealthManager) updateHealth() {
	ok, err := h.Redis.SetNX(context.Background(), "lock:health:"+h.Processor, "1", 2*time.Second).Result()
	if err != nil || !ok {
		return
	}

	status, body, err := h.Client.Get(h.Endpoint)
	if err != nil || status != fasthttp.StatusOK {
		h.SaveHealthToRedis(true, 9999)
		return
	}

	var health types.HealthResponse
	if err := health.UnmarshalJSON(body); err != nil {
		h.SaveHealthToRedis(true, 9999)
		return
	}

	h.SaveHealthToRedis(health.Failing, health.MinResponseTime)
}

func (h *HealthManager) SaveHealthToRedis(failing bool, minResp int) {
	health := types.ProcessorHealth{
		Failing:         failing,
		MinResponseTime: minResp,
		LastChecked:     time.Now().UTC(),
	}
	data, _ := health.MarshalJSON()
	h.Redis.Set(context.Background(), "health:"+h.Processor, data, 10*time.Second)
}

func (h *HealthManager) GetHealth() (*types.ProcessorHealth, error) {
	data, err := h.Redis.Get(context.Background(), "health:"+h.Processor).Result()
	if err != nil {
		return nil, err
	}
	var hp types.ProcessorHealth
	err = hp.UnmarshalJSON([]byte(data))
	return &hp, err
}

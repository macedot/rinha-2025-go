package health

import (
	"log"
	"math/rand"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

type HealthManager struct {
	Client       *client.HttpClient
	Storage      *client.SocketClient
	HealthStore  string
	HealthClient string
	HealthCache  *types.ProcessorHealth
	Processor    string
}

func (h *HealthManager) CheckAndUpdateHealth() {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	for {
		h.updateHealth()
		time.Sleep(time.Duration(5000+rand.Intn(1000)) * time.Millisecond)
	}
}

func (h *HealthManager) updateHealth() {
	health := h.GetHealthStorage()
	if health != nil && time.Since(health.LastChecked) < 5*time.Second {
		return
	}
	health = h.GetHealthClient()
	if health == nil {
		return
	}
	health.LastChecked = time.Now()
	//log.Println("updateHealth:", h.Processor, health)
	// ok, err := h.Redis.SetNX(h.Ctx, "lock:health:"+h.Processor, "1", 2*time.Second).Result()
	h.SaveHealth(health.Failing, health.MinResponseTime)
}

func (h *HealthManager) GetHealthClient() *types.ProcessorHealth {
	client := client.NewHttpClient().Init()
	status, body, err := client.GetTimeout(h.HealthClient, 6*time.Second)
	if err != nil {
		log.Println("GetHealthClient: error getting health:", err)
		return nil
	}
	if status == fasthttp.StatusTooManyRequests {
		return nil
	}
	if status != fasthttp.StatusOK {
		log.Println("GetHealthClient: unexpected status:", status)
		return nil
	}
	var health types.ProcessorHealth
	if err := health.UnmarshalJSON(body); err != nil {
		log.Println("GetHealthClient: error unmarshalling health:", err)
		return nil
	}
	h.HealthCache = &health
	return &health
}

func (h *HealthManager) SaveHealth(failing bool, minResponseTime int) {
	// h.Redis.Set(h.Ctx, "health:"+h.Processor, data, 10*time.Second)
	health := types.ProcessorHealth{
		Failing:         failing,
		MinResponseTime: minResponseTime,
		LastChecked:     time.Now().Truncate(time.Second),
	}
	if h.HealthCache == nil || h.HealthCache.LastChecked.Before(health.LastChecked) {
		h.HealthCache = &health
	}
	data, _ := health.MarshalJSON()
	_, err := h.Storage.Post(h.HealthStore, data)
	if err != nil {
		log.Println("SaveHealth: error saving health:", err)
	} else {
		log.Println("SaveHealth: saved health:", health)
	}
}

func (h *HealthManager) GetHealthStorage() *types.ProcessorHealth {
	if h.HealthCache != nil && time.Since(h.HealthCache.LastChecked) < 5*time.Second {
		return h.HealthCache
	}
	// data, err := h.Redis.Get(h.Ctx, "health:"+h.Processor).Result()
	status, body, err := h.Storage.Get(h.HealthStore)
	if err != nil {
		log.Println("GetHealthStorage: error getting health:", err)
		return h.HealthCache
	}
	if status != fasthttp.StatusOK {
		log.Println("GetHealthStorage: unexpected status:", status)
		return h.HealthCache
	}
	var health types.ProcessorHealth
	if err := health.UnmarshalJSON(body); err != nil {
		log.Println("GetHealthStorage: error unmarshalling health:", err)
		return h.HealthCache
	}
	if h.HealthCache != nil && h.HealthCache.LastChecked.After(health.LastChecked) {
		return h.HealthCache
	}
	h.HealthCache = &health
	return &health
}

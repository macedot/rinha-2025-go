package services

import (
	"log"
	"math/rand"
	"reflect"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/models"
	"sync"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

const (
	HEALTH_REDIS_KEY       = "health"
	HEALTH_REDIS_TIMEOUT   = "timeout"
	HEALTH_REDIS_INSTANCES = "instances"
)

type Health struct {
	cfg    *config.Config
	redis  *database.Redis
	client *HttpClient
}

func NewHealth(
	config *config.Config,
	redis *database.Redis,
	client *HttpClient,
) *Health {
	return &Health{
		cfg:    config,
		redis:  redis,
		client: client,
	}
}

func (h *Health) Close() {
	h.redis.Close()
}

func (h *Health) ResetHealthTimeout() {
	h.redis.SetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT, 0)
}

func (h *Health) SetHealthTimeout(duration time.Duration) error {
	value := time.Now().UTC().Add(duration).UnixMilli()
	return h.redis.SetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT, value)
}

func (h *Health) GetHealthTimeout() int64 {
	return h.redis.GetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT)
}

func (h *Health) GetActiveInstance() *config.ServiceStatus {
	var activeService config.ServiceStatus
	jsonData := h.redis.GetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES)
	if jsonData != "" {
		err := oj.Unmarshal([]byte(jsonData), &activeService)
		if err != nil {
			log.Print("GetActiveInstance:", err, jsonData)
		}
	}
	return &activeService
}

func (h *Health) SetInstancesCache(activeInstance *config.ServiceStatus) error {
	activeInstance.LastUpdate = time.Now()
	bytes, err := oj.Marshal(activeInstance)
	if err == nil {
		err = h.redis.SetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES, string(bytes))
	}
	if err != nil {
		log.Print("SetInstancesCache:", err, activeInstance)
	}
	return err
}

func (h *Health) ResetInstancesCache(activeInstance *config.ServiceStatus) error {
	currentInstance := h.GetActiveInstance()
	if currentInstance == nil {
		return h.SetInstancesCache(activeInstance)
	}
	return nil
}

func (h *Health) RefreshServiceStatus() {
	currStatus := h.cfg.GetActiveService()
	currServices := h.cfg.GetServices()
	h.updateServicesHealth(currServices)
	h.cfg.UpdateServices(currServices).UpdateActiveInstance()
	activeStatus := h.cfg.GetActiveService()
	h.SetInstancesCache(activeStatus)
	if !reflect.DeepEqual(currStatus, activeStatus) {
		log.Print("RefreshServiceStatus:", activeStatus)
	}
}

func (h *Health) UpdateServicesHealth() {
	now := time.Now().UTC().UnixMilli()
	expiration := h.GetHealthTimeout()
	if expiration < now {
		h.SetHealthTimeout(time.Hour)
		h.RefreshServiceStatus()
		h.SetHealthTimeout(h.cfg.ServiceRefreshInterval)
	}
}

func (h *Health) updateServicesHealth(services []config.Service) {
	var wg sync.WaitGroup
	for i, service := range services {
		wg.Add(1)
		go func() {
			defer wg.Done()
			health := h.getServiceHealth(&service)
			services[i].Failing = health.Failing
			services[i].MinResponseTime = health.MinResponseTime
		}()
	}
	wg.Wait()
}

func (h *Health) getServiceHealth(service *config.Service) *models.HealthResponse {
	health := models.HealthResponse{Failing: true}
	statusCode, body := h.client.Get(service.URL + "/payments/service-health")
	if statusCode == fasthttp.StatusOK {
		if err := oj.Unmarshal(body, &health); err != nil {
			health.Failing = true
			log.Print("getServiceHealth:", service.URL, err)
		}
	}
	return &health
}

func (h *Health) ProcessServicesHealth() {
	h.ResetHealthTimeout()
	sleep := time.Duration(rand.Intn(3))
	log.Printf("Sleep for %d seconds...", sleep)
	time.Sleep(sleep * time.Second)
	for {
		h.UpdateServicesHealth()
		time.Sleep(h.cfg.ServiceRefreshInterval)
	}
}

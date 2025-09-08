package services

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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
	HEALTH_REDIS_LOCK      = "health_lock"
	HEALTH_REDIS_LOCK_TIME = "health_lock_time"
	HEALTH_REDIS_INSTANCES = "instances"
)

type Health struct {
	cfg      *config.Config
	redis    *database.Redis
	client   *HttpClient
	services *config.Services
}

func NewHealth(
	config *config.Config,
	redis *database.Redis,
	client *HttpClient,
) *Health {
	services := config.GetServices()
	return &Health{
		cfg:      config,
		redis:    redis,
		client:   client,
		services: services,
	}
}

func (h *Health) Close() {
	h.redis.Close()
}

func (h *Health) GetActiveInstance() *config.Service {
	jsonData := h.redis.GetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES)
	if jsonData == "" {
		return nil
	}
	var activeService config.Service
	if err := oj.Unmarshal([]byte(jsonData), &activeService); err != nil {
		log.Print("GetActiveInstance:", err, jsonData)
		return nil
	}
	return &activeService
}

func (h *Health) setActiveInstance(activeService *config.Service) error {
	bytes, err := oj.Marshal(activeService)
	if err != nil {
		log.Print("SetActiveInstance:", err, activeService)
		return err
	}
	return h.redis.SetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES, string(bytes))
}

func (h *Health) selectActiveInstance() *config.Service {
	d := &h.services.Default
	f := &h.services.Fallback
	if d.Failing {
		if f.Failing {
			return nil
		}
		return f
	}
	if d.MinResponseTime <= 150 || f.Failing || d.MinResponseTime == f.MinResponseTime {
		return d
	}
	if f.MinResponseTime <= 150 {
		return f
	}
	return d
}

func (h *Health) refreshServiceStatus() {
	start := time.Now()
	currentActive := h.GetActiveInstance()
	h.updateServicesHealth(h.services)
	activeStatus := h.selectActiveInstance()
	h.setActiveInstance(activeStatus)
	from, to := "nil", "nil"
	if currentActive != nil {
		from = fmt.Sprintf("[%s %d]", currentActive.Table, currentActive.MinResponseTime)
	}
	if activeStatus != nil {
		to = fmt.Sprintf("[%s %d]", activeStatus.Table, activeStatus.MinResponseTime)
	}
	if from == to {
		return
	}
	log.Println(from, "->", to)
	log.Println("refreshServiceStatus:", time.Since(start))
}

func (h *Health) updateServicesHealth(services *config.Services) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.getServiceHealth(&services.Default)
		services.Default.Failing = health.Failing
		services.Default.MinResponseTime = health.MinResponseTime
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		health := h.getServiceHealth(&services.Fallback)
		services.Fallback.Failing = health.Failing
		services.Fallback.MinResponseTime = health.MinResponseTime
	}()
	wg.Wait()
}

func (h *Health) getServiceHealth(service *config.Service) *models.HealthResponse {
	health := models.HealthResponse{Failing: true}
	statusCode, body, err := h.client.Get(service.URL+"/payments/service-health", service)
	if err != nil {
		log.Print("getServiceHealth:Get:", service.URL, err)
		return &health
	}
	if statusCode == fasthttp.StatusOK {
		if err := oj.Unmarshal(body, &health); err != nil {
			health.Failing = true
			log.Print("getServiceHealth:Unmarshal:", service.URL, err)
		}
	}
	return &health
}

func (h *Health) ProcessServicesHealth() {
	lockValue, _ := os.Hostname()
	lockTTL := time.Second + h.cfg.ServiceRefreshInterval
	sleep := time.Duration(rand.Intn(3000))
	log.Printf("Sleep for %d ms...", sleep)
	time.Sleep(sleep * time.Millisecond)
	for {
		waitTime := h.cfg.ServiceRefreshInterval
		if !h.redis.TryLock(HEALTH_REDIS_LOCK, lockValue, lockTTL) {
			time.Sleep(time.Second)
			continue
		}
		lastRun, err := h.redis.GetLastRunTime(HEALTH_REDIS_LOCK_TIME)
		if err == nil {
			waitTime = h.cfg.ServiceRefreshInterval - time.Since(lastRun)
			if waitTime < 0 {
				h.refreshServiceStatus()
				h.redis.SetLastRunTime(HEALTH_REDIS_LOCK_TIME, time.Now())
				waitTime = h.cfg.ServiceRefreshInterval
			}
		} else {
			log.Println("ProcessServicesHealth:GetLastRunTime:", err)
		}
		h.redis.Unlock(HEALTH_REDIS_LOCK)
		time.Sleep(waitTime)
	}
}

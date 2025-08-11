package services

import (
	"log"
	"net/http"
	"reflect"
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/models"
	"sync"
	"time"

	"github.com/ohler55/ojg/oj"
)

const (
	HEALTH_REDIS_KEY       = "health"
	HEALTH_REDIS_TIMEOUT   = "timeout"
	HEALTH_REDIS_INSTANCES = "instances"
)

func ResetHealthTimeout() {
	redis := database.RedisInstance()
	redis.SetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT, 0)
}

func SetHealthTimeout(duration time.Duration) error {
	redis := database.RedisInstance()
	value := time.Now().UTC().Add(duration).UnixNano()
	return redis.SetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT, value)
}

func GetHealthTimeout() int64 {
	redis := database.RedisInstance()
	return redis.GetInt(HEALTH_REDIS_KEY, HEALTH_REDIS_TIMEOUT)
}

func GetInstancesCache() []config.Service {
	var instances []config.Service
	redis := database.RedisInstance()
	jsonData := redis.GetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES)
	if jsonData != "" {
		err := oj.Unmarshal([]byte(jsonData), &instances)
		if err != nil {
			log.Print("GetInstancesCache:", err, jsonData)
		}
	}
	return instances
}

func SetInstancesCache(instances []config.Service) error {
	bytes, err := oj.Marshal(instances)
	if err == nil {
		redis := database.RedisInstance()
		err = redis.SetString(HEALTH_REDIS_KEY, HEALTH_REDIS_INSTANCES, string(bytes))
	}
	if err != nil {
		log.Print("SetInstancesCache:", err, instances)
	}
	return err
}

func ResetInstancesCache(instances []config.Service) error {
	cache := GetInstancesCache()
	if len(cache) == 0 {
		return SetInstancesCache(instances)
	}
	return nil
}

func RefreshServiceStatus(cfg *config.Config) {
	currInstances := cfg.GetInstances() //GetInstancesCache()
	currServices := cfg.GetServices()
	updateServicesHealth(currServices)
	cfg.UpdateServices(currServices).UpdateInstances()
	instances := cfg.GetInstances()
	SetInstancesCache(instances)
	if !reflect.DeepEqual(currInstances, instances) {
		type Result struct {
			Fee   float64
			Delay uint32
		}
		var arr []Result
		for _, instance := range instances {
			arr = append(arr, Result{
				Fee:   instance.Fee,
				Delay: instance.MinResponseTime,
			})
		}
		log.Print("RefreshServiceStatus:", arr)
	}
}

func updateServicesHealth(services []config.Service) {
	var wg sync.WaitGroup
	for i, service := range services {
		wg.Add(1)
		go func() {
			defer wg.Done()
			health := getServiceHealth(service)
			services[i].Failing = health.Failing
			services[i].MinResponseTime = health.MinResponseTime
		}()
	}
	wg.Wait()
}

func getServiceHealth(service config.Service) models.HealthResponse {
	health := models.HealthResponse{Failing: true}
	client := HttpClientInstance()
	statusCode, body := client.Get(service.URL + "/payments/service-health")
	if statusCode == http.StatusOK {
		if err := oj.Unmarshal(body, &health); err != nil {
			health.Failing = true
			log.Print("getServiceHealth:", service.URL, err)
		}
	}
	return health
}

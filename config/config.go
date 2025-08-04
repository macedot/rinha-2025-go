package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"rinha-2025/utils"
	"time"
)

type Service struct {
	URL   string  `json:"url"`
	Table string  `json:"table"`
	Token string  `json:"token"`
	Fee   float64 `json:"fee"`

	Failing         bool
	MinResponseTime uint32
	Timeout         time.Duration
	KeyAmount       string
	KeyTime         string
}

type Config struct {
	DebugMode              bool
	Services               []Service
	ServiceRefreshInterval time.Duration
	Instances              []Service
	ServerURL              string
	RedisURL               string
}

type ConfigCache struct {
	Services []Service
}

var config Config

func ConfigInstance() *Config {
	return &config
}

func (c *Config) GetServices() []Service {
	return c.Services
}

func (c *Config) UpdateServices(services []Service) {
	c.Services = services
}

func (c *Config) GetInstances() []Service {
	return c.Instances
}

// https://github.com/JosineyJr/rdb25_02/blob/ae8517f398e4261890bbfe0bd57ca986642a34e5/internal/routing/router.go#L121
// https://github.com/zanfranceschi/rinha-de-backend-2025/blob/main/participantes/andersongomes001/partial-results.json
func (c *Config) UpdateInstances() {
	c.Instances = nil
	if len(c.Services) == 1 {
		if c.Services[0].Failing {
			return
		}
		c.Instances = append(c.Instances, c.Services[0])
		return
	}
	if c.Services[0].Failing && c.Services[1].Failing {
		return
	}
	if c.Services[0].Failing {
		c.Instances = append(c.Instances, c.Services[1])
		return
	}
	if c.Services[1].Failing {
		c.Instances = append(c.Instances, c.Services[0])
		return
	}
	dl := c.Services[0].MinResponseTime
	fl := c.Services[1].MinResponseTime
	if 0 < fl && fl <= 20 && (3*fl) < dl {
		c.Instances = append(c.Instances, c.Services[1])
		return
	}
	c.Instances = append(c.Instances, c.Services[0])
}

func (c *Config) Init() {
	c.DebugMode = utils.GetEnvBool("API_DEBUG_MODE", false)
	var services []Service
	envServices := os.Getenv("SERVICES")
	if envServices != "" {
		err := json.Unmarshal([]byte(envServices), &services)
		if err != nil {
			log.Fatal("Error unmarshaling JSON:", err)
		}
	} else {
		services = []Service{
			{
				URL:   "http://payment-processor-default:8080",
				Table: "d",
				Fee:   0.05,
			},
		}
	}
	for _, service := range services {
		if service.URL == "" {
			log.Fatal("Missing URL from Service")
		}
		if service.Table == "" {
			log.Fatal("Missing Table from Service")
		}
		if service.Token == "" {
			service.Token = "123"
		}
		//service.Fee = 0.0
		service.Failing = false
		service.MinResponseTime = 0
		service.Timeout = 10 * time.Second
		service.KeyAmount = fmt.Sprintf("summary:%s:data", service.Table)
		service.KeyTime = fmt.Sprintf("summary:%s:history", service.Table)
		c.Services = append(c.Services, service)
		log.Print("Service:", service)
	}
	if len(c.Services) < 1 {
		log.Fatal("at least one services is required")
	}
	c.ServiceRefreshInterval = utils.GetEnvDuration("SERVICE_REFRESH_INTERVAL", "5s")
	c.ServerURL = utils.GetEnv("SERVER_URL", ":5000")
	c.RedisURL = utils.GetEnv("REDIS_URL", "redis:6379")
}

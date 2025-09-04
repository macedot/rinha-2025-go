package config

import (
	"fmt"
	"log"
	"os"
	"rinha-2025-go/pkg/utils"
	"time"

	"github.com/ohler55/ojg/oj"
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
	LastUpdate      time.Time
}

type ServiceMode int

const (
	None ServiceMode = iota - 1
	Default
	Fallback
)

type ServiceStatus struct {
	Mode       ServiceMode
	LastUpdate time.Time
}

type Config struct {
	ServerSocket           string
	RedisSocket            string
	ActiveService          ServiceStatus
	Services               []Service
	ServiceRefreshInterval time.Duration
	NumWorkers             int
}

var appConfig Config

func ConfigInstance() *Config {
	return &appConfig
}

func (c *Config) GetServices() []Service {
	return c.Services
}

func (c *Config) GetActiveService() *ServiceStatus {
	return &c.ActiveService
}

func (c *Config) UpdateServices(services []Service) *Config {
	c.Services = services
	return c
}

// https://github.com/JosineyJr/rdb25_02/blob/ae8517f398e4261890bbfe0bd57ca986642a34e5/internal/routing/router.go#L121
// https://github.com/zanfranceschi/rinha-de-backend-2025/blob/main/participantes/andersongomes001/partial-results.json
func (c *Config) UpdateActiveInstance() *Config {
	c.ActiveService = ServiceStatus{Mode: None}
	if len(c.Services) == 1 {
		if c.Services[0].Failing {
			return c
		}
		c.ActiveService.Mode = Default
		return c
	}
	if c.Services[0].Failing {
		if c.Services[1].Failing {
			return c
		}
		c.ActiveService.Mode = Fallback
		return c
	}
	dl := float32(c.Services[0].MinResponseTime)
	if dl <= 100 || c.Services[1].Failing {
		c.ActiveService.Mode = Default
		return c
	}
	fl := float32(c.Services[1].MinResponseTime)
	if dl-fl < 1000 {
		c.ActiveService.Mode = Default
		return c
	}
	c.ActiveService.Mode = Fallback
	return c
}

func (c *Config) Init() *Config {
	var services []Service
	if envServices := os.Getenv("SERVICES"); envServices != "" {
		if err := oj.Unmarshal([]byte(envServices), &services); err != nil {
			log.Fatal("error unmarshaling JSON:", err)
		}
	}
	if len(services) < 1 {
		log.Fatal("Config: at least one services is required")
	}
	for _, service := range services {
		if service.URL == "" {
			log.Fatal("Config: missing URL from Service")
		}
		if service.Table == "" {
			log.Fatal("Config: missing Table from Service")
		}
		if service.Token == "" {
			service.Token = "123"
		}
		//service.Fee = 0.0
		service.Failing = false
		service.MinResponseTime = 0
		service.Timeout = 1 * time.Second
		service.KeyAmount = fmt.Sprintf("summary:%s:data", service.Table)
		service.KeyTime = fmt.Sprintf("summary:%s:history", service.Table)
		c.Services = append(c.Services, service)
		log.Print("Service:", service)
	}
	c.ServerSocket = utils.GetEnv("SERVER_SOCKET", "")
	c.RedisSocket = utils.GetEnv("REDIS_SOCKET")
	c.ServiceRefreshInterval = utils.GetEnvDuration("SERVICE_REFRESH_INTERVAL", "5001ms")
	c.NumWorkers = utils.GetEnvInt("NUM_WORKERS", "5")
	return c
}

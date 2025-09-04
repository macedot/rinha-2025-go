package config

import (
	"fmt"
	"log"
	"rinha-2025-go/pkg/utils"
	"strconv"
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
}

type Services struct {
	Default  Service
	Fallback Service
}

type Config struct {
	ServerSocket           string
	RedisSocket            string
	ActiveInstance         *Service
	Services               Services
	ServiceRefreshInterval time.Duration
	NumWorkers             int
}

var appConfig Config

func ConfigInstance() *Config {
	return &appConfig
}

func (c *Config) GetServices() *Services {
	return &c.Services
}

func (c *Config) SetServices(services *Services) {
	c.Services = *services
}

func (c *Config) GetActiveInstance() *Service {
	return c.ActiveInstance
}
func (c *Config) SetActiveInstance(activeService *Service) {
	c.ActiveInstance = activeService
}

func (c *Config) Init() *Config {
	var services []Service
	var err error
	envServices := utils.GetEnv("SERVICES")
	if err := oj.Unmarshal([]byte(envServices), &services); err != nil {
		log.Fatal("error unmarshaling JSON:", err)
	}
	if len(services) != 2 {
		log.Fatal("Config: invalid Services configuration")
	}
	for idx, service := range services {
		if service.URL == "" {
			log.Fatal("Config: missing URL from Service")
		}
		if service.Table == "" {
			log.Fatal("Config: missing Table from Service")
		}
		if service.Token == "" {
			service.Token = "123"
		}
		service.Timeout = 3 * time.Second
		service.KeyAmount = fmt.Sprintf("summary:%s:data", service.Table)
		service.KeyTime = fmt.Sprintf("summary:%s:history", service.Table)
		log.Print("Service:", service)
		if idx == 0 {
			c.Services.Default = service
		} else {
			c.Services.Fallback = service
		}
	}
	c.ActiveInstance = &c.Services.Default
	c.ServerSocket = utils.GetEnvOr("SERVER_SOCKET", "")
	c.RedisSocket = utils.GetEnv("REDIS_SOCKET")
	c.ServiceRefreshInterval, err = time.ParseDuration(utils.GetEnvOr("SERVICE_REFRESH_INTERVAL", "5s"))
	if err != nil {
		log.Fatal("error parsing SERVICE_REFRESH_INTERVAL:", err)
	}
	c.NumWorkers, err = strconv.Atoi(utils.GetEnvOr("NUM_WORKERS", "1"))
	if err != nil {
		log.Fatal("error parsing NUM_WORKERS:", err)
	}
	return c
}

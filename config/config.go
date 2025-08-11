package config

import (
	"fmt"
	"log"
	"os"
	"rinha-2025/utils"
	"time"

	"github.com/joho/godotenv"
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

type Config struct {
	DebugMode              bool
	ServerType             string
	ServerURL              string
	ServerSocket           string
	RedisURL               string
	Instances              []Service
	Services               []Service
	ServiceRefreshInterval time.Duration
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

func (c *Config) GetInstances() []Service {
	return c.Instances
}

func (c *Config) UpdateServices(services []Service) *Config {
	c.Services = services
	return c
}

// https://github.com/JosineyJr/rdb25_02/blob/ae8517f398e4261890bbfe0bd57ca986642a34e5/internal/routing/router.go#L121
// https://github.com/zanfranceschi/rinha-de-backend-2025/blob/main/participantes/andersongomes001/partial-results.json
func (c *Config) UpdateInstances() *Config {
	c.Instances = nil
	if len(c.Services) == 1 {
		if c.Services[0].Failing {
			return c
		}
		c.Instances = append(c.Instances, c.Services[0])
		return c
	}
	if c.Services[0].Failing && c.Services[1].Failing {
		return c
	}
	if c.Services[0].Failing {
		c.Instances = append(c.Instances, c.Services[1])
		return c
	}
	if c.Services[1].Failing {
		c.Instances = append(c.Instances, c.Services[0])
		return c
	}
	dl := c.Services[0].MinResponseTime
	fl := c.Services[1].MinResponseTime
	if 0 < fl && fl <= 20 && (3*fl) < dl {
		c.Instances = append(c.Instances, c.Services[1])
		return c
	}
	c.Instances = append(c.Instances, c.Services[0])
	return c
}

func (c *Config) Init() *Config {
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded ENV from .env file")
	}
	c.DebugMode = utils.GetEnvBool("API_DEBUG_MODE", "false")
	var services []Service
	if envServices := os.Getenv("SERVICES"); envServices != "" {
		if err := oj.Unmarshal([]byte(envServices), &services); err != nil {
			log.Fatal("error unmarshaling JSON:", err)
		}
	}
	if len(services) < 1 {
		log.Fatal("at least one services is required")
	}
	for _, service := range services {
		if service.URL == "" {
			log.Fatal("missing URL from Service")
		}
		if service.Table == "" {
			log.Fatal("missing Table from Service")
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
	c.ServerType = utils.GetEnv("SERVER_TYPE", "")
	c.ServerSocket = utils.GetEnv("SERVER_SOCKET", "")
	c.ServerURL = utils.GetEnv("SERVER_URL")
	c.RedisURL = utils.GetEnv("REDIS_URL")
	c.ServiceRefreshInterval = utils.GetEnvDuration("SERVICE_REFRESH_INTERVAL", "5s")
	if c.DebugMode {
		fmt.Printf("\n%+v\n\n", c)
	}
	return c
}

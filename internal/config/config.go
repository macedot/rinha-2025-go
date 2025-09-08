package config

import (
	"fmt"
	"log"
	"rinha-2025-go/pkg/utils"
	"runtime"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
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
	c.Services.Default.URL = utils.GetEnvOr("DEFAULT_URL", "http://payment-processor-default:8080")
	c.Services.Default.Table = "d"
	c.Services.Default.Token = "123"
	c.Services.Default.KeyAmount = fmt.Sprintf("summary:%s:data", c.Services.Default.Table)
	c.Services.Default.KeyTime = fmt.Sprintf("summary:%s:history", c.Services.Default.Table)
	c.Services.Default.Timeout = 10 * time.Second

	c.Services.Fallback.URL = utils.GetEnvOr("FALLBACK_URL", "http://payment-processor-fallback:8080")
	c.Services.Fallback.Table = "f"
	c.Services.Fallback.Token = "123"
	c.Services.Fallback.KeyAmount = fmt.Sprintf("summary:%s:data", c.Services.Fallback.Table)
	c.Services.Fallback.KeyTime = fmt.Sprintf("summary:%s:history", c.Services.Fallback.Table)
	c.Services.Fallback.Timeout = 10 * time.Second

	c.ServiceRefreshInterval = 5 * time.Second
	c.ActiveInstance = &c.Services.Default
	c.RedisSocket = utils.GetEnvOr("REDIS_SOCKET", "/sockets/redis.sock")
	c.ServerSocket = utils.GetEnvOr("SERVER_SOCKET", "")

	workers, err := strconv.Atoi(utils.GetEnvOr("NUM_WORKERS", "50"))
	if err != nil {
		log.Fatal("error parsing NUM_WORKERS:", err)
	}
	c.NumWorkers = workers

	GOMAXPROCS, err := strconv.Atoi(utils.GetEnvOr("GOMAXPROCS", "3"))
	if err != nil {
		log.Fatal("error parsing GOMAXPROCS:", err)
	}
	runtime.GOMAXPROCS(GOMAXPROCS)

	return c
}

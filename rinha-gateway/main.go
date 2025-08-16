package main

import (
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

type LoadBalancer struct {
	requestCount uint64
	clients      []*fasthttp.Client
}

func (lb *LoadBalancer) nextIndex() int {
	return int(atomic.AddUint64(&lb.requestCount, 1) % uint64(len(lb.clients)))
}

func (lb *LoadBalancer) Handler(ctx *fasthttp.RequestCtx) {
	next := lb.nextIndex()
	req := &ctx.Request
	resp := &ctx.Response
	if err := lb.clients[next].Do(req, resp); err != nil {
		log.Println(resp.StatusCode(), string(resp.Body()))
		ctx.Error(err.Error(), fasthttp.StatusBadGateway)
	}
}

func NewUnixClient(unixSocketPath string) *fasthttp.Client {
	if _, err := os.Stat(unixSocketPath); os.IsNotExist(err) {
		log.Fatalf("UNIX socket %s does not exist", unixSocketPath)
	}
	return &fasthttp.Client{
		MaxConnsPerHost:               4096,
		ReadTimeout:                   700 * time.Millisecond,
		WriteTimeout:                  700 * time.Millisecond,
		ReadBufferSize:                1024,
		WriteBufferSize:               1024,
		MaxIdleConnDuration:           10 * time.Second,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: func(_ string) (net.Conn, error) {
			return net.Dial("unix", unixSocketPath)
		},
	}
}

func NewUnixClients(param []string) []*fasthttp.Client {
	var clients []*fasthttp.Client
	for _, unixSocketPath := range param {
		clients = append(clients, NewUnixClient(unixSocketPath))
	}
	return clients
}

func NewLoadBalancer(clients []string) *LoadBalancer {
	return &LoadBalancer{
		requestCount: 0,
		clients:      NewUnixClients(clients),
	}
}

func main() {
	apiEnv := os.Getenv("API_SOCKETS")
	if apiEnv == "" {
		log.Fatalln("API_SOCKETS environment variable not set")
	}
	apis := strings.Split(apiEnv, ",")
	if len(apis) == 0 {
		log.Fatalln("invalid API_SOCKETS environment variable")
	}
	log.Println("Gateway clients:", apis)
	lb := NewLoadBalancer(apis)
	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = ":9999"
	}
	log.Println("Starting Gateway server: " + addr)
	log.Fatalln(fasthttp.ListenAndServe(addr, lb.Handler))
}

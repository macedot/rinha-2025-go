package gateway

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"

	"github.com/valyala/fasthttp"
)

type LoadBalancer struct {
	clients  []*fasthttp.Client
	requests atomic.Int32
}

func NewLoadBalancer(clients []string) *LoadBalancer {
	return &LoadBalancer{
		clients: util.NewUnixClients(clients),
	}
}

func (lb *LoadBalancer) toggleIndex() int {
	if len(lb.clients) == 1 {
		return 0
	}
	return int(lb.requests.Add(1)-1) % len(lb.clients)
}

func (lb *LoadBalancer) Handler(ctx *fasthttp.RequestCtx) {
	req := &ctx.Request
	resp := &ctx.Response
	defer req.CloseBodyStream()
	defer resp.CloseBodyStream()
	idx := lb.toggleIndex()
	if err := lb.clients[idx].Do(req, resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %d %s", resp.StatusCode(), string(resp.Body()))
		ctx.Error(err.Error(), fasthttp.StatusBadGateway)
	}
	// log.Println("Gateway:",
	// 	string(req.Header.RequestURI()),
	// 	fmt.Sprintf("%d %d %s", idx, resp.StatusCode(), string(resp.Body())),
	// )
}

func Run() error {
	apis := strings.Split(util.GetEnv("SERVICES_SOCKETS"), ",")
	if len(apis) == 0 {
		log.Fatalln("invalid SERVICES_SOCKETS environment variable")
	}
	log.Println("Gateway clients:", apis)
	lb := NewLoadBalancer(apis)
	addr := util.GetEnvOr("SERVER_ADDRESS", ":9999")
	log.Println("Starting Gateway server: " + addr)
	return server.RunHttpServer(addr, lb.Handler)
}

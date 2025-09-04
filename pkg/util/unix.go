package util

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/valyala/fasthttp"
)

func NewSocketFromEnv(socketEnv string) string {
	return NewSocketFile(GetEnv(socketEnv))
}

func NewSocketFile(socketPath string) string {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	// fp, err := os.Create(socketPath)
	// if err != nil {
	// 	log.Fatalf("Failed to create socket file: %v", err)
	// }
	// fp.Close()
	return socketPath
}

func NewListenUnix(socketPath string) net.Listener {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on Unix socket: %v", err)
	}
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set socket permissions: %v", err)
	}
	return listener
}

func NewUnixClient(socketPath string) *fasthttp.Client {
	for i := range 10 {
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			if i == 9 {
				log.Fatalf("UNIX socket %s does not exist", socketPath)
			}
		}
		time.Sleep(100 * time.Millisecond)
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
			return net.Dial("unix", socketPath)
		},
	}
}

func NewUnixClients(param []string) []*fasthttp.Client {
	var clients []*fasthttp.Client
	for _, socketPath := range param {
		clients = append(clients, NewUnixClient(socketPath))
	}
	return clients
}

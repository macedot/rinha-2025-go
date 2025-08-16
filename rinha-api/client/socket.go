package client

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/valyala/fasthttp"
)

type SocketClient struct {
	client *fasthttp.Client
}

func NewSocketClient() *SocketClient {
	return &SocketClient{}
}

func (c *SocketClient) Init(unixSocketPath string) *SocketClient {
	if _, err := os.Stat(unixSocketPath); os.IsNotExist(err) {
		log.Fatalf("UNIX socket %s does not exist", unixSocketPath)
	}
	c.client = &fasthttp.Client{
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
	return c
}

func (c *SocketClient) Get(url string) (int, []byte) {
	statusCode, body, err := c.client.Get([]byte{}, url)
	if err != nil {
		log.Fatalf("HTTP GET request failed: %v", err)
	}
	if statusCode != fasthttp.StatusOK {
		log.Fatalf("HTTP GET request failed: %v", err)
	}
	return statusCode, body
}

func (c *SocketClient) Post(url string, payload []byte) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.SetBodyRaw(payload)
	err := c.client.Do(req, resp)
	if err != nil {
		return nil, err
	}
	statusCode := resp.StatusCode()
	if statusCode >= fasthttp.StatusMultipleChoices {
		return nil, fmt.Errorf("[DB] Unexpected status: %d", statusCode)
	}
	return resp.Body(), nil
}

func (c *SocketClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

func NewSocket(socketPath string) {
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	fp, err := os.Create(socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket file: %v", err)
	}
	fp.Close()
}

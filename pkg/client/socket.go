package client

import (
	"fmt"
	"log"
	"net"
	"os"

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
		// MaxConnsPerHost:               500,
		// ReadTimeout:                   700 * time.Millisecond,
		// WriteTimeout:                  700 * time.Millisecond,
		// ReadBufferSize:                1024,
		// WriteBufferSize:               1024,
		// MaxIdleConnDuration:           10 * time.Second,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: func(_ string) (net.Conn, error) {
			return net.Dial("unix", unixSocketPath)
		},
	}
	return c
}

func (c *SocketClient) Get(url string) (int, []byte, error) {
	statusCode, body, err := c.client.Get([]byte{}, url)
	if err != nil {
		log.Printf("HTTP GET request failed(%d): %v", statusCode, err)
		return 500, nil, err
	}
	if statusCode != fasthttp.StatusOK {
		log.Printf("HTTP GET request failed(%d)", statusCode)
		return 500, nil, fmt.Errorf("unexpected status: %d", statusCode)
	}
	return statusCode, body, nil
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
		return nil, fmt.Errorf("unexpected status: %d", statusCode)
	}
	return resp.Body(), nil
}

func (c *SocketClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

package services

import (
	"log"
	"rinha-2025-go/internal/config"
	"time"

	"github.com/valyala/fasthttp"
)

type HttpClient struct {
	client *fasthttp.Client
}

var headerContentTypeJSON = []byte("application/json")

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &fasthttp.Client{
			ReadTimeout:                   5 * time.Second,
			WriteTimeout:                  5 * time.Second,
			MaxIdleConnDuration:           1 * time.Hour,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		},
	}
}

func (c *HttpClient) Get(url string) (int, []byte) {
	statusCode, body, err := c.client.Get([]byte{}, url)
	if err != nil {
		log.Fatalf("HTTP GET request failed: %v", err)
	}
	if statusCode != fasthttp.StatusOK {
		log.Fatalf("HTTP GET request failed: %v", err)
	}
	return statusCode, body
}

func (c *HttpClient) Post(url string, payload []byte, instance *config.Service) (int, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentTypeBytes(headerContentTypeJSON)
	req.Header.Set("X-Rinha-Token", instance.Token)
	req.SetBodyRaw(payload)
	resp := fasthttp.AcquireResponse()
	err := c.client.DoTimeout(req, resp, instance.Timeout+time.Duration(instance.MinResponseTime)*time.Millisecond)
	fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

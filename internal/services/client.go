package services

import (
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/pkg/http"
	"time"

	"github.com/valyala/fasthttp"
)

type HttpClient struct {
	client *fasthttp.Client
}

var headerContentTypeJSON = []byte("application/json")

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: http.NewFastHttpClient(),
	}
}

func (c *HttpClient) Get(url string, instance *config.Service) (int, []byte) {
	statusCode, body, err := c.client.GetTimeout([]byte{}, url, instance.Timeout)
	if err != nil {
		log.Printf("HTTP GET request failed: %v", err)
		return 0, nil
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

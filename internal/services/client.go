package services

import (
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

func (c *HttpClient) Get(url string, instance *config.Service) (int, []byte, error) {
	return c.makeRequest(fasthttp.MethodGet, url, nil, instance)
}

func (c *HttpClient) Post(url string, payload []byte, instance *config.Service) (int, error) {
	status, _, err := c.makeRequest(fasthttp.MethodPost, url, payload, instance)
	if err != nil {
		return 0, err
	}
	return status, nil
}

func (c *HttpClient) makeRequest(method string, url string, payload []byte, instance *config.Service) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.Set("X-Rinha-Token", instance.Token)
	if method == fasthttp.MethodPost {
		req.Header.SetContentTypeBytes(headerContentTypeJSON)
		req.SetBodyRaw(payload)
	}
	timeout := instance.Timeout + time.Duration(instance.MinResponseTime)*time.Millisecond
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

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
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("X-Rinha-Token", instance.Token)
	timeout := instance.Timeout + time.Duration(instance.MinResponseTime)*time.Millisecond
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (c *HttpClient) Post(url string, payload []byte, instance *config.Service) (int, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentTypeBytes(headerContentTypeJSON)
	req.Header.Set("X-Rinha-Token", instance.Token)
	req.SetBodyRaw(payload)
	timeout := instance.Timeout + time.Duration(instance.MinResponseTime)*time.Millisecond
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

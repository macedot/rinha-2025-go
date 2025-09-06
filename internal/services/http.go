package services

import (
	"rinha-2025-go/internal/config"
	"time"

	"github.com/valyala/fasthttp"
)

type HttpClient struct {
	client *fasthttp.Client
}

func NewHttpClient(client *fasthttp.Client) *HttpClient {
	return &HttpClient{
		client: client,
	}
}

func (c *HttpClient) Get(url string, instance *config.Service) (int, []byte, error) {
	return c.buildRequest(url, fasthttp.MethodGet, nil, instance)
}

func (c *HttpClient) Post(url string, payload []byte, instance *config.Service) (int, error) {
	status, _, err := c.buildRequest(url, fasthttp.MethodPost, payload, instance)
	return status, err
}

func (c *HttpClient) buildRequest(url string, method string, payload []byte, instance *config.Service) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.Set("X-Rinha-Token", instance.Token)
	if method == fasthttp.MethodPost {
		req.Header.SetContentTypeBytes([]byte("application/json"))
		req.SetBodyRaw(payload)
	}
	timeout := instance.Timeout + time.Duration(instance.MinResponseTime)*time.Millisecond
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

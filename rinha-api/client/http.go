package client

import (
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

type HttpClient struct {
	client *fasthttp.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{}
}

func (c *HttpClient) Init() *HttpClient {
	c.client = &fasthttp.Client{
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
	}
	return c
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

func (c *HttpClient) Post(url string, payload []byte) (int, []byte, error) {
	return c.PostTimeout(url, payload, 10*time.Second)
}

func (c *HttpClient) PostTimeout(url string, payload []byte, timeout time.Duration) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.SetBodyRaw(payload)
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 500, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (c *HttpClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

func (c *HttpClient) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return c.client.DoTimeout(req, resp, timeout)
}

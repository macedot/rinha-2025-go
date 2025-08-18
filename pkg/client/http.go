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
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		NoDefaultUserAgentHeader:      true,
		// MaxConnsPerHost:               500,
		// MaxIdleConnDuration:           1 * time.Hour,
		// ReadBufferSize:                1024,
		// ReadTimeout:                   6 * time.Second,
		// WriteBufferSize:               1024,
		// WriteTimeout:                  6 * time.Second,
		// Dial: (&fasthttp.TCPDialer{
		// 	Concurrency:      500,
		// 	DNSCacheDuration: time.Hour,
		// }).Dial,
	}
	return c
}

func (c *HttpClient) Get(url string) (int, []byte, error) {
	return c.GetTimeout(url, 6*time.Second)
}

func (c *HttpClient) GetTimeout(url string, timeout time.Duration) (int, []byte, error) {
	statusCode, body, err := c.client.GetTimeout([]byte{}, url, timeout)
	if err != nil {
		log.Printf("GET: %v | %s", err, url)
		return fasthttp.StatusInternalServerError, body, err
	}
	return statusCode, body, nil
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
		log.Printf("POST: %v | %s", err, url)
		return fasthttp.StatusInternalServerError, nil, err
	}
	return resp.StatusCode(), resp.Body(), nil
}

func (c *HttpClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

func (c *HttpClient) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return c.client.DoTimeout(req, resp, timeout)
}

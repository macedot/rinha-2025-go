package http

import (
	"time"

	"github.com/valyala/fasthttp"
)

func NewFastHttpClient() *fasthttp.Client {
	return &fasthttp.Client{
		ReadBufferSize:                8192,
		WriteBufferSize:               8192,
		MaxConnsPerHost:               4096,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		MaxIdleConnDuration:           10 * time.Second,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
}

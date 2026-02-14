package http

import (
	"time"

	"github.com/valyala/fasthttp"
)

func NewFastHttpClient() *fasthttp.Client {
	return &fasthttp.Client{
		ReadBufferSize:                2048, // 2KB - sufficient for payment payloads (~500 bytes)
		WriteBufferSize:               2048, // 2KB - reduces memory vs 8KB default
		MaxConnsPerHost:               512,  // Right-sized for 4 API instances (128 each)
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		MaxIdleConnDuration:           10 * time.Second,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // Headers set correctly, skip normalization
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      512, // Match MaxConnsPerHost
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
}

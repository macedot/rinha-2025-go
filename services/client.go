package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// DNSCache holds cached DNS resolutions
type DNSCache struct {
	cache map[string][]string
	mutex sync.RWMutex
	ttl   time.Duration
}

type Client struct {
	dnsCache *DNSCache
	http     *http.Client
}

var client Client

func HttpClientInstance() *Client {
	return &client
}

func (c *Client) Init() {
	c.dnsCache = NewDNSCache(5 * time.Minute)
	c.http = &http.Client{
		Transport: CustomTransport(client.dnsCache),
		Timeout:   30 * time.Second,
	}
}

// NewDNSCache creates a new DNS cache with specified TTL
func NewDNSCache(ttl time.Duration) *DNSCache {
	return &DNSCache{
		cache: make(map[string][]string),
		mutex: sync.RWMutex{},
		ttl:   ttl,
	}
}

// Resolve looks up IP addresses for a hostname, using cache if available
func (dc *DNSCache) Resolve(host string) ([]string, error) {
	dc.mutex.RLock()
	if addrs, found := dc.cache[host]; found {
		dc.mutex.RUnlock()
		return addrs, nil
	}
	dc.mutex.RUnlock()

	// Perform DNS lookup
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}

	// Cache the results
	dc.mutex.Lock()
	dc.cache[host] = addrs
	dc.mutex.Unlock()

	// Schedule cache eviction
	go func() {
		<-time.After(dc.ttl)
		dc.mutex.Lock()
		delete(dc.cache, host)
		dc.mutex.Unlock()
	}()

	return addrs, nil
}

// CustomDialer creates a dialer with DNS cache
func CustomDialer(dc *DNSCache) *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}

// CustomTransport creates an HTTP transport with custom dialer
func CustomTransport(dc *DNSCache) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// Resolve using DNS cache
			addrs, err := dc.Resolve(host)
			if err != nil {
				return nil, err
			}

			// Try connecting to each resolved IP
			dialer := CustomDialer(dc)
			for _, ip := range addrs {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("failed to connect to %s", addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func (c *Client) Get(url string) (int, []byte) {
	resp, err := c.http.Get(url)
	if err != nil {
		log.Fatalf("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		return resp.StatusCode, body
	}
	return resp.StatusCode, nil
}

func (c *Client) Post(url string, payload []byte) error {
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("HTTP POST request failed: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

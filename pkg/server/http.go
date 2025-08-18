package server

import (
	"github.com/valyala/fasthttp"
)

func RunSocketServer(socketPath string, handler func(ctx *fasthttp.RequestCtx)) error {
	// Configure server for low latency
	server := &fasthttp.Server{
		// Concurrency:        1000, // Adjust based on system resources
		// DisableKeepalive:   true, // Disable for lowest latency
		// Logger:             log.New(os.Stdout, "", log.LstdFlags), // Minimal logging
		// MaxConnsPerIP:      500,                                   // Prevent abuse from single IP
		// MaxRequestsPerConn: 1000,                                  // Limit requests per connection
		// ReadTimeout:        100 * time.Millisecond,                // Short timeout for reads
		// WriteTimeout:       100 * time.Millisecond,                // Short timeout for writes
		Handler:            handler,
		ReduceMemoryUsage:  true, // Minimize memory allocations
		ReadBufferSize:     1024, // Match typical payload size
		WriteBufferSize:    1024, // Match typical payload size
		MaxRequestBodySize: 1024, // 1KiB limit to prevent abuse
	}
	return server.ListenAndServeUNIX(socketPath, 0666)
}

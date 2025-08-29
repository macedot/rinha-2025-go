package server

import (
	"log"
	"net"

	"github.com/valyala/fasthttp"
)

// tcpNoDelayListener wraps a net.Listener to set TCPNoDelay on accepted connections
type tcpNoDelayListener struct {
	net.Listener
}

func (ln *tcpNoDelayListener) Accept() (net.Conn, error) {
	conn, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true) // Disable Nagle's algorithm for low latency
	}
	return conn, nil
}

func RunHttpServer(addr string, handler func(ctx *fasthttp.RequestCtx)) error {
	// Configure server for low latency
	server := &fasthttp.Server{
		Concurrency:   1000, // Adjust based on system resources
		MaxConnsPerIP: 1000, // Prevent abuse from single IP
		// DisableKeepalive: false, // Disable for lowest latency
		// Logger:           log.New(os.Stdout, "", log.LstdFlags), // Minimal logging
		// MaxRequestsPerConn: 1000,                                  // Limit requests per connection
		// ReadTimeout:        100 * time.Millisecond,                // Short timeout for reads
		// WriteTimeout:       100 * time.Millisecond,                // Short timeout for writes
		Handler:            handler,
		ReduceMemoryUsage:  true, // Minimize memory allocations
		ReadBufferSize:     1024, // Match typical payload size
		WriteBufferSize:    1024, // Match typical payload size
		MaxRequestBodySize: 1024, // 1KiB limit to prevent abuse
	}

	// Create TCP listener
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Wrap listener to set TCPNoDelay on accepted connections
	noDelayLn := &tcpNoDelayListener{Listener: ln}

	// Start server
	return server.Serve(noDelayLn)
}

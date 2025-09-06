package client

import (
	"log"
	"net"
	"os"
	"path/filepath"
)

func NewListenSocket(socketPath string) net.Listener {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on Unix socket: %v", err)
	}
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set socket permissions: %v", err)
	}
	return listener
}

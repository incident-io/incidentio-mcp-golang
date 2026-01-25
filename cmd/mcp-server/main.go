package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/incident-io/incidentio-mcp-golang/internal/server"
)

func main() {
	// Parse command-line flags
	transport := flag.String("transport", "stdio", "Transport type: stdio or http")
	port := flag.Int("port", 8080, "Port for HTTP transport (default: 8080)")
	flag.Parse()

	// Also check environment variables (useful for Docker)
	if envTransport := os.Getenv("MCP_TRANSPORT"); envTransport != "" {
		*transport = envTransport
	}
	if envPort := os.Getenv("MCP_PORT"); envPort != "" {
		var p int
		if _, err := fmt.Sscanf(envPort, "%d", &p); err == nil {
			*port = p
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Configure and start server
	cfg := server.Config{
		Transport: server.TransportType(*transport),
		Port:      *port,
	}

	srv := server.NewWithConfig(cfg)
	if err := srv.Start(ctx); err != nil {
		log.Printf("Server error: %v", err)
	}
}

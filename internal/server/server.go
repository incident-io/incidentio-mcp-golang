package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
	"github.com/incident-io/incidentio-mcp-golang/internal/handlers"
	"github.com/incident-io/incidentio-mcp-golang/pkg/mcp"
)

type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

type Server struct {
	tools     map[string]Handler
	transport TransportType
	port      int
	mu        sync.RWMutex
}

// Handler interface for tool handlers
type Handler interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(args map[string]interface{}) (string, error)
}

type Config struct {
	Transport TransportType
	Port      int
}

func New() *Server {
	return NewWithConfig(Config{Transport: TransportStdio})
}

func NewWithConfig(cfg Config) *Server {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	return &Server{
		tools:     make(map[string]Handler),
		transport: cfg.Transport,
		port:      cfg.Port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.registerTools()

	switch s.transport {
	case TransportHTTP:
		return s.startHTTP(ctx)
	default:
		return s.startStdio(ctx)
	}
}

func (s *Server) startStdio(ctx context.Context) error {
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var msg mcp.Message
			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					return nil
				}
				continue
			}

			response, err := s.handleMessage(&msg)
			if err != nil {
				response = s.createErrorResponse(msg.ID, err)
			}

			if response != nil {
				if err := encoder.Encode(response); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
				}
			}
		}
	}
}

func (s *Server) startHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	// MCP endpoint - handles JSON-RPC over HTTP POST
	mux.HandleFunc("/mcp", s.handleHTTPMCP)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		server.Shutdown(context.Background())
	}()

	log.Printf("Starting MCP HTTP server on port %d", s.port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

func (s *Server) handleHTTPMCP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for Claude Code compatibility
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg mcp.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(s.createErrorResponse(nil, fmt.Errorf("invalid JSON: %w", err)))
		return
	}

	response, err := s.handleMessage(&msg)
	if err != nil {
		response = s.createErrorResponse(msg.ID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	if response != nil {
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) registerTools() {
	// Initialize incident.io client
	c, err := client.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize incident.io client: %v", err)
		return
	}

	// Use the new registry to register all tools
	registry := handlers.NewToolRegistry()
	registry.RegisterAllTools(c)

	// Copy tools from registry to server
	s.mu.Lock()
	for name, tool := range registry.GetTools() {
		s.tools[name] = tool
	}
	s.mu.Unlock()
}

func (s *Server) handleMessage(msg *mcp.Message) (*mcp.Message, error) {
	// Handle notifications (no ID means it's a notification)
	if msg.ID == nil {
		return nil, nil
	}

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolCall(msg)
	default:
		return &mcp.Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", msg.Method),
			},
		}, nil
	}
}

func (s *Server) handleInitialize(msg *mcp.Message) (*mcp.Message, error) {
	response := &mcp.Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "incidentio-mcp-server",
				"version": "0.1.0",
			},
		},
	}
	return response, nil
}

func (s *Server) handleToolsList(msg *mcp.Message) (*mcp.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var toolsList []map[string]interface{}
	for _, tool := range s.tools {
		toolsList = append(toolsList, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		})
	}

	response := &mcp.Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"tools": toolsList,
		},
	}
	return response, nil
}

func (s *Server) handleToolCall(msg *mcp.Message) (*mcp.Message, error) {
	params, ok := msg.Params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params")
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	s.mu.RLock()
	tool, exists := s.tools[toolName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	args, _ := params["arguments"].(map[string]interface{})

	result, err := tool.Execute(args)
	if err != nil {
		return nil, err
	}

	response := &mcp.Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
	return response, nil
}

func (s *Server) createErrorResponse(id interface{}, err error) *mcp.Message {
	return &mcp.Message{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &mcp.Error{
			Code:    -32603,
			Message: err.Error(),
		},
	}
}

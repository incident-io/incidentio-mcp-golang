package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// ErrSessionAlreadyExists is returned when attempting to create a session that already exists
var ErrSessionAlreadyExists = errors.New("session already exists")

// MCPMessage represents a JSON-RPC message
type MCPMessage struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// MCPProxy manages the connection between HTTP/SSE and stdio MCP server
type MCPProxy struct {
	mcpServerPath string
	mu            sync.Mutex
	sessions      map[string]*Session
}

// ProtocolState represents the initialization state of an MCP session
type ProtocolState int

const (
	StateUninitialized ProtocolState = iota
	StateInitializing
	StateReady
)

// Session represents a single MCP session
type Session struct {
	ID                   string
	cmd                  *exec.Cmd
	stdin                io.WriteCloser
	stdout               io.ReadCloser
	stderr               io.ReadCloser
	messages             chan MCPMessage
	ctx                  context.Context
	cancel               context.CancelFunc
	mu                   sync.Mutex
	wg                   sync.WaitGroup // Tracks active goroutines for safe cleanup
	refCount             int32          // Atomic reference counter for active connections
	protocolState        ProtocolState  // Tracks MCP protocol initialization state
	stateMu              sync.RWMutex   // Protects protocolState and initializeRequestID
	initializeRequestID  interface{}    // Tracks the ID of the initialize request for matching responses
}

func NewMCPProxy(mcpServerPath string) *MCPProxy {
	return &MCPProxy{
		mcpServerPath: mcpServerPath,
		sessions:      make(map[string]*Session),
	}
}

func (p *MCPProxy) createSession(sessionID string) (*Session, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.sessions[sessionID]; exists {
		return nil, ErrSessionAlreadyExists
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, p.mcpServerPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	session := &Session{
		ID:            sessionID,
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		stderr:        stderr,
		messages:      make(chan MCPMessage, 100),
		ctx:           ctx,
		cancel:        cancel,
		protocolState: StateUninitialized,
	}

	// Start reading from MCP server stdout
	session.wg.Add(1)
	go session.readLoop()

	// Forward stderr to logs
	session.wg.Add(1)
	go func() {
		defer session.wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[MCP Server %s] %s", sessionID, scanner.Text())
		}
	}()

	// Wait for process to exit and clean up zombie process
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("[MCP Server %s] Process exited with error: %v", sessionID, err)
		} else {
			log.Printf("[MCP Server %s] Process exited cleanly", sessionID)
		}
	}()

	p.sessions[sessionID] = session
	return session, nil
}

func (p *MCPProxy) getSession(sessionID string) (*Session, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (p *MCPProxy) destroySession(sessionID string) {
	p.mu.Lock()
	session, exists := p.sessions[sessionID]
	if !exists {
		p.mu.Unlock()
		return
	}
	// Remove from session map immediately while holding lock
	delete(p.sessions, sessionID)
	p.mu.Unlock()

	log.Printf("Destroying session: %s", sessionID)

	// Cancel context to signal goroutines to exit
	session.cancel()

	// Close all pipes to unblock any pending I/O operations
	session.stdin.Close()
	session.stdout.Close()
	session.stderr.Close()

	// Wait for all goroutines (readLoop and stderr scanner) to complete
	// This ensures no goroutine will try to send to the channel after we close it
	log.Printf("Waiting for goroutines to exit for session: %s", sessionID)
	session.wg.Wait()

	// NOW it's safe to close the channel - all senders have exited
	close(session.messages)
	log.Printf("Session cleanup complete: %s", sessionID)

	// Note: cmd.Wait() is handled by the goroutine in createSession
	// The process will be reaped automatically when it exits
}

func (s *Session) readLoop() {
	defer s.wg.Done()
	decoder := json.NewDecoder(s.stdout)
	for {
		// Check context cancellation before attempting to decode
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		var msg MCPMessage
		if err := decoder.Decode(&msg); err != nil {
			if err != io.EOF {
				log.Printf("Error decoding message from MCP server: %v", err)
			}
			return
		}

		// Track protocol state transitions based on responses
		// Responses have Method == "" and either Result or Error set
		if msg.Method == "" && msg.ID != nil {
			s.stateMu.Lock()
			// Check if this is a response to our initialize request
			if s.protocolState == StateInitializing && s.initializeRequestID != nil {
				// Compare request IDs (handles both string and numeric IDs)
				if fmt.Sprintf("%v", msg.ID) == fmt.Sprintf("%v", s.initializeRequestID) {
					if msg.Error != nil {
						// Initialize failed - reset to uninitialized
						s.protocolState = StateUninitialized
						s.initializeRequestID = nil
						log.Printf("[Session %s] MCP initialization failed: %v", s.ID, msg.Error)
					} else if msg.Result != nil {
						// Initialize succeeded
						s.protocolState = StateReady
						s.initializeRequestID = nil
						log.Printf("[Session %s] MCP protocol initialized successfully", s.ID)
					}
				}
			}
			s.stateMu.Unlock()
		}

		// Safe send: context is checked before decode, so if we get here
		// we know destroySession hasn't closed the channel yet (it waits for wg)
		select {
		case s.messages <- msg:
			// Message sent successfully
		case <-s.ctx.Done():
			// Context canceled while trying to send
			return
		}
	}
}

func (s *Session) sendMessage(msg MCPMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for concurrent initialize requests before sending
	if msg.Method == "initialize" {
		s.stateMu.Lock()
		currentState := s.protocolState
		s.stateMu.Unlock()

		// Reject concurrent initialize requests
		if currentState == StateInitializing {
			return fmt.Errorf("initialization already in progress")
		}

		log.Printf("[Session %s] Sending initialize request with ID: %v", s.ID, msg.ID)
	} else if msg.Method == "notifications/initialized" {
		log.Printf("[Session %s] Forwarding notifications/initialized to MCP server", s.ID)
	}

	// Send the message first
	encoder := json.NewEncoder(s.stdin)
	if err := encoder.Encode(msg); err != nil {
		return err
	}

	// Only update state AFTER successful send
	if msg.Method == "initialize" {
		s.stateMu.Lock()
		s.protocolState = StateInitializing
		s.initializeRequestID = msg.ID // Store the request ID to match with response
		s.stateMu.Unlock()
		log.Printf("[Session %s] Initialize request sent successfully, state → StateInitializing", s.ID)
	} else if msg.Method == "notifications/initialized" {
		log.Printf("[Session %s] notifications/initialized forwarded successfully", s.ID)
	}

	return nil
}

func (s *Session) getState() ProtocolState {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.protocolState
}

func (p *MCPProxy) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	session, err := p.createSession(sessionID)
	if err != nil {
		// Only retry with getSession if the session already exists
		if errors.Is(err, ErrSessionAlreadyExists) {
			session, err = p.getSession(sessionID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Session exists but could not be retrieved: %v", err), http.StatusInternalServerError)
				return
			}
			log.Printf("Reusing existing session: %s", sessionID)
		} else {
			// All other errors are genuine creation failures - report them directly
			http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Increment reference count for this connection
	refCount := atomic.AddInt32(&session.refCount, 1)
	log.Printf("Client connected to session %s (refCount: %d)", sessionID, refCount)

	// Ensure cleanup when client disconnects (always run, regardless of who created the session)
	defer func() {
		// Decrement reference count
		newRefCount := atomic.AddInt32(&session.refCount, -1)
		log.Printf("Client disconnected from session %s (refCount: %d)", sessionID, newRefCount)

		// If this was the last client, destroy the session
		if newRefCount == 0 {
			log.Printf("Last client disconnected, cleaning up session: %s", sessionID)
			p.destroySession(sessionID)
		}
	}()

	log.Printf("SSE connection established for session: %s", sessionID)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send endpoint event (required by MCP Inspector and clients)
	// This tells the client where to POST messages
	messageEndpoint := fmt.Sprintf("/message?session=%s", sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
	flusher.Flush()

	// Also send session ID for compatibility
	fmt.Fprintf(w, "event: session\ndata: %s\n\n", sessionID)
	flusher.Flush()

	// Stream messages from MCP server
	for {
		select {
		case <-r.Context().Done():
			log.Printf("SSE connection closed for session: %s", sessionID)
			return
		case msg := <-session.messages:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}
			// Send as 'message' event per MCP SSE spec
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			// Send keepalive
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (p *MCPProxy) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Missing session parameter", http.StatusBadRequest)
		return
	}

	session, err := p.getSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
		return
	}

	var msg MCPMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Handle ping directly in the proxy by sending response through SSE stream
	if msg.Method == "ping" {
		log.Printf("[Session %s] Handling ping request, sending response via SSE", sessionID)
		// Send ping response through the SSE message channel
		pingResponse := MCPMessage{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Result:  map[string]interface{}{}, // Empty result per MCP spec
		}
		// Use context-aware send to prevent panic if session is being destroyed
		select {
		case session.messages <- pingResponse:
			// Response sent successfully via SSE
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
		case <-session.ctx.Done():
			// Session destroyed during ping - return error
			log.Printf("[Session %s] Ping failed: session destroyed", sessionID)
			http.Error(w, "Session destroyed", http.StatusGone)
		case <-time.After(5 * time.Second):
			http.Error(w, "Timeout sending ping response", http.StatusInternalServerError)
		}
		return
	}

	// Empty method indicates a malformed request or response (responses should come via SSE only)
	if msg.Method == "" {
		errMsg := "Invalid request: method field is required"
		log.Printf("[Session %s] Rejected request with empty method", sessionID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MCPMessage{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: map[string]interface{}{
				"code":    -32600,
				"message": errMsg,
			},
		})
		return
	}

	// Validate protocol state based on method
	state := session.getState()

	// 'initialize' is only allowed when StateUninitialized
	if msg.Method == "initialize" {
		if state != StateUninitialized {
			errMsg := fmt.Sprintf("Initialize not allowed in state %d. Session already initialized or initialization in progress.", state)
			log.Printf("[Session %s] Rejected initialize request in state %d", sessionID, state)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(MCPMessage{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Error: map[string]interface{}{
					"code":    -32002,
					"message": errMsg,
					"data": map[string]interface{}{
						"state": state,
					},
				},
			})
			return
		}
	} else if msg.Method == "notifications/initialized" {
		// 'notifications/initialized' is only allowed when StateReady (after initialize response)
		if state != StateReady {
			var errMsg string
			if state == StateInitializing {
				errMsg = "notifications/initialized not allowed yet. Wait for initialize response first."
			} else {
				errMsg = "notifications/initialized not allowed. Must send 'initialize' request and receive response first."
			}
			log.Printf("[Session %s] Rejected notifications/initialized in state %d", sessionID, state)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(MCPMessage{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Error: map[string]interface{}{
					"code":    -32002,
					"message": errMsg,
					"data": map[string]interface{}{
						"state": state,
					},
				},
			})
			return
		}
	} else {
		// All other methods require StateReady
		if state != StateReady {
			var errMsg string
			if state == StateInitializing {
				errMsg = "Initialization in progress. Please wait for initialize response and send notifications/initialized before other requests."
			} else {
				errMsg = "Protocol not initialized. Must send 'initialize' request first."
			}

			log.Printf("[Session %s] Rejected %s request in state %d: %s", sessionID, msg.Method, state, errMsg)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(MCPMessage{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Error: map[string]interface{}{
					"code":    -32002,
					"message": errMsg,
					"data": map[string]interface{}{
						"state": state,
					},
				},
			})
			return
		}
	}

	if err := session.sendMessage(msg); err != nil {
		// Check if this is a concurrent initialize error
		if msg.Method == "initialize" && err.Error() == "initialization already in progress" {
			log.Printf("[Session %s] Rejected concurrent initialize request", sessionID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(MCPMessage{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Error: map[string]interface{}{
					"code":    -32001,
					"message": "Initialization already in progress. Please wait for the current initialization to complete.",
				},
			})
			return
		}

		// Other send errors
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (p *MCPProxy) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"sessions": len(p.sessions),
	})
}

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	mcpServerPath := flag.String("mcp-server", "./mcp-server", "Path to MCP server binary")
	flag.Parse()

	// Check if MCP server exists
	if _, err := os.Stat(*mcpServerPath); os.IsNotExist(err) {
		log.Fatalf("MCP server not found at: %s", *mcpServerPath)
	}

	proxy := NewMCPProxy(*mcpServerPath)

	http.HandleFunc("/sse", proxy.handleSSE)
	http.HandleFunc("/message", proxy.handleMessage)
	http.HandleFunc("/health", proxy.handleHealth)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting MCP HTTP/SSE proxy server on %s", addr)
	log.Printf("SSE endpoint: http://localhost%s/sse", addr)
	log.Printf("Message endpoint: http://localhost%s/message", addr)
	log.Printf("Health endpoint: http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

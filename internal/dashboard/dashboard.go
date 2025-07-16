package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"Yamcha/internal/attacker"
	"Yamcha/internal/config"
	"Yamcha/internal/metrics"
)

// TestSession represents a running test session
type TestSession struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Config      *config.Config             `json:"config"`
	Metrics     *metrics.Metrics           `json:"-"`
	RTCollector *metrics.RealTimeCollector `json:"-"`
	Status      string                     `json:"status"` // "running", "completed", "failed", "cancelled"
	StartTime   time.Time                  `json:"start_time"`
	EndTime     *time.Time                 `json:"end_time,omitempty"`
	LiveStats   *metrics.Statistics        `json:"live_stats,omitempty"`
	ctx         context.Context
	cancel      context.CancelFunc
}

// Dashboard manages the web dashboard and test sessions
type Dashboard struct {
	server     *http.Server
	router     *mux.Router
	upgrader   websocket.Upgrader
	sessions   map[string]*TestSession
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	sessionsMu sync.RWMutex
	broadcast  chan []byte
	port       int
}

// NewDashboard creates a new dashboard instance
func NewDashboard(port int) *Dashboard {
	router := mux.NewRouter()

	dashboard := &Dashboard{
		router: router,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin in development
			},
		},
		sessions:  make(map[string]*TestSession),
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		port:      port,
	}

	dashboard.setupRoutes()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	dashboard.server = server

	return dashboard
}

// setupRoutes configures all HTTP routes
func (d *Dashboard) setupRoutes() {
	// Static files
	d.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// Main dashboard page
	d.router.HandleFunc("/", d.handleDashboard).Methods("GET")

	// API routes
	api := d.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/sessions", d.handleGetSessions).Methods("GET")
	api.HandleFunc("/sessions", d.handleCreateSession).Methods("POST")
	api.HandleFunc("/sessions/{id}/start", d.handleStartSession).Methods("POST")
	api.HandleFunc("/sessions/{id}/stop", d.handleStopSession).Methods("POST")
	api.HandleFunc("/sessions/{id}/status", d.handleGetSessionStatus).Methods("GET")
	api.HandleFunc("/sessions/{id}", d.handleDeleteSession).Methods("DELETE")

	// WebSocket endpoint
	d.router.HandleFunc("/ws", d.handleWebSocket)
}

// Start starts the dashboard server
func (d *Dashboard) Start() error {
	// Start the broadcast handler
	go d.handleBroadcast()

	log.Printf("🌐 Dashboard starting on http://localhost:%d", d.port)
	return d.server.ListenAndServe()
}

// Stop stops the dashboard server
func (d *Dashboard) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.server.Shutdown(ctx)
}

// handleDashboard serves the main dashboard HTML page
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

// handleWebSocket handles WebSocket connections
func (d *Dashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 WebSocket connection attempt from %s", r.RemoteAddr)
	conn, err := d.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	d.clientsMu.Lock()
	d.clients[conn] = true
	clientCount := len(d.clients)
	d.clientsMu.Unlock()

	log.Printf("✅ WebSocket client connected. Total clients: %d", clientCount)

	// Send initial session data
	d.sendSessionsUpdate(conn)

	// Listen for client messages (mainly for keepalive)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			d.clientsMu.Lock()
			delete(d.clients, conn)
			clientCount := len(d.clients)
			d.clientsMu.Unlock()
			log.Printf("📤 WebSocket client disconnected. Remaining clients: %d", clientCount)
			break
		}
	}
}

// handleBroadcast handles broadcasting messages to all connected clients
func (d *Dashboard) handleBroadcast() {
	log.Printf("📻 Broadcast handler started")
	for message := range d.broadcast {
		d.clientsMu.RLock()
		clientCount := len(d.clients)
		d.clientsMu.RUnlock()

		log.Printf("📤 Broadcasting message to %d clients", clientCount)

		d.clientsMu.RLock()
		for client := range d.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("❌ Error sending to WebSocket client: %v", err)
				client.Close()
				delete(d.clients, client)
			}
		}
		d.clientsMu.RUnlock()
	}
}

// broadcastUpdate sends an update to all connected clients
func (d *Dashboard) broadcastUpdate(updateType string, data interface{}) {
	log.Printf("📡 Broadcasting update type: %s", updateType)

	message := map[string]interface{}{
		"type": updateType,
		"data": data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Error marshaling broadcast data: %v", err)
		return
	}

	select {
	case d.broadcast <- jsonData:
		log.Printf("✅ Update queued for broadcast: %s", updateType)
	default:
		log.Printf("⚠️ Broadcast channel full, skipping update: %s", updateType)
	}
}

// sendSessionsUpdate sends current sessions to a specific client
func (d *Dashboard) sendSessionsUpdate(conn *websocket.Conn) {
	d.sessionsMu.RLock()
	sessions := make([]*TestSession, 0, len(d.sessions))
	for _, session := range d.sessions {
		// Create a copy without the context and cancel function
		sessionCopy := *session
		sessionCopy.ctx = nil
		sessionCopy.cancel = nil
		sessions = append(sessions, &sessionCopy)
	}
	d.sessionsMu.RUnlock()

	message := map[string]interface{}{
		"type": "sessions_update",
		"data": sessions,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling sessions data: %v", err)
		return
	}

	conn.WriteMessage(websocket.TextMessage, jsonData)
}

// handleGetSessions returns all test sessions
func (d *Dashboard) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	d.sessionsMu.RLock()
	sessions := make([]*TestSession, 0, len(d.sessions))
	for _, session := range d.sessions {
		// Create a copy without the context
		sessionCopy := *session
		sessionCopy.ctx = nil
		sessionCopy.cancel = nil
		sessions = append(sessions, &sessionCopy)
	}
	d.sessionsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// handleCreateSession creates a new test session
func (d *Dashboard) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Received session creation request from %s", r.RemoteAddr)
	log.Printf("🔍 Content-Type: %s", r.Header.Get("Content-Type"))
	log.Printf("🔍 Content-Length: %s", r.Header.Get("Content-Length"))

	// Read the raw body first for logging
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	log.Printf("🔍 Raw request body: %s", string(bodyBytes))

	var req struct {
		Name   string         `json:"name"`
		Config *config.Config `json:"config"`
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("❌ Failed to decode JSON: %v", err)
		log.Printf("❌ Body that failed to parse: %s", string(bodyBytes))
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("✅ Successfully decoded request - Name: %s, URL: %s", req.Name, req.Config.Target.URL)

	// Set defaults for config
	log.Printf("🔍 Setting configuration defaults...")
	log.Printf("🔍 Before defaults - SpikeHeight: %d", req.Config.Load.SpikeHeight)
	if err := req.Config.SetDefaults(); err != nil {
		log.Printf("❌ Failed to set config defaults: %v", err)
		http.Error(w, fmt.Sprintf("Config defaults error: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("🔍 After defaults - SpikeHeight: %d", req.Config.Load.SpikeHeight)

	// Validate config
	log.Printf("🔍 Validating configuration...")
	if err := req.Config.Validate(); err != nil {
		log.Printf("❌ Config validation failed: %v", err)
		http.Error(w, fmt.Sprintf("Invalid config: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("✅ Configuration validated successfully")

	sessionID := fmt.Sprintf("test_%d", time.Now().Unix())
	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("🆔 Created session ID: %s", sessionID)

	session := &TestSession{
		ID:        sessionID,
		Name:      req.Name,
		Config:    req.Config,
		Status:    "created",
		StartTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	d.sessionsMu.Lock()
	d.sessions[sessionID] = session
	d.sessionsMu.Unlock()

	log.Printf("📝 Session stored successfully")

	// Broadcast update
	d.broadcastUpdate("session_created", session)
	log.Printf("📡 Broadcasted session creation update")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
	log.Printf("✅ Session creation completed: %s", sessionID)
}

// handleStartSession starts a test session
func (d *Dashboard) handleStartSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	log.Printf("🚀 Received start request for session: %s", sessionID)

	d.sessionsMu.Lock()
	session, exists := d.sessions[sessionID]
	if !exists {
		log.Printf("❌ Session not found: %s", sessionID)
		d.sessionsMu.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.Status == "running" {
		log.Printf("⚠️ Session already running: %s", sessionID)
		d.sessionsMu.Unlock()
		http.Error(w, "Session already running", http.StatusBadRequest)
		return
	}

	session.Status = "running"
	session.StartTime = time.Now()
	session.Metrics = metrics.NewMetrics()
	session.RTCollector = metrics.NewRealTimeCollector(time.Second)
	d.sessionsMu.Unlock()

	log.Printf("✅ Session status updated to running: %s", sessionID)

	// Start the test in a separate goroutine
	log.Printf("🔄 Starting test execution goroutine for session: %s", sessionID)
	go d.runTestSession(session)

	// Start real-time updates for this session
	log.Printf("📊 Starting real-time monitoring for session: %s", sessionID)
	go d.monitorSession(session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	log.Printf("✅ Session start completed: %s", sessionID)
}

// handleStopSession stops a test session
func (d *Dashboard) handleStopSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	d.sessionsMu.RLock()
	session, exists := d.sessions[sessionID]
	d.sessionsMu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.Status != "running" {
		http.Error(w, "Session not running", http.StatusBadRequest)
		return
	}

	session.cancel()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

// handleGetSessionStatus returns the current status of a session
func (d *Dashboard) handleGetSessionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	d.sessionsMu.RLock()
	session, exists := d.sessions[sessionID]
	d.sessionsMu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update live stats if available
	if session.RTCollector != nil {
		stats := session.RTCollector.GetMetrics().GetStatistics()
		session.LiveStats = &stats
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// handleDeleteSession deletes a test session
func (d *Dashboard) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	log.Printf("🗑️ Received delete request for session: %s", sessionID)

	d.sessionsMu.Lock()
	session, exists := d.sessions[sessionID]
	if exists {
		log.Printf("🗑️ Session found, cancelling context for: %s", sessionID)
		if session.cancel != nil {
			session.cancel()
		}
		delete(d.sessions, sessionID)
		log.Printf("🗑️ Session removed from memory: %s", sessionID)
	}
	d.sessionsMu.Unlock()

	if !exists {
		log.Printf("❌ Session not found for deletion: %s", sessionID)
		http.Error(w, "Session not found - it may have already been deleted", http.StatusNotFound)
		return
	}

	// Broadcast update
	log.Printf("📡 Broadcasting session deletion: %s", sessionID)
	d.broadcastUpdate("session_deleted", map[string]string{"id": sessionID})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	log.Printf("✅ Session deletion completed: %s", sessionID)
}

// runTestSession executes the actual load test
func (d *Dashboard) runTestSession(session *TestSession) {
	defer func() {
		if session.RTCollector != nil {
			session.RTCollector.Stop()
		}

		endTime := time.Now()
		session.EndTime = &endTime

		if session.Status == "running" {
			session.Status = "completed"
		}

		// Broadcast final update
		d.broadcastUpdate("session_completed", session)
	}()

	// Create attacker
	factory := attacker.NewAttackerFactory()
	att, err := factory.CreateAttacker(session.Config.Load.AttackType)
	if err != nil {
		session.Status = "failed"
		log.Printf("Failed to create attacker for session %s: %v", session.ID, err)
		return
	}

	// Start real-time collector
	session.RTCollector.Start()

	// Execute the attack
	err = att.Attack(session.ctx, session.Config, session.RTCollector.GetMetrics())
	if err != nil && session.ctx.Err() == nil {
		session.Status = "failed"
		log.Printf("Attack failed for session %s: %v", session.ID, err)
		return
	}

	if session.ctx.Err() != nil {
		session.Status = "cancelled"
		log.Printf("Session %s was cancelled", session.ID)
	}

	// Copy results to main metrics
	rtResults := session.RTCollector.GetMetrics().GetResults()
	for _, result := range rtResults {
		session.Metrics.AddResult(result)
	}
}

// monitorSession monitors a session and sends live updates
func (d *Dashboard) monitorSession(session *TestSession) {
	log.Printf("📊 Starting real-time monitoring for session: %s", session.ID)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if session.Status != "running" {
				log.Printf("📊 Stopping monitoring for session %s (status: %s)", session.ID, session.Status)
				return
			}

			if session.RTCollector != nil {
				stats := session.RTCollector.GetMetrics().GetStatistics()
				session.LiveStats = &stats
				log.Printf("📊 Live stats for session %s: %d requests, %.2fms avg",
					session.ID, stats.TotalRequests, float64(stats.AvgResponseTime.Nanoseconds())/1000000.0)

				// Broadcast live update
				d.broadcastUpdate("session_update", session)
			} else {
				log.Printf("⚠️ RTCollector is nil for session %s", session.ID)
			}

		case <-session.ctx.Done():
			log.Printf("📊 Session context cancelled for: %s", session.ID)
			return
		}
	}
}

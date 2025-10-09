package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	CompanyID string      `json:"company_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID        string
	CompanyID string
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *WebSocketHub
}

// WebSocketHub manages WebSocket connections with enhanced features
type WebSocketHub struct {
	// Registered clients
	clients map[*Client]bool
	
	// Channels for hub operations
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	
	// Company-specific channels
	companyChannels map[string]chan []byte
	
	// Redis client for pub/sub
	redis *redis.Client
	
	// Mutex for thread safety
	mutex sync.RWMutex
	
	// Configuration
	config *WebSocketConfig
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	PingPeriod      time.Duration
	PongWait        time.Duration
	WriteWait       time.Duration
	MaxMessageSize  int64
}

// DefaultWebSocketConfig returns default WebSocket configuration
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingPeriod:      54 * time.Second,
		PongWait:        60 * time.Second,
		WriteWait:       10 * time.Second,
		MaxMessageSize:  512,
	}
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(redis *redis.Client, config *WebSocketConfig) *WebSocketHub {
	if config == nil {
		config = DefaultWebSocketConfig()
	}
	
	hub := &WebSocketHub{
		clients:         make(map[*Client]bool),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan []byte),
		companyChannels: make(map[string]chan []byte),
		redis:           redis,
		config:          config,
	}
	
	// Start the hub
	go hub.run()
	
	// Start Redis pub/sub for cross-instance communication
	go hub.startRedisPubSub()
	
	return hub
}

// run starts the WebSocket hub
func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			
			// Send welcome message
			welcomeMsg := WebSocketMessage{
				Type:      "connection_established",
				Data:      map[string]string{"message": "Connected to FleetTracker Pro"},
				Timestamp: time.Now(),
			}
			client.sendMessage(welcomeMsg)
			
			log.Printf("Client %s connected. Total clients: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			
			log.Printf("Client %s disconnected. Total clients: %d", client.ID, len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// startRedisPubSub starts Redis pub/sub for cross-instance communication
func (h *WebSocketHub) startRedisPubSub() {
	pubsub := h.redis.Subscribe(context.Background(), "fleet_tracker:websocket")
	defer pubsub.Close()
	
	ch := pubsub.Channel()
	for msg := range ch {
		h.broadcast <- []byte(msg.Payload)
	}
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
	// Get company ID and user ID from query parameters or headers
	companyID := c.Query("company_id")
	userID := c.Query("user_id")
	
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  h.config.ReadBufferSize,
		WriteBufferSize: h.config.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	
	// Create client
	client := &Client{
		ID:        fmt.Sprintf("%s_%s_%d", companyID, userID, time.Now().UnixNano()),
		CompanyID: companyID,
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h,
	}
	
	// Register client
	h.register <- client
	
	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *WebSocketHub) BroadcastMessage(message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	
	h.broadcast <- data
}

// BroadcastToCompany broadcasts a message to all clients of a specific company
func (h *WebSocketHub) BroadcastToCompany(companyID string, message WebSocketMessage) {
	message.CompanyID = companyID
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	
	h.mutex.RLock()
	for client := range h.clients {
		if client.CompanyID == companyID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// BroadcastToUser broadcasts a message to a specific user
func (h *WebSocketHub) BroadcastToUser(companyID, userID string, message WebSocketMessage) {
	message.CompanyID = companyID
	message.UserID = userID
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	
	h.mutex.RLock()
	for client := range h.clients {
		if client.CompanyID == companyID && client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// GetConnectedClients returns the number of connected clients
func (h *WebSocketHub) GetConnectedClients() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetCompanyClients returns the number of connected clients for a company
func (h *WebSocketHub) GetCompanyClients(companyID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	count := 0
	for client := range h.clients {
		if client.CompanyID == companyID {
			count++
		}
	}
	return count
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(c.Hub.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.config.PongWait))
		return nil
	})
	
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message for client %s: %v", c.ID, err)
		return
	}
	
	select {
	case c.Send <- data:
	default:
		close(c.Send)
	}
}

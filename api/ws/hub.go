package ws

import (
	"context"
	"encoding/json"
	"helios-api/db"
	"helios-api/models"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client represents a WebSocket client connection
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	tradingPair string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients, grouped by trading pair
	clients map[string]map[*Client]bool

	// Inbound messages from clients
	broadcast chan Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Database connection
	db *db.Database

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Message represents a broadcast message
type Message struct {
	TradingPair string
	Data        []byte
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// NewHub creates a new WebSocket hub
func NewHub(database *db.Database) *Hub {
	return &Hub{
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
		db:         database,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.tradingPair] == nil {
				h.clients[client.tradingPair] = make(map[*Client]bool)
			}
			h.clients[client.tradingPair][client] = true
			h.mu.Unlock()
			log.Printf("✅ Client registered for pair: %s (Total: %d)", 
				client.tradingPair, len(h.clients[client.tradingPair]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.tradingPair]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					log.Printf("❌ Client unregistered from pair: %s (Remaining: %d)", 
						client.tradingPair, len(h.clients[client.tradingPair]))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.TradingPair]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- message.Data:
				default:
					// Client's send channel is full, unregister it
					h.mu.Lock()
					close(client.send)
					delete(h.clients[message.TradingPair], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

// BroadcastOrderBook fetches and broadcasts the order book for a trading pair
func (h *Hub) BroadcastOrderBook(ctx context.Context, tradingPairID int, tradingPairSymbol string) {
	orderBook, err := h.db.GetOrderBook(ctx, tradingPairID)
	if err != nil {
		log.Printf("❌ Failed to fetch order book for pair %s: %v", tradingPairSymbol, err)
		return
	}

	wsMessage := models.WSMessage{
		Type: "orderbook",
		Data: orderBook,
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("❌ Failed to marshal order book: %v", err)
		return
	}

	h.broadcast <- Message{
		TradingPair: tradingPairSymbol,
		Data:        data,
	}

	log.Printf("📡 Broadcast order book update for %s to %d clients", 
		tradingPairSymbol, len(h.clients[tradingPairSymbol]))
}

// ListenForNotifications listens for PostgreSQL NOTIFY events
func (h *Hub) ListenForNotifications(ctx context.Context, pool *pgxpool.Pool) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to acquire connection for LISTEN: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN market_update")
	if err != nil {
		log.Fatalf("❌ Failed to LISTEN on market_update: %v", err)
	}

	log.Println("🔔 Started listening for market updates from PostgreSQL")

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("⚠️ Context cancelled, stopping LISTEN goroutine")
				return
			}
			log.Printf("❌ Error waiting for notification: %v", err)
			continue
		}

		// Parse notification payload
		var update models.MarketUpdate
		if err := json.Unmarshal([]byte(notification.Payload), &update); err != nil {
			log.Printf("❌ Failed to parse notification payload: %v", err)
			continue
		}

		log.Printf("🔔 Received market update for pair ID: %d", update.PairID)

		// Get trading pair by ID
		pair, err := h.db.GetTradingPairByID(ctx, update.PairID)
		if err != nil {
			log.Printf("❌ Failed to get trading pair: %v", err)
			continue
		}

		// Broadcast updated order book
		go h.BroadcastOrderBook(ctx, update.PairID, pair.Symbol)
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles WebSocket requests from clients
func ServeWS(hub *Hub, tradingPair string, conn *websocket.Conn) {
	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		tradingPair: tradingPair,
	}

	client.hub.register <- client

	// Send initial order book
	ctx := context.Background()
	tradingPairID, err := hub.db.GetTradingPairIDBySymbol(ctx, tradingPair)
	if err == nil {
		go hub.BroadcastOrderBook(ctx, tradingPairID, tradingPair)
	}

	// Start pumps
	go client.writePump()
	go client.readPump()
}

package ws

import (
	"helios-api/models"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	// "github.com/gorilla/websocket"
)

// WSHandler handles WebSocket endpoint
type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
// GET /ws/v1/market/:pair
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	pair := c.Param("pair")
	// Wildcard params include leading slash, remove it
	if len(pair) > 0 && pair[0] == '/' {
		pair = pair[1:]
	}
	log.Printf("🔍 WebSocket connection attempt - Raw pair param: '%s'", pair)
	
	// URL decode the pair (Gin should do this automatically, but let's be explicit)
	decodedPair, err := url.QueryUnescape(pair)
	if err != nil {
		log.Printf("❌ Failed to decode trading pair: %v", err)
		decodedPair = pair // Fall back to original
	}
	
	log.Printf("🔍 WebSocket decoded pair: '%s'", decodedPair)
	
	if decodedPair == "" {
		log.Printf("❌ WebSocket rejected - Empty trading pair")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Trading pair is required",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade WebSocket: %v", err)
		log.Printf("   Request URL: %s", c.Request.URL.String())
		log.Printf("   Request Headers: %v", c.Request.Header)
		return
	}


	log.Printf("✅ WebSocket upgraded successfully for pair: %s", decodedPair)

	// Serve the WebSocket connection
	ServeWS(h.hub, decodedPair, conn)
}

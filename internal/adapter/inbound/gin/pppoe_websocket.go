package gin_inbound_adapter

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func (h *pppoeAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Use context from request to handle cancellation
	ctx := c.Request.Context()

	// Subscribe to domain events
	// Note: This creates a new Redis subscription for each client.
	eventChan, err := h.domain.Pppoe().SubscribeToEvents(ctx)
	if err != nil {
		log.Printf("Failed to subscribe to events: %v", err)
		return
	}

	// Loop to send events to client
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-eventChan:
			if !ok {
				return
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

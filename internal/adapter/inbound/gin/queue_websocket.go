package gin_inbound_adapter

import (
	"log"

	"github.com/gin-gonic/gin"
)

func (h *queueAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	eventChan, err := h.domain.Queue().SubscribeToStats(ctx)
	if err != nil {
		log.Printf("Failed to subscribe to queue stats: %v", err)
		return
	}

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

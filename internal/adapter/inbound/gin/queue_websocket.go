package gin_inbound_adapter

import (
	"log"

	"github.com/gin-gonic/gin"

	"go-template/internal/model"
)

func (h *queueAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	// Check if filtering by queue name
	queueName := c.Query("name")

	var eventChan <-chan model.WebSocketMessage

	if queueName != "" {
		// Subscribe to specific queue stats
		eventChan, err = h.domain.Queue().SubscribeToStatsByName(ctx, queueName)
		if err != nil {
			log.Printf("Failed to subscribe to queue stats for %s: %v", queueName, err)
			return
		}
	} else {
		// Subscribe to all queue stats
		eventChan, err = h.domain.Queue().SubscribeToStats(ctx)
		if err != nil {
			log.Printf("Failed to subscribe to queue stats: %v", err)
			return
		}
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

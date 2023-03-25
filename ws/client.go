package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"nerajima.com/NeraJima/models"
)

type client struct {
	ConnectionId uuid.UUID // This allows us to distinguish the connections associated to a single user because one user can connect from multiple devices meaning one user can have multiple connections. This id helps us differentiate them
	Conn         *websocket.Conn
	Message      chan *Message
	Profile      models.Profile
	mu           sync.Mutex
}

type Message struct {
	BelongsTo []string  `json:"belongs_to"`
	Body      string    `json:"body"`
	Received  time.Time `json:"received"`
}

func (c *client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok { // if no message was received
			return
		}

		c.Conn.WriteJSON(message)
	}
}

func (c *client) readMessage(h *Hub) {
	defer func() {
		h.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(m, &msg); err != nil {
			c.Conn.WriteJSON(&fiber.Map{"error": "Malformed data..."})
			continue
		}

		if len(msg.BelongsTo) > 0 {
			h.NewBroadcast(&msg)
		} else {
			c.Conn.WriteJSON(&fiber.Map{"error": "Please include belongs_to: [...ids]"})
		}
	}
}

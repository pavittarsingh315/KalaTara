package ws

import (
	"encoding/json"
	"errors"
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
	To       []string  `json:"to,omitempty"`
	From     string    `json:"from"`
	ChatId   string    `json:"chat_id"`
	Body     string    `json:"body"`
	Received time.Time `json:"received"`
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

		message.To = []string{} // make empty to omit in response

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

		if err := msg.Format(c); err != nil {
			c.Conn.WriteJSON(&fiber.Map{"error": err.Error()})
			continue
		}

		h.NewBroadcast(&msg)
	}
}

func (m *Message) Format(c *client) error {
	m.From = c.Profile.UserId
	m.Received = time.Now()
	if len(m.To) <= 0 {
		return errors.New("to not provided")
	}
	if m.ChatId == "" {
		return errors.New("chat id not provided")
	}
	return nil
}

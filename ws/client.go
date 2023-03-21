package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"nerajima.com/NeraJima/models"
)

type client struct {
	ConnectionId uuid.UUID // This allows us to distinguish the connections associated to a single user because one user can connect from multiple devices meaning one user can have multiple connections. This id helps us differentiate them
	Conn         *websocket.Conn
	Message      chan *fiber.Map
	Profile      models.Profile
}

func (c *client) writeMessage(wg *sync.WaitGroup) {
	defer func() {
		c.Conn.Close()
		wg.Done()
	}()

	for {
		message, ok := <-c.Message
		if !ok { // if no message was received
			return
		}

		c.Conn.WriteJSON(message)
	}
}

func (c *client) readMessage(h *Hub, wg *sync.WaitGroup) {
	defer func() {
		h.unregister <- c
		c.Conn.Close()
		wg.Done()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var data fiber.Map
		if err := json.Unmarshal(m, &data); err != nil {
			log.Printf("error: %v", err)
			break
		}

		h.broadcast <- &data
	}
}

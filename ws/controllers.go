package ws

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"nerajima.com/NeraJima/models"
)

// Connect client to ws hub
func (h *Hub) Connect(c *websocket.Conn) {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	h.mu.Lock()
	numDevices := len(h.clients[reqProfile.UserId])
	h.mu.Unlock()
	if numDevices >= maxNumberOfDevices {
		c.WriteJSON(&fiber.Map{"error": "You have reached the maximum number of devices you can connect to the server from. Please disconnect from one of your other devices and try again"})
		c.Close()
		return
	}

	cl := &client{
		ConnectionId: uuid.New(),
		Conn:         c,
		Message:      make(chan *Message, 10), // channel is buffered with capacity = 10
		Profile:      reqProfile,
	}

	h.register <- cl

	go cl.writeMessage()
	cl.readMessage(h) // we don't run this in a go routine cause fiber spawns a goroutine for each request therefore this func runs in the goroutine spawned by fiber for each instance of the request
}

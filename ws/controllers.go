package ws

import (
	"github.com/gofiber/websocket/v2"
	"nerajima.com/NeraJima/models"
)

// Connect client to ws hub
func (h *Hub) Connect(c *websocket.Conn) {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	cl := &client{
		Conn:    c,
		Message: make(chan string, 10), // channel is buffered with capacity = 10
		Profile: reqProfile,
	}

	h.register <- cl

	go cl.writeMessage()
	cl.readMessage(h)
}

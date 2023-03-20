package ws

import (
	"strconv"

	"github.com/gofiber/websocket/v2"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/utils"
)

// Connect client to ws hub
func (h *Hub) Connect(c *websocket.Conn) {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var cid int
	for {
		id, _ := strconv.Atoi(utils.GenerateRandomCode(4))
		if _, exists := h.Clients[reqProfile.UserId][id]; !exists {
			cid = id
			break
		}
	}

	cl := &client{
		ConnectionId: cid,
		Conn:         c,
		Message:      make(chan string, 10), // channel is buffered with capacity = 10
		Profile:      reqProfile,
	}

	h.register <- cl

	go cl.writeMessage()
	cl.readMessage(h)
}

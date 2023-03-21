package ws

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"nerajima.com/NeraJima/models"
)

// Connect client to ws hub
func (h *Hub) Connect(c *websocket.Conn) {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	cl := &client{
		ConnectionId: uuid.New(),
		Conn:         c,
		Message:      make(chan *fiber.Map, 10), // channel is buffered with capacity = 10
		Profile:      reqProfile,
	}

	h.register <- cl

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go cl.writeMessage(wg)
	go cl.readMessage(h, wg)
	wg.Wait()
}

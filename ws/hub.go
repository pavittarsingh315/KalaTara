package ws

import "github.com/google/uuid"

type Hub struct {
	// Map of all the connected clients via websockets
	//
	// Key: user id
	//
	// Value: list of clients associated with one user i.e. a user connected to the server via multiple devices
	Clients    map[string]map[uuid.UUID]*client
	register   chan *client
	unregister chan *client
	broadcast  chan string
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]map[uuid.UUID]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan string, 10), // channel is buffered with capacity = 10
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.register:
			if _, exists := h.Clients[cl.Profile.UserId]; !exists { // if client is not already in Clients
				h.Clients[cl.Profile.UserId] = map[uuid.UUID]*client{cl.ConnectionId: cl}
			} else {
				h.Clients[cl.Profile.UserId][cl.ConnectionId] = cl
			}
		case cl := <-h.unregister:
			delete(h.Clients[cl.Profile.UserId], cl.ConnectionId)
			close(cl.Message)
		case message := <-h.broadcast:
			for _, client := range h.Clients {
				for _, cl := range client {
					cl.Message <- message
				}
			}
		}
	}
}

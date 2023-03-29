package ws

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	// Map of all the connected clients via websockets
	//
	// Key: user id
	//
	// Value: list of clients associated with one user i.e. a user connected to the server via multiple devices
	clients    map[string]map[uuid.UUID]*client
	register   chan *client
	unregister chan *client
	broadcast  chan *Message
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[uuid.UUID]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan *Message, 10), // channel is buffered with capacity = 10
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.register:
			h.mu.Lock()
			if _, exists := h.clients[cl.Profile.UserId]; !exists { // if client is not already in Clients
				h.clients[cl.Profile.UserId] = map[uuid.UUID]*client{cl.ConnectionId: cl}
			} else {
				h.clients[cl.Profile.UserId][cl.ConnectionId] = cl
			}
			h.mu.Unlock()
		case cl := <-h.unregister:
			cl.mu.Lock()
			h.mu.Lock()
			if len(h.clients[cl.Profile.UserId]) == 1 {
				delete(h.clients, cl.Profile.UserId)
			} else {
				delete(h.clients[cl.Profile.UserId], cl.ConnectionId)
			}
			h.mu.Unlock()
			close(cl.Message)
			cl.mu.Unlock()
		case msg := <-h.broadcast:
			go func(m *Message) { // run in parallel so we don't block on messages belonging to many users
				for _, id := range m.To {
					h.mu.Lock()
					clients, ok := h.clients[id]
					h.mu.Unlock()
					if !ok {
						continue
					}

					for _, client := range clients {
						client.mu.Lock()
						// Check if client is still here because its possible that a client unregisters as message is being broadcasted to them but since they unregistered and their message chan closed, it'll cause a panic.
						// This prevents that case.
						h.mu.Lock()
						_, exists := h.clients[id][client.ConnectionId]
						h.mu.Unlock()
						if !exists {
							client.mu.Unlock()
							continue
						}
						client.Message <- m
						client.mu.Unlock() // we don't defer this because then it would only run when the entire anonymous goroutine returns
					}
				}
			}(msg)
		}
	}
}

func (h *Hub) NewBroadcast(msg *Message) {
	h.broadcast <- msg
}

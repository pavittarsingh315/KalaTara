package ws

type Hub struct {
	// Map of all the connected clients via websockets
	//
	// Key: user id
	Clients    map[string]*client
	register   chan *client
	unregister chan *client
	broadcast  chan string
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan string, 10), // channel is buffered with capacity = 10
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, exists := h.Clients[client.Profile.UserId]; !exists { // if client is not already in Clients
				h.Clients[client.Profile.UserId] = client
			}
		case client := <-h.unregister:
			delete(h.Clients, client.Profile.UserId)
			close(client.Message)
		case message := <-h.broadcast:
			for _, client := range h.Clients {
				client.Message <- message
			}
		}
	}
}

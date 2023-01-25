package domain

type Hub struct {
	Clients      map[*Client]bool
	RegisterCh   chan *Client
	UnregisterCh chan *Client
	BroadcastCh  chan []byte
	Rooms        map[*Room]bool
}

func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[*Client]bool),
		RegisterCh:   make(chan *Client),
		UnregisterCh: make(chan *Client),
		BroadcastCh:  make(chan []byte),
		Rooms:        make(map[*Room]bool),
	}
}

func (h *Hub) RunLoop() {
	for {
		select {
		case client := <-h.RegisterCh:
			h.register(client)

		case client := <-h.UnregisterCh:
			h.unregister(client)

		case msg := <-h.BroadcastCh:
			h.broadcastToAllClient(msg)
		}
	}

}

func (h *Hub) register(c *Client) {
	h.Clients[c] = true
}

func (h *Hub) unregister(c *Client) {
	delete(h.Clients, c)
}

func (h *Hub) broadcastToAllClient(msg []byte) {
	for c := range h.Clients {
		c.sendCh <- msg
	}
}

func (h *Hub) NotiftyClientJoined(client *Client) {
	message := &Message{
		Action: UserJoinedAction,
		Sender: client,
	}
	h.broadcastToAllClient(message.Encode())
}

func (h *Hub) NotifyClientLeft(client *Client) {
	message := &Message{
		Action: UserLeftAction,
		Sender: client,
	}
	h.broadcastToAllClient(message.Encode())
}

func (h *Hub) ListOnlineClients(client *Client) {
	for existingClient := range h.Clients {
		message := &Message{
			Action: UserJoinedAction,
			Sender: existingClient,
		}
		client.sendCh <- message.Encode()
	}
}

func (h *Hub) FindRoomByID(ID string) *Room {
	var foundRoom *Room
	for room := range h.Rooms {
		if room.GetId() == ID {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (h *Hub) CreateRoom(name string, private bool) *Room {
	room := NewRoom(name, private)
	go room.RunRoom()
	h.Rooms[room] = true
	return room
}

func (h *Hub) FindClientByID(ID string) *Client {
	var foundClient *Client
	for client := range h.Clients {
		if client.ID.String() == ID {
			foundClient = client
			break
		}
	}
	return foundClient
}

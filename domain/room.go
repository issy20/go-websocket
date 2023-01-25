package domain

import (
	"fmt"

	"github.com/google/uuid"
)

const welcomeMessage = "%s joined the room"

type Room struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Clients      map[*Client]bool
	RegisterCh   chan *Client
	UnregisterCh chan *Client
	BroadcastCh  chan *Message
	Private      bool `json:"private"`
}

func NewRoom(name string, private bool) *Room {
	return &Room{
		ID:           uuid.New(),
		Name:         name,
		Clients:      make(map[*Client]bool),
		RegisterCh:   make(chan *Client),
		UnregisterCh: make(chan *Client),
		BroadcastCh:  make(chan *Message),
		Private:      private,
	}
}

func (room *Room) RunRoom() {
	for {
		select {
		case client := <-room.RegisterCh:
			room.RegisterClientInRoom(client)
		case client := <-room.UnregisterCh:
			room.UnregisterClientInRoom(client)
		case message := <-room.BroadcastCh:
			room.BroadcastToClientsInRoom(message.Encode())
		}
	}
}

func (room *Room) RegisterClientInRoom(client *Client) {
	if !room.Private {
		room.NotiftyClientJoined(client)
	}
	room.Clients[client] = true
}

func (room *Room) UnregisterClientInRoom(client *Client) {
	if _, ok := room.Clients[client]; ok {
		delete(room.Clients, client)
	}
}

func (room *Room) BroadcastToClientsInRoom(message []byte) {
	for client := range room.Clients {
		client.sendCh <- message
	}
}

func (room *Room) NotiftyClientJoined(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  room,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}
	room.BroadcastToClientsInRoom(message.Encode())
}

func (room *Room) GetId() string {
	return room.ID.String()
}

func (room *Room) GetName() string {
	return room.Name
}

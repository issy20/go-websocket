package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/issy20/go-websocket/config"
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

func (r *Room) RunRoom() {
	go r.SubscribeToRoomMessages()

	for {
		select {
		case client := <-r.RegisterCh:
			r.RegisterClientInRoom(client)
		case client := <-r.UnregisterCh:
			r.UnregisterClientInRoom(client)
		case message := <-r.BroadcastCh:
			r.BroadcastToClientsInRoom(message.Encode())
		}
	}
}

func (r *Room) RegisterClientInRoom(client *Client) {
	if !r.Private {
		r.NotiftyClientJoined(client)
	}
	r.Clients[client] = true
}

func (r *Room) UnregisterClientInRoom(client *Client) {
	delete(r.Clients, client)
}

func (r *Room) BroadcastToClientsInRoom(message []byte) {
	for client := range r.Clients {
		client.sendCh <- message
	}
}

func (r *Room) PublishRoomMessage(message []byte) {
	err := config.Redis.Publish(ctx, r.GetName(), message).Err()
	if err != nil {
		log.Println(err)
	}
}

func (r *Room) SubscribeToRoomMessages() {
	pubsub := config.Redis.Subscribe(ctx, r.GetName())
	ch := pubsub.Channel()
	for msg := range ch {
		r.BroadcastToClientsInRoom([]byte(msg.Payload))
	}
}

func (r *Room) NotiftyClientJoined(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  r,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}
	r.PublishRoomMessage(message.Encode())
}

func (r *Room) GetId() string {
	return r.ID.String()
}

func (r *Room) GetName() string {
	return r.Name
}

func (r *Room) GetPrivate() bool {
	return r.Private
}

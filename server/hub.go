package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/issy20/go-websocket/config"
	"github.com/issy20/go-websocket/models"
)

const PubSubGeneralChannel = "general"

type Hub struct {
	Users        []models.IUser
	Clients      map[*Client]bool
	RegisterCh   chan *Client
	UnregisterCh chan *Client
	// BroadcastCh    chan []byte
	Rooms          map[*Room]bool
	RoomRepository models.RoomRepository
	UserRepository models.UserRepository
}

func NewHub(roomRepository models.RoomRepository, userRepository models.UserRepository) *Hub {
	hub := &Hub{
		Clients:      make(map[*Client]bool),
		RegisterCh:   make(chan *Client),
		UnregisterCh: make(chan *Client),
		// BroadcastCh:    make(chan []byte),
		Rooms:          make(map[*Room]bool),
		RoomRepository: roomRepository,
		UserRepository: userRepository,
	}
	hub.Users = userRepository.GetAllUsers()
	return hub
}

func (h *Hub) RunLoop() {

	go h.ListenPubSubChannel()
	for {
		select {
		case client := <-h.RegisterCh:
			h.RegisterClient(client)

		case client := <-h.UnregisterCh:
			h.UnregisterClient(client)
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	if user := h.FindUserByID(client.ID.String()); user == nil {
		log.Println("cannot register client")
	}

	h.PublishClientJoined(client)
	h.ListOnlineClients(client)
	h.Clients[client] = true
}

func (h *Hub) UnregisterClient(client *Client) {
	log.Print("UnregisterClient()")
	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		h.PublishClientLeft(client)
	}
}

func (h *Hub) PublishClientJoined(client *Client) {
	message := &Message{
		Action: UserJoinedAction,
		Sender: client,
	}
	if err := config.Redis.Publish(ctx, PubSubGeneralChannel, message.Encode()).Err(); err != nil {
		log.Println(err, "PublishClientJoined()")
	}
}

func (h *Hub) PublishClientLeft(client *Client) {
	message := &Message{
		Action: UserLeftAction,
		Sender: client,
	}
	if err := config.Redis.Publish(ctx, PubSubGeneralChannel, message.Encode()).Err(); err != nil {
		log.Println(err, "PublishClientLeft()")
	}
}

func (h *Hub) ListenPubSubChannel() {
	pubsub := config.Redis.Subscribe(ctx, PubSubGeneralChannel)
	ch := pubsub.Channel()

	log.Print(&ch)

	for msg := range ch {
		var message Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("ListenPubSubChannel: Error on unmarshal JSON message %s", err)
			return
		}

		switch message.Action {
		case UserJoinedAction:
			h.HandleUserJoined(message)
		case UserLeftAction:
			h.HandleUserLeft(message)
		case JoinRoomPrivateAction:
			h.HandleUserJoinPrivate(message)
		}
	}
}

func (h *Hub) HandleUserJoined(message Message) {
	h.Users = append(h.Users, message.Sender)
	log.Print("hub.go", h)
	h.broadcastToAllClient(message.Encode())
}

func (h *Hub) HandleUserLeft(message Message) {
	log.Print("HandleUserLeft")
	for i, user := range h.Users {
		if user.GetId() == message.Sender.GetId() {
			h.Users[i] = h.Users[len(h.Users)-1]
			h.Users = h.Users[:len(h.Users)-1]
			break
		}
	}
	h.broadcastToAllClient(message.Encode())
}

// messageのidとは？
func (h *Hub) HandleUserJoinPrivate(message Message) {
	targetClients := h.FindClientsByID(message.Message)
	for _, targetClient := range targetClients {
		targetClient.JoinRoom(message.Target.GetName(), message.Sender)
	}
}

func (h *Hub) ListOnlineClients(client *Client) {
	var uniqueUsers = make(map[string]bool)
	fmt.Print(uniqueUsers)
	for _, user := range h.Users {
		if ok := uniqueUsers[user.GetId()]; !ok {
			message := &Message{
				Action: UserJoinedAction,
				Sender: user,
			}
			uniqueUsers[user.GetId()] = true
			client.sendCh <- message.Encode()
		}
	}
}

func (h *Hub) broadcastToAllClient(msg []byte) {
	for c := range h.Clients {
		c.sendCh <- msg
	}
}

func (h *Hub) FindRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range h.Rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}
	if foundRoom == nil {
		foundRoom = h.RunRoomFromRepository(name)
	}
	return foundRoom
}

func (h *Hub) RunRoomFromRepository(name string) *Room {
	var room *Room
	dbRoom := h.RoomRepository.FindRoomByName(name)
	if dbRoom != nil {
		room = NewRoom(dbRoom.GetName(), dbRoom.GetPrivate())
		room.ID, _ = uuid.Parse(dbRoom.GetId())
		go room.RunRoom()
		h.Rooms[room] = true
	}
	return room
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
	h.RoomRepository.AddRoom(room)
	go room.RunRoom()
	h.Rooms[room] = true
	return room
}

func (h *Hub) FindUserByID(ID string) models.IUser {
	var foundUser models.IUser
	for _, client := range h.Users {
		if client.GetId() == ID {
			foundUser = client
			break
		}
	}
	return foundUser
}

func (h *Hub) FindClientsByID(ID string) []*Client {
	var foundClients []*Client
	for client := range h.Clients {
		if client.ID.String() == ID {
			foundClients = append(foundClients, client)

		}
	}
	return foundClients
}

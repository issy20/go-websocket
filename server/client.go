package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/issy20/go-websocket/auth"
	"github.com/issy20/go-websocket/config"
	"github.com/issy20/go-websocket/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	// space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ws     *websocket.Conn
	hub    *Hub
	sendCh chan []byte
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Rooms  map[*Room]bool
}

func NewClient(ws *websocket.Conn, hub *Hub, name string, ID string) *Client {
	client := &Client{
		Name:   name,
		ws:     ws,
		hub:    hub,
		sendCh: make(chan []byte),
		Rooms:  make(map[*Room]bool),
	}
	if ID != "" {
		client.ID, _ = uuid.Parse(ID)
	}
	return client
}

func (c *Client) ReadLoop() {
	defer func() {
		c.Disconnect()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, jsonMsg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		c.HandleNewMessage(jsonMsg)
	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.sendCh:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.sendCh)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.sendCh)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Disconnect() {
	c.hub.UnregisterCh <- c
	log.Print("disconnect", c.hub.UnregisterCh)
	for room := range c.Rooms {
		room.UnregisterCh <- c
	}
	close(c.sendCh)
	c.ws.Close()
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	userCtxValue := r.Context().Value(auth.UserContextKey)
	if userCtxValue == nil {
		log.Println("Not authenticated")
		return
	}

	user := userCtxValue.(models.IUser)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, hub, user.GetName(), user.GetId())

	go client.WriteLoop()
	go client.ReadLoop()

	hub.RegisterCh <- client

}

func (c *Client) HandleNewMessage(jsonMessage []byte) {
	var m Message
	log.Print(jsonMessage)
	if err := json.Unmarshal(jsonMessage, &m); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	m.Sender = c

	switch m.Action {
	case SendMessageAction:
		roomID := m.Target.GetId()
		if room := c.hub.FindRoomByID(roomID); room != nil {
			room.BroadcastCh <- &m
		}
	case JoinRoomAction:
		c.HandleJoinRoomMessage(m)
	case LeaveRoomAction:
		c.HandleLeaveRoomMessage(m)
	case JoinRoomPrivateAction:
		c.HandleJoinRoomPrivateMessage(m)
	}
}

func (c *Client) HandleJoinRoomMessage(message Message) {
	roomName := message.Message
	c.JoinRoom(roomName, nil)
}

func (c *Client) HandleLeaveRoomMessage(message Message) {
	room := c.hub.FindRoomByID(message.Message)
	if room == nil {
		return
	}
	if _, ok := c.Rooms[room]; ok {
		delete(c.Rooms, room)
	}
	room.UnregisterCh <- c
}

func (c *Client) HandleJoinRoomPrivateMessage(message Message) {
	target := c.hub.FindUserByID(message.Message)
	if target == nil {
		return
	}

	roomName := message.Message + c.ID.String()

	joinedRoom := c.JoinRoom(roomName, target)
	if joinedRoom != nil {
		c.inviteTargetUser(target, joinedRoom)
	}
}

func (c *Client) JoinRoom(roomName string, sender models.IUser) *Room {
	room := c.hub.FindRoomByID(roomName)
	if room == nil {
		room = c.hub.CreateRoom(roomName, sender != nil)
	}
	if sender == nil && room.Private {
		return nil
	}
	if !c.IsInRoom(room) {
		c.Rooms[room] = true
		room.RegisterCh <- c
		c.NotifyRoomJoined(room, sender)
	}
	return room
}

func (client *Client) IsInRoom(room *Room) bool {
	if _, ok := client.Rooms[room]; ok {
		return true
	}
	return false
}

func (c *Client) inviteTargetUser(target models.IUser, room *Room) {
	inviteMessage := &Message{
		Action:  JoinRoomPrivateAction,
		Message: target.GetId(),
		Target:  room,
		Sender:  c,
	}

	if err := config.Redis.Publish(ctx, PubSubGeneralChannel, inviteMessage.Encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (client *Client) NotifyRoomJoined(room *Room, sender models.IUser) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}
	client.sendCh <- message.Encode()
}

func (c *Client) GetId() string {
	return c.ID.String()
}

func (client *Client) GetName() string {
	return client.Name
}

package domain

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Client struct {
	ws     *websocket.Conn
	sendCh chan []byte
	rooms  map[*Room]bool
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Rooms  map[*Room]bool
	hub    *Hub
}

func NewClient(ws *websocket.Conn, hub *Hub, name string) *Client {
	return &Client{
		ws:     ws,
		sendCh: make(chan []byte),
		ID:     uuid.New(),
		Name:   name,
		Rooms:  make(map[*Room]bool),
	}
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

func (c *Client) Disconnect() {
	c.hub.UnregisterCh <- c
	for room := range c.rooms {
		room.UnregisterCh <- c
	}
	close(c.sendCh)
	c.ws.Close()
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, hub, name[0])

	go client.WriteLoop()
	go client.ReadLoop()

	hub.RegisterCh <- client

}

func (client *Client) HandleNewMessage(jsonMessage []byte) {
	var m Message

	if err := json.Unmarshal(jsonMessage, &m); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	m.Sender = client

	switch m.Action {
	case SendMessageAction:
		roomID := m.Target.GetId()
		if room := client.hub.FindRoomByID(roomID); room != nil {
			room.BroadcastCh <- &m
		}
	case JoinRoomAction:
		client.HandleJoinRoomMessage(m)
	case LeaveRoomAction:
		client.HandleLeaveRoomMessage(m)
	case JoinRoomPrivateAction:
		client.HandleJoinRoomPrivateMessage(m)
	}
}

func (client *Client) HandleJoinRoomMessage(message Message) {
	roomName := message.Message
	client.JoinRoom(roomName, nil)
}

func (client *Client) HandleLeaveRoomMessage(message Message) {
	room := client.hub.FindRoomByID(message.Message)
	if room == nil {
		return
	}
	if _, ok := client.Rooms[room]; ok {
		delete(client.Rooms, room)
	}
	room.UnregisterCh <- client
}

func (client *Client) HandleJoinRoomPrivateMessage(message Message) {
	target := client.hub.FindClientByID(message.Message)
	if target == nil {
		return
	}

	roomName := message.Message + client.ID.String()

	client.JoinRoom(roomName, target)
	target.JoinRoom(roomName, client)
}

func (client *Client) JoinRoom(roomName string, sender *Client) {
	room := client.hub.FindRoomByID(roomName)
	if room == nil {
		room = client.hub.CreateRoom(roomName, sender != nil)
	}
	if sender == nil && room.Private {
		return
	}
	if !client.IsInRoom(room) {
		client.Rooms[room] = true
		room.RegisterCh <- client
		client.NotifyRoomJoined(room, sender)
	}
}

func (client *Client) IsInRoom(room *Room) bool {
	if _, ok := client.Rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) NotifyRoomJoined(room *Room, sender *Client) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}
	client.sendCh <- message.Encode()
}

func (client *Client) GetName() string {
	return client.Name
}

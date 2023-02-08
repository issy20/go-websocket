package main

import (
	"log"

	"github.com/issy20/go-websocket/models"
	jsoniter "github.com/json-iterator/go"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Message struct {
	Action  string       `json:"action"`
	Message string       `json:"message"`
	Target  *Room        `json:"target"`
	Sender  models.IUser `json:"sender"`
}

func (message *Message) Encode() []byte {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	j, err := json.Marshal(&message)
	if err != nil {
		log.Println(err)
	}

	return j
}

func (message *Message) UnMarshalJSON(data []byte) error {
	log.Print(data)

	type Alias Message
	msg := &struct {
		Sender Client `json:"sender"`
		*Alias
	}{
		Alias: (*Alias)(message),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	message.Sender = &msg.Sender
	return nil
}

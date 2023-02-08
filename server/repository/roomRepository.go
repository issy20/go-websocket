package repository

import (
	"database/sql"

	"github.com/issy20/go-websocket/models"
)

type Room struct {
	Id      string
	Name    string
	Private bool
}

func (r *Room) GetId() string {
	return r.Id
}

func (r *Room) GetName() string {
	return r.Name
}

func (r *Room) GetPrivate() bool {
	return r.Private
}

type RoomRepository struct {
	Db *sql.DB
}

func (rr *RoomRepository) AddRoom(room models.Room) {
	stmt, err := rr.Db.Prepare("INSERT INTO rooms(id, name, private) values (?, ?, ?)")
	checkErr(err)

	_, err = stmt.Exec(room.GetId(), room.GetName(), room.GetPrivate())
	checkErr(err)
}

func (rr *RoomRepository) FindRoomByName(name string) models.Room {
	row := rr.Db.QueryRow("SELECT id, name, private FROM rooms where name = ? LIMIT 1", name)
	var room Room

	if err := row.Scan(&room.Id, &room.Name, &room.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	return &room
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

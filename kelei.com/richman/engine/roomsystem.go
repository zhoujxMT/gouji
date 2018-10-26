package engine

import (
	"fmt"

	"github.com/satori/go.uuid"
)

type RoomSystem struct {
	rooms map[string]*Room
}

func NewRoomSystem() *RoomSystem {
	roomSystem := RoomSystem{}
	roomSystem.rooms = make(map[string]*Room)
	return &roomSystem
}

func (this *RoomSystem) AddRoom() *Room {
	room := NewRoom()
	roomid := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
	room.setRoomID(roomid)
	this.rooms[roomid] = room
	return room
}

func (this *RoomSystem) Clear() {
	this.rooms = make(map[string]*Room)
}

func (this *RoomSystem) GetRooms() map[string]*Room {
	return this.rooms
}

func (this *RoomSystem) GetRoomCount() int {
	return len(this.GetRooms())
}

func (this *RoomSystem) matchingRoom() *Room {
	var room *Room
	for _, r := range this.GetRooms() {
		if len(r.GetUsers()) < r.GetPCount() {
			room = r
			break
		}
	}
	if room == nil {
		room = this.AddRoom()
	}
	return room
}

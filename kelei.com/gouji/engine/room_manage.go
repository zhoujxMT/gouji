/*
房间管理
*/

package engine

import (
	"strconv"
	"sync"
	"time"
)

type roomManage struct {
	rooms map[string]*Room //房间列表
	lock  sync.Mutex
}

func ctor() *roomManage {
	RoomManage := roomManage{}
	RoomManage.rooms = make(map[string]*Room)
	return &RoomManage
}

var (
	RoomManage = *ctor()
	_          = RoomManage.Init()
)

//初始化
func (r *roomManage) Init() int {
	return 1
}

//重置房间
func (r *roomManage) Reset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rooms = make(map[string]*Room)
}

//裁判端定制
func (r *roomManage) Judgment() {
	r.createThreeEmptyRoom()
}

//生成三个空房间
func (r *roomManage) createThreeEmptyRoom() {
	//生成3个空房间
	for i := 0; i < 3; i++ {
		time.Sleep(time.Millisecond)
		r.AddRoomWithRoomID(strconv.Itoa(i + 1))
	}
}

//获取所有的房间
func (r *roomManage) GetRooms() map[string]*Room {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.rooms
}

//创建一个房间
func (r *roomManage) createRoom() *Room {
	room := Room{}
	cur := time.Now()
	timestamp := cur.UnixNano() / 1000000 //毫秒ranking
	room.id = strconv.FormatInt(timestamp, 10)
	room.cards = []Card{}
	room.idleusers = make(map[string]*User)
	room.users = make([]*User, pcount)
	room.ranking = make([]*User, pcount)
	room.users_cards = make(map[string]string, pcount)
	room.isrevolution = true
	room.istribute = false
	room.surplusBKingCount = 4
	room.surplusSKingCount = 4
	room.surplusTwoCount = 16
	room.configRoomByGameRule()
	return &room
}

//添加一个房间
func (r *roomManage) AddRoom() *Room {
	room := r.createRoom()
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rooms[room.id] = room
	return room
}

//添加一个房间
func (r *roomManage) AddRoomWithRoomID(roomid string) *Room {
	room := r.createRoom()
	room.SetRoomID(roomid)
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rooms[room.id] = room
	return room
}

//删除一个房间
func (r *roomManage) removeRoom(room *Room) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.rooms, room.id)
}

//获取房间
func (r *roomManage) GetRoom(roomid string) *Room {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.rooms[roomid]
}

//获取房间数量
func (r *roomManage) GetRoomCount() int {
	return len(r.GetRooms())
}

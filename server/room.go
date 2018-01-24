package main

import (
	"sync"
)

var (
	RoomsManager *Rooms
)

func NewRooms() *Rooms {
	return &Rooms{
		Rooms: map[string]*Room{},
		Mu:    sync.Mutex{},
	}
}

type Rooms struct {
	Rooms map[string]*Room
	Mu    sync.Mutex
}

func (rs *Rooms) Exist(name string) bool {
	_, ok := rs.Rooms[name]
	return ok
}

func (rs *Rooms) Join(name string, client *Client) *Room {
	rs.Mu.Lock()
	if !rs.Exist(name) {
		room := NewRoom(name, rs)
		go room.Run()
		rs.Rooms[name] = room
	}
	room := rs.Rooms[name]
	rs.Mu.Unlock()
	room.Add(client)
	return room
}

func NewRoom(name string, rooms *Rooms) *Room {
	return &Room{
		Name:          name,
		Clients:       sync.Map{},
		BroadcastChan: make(chan string),
		Rooms:         rooms,
	}
}

type Room struct {
	Name          string
	Rooms         *Rooms
	Clients       sync.Map
	BroadcastChan chan string
}

func (r *Room) Add(client *Client) {
	r.Clients.LoadOrStore(client.Name, client)
}

func (r *Room) Delete(name string) {
	r.Clients.Delete(name)
}

func (r *Room) Close() {

}

func (r *Room) Run() {
	for {
		msg, ok := <-r.BroadcastChan
		if !ok {
			break
		}
		r.Clients.Range(func(key interface{}, value interface{}) bool {
			client := value.(*Client)
			client.Out <- msg
			return true
		})
	}
}

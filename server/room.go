package main

import (
	"fmt"
	"log"
	"sync"
)

var (
	RoomsManager *Rooms = NewRooms()
)

func NewRooms() *Rooms {
	return &Rooms{
		Rooms: map[string]*Room{},
		Mu:    &sync.Mutex{},
	}
}

type Rooms struct {
	Rooms map[string]*Room
	Mu    *sync.Mutex
}

func (rs *Rooms) Join(name string, client *Client) {
	rs.Mu.Lock()
	client.Leave()
	_, exist := rs.Rooms[name]
	if !exist {
		room := NewRoom(name, rs)
		log.Printf("room: %s created\n", room.Name)
		go room.Run()
		rs.Rooms[name] = room
	}
	room := rs.Rooms[name]
	room.Clients[client.Name] = client
	client.Room = room
	rs.Mu.Unlock()
	log.Printf("name: %s joined room: %s\n", client.Name, room.Name)
}

func NewRoom(name string, rooms *Rooms) *Room {
	return &Room{
		Mu:               &sync.Mutex{},
		Name:             name,
		Clients:          map[string]*Client{},
		Rooms:            rooms,
		IsClosed:         make(chan struct{}),
		BroadcastChan:    make(chan *Message),
		RawBroadcastChan: make(chan string),
	}
}

type Room struct {
	Mu               *sync.Mutex
	Name             string
	Rooms            *Rooms
	Clients          map[string]*Client
	IsClosed         chan struct{}
	BroadcastChan    chan *Message
	RawBroadcastChan chan string
}

func (r *Room) HasClients() bool {
	return len(r.Clients) != 0
}

func (r *Room) Delete(name string) {
	r.Mu.Lock()
	delete(r.Clients, name)
	if !r.HasClients() {
		r.IsClosed <- struct{}{}
	}
	r.Mu.Unlock()
}

func (r *Room) Users() []string {
	r.Mu.Lock()
	names := make([]string, 0, len(r.Clients))
	for _, c := range r.Clients {
		names = append(names, c.Name)
	}
	r.Mu.Unlock()
	return names
}

func (r *Room) Run() {
	for {
		select {
		case <-r.IsClosed:
			log.Printf("room %s closed\n", r.Name)
			return
		case msg := <-r.BroadcastChan:
			for _, c := range r.Clients {
				if c.Name != msg.From.Name {
					c.Out <- fmt.Sprintf("%s says: %s", msg.From.Name, msg.Text)
				}
			}
		case msg := <-r.RawBroadcastChan:
			for _, c := range r.Clients {
				c.Out <- msg
			}
		}
	}
}

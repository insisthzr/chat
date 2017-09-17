package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	timeout = 30 * time.Second
)

var (
	entering = make(chan handler)
	leaving  = make(chan handler)
	messages = make(chan string)
)

func broadcaster() {
	handlers := make(map[int]handler)
	for {
		select {
		case msg := <-messages:
			for _, h := range handlers {
				h.out <- msg
			}
		case h := <-entering:
			handlers[h.user.Id] = h
			msg := "online clients:\n"
			for _, hand := range handlers {
				msg += fmt.Sprintf("- %s\n", hand.user.Name)
			}
			h.out <- msg
		case h := <-leaving:
			delete(handlers, h.user.Id)
			close(h.out)
			close(h.in)
		}
	}
}

type handler struct {
	conn net.Conn
	user *User
	in   chan string
	out  chan string
}

func (h *handler) handle() {
	go h.clientReader()
	go h.clientWriter()

	timer := time.NewTimer(timeout)
	h.out <- "input your name:"
LOGIN:
	for {
		select {
		case name := <-h.in:
			user := &User{Name: name}
			err := user.Login()
			if err != nil {
				h.out <- fmt.Sprintf("your name(%s) exists\npick another name", name)
			} else {
				h.user = user
				break LOGIN
			}
		case <-timer.C:
			h.conn.Close()
			return
		}
	}

	h.out <- "You are " + h.user.Name
	messages <- h.user.Name + " has arrived"
	entering <- *h

MSG:
	for {
		select {
		case msg := <-h.in:
			switch msg {
			case `\quit`:
				break MSG
			default:
				messages <- h.user.Name + ": " + msg
			}
			timer.Reset(timeout)
		case <-timer.C:
			break MSG
		}
	}

	h.out <- `\quit`
	h.user.Logout()
	leaving <- *h
	messages <- h.user.Name + " has left"
	log.Printf("client: %s disconnected", h.user.Name)
	h.conn.Close()
}

func (h *handler) clientWriter() {
	for msg := range h.out {
		fmt.Fprintln(h.conn, msg)
	}
}

func (h *handler) clientReader() {
	scanner := bufio.NewScanner(h.conn)
	for scanner.Scan() {
		h.in <- scanner.Text()
	}
}

func newHandler(conn net.Conn) *handler {
	h := &handler{
		conn: conn,
		in:   make(chan string),
		out:  make(chan string),
	}
	return h
}

var (
	port = flag.String("port", "20000", "port")
)

func main() {
	addr := "localhost:" + *port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening at", addr)
	go watchUsers()
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		handler := newHandler(conn)
		go handler.handle()
	}
}

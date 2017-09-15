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
	clients := make(map[handler]struct{})
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.out <- msg
			}
		case cli := <-entering:
			clients[cli] = struct{}{}
			msg := "online clients:\n"
			for c := range clients {
				msg += fmt.Sprintf("- %s\n", c.name)
			}
			cli.out <- msg
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.out)
			close(cli.in)
		}
	}
}

type handler struct {
	conn net.Conn
	name string
	in   chan string
	out  chan string
}

func (h *handler) handle() {
	go h.clientReader()
	go h.clientWriter()

	timer := time.NewTimer(timeout)
	h.out <- "input your name:"
	select {
	case name := <-h.in:
		h.name = name
	case <-timer.C:
		h.conn.Close()
		return
	}

	h.out <- "You are " + h.name
	messages <- h.name + " has arrived"
	entering <- *h

LOOP:
	for {
		select {
		case msg := <-h.in:
			messages <- h.name + ": " + msg
			timer.Reset(timeout)
		case <-timer.C:
			break LOOP
		}
	}

	leaving <- *h
	messages <- h.name + " has left"

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

func (h *handler) user() {

}

func newHandler(conn net.Conn) *handler {
	return &handler{
		conn: conn,
		in:   make(chan string),
		out:  make(chan string),
	}
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

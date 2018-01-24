package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

func NewClient(conn net.Conn) *Client {
	return &Client{
		Connection: conn,
		In:         make(chan string),
		Out:        make(chan string),
	}
}

type Client struct {
	Name       string
	Connection net.Conn
	In         chan string
	Out        chan string
	Room       *Room
}

//todo close write socket
func (c *Client) writing() {
	for {
		msg, ok := <-c.Out
		if !ok {
			break
		}
		_, err := io.WriteString(c.Connection, msg+"\n")
		if err != nil {
			log.Println(err)
		}
	}
}

//todo close read socket
func (c *Client) reading() {
	scanner := bufio.NewScanner(c.Connection)
	for scanner.Scan() {
		c.In <- scanner.Text()
	}
}

func (c *Client) Run() {
	go c.writing()
	go c.reading()

	c.Out <- "your name?"
	name := <-c.In
	c.Name = name
	c.Out <- "welcome " + c.Name

	for {
		cmd := <-c.In
		args := strings.Split(cmd, " ")
		switch args[0] {
		case "join":
			if len(args) >= 2 && args[1] != "" {
				name := args[1]
				room := RoomsManager.Join(name, c)
				c.Room = room
			}
		case "send":
			if len(args) >= 2 {
				msg := args[1]
				c.Room.BroadcastChan <- msg
			}
		default:
			c.Out <- "known command"
		}
	}
}

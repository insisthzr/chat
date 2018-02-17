package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func NewClient(conn net.Conn) *Client {
	return &Client{
		Connection: conn,
		IsClosed:   make(chan struct{}),
		In:         make(chan string),
		Out:        make(chan string),
	}
}

type Client struct {
	Name       string
	Connection net.Conn
	IsClosed   chan struct{}
	In         chan string
	Out        chan string
	Room       *Room
}

func (c *Client) Close() {
	c.Leave()
	c.IsClosed <- struct{}{}
	if err := c.Connection.Close(); err != nil {
		log.Println(err)
	}
}

func (c *Client) IsInRoom() bool {
	return c.Room != nil
}

func (c *Client) Leave() {
	if c.IsInRoom() {
		c.Room.Delete(c.Name)
		log.Printf("name: %s leave room: %s\n", c.Name, c.Room.Name)
	}
}

//TODO stream write
func (c *Client) writing() {
	for {
		select {
		case msg := <-c.Out:
			buf := []byte(msg + "\n")
			if _, err := c.Connection.Write(buf); err != nil {
				log.Println(err)
			}
		case <-c.IsClosed:
			return
		}
	}
}

//TODO initiative close
//TODO stream read
func (c *Client) reading() {
	buf := make([]byte, 1024)
	for {
		n, err := c.Connection.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		c.In <- strings.TrimSpace(string(buf[:n]))
	}
	c.Close()
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
		log.Printf("client %s command %s\n", c.Name, cmd)
		args := strings.Split(cmd, " ")
		switch args[0] {
		case "/join":
			if len(args) >= 2 && args[1] != "" {
				name := args[1]
				RoomsManager.Join(name, c)
				c.Room.RawBroadcastChan <- fmt.Sprintf("%s joined", c.Name)
			}
		case "/send":
			if c.IsInRoom() {
				if len(args) >= 2 {
					msg := args[1]
					c.Room.BroadcastChan <- &Message{Text: msg, From: c}
				}
			} else {
				c.Out <- "has not joined a room yet"
			}
		case "/users":
			if c.IsInRoom() {
				msg := "users:\n"
				for _, name := range c.Room.Users() {
					msg += (name + "\n")
				}
				c.Out <- msg
			} else {
				c.Out <- "has not joined a room yet"
			}
		case "/leave":
			c.Leave()
		case "/quit":
			msg := c.Name + " quited"
			if c.IsInRoom() {
				c.Room.RawBroadcastChan <- msg
			}
			c.Close()
		default:
			c.Out <- "known command"
		}
	}
}

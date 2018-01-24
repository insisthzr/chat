package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

var (
	port = flag.String("port", "20001", "port")
)

func main() {
	go func() {
		err := http.ListenAndServe(":50001", nil)
		log.Println(err)
	}()
	addr := "localhost:" + *port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening at", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		client := NewClient(conn)
		log.Println("client connected")
		go client.Run()
	}
}

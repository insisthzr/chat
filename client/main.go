package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	port = flag.String("port", "20001", "port")
)

func main() {
	conn, err := net.Dial("tcp", "localhost:"+*port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("connected")
	defer conn.Close()
	go reader(os.Stdout, conn)
	write(conn, os.Stdin)
	log.Println("disconnected")
}

func reader(writer io.Writer, reader io.Reader) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		panic(err)
	}
}

func write(writer io.Writer, reader io.Reader) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		panic(err)
	}
}

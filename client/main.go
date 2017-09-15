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
	port = flag.String("port", "20000", "port")
)

func main() {
	conn, err := net.Dial("tcp", "localhost:"+*port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("connected")
	defer conn.Close()
	go mustCopy(os.Stdout, conn)
	mustCopy(conn, os.Stdin)
	log.Println("disconnected")
}

func mustCopy(writer io.Writer, reader io.Reader) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		panic(err)
	}
}

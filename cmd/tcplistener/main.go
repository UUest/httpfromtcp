package main

import (
	"fmt"
	"log"
	"net"

	"github.com/UUest/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()
	fmt.Printf("Listening for TCP traffic on port: %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			fmt.Printf("Connection accepted from: %s\n", c.RemoteAddr())
			requestLine, err := request.RequestFromReader(c)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Request line:")
			fmt.Printf("- Method: %s\n", requestLine.RequestLine.Method)
			fmt.Printf("- Target: %s\n", requestLine.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", requestLine.RequestLine.HttpVersion)
			c.Close()
			fmt.Println("Connection to", c.RemoteAddr(), "closed")
		}(conn)

	}

}

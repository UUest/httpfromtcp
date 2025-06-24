package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
			lines := getLinesChannel(c)
			for line := range lines {
				fmt.Println(line)
			}
			c.Close()
			fmt.Println("Connection to", c.RemoteAddr(), "closed")
		}(conn)

	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		currentLineContents := ""
		for {
			b := make([]byte, 8)
			n, err := f.Read(b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}

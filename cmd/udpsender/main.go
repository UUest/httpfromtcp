package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const address = "localhost:42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Error resolving UDP address: %s\n", err.Error())
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Error dialing UDP address: %s\n", err.Error())
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Fatalf("Error reading line: %s\n", err.Error())
		}
		n, err := conn.Write(line)
		if err != nil {
			log.Fatalf("Error writing line: %s\n", err.Error())
		}
		fmt.Printf("Sent %d bytes\n", n)
	}
}

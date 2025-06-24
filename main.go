package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	sob := make([]byte, 8)
	var line string
	for {
		n, err := file.Read(sob)
		if n > 0 {
			parts := strings.Split(string(sob[:n]), "\n")
			if len(parts) == 1 {
				line += string(parts[0])
			} else if len(parts) > 1 {
				line += string(parts[0])
				fmt.Printf("read: %s\n", line)
				line = string(parts[1])
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

	}

}

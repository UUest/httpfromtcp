package main

import (
	"fmt"
	"io"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		sob := make([]byte, 8)
		var line string
		for {
			n, err := f.Read(sob)
			if n > 0 {
				parts := strings.Split(string(sob[:n]), "\n")
				if len(parts) == 1 {
					line += string(parts[0])
				} else if len(parts) > 1 {
					line += string(parts[0])
					if line != "" {
						lines <- line
					}
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
		close(lines)
	}()
	return lines
}

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const BYTES_TO_READ = 8

func main() {
	fp, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("File does not exist")
	}
	eightByteBuffer := make([]byte, 8)
	currentLine := ""
	for {
		n, err := fp.Read(eightByteBuffer)
		if err != nil {
			break
		}
		readString := string(eightByteBuffer[:n])
		splitStrings := strings.Split(readString, "\n")
		numParts := len(splitStrings)
		if numParts == 1 {
			currentLine += string(splitStrings[0])
			continue
		}
		for i, val := range(splitStrings) {
			if (i == 1) {
				fmt.Printf("read: %s\n", currentLine)
				currentLine = ""
			}
			currentLine += string(val)
		}
	}

	if(currentLine != "") {
		fmt.Printf("read: %s", currentLine)
	}
}

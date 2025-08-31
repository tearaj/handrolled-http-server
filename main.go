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
	allLines := make([][]byte, 7)
	currLineIndex := 0
	for {
		n, err := fp.Read(eightByteBuffer)

		if err != nil {
			log.Println("Error: ", err)
			break
		}
		readString := string(eightByteBuffer[:n])
		splitStrings := strings.Split(readString, "\n")
		if len(splitStrings) == 1 {
			allLines[currLineIndex] = append(allLines[currLineIndex], []byte(splitStrings[0])...)
		} else {
			// Newlines found, handle each part
			for i, val := range splitStrings {
				allLines[currLineIndex] = append(allLines[currLineIndex], []byte(val)...)
				if i < len(splitStrings)-1 {
					currLineIndex++
				}
			}
		}
	}

	for _, val := range allLines {
		fmt.Println(string(val))
	}
}

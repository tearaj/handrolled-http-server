package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

const BYTES_TO_READ = 8

func main() {
	fp, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("File does not exist")
	}
	ch := getLinesChannel(fp)
	for val := range ch {
		fmt.Printf("read: %v\n", val)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	dataCh := make(chan string)
	currLine := ""
	go func() {
		defer f.Close()
		defer close(dataCh)
		for {
			data := make([]byte, BYTES_TO_READ)
			_, err := f.Read(data)
			if err == io.EOF {
				dataCh <- currLine
				log.Println("Data read completed. Closing file.")
				break
			}
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				currLine += string(data[:i])
				dataCh <- currLine
				currLine = ""
				data = data[i+1:]
			}
			currLine += string(data[:])
		}
	}()
	return dataCh
}

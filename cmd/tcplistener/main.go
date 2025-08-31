package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

const BYTES_TO_READ = 8

func main() {
	socket, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer socket.Close()
	for {
		conn, err := socket.Accept()
		fmt.Println("Accepted connection: ", conn)
		if err != nil {
			log.Println("Warning accepting failed: ", err)
		}
		ch := getLinesChannel(conn)
		for val := range ch {
			fmt.Printf("read: %v\n", val)
		}
		fmt.Println("Closing since channel has closed!")
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
			_, err := f.Read(data) // Read is a blocking operation
			if err == io.EOF {
				dataCh <- currLine
				log.Println("Data read completed")
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

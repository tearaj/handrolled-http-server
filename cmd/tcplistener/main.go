package main

import (
	"fmt"
	"httpFromTCP/internal/request"
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
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error occured: ", err)
		}
		log.Println("Request Line:")
		log.Printf("- Method: %v\n", req.RequestLine.Method)
		log.Printf("- Target: %v\n", req.RequestLine.RequestTarget)
		log.Printf("- Version: %v\n", req.RequestLine.HttpVersion)
		fmt.Println("Closing since channel has closed!")
	}
}

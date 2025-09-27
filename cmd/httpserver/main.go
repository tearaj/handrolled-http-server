package main

import (
	"fmt"
	"httpFromTCP/internal/headers"
	"httpFromTCP/internal/request"
	"httpFromTCP/internal/response"
	"httpFromTCP/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{Code: response.STATUS_CODE_BAD_REQUEST, Message: fmt.Sprint("Your problem is not my problem\n")}
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: fmt.Sprint("Woopsie, my bad\n")}
	}
	message := "All good, frfr\n"
	headers := headers.GetDefaultHeaders(len(message))
	err := response.WriteStatusLine(w, response.STATUS_CODE_OK)
	err = response.WriteHeaders(w, headers)
	err = response.WriteBody(w, []byte(message))
	if err != nil {
		log.Println("ERROR: Writing handler", err)
	}
	return nil
}

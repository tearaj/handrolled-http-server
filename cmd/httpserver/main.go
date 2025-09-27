package main

import (
	"httpFromTCP/internal/headers"
	"httpFromTCP/internal/request"
	"httpFromTCP/internal/response"
	"httpFromTCP/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069
const badRequestHtml = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
const internalServerErrorHtml = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
const okHtml = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

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

func handler(w *response.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{Code: response.STATUS_CODE_BAD_REQUEST, Message: badRequestHtml}
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: internalServerErrorHtml}
	}
	headers := headers.GetDefaultHeaders(len(okHtml))
	err := w.WriteStatusLine(response.STATUS_CODE_OK)
	err = w.WriteHeaders(headers)
	err = w.WriteBody([]byte(okHtml))
	if err != nil {
		log.Println("ERROR: Writing handler", err)
	}
	return nil
}

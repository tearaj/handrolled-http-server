package main

import (
	"httpFromTCP/internal/headers"
	"httpFromTCP/internal/request"
	"httpFromTCP/internal/response"
	"httpFromTCP/internal/server"
	"io"
	"log"
	"net/http"
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
	if req.RequestLine.RequestTarget == "/httpbin/stream/100" {
		return handleStreaming(w)
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

func handleStreaming(w *response.Writer) *server.HandlerError {
	res, err := http.Get("https://httpbin.org/stream/100")
	if err != nil {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
	}
	bufSize := 32
	arr := make([]byte, bufSize)
	headers := headers.GetDefaultHeaders(0)
	headers.Replace("Transfer-Encoding", "chunked")
	headers.Remove("content-length")
	headers.Remove("connection")
	err = w.WriteStatusLine(response.STATUS_CODE_OK)
	if err != nil {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
	}
	for {
		n, err := res.Body.Read(arr)
		if err != nil {
			if err == io.EOF {
				break
			}
			return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
		}
		_, err = w.WriteChunkedBody(arr[:n])
		if err != nil {
			return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
		}
		copy(arr, arr[n:])
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		return &server.HandlerError{Code: response.STATUS_CODE_INTERNAL_SERVER_ERROR, Message: err.Error()}
	}

	return nil
}

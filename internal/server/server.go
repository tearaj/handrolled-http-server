package server

import (
	"fmt"
	"httpFromTCP/internal/headers"
	"httpFromTCP/internal/request"
	"httpFromTCP/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}
type Handler func(w *response.Writer, request *request.Request) *HandlerError

type Server struct {
	connections map[net.Conn]net.Conn
	closed      atomic.Bool
	listener    net.Listener
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	server := Server{listener: listener, connections: make(map[net.Conn]net.Conn)}
	go server.listen(handler)
	return &server, err
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return fmt.Errorf("server: closing an already closed server.")
	}
	err := s.listener.Close()
	if err != nil {
		s.closed.Store(true)
	}
	return err
}

func (s *Server) listen(handler Handler) {

	for {
		if s.closed.Load() {
			break
		}
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("ERROR: something went wrong while connection establishment", err)
		}
		s.handle(conn, handler)
	}
}

func (h *HandlerError) writeToConn(w response.Writer) error {
	res := h.Code
	errHeaders := headers.GetDefaultHeaders(len(h.Message))
	err := w.WriteStatusLine(res)
	if err != nil {
		log.Printf("ERROR: %v\n", h.Message)
	}
	err = w.WriteHeaders(errHeaders)
	err = w.WriteBody([]byte(h.Message))
	if err != nil {
		log.Printf("ERROR: %v\n", h.Message)
	}
	return err
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	defer conn.Close()
	s.connections[conn] = conn
	log.Println("Handler acceped!")
	req, err := request.RequestFromReader(conn)
	responseWriter := response.NewWriter(conn)
	if err != nil {
		handlerErr := HandlerError{Code: 400, Message: err.Error()}
		handlerErr.writeToConn(*responseWriter)
		return
	}
	handlerErr := handler(responseWriter, &req)
	if handlerErr != nil {
		log.Println("Serve Errors: ", handlerErr)
		handlerErr.writeToConn(*responseWriter)
	}
}

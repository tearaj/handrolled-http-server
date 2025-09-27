package server

import (
	"fmt"
	"httpFromTCP/internal/headers"
	"httpFromTCP/internal/request"
	"httpFromTCP/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}
type Handler func(w io.Writer, request *request.Request) *HandlerError

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

func (h *HandlerError) writeToConn(w io.Writer) error {
	res := h.Code
	errHeaders := headers.GetDefaultHeaders(len(h.Message))
	err := response.WriteStatusLine(w, res)
	if err != nil {
		log.Printf("ERROR: %v\n", h.Message)
	}
	err = response.WriteHeaders(w, errHeaders)
	err = response.WriteBody(w, []byte(h.Message))
	if err != nil {
		log.Printf("ERROR: %v\n", h.Message)
	}
	return err
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	s.connections[conn] = conn
	log.Println("Handler acceped!")
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := HandlerError{Code: 400, Message: err.Error()}
		handlerErr.writeToConn(conn)
		return
	}
	handlerErr := handler(conn, &req)
	if handlerErr != nil {
		log.Println("Serve Errors: ", handlerErr)
		handlerErr.writeToConn(conn)
	}
}

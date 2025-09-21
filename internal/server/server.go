package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	connections map[net.Conn]net.Conn
	closed      atomic.Bool
	listener    net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	server := Server{listener: listener, connections: make(map[net.Conn]net.Conn)}
	go server.listen()
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
func (s *Server) listen() {
	for {
		if s.closed.Load() {
			break
		}
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("ERROR: something went wrong while connection establishment", err)
		}
		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	s.connections[conn] = conn
	conn.Write([]byte(`HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!`))
}

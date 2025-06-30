package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/UUest/httpfromtcp/internal/request"
	"github.com/UUest/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: l,
		handler:  handler,
	}
	go s.Listen()
	return s, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}
	s.closed.Store(true)
	return nil
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		return
	}
	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	he := s.handler(buff, req)
	if he != nil {
		WriteHandlerError(conn, he)
		return
	}
	defaultHeaders := response.GetDefaultHeaders(len(buff.Bytes()))
	response.WriteStatusLine(conn, response.StatusCode(response.OK))
	response.WriteHeaders(conn, defaultHeaders)
	conn.Write(buff.Bytes())
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode int
	Message    string
}

func WriteHandlerError(w io.Writer, he *HandlerError) {
	response.WriteStatusLine(w, response.StatusCode(he.StatusCode))
	response.WriteHeaders(w, response.GetDefaultHeaders(len(he.Message)))
	w.Write([]byte(he.Message))
}

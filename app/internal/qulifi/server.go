package qulifi

import (
	"context"
	"fmt"
	"net"
)

type contextKey string

const (
	ctxKeyConnID = contextKey("github.com/codecrafters-io/redis-starter-go:server:connectionID")
)

type Server struct {
	l net.Listener
}

func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.l = l
	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	if err := s.Listen(addr); err != nil {
		return fmt.Errorf("s.Listen: %w", err)
	}

	connID := 1
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return fmt.Errorf("s.l.Accept(): %w", err)
		}
		ctx := context.WithValue(context.Background(), ctxKeyConnID, connID)
		go handleConnection(ctx, conn)
		connID++
	}

}

func handleConnection(ctx context.Context, conn net.Conn) {}

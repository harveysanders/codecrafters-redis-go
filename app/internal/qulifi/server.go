package qulifi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
)

type contextKey string

const (
	ctxKeyConnID = contextKey("github.com/codecrafters-io/redis-starter-go:server:connectionID")
)

type Server struct {
	l   net.Listener
	Log *slog.Logger
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

	defer func() {
		_ = s.l.Close()
	}()

	connID := 1
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return fmt.Errorf("s.l.Accept(): %w", err)
		}
		ctx := context.WithValue(context.Background(), ctxKeyConnID, connID)
		go s.handleConnection(ctx, conn)
		connID++
	}

}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	rawConnID := ctx.Value(ctxKeyConnID)
	connID, ok := rawConnID.(int)
	if !ok {
		s.Log.ErrorContext(ctx, "could not retrieved connection ID from context")
		return
	}

	log := s.Log.With(slog.Attr{Key: "connID", Value: slog.IntValue(connID)})

	_, err := conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		log.ErrorContext(ctx, fmt.Sprintf("write PONG: %v", err))
		return
	}
}

package qulifi

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

type contextKey string

const (
	ctxKeyConnID = contextKey("github.com/codecrafters-io/redis-starter-go:server:connectionID")
)

const (
	msgDelimiter = "\r\n"
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

	rdr := bufio.NewReader(conn)

	for {
		msg, err := rdr.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.ErrorContext(ctx, fmt.Sprintf("conn.ReadBytes: %v", err))
				return
			}
		}

		// Message must be at least 3 bytes (type:1, ending: 2)
		if len(msg) < 3 {
			log.ErrorContext(ctx, "short read", slog.Int("messageLen", len(msg)))
			return
		}

		// double checking the line ending since I can only pass a byte (\n) to io.ReadBytes.
		// TODO: See if there is a better solution
		if !validateEnding(msg) {
			log.ErrorContext(ctx, "invalid line ending", slog.String("message", string(msg)))
			return // Or keep reading?
		}

		log.InfoContext(ctx, string(msg))

		// Write PONG response
		if _, err = conn.Write([]byte("+PONG" + msgDelimiter)); err != nil {
			log.ErrorContext(ctx, fmt.Sprintf("write PONG: %v", err))
			return
		}
		return
	}
}

// validateEnding checks if the line ends with the correct delimiter.
func validateEnding(msg []byte) bool {
	oneFromEnd := msg[len(msg)-2:]
	return bytes.Equal(oneFromEnd, []byte(msgDelimiter))
}

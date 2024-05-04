package qulifi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/internal/resp"
)

type contextKey string

const (
	ctxKeyConnID = contextKey("github.com/codecrafters-io/redis-starter-go:server:connectionID")
	logKeyErr    = "error"
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

	for {
		var arrOffWire resp.TypeArray
		if _, err := arrOffWire.ReadFrom(conn); err != nil {
			if !errors.Is(err, io.EOF) {
				log.ErrorContext(ctx, "read commands", logKeyErr, err)
				return
			}
		}
		commands := make([]resp.TypeBulkString, 0, len(arrOffWire))
		for i, v := range arrOffWire {
			cmd, ok := v.(resp.TypeBulkString)
			if !ok {
				log.Error(fmt.Sprintf("expected BULK STRING, got %v", v))
				return
			}
			commands[i] = cmd
		}

		for i := 0; i < len(commands); i++ {
			cmd := commands[i]
			log.InfoContext(ctx, "incoming", "command", string(cmd))

			switch resp.Command(cmd) {
			case resp.CmdEcho:
				msg := commands[i+1]
				if _, err := conn.Write([]byte(msg)); err != nil {
					log.ErrorContext(ctx, "handleEcho", logKeyErr, err)
					return
				}
			case resp.CmdPing:
				// Write PONG response
				if _, err := conn.Write([]byte("+PONG" + resp.MsgDelimiter)); err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("write PONG: %v", err))
					return
				}
			default:
				log.Warn("unknown command", slog.String("message", string(cmd)))
			}
		}
	}
}

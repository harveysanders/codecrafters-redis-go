package qulifi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/internal/inmem"
	"github.com/codecrafters-io/redis-starter-go/app/internal/resp"
)

type contextKey string

const (
	ctxKeyConnID = contextKey("github.com/codecrafters-io/redis-starter-go:server:connectionID")
	logKeyErr    = "error"
)

type Option func(s *Server)

func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.Log = l
	}
}

func WithStore(store *inmem.Store) Option {
	return func(s *Server) {
		s.store = store
	}
}

type Server struct {
	l     net.Listener
	Log   *slog.Logger
	store *inmem.Store
}

func NewServer(opts ...Option) *Server {
	srv := &Server{}
	for _, o := range opts {
		o(srv)
	}
	return srv
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
		list := make([]resp.TypeBulkString, 0, len(arrOffWire))
		for _, v := range arrOffWire {
			cmd, ok := v.(resp.TypeBulkString)
			if !ok {
				log.Error(fmt.Sprintf("expected BULK STRING, got %v", v))
				return
			}
			list = append(list, cmd)
		}

		commands := resp.NewCommands(list)

		for commands.Next() {
			cmd := commands.Cur()
			log.InfoContext(ctx, "incoming", "command", string(cmd))

			switch resp.Command(cmd) {
			case resp.CmdEcho:
				if !commands.Next() {
					log.ErrorContext(ctx, "missing ECHO arg")
					return
				}
				msg := commands.Cur()
				data, err := msg.MarshalBinary()
				if err != nil {
					log.ErrorContext(ctx, "msg.MarshalBinary", logKeyErr, err)
					return
				}
				if _, err := conn.Write([]byte(data)); err != nil {
					log.ErrorContext(ctx, "handleEcho", logKeyErr, err)
					return
				}
			case resp.CmdGet:
				if !commands.Next() {
					log.ErrorContext(ctx, "missing GET key")
				}
				v, err := s.store.Get(string(commands.Cur()))
				if err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("store.Get(): %v", err))
					return
				}
				val, ok := v.(resp.TypeSimpleString)
				if !ok {
					log.ErrorContext(ctx, fmt.Sprintf("expected value to be a string, got: %+v", val))
				}

				data, err := val.MarshalBinary()
				if err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("write GET value marshal binary: %v", err))
					return
				}
				if _, err := conn.Write(data); err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("write Get response: %v", err))
					return
				}
			case resp.CmdPing:
				data, err := resp.TypeSimpleString("PONG").MarshalBinary()
				if err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("pong.MarshalBinary: %v", err))
					return
				}

				if _, err := conn.Write(data); err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("write PONG: %v", err))
					return
				}
			case resp.CmdSet:
				if !commands.Next() {
					log.ErrorContext(ctx, "missing SET key")
				}
				key := commands.Cur()
				if !commands.Next() {
					log.ErrorContext(ctx, "missing SET value")
				}
				err := s.store.Set(string(key), commands.Cur())
				if err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("store.Set(): %v", err))
					return
				}
				okResp, err := resp.RespOk().MarshalBinary()
				if err != nil {
					log.ErrorContext(ctx, "marshall OK response: %v", err)
					return
				}
				if _, err := conn.Write(okResp); err != nil {
					log.ErrorContext(ctx, fmt.Sprintf("write OK: %v", err))
					return
				}

			default:
				log.Warn("unknown command", slog.String("message", string(cmd)))
			}
		}
	}
}

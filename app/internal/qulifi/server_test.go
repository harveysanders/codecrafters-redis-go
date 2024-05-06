package qulifi_test

import (
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/internal/inmem"
	"github.com/codecrafters-io/redis-starter-go/app/internal/qulifi"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	addr := "0.0.0.0:9878"
	srv := qulifi.NewServer(
		qulifi.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
		qulifi.WithStore(inmem.New()),
	)

	go func() {
		err := srv.ListenAndServe(addr)
		require.NoError(t, err)
	}()

	time.Sleep(250 * time.Millisecond)

	client, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	_, err = client.Write([]byte("*3\r\n$3\r\nSET\r\n$4\r\npear\r\n$5\r\ngrape\r\n"))
	require.NoError(t, err)

	resp := make([]byte, 1024)
	n, err := client.Read(resp)
	require.NoError(t, err)
	require.Greater(t, n, 0, "expected some response")

	want := "$2\r\nOK\r\n"
	got := resp[:n]

	require.Equal(t, want, string(got))
}

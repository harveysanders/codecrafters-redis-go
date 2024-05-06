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

	type reqResp struct {
		req      string
		wantResp string
		desc     string
	}
	testCases := []reqResp{
		{
			desc:     "Send SET request",
			req:      "*3\r\n$3\r\nSET\r\n$4\r\npear\r\n$5\r\ngrape\r\n",
			wantResp: "$2\r\nOK\r\n",
		},
		{
			desc:     "Receive set value with GET request",
			req:      "*2\r\n$3\r\nGET\r\n$4\r\npear\r\n",
			wantResp: "$5\r\ngrape\r\n",
		},
	}
	client, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	resp := make([]byte, 1024)
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = client.Write([]byte(tc.req))
			require.NoError(t, err)

			n, err := client.Read(resp)
			require.NoError(t, err)
			require.Greater(t, n, 0, "expected some response")

			got := resp[:n]

			require.Equal(t, tc.wantResp, string(got))
		})
	}

}

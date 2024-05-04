package resp_test

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/internal/resp"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		data []byte
		typ  resp.TypeByte
		want any
	}{
		{
			data: []byte("*2\r\n$4\r\nECHO\r\n$10\r\nstrawberry\r\n"),
			typ:  resp.ByteArray,
			want: interface{}(resp.TypeArray{
				resp.TypeBulkString("ECHO"),
				resp.TypeBulkString("strawberry"),
			}),
		},
	}

	for _, tc := range testCases {
		switch tc.typ {
		case resp.ByteArray:
			want, ok := tc.want.(resp.TypeArray)
			require.True(t, ok)

			var got resp.TypeArray
			nRead, err := got.ReadFrom(bytes.NewReader(tc.data))
			require.NoError(t, err)
			require.Equal(t, int64(len(tc.data)), nRead)

			for i, v := range want {
				switch v.(type) {
				case resp.TypeBulkString:
					gotItem, ok := got[i].(resp.TypeBulkString)
					require.True(t, ok)
					require.Equal(t, v, gotItem)
				}
			}
		}
	}
}

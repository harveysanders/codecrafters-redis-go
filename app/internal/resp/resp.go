package resp

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"net/textproto"
	"strconv"
)

const (
	MsgDelimiter = "\r\n"
)

type Command string

const (
	CmdEcho Command = "ECHO"
	CmdPing Command = "PING"
)

type TypeByte rune

const (
	ByteSimpleString   TypeByte = '+'
	ByteSimpleError    TypeByte = '-'
	ByteInteger        TypeByte = ':'
	ByteBulkString     TypeByte = '$'
	ByteArray          TypeByte = '*'
	ByteNull           TypeByte = '_'
	ByteBoolean        TypeByte = '#'
	ByteDouble         TypeByte = ','
	ByteBigNumber      TypeByte = '('
	ByteBulkError      TypeByte = '!'
	ByteVerbatimString TypeByte = '='
	ByteMap            TypeByte = '%'
	ByteSet            TypeByte = '~'
	BytePush           TypeByte = '>'
)

type (
	TypeSimpleString   string
	TypeSimpleError    error
	TypeInteger        int64
	TypeBulkString     string
	TypeArray          []any
	TypeNull           struct{}
	TypeBoolean        bool
	TypeDouble         float64
	TypeBigNumber      big.Int
	TypeBulkError      error
	TypeVerbatimString struct {
		enc  string
		data []byte
	}
	TypeMap  map[any]any
	TypeSet  []any
	TypePush []interface{}
)

func (a *TypeArray) ReadFrom(r io.Reader) (int64, error) {
	rdr := textproto.NewReader(bufio.NewReader(r))
	typeByte, err := rdr.R.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("read type byte: %w", err)
	}
	nRead := int64(1)
	if typeByte != byte(ByteArray) {
		return nRead, fmt.Errorf("expected type %q, got %q", ByteArray, typeByte)
	}

	lenRaw, err := rdr.ReadLine()
	if err != nil {
		return nRead, fmt.Errorf("read length: %w", err)
	}

	nRead += int64(len(lenRaw) + len(MsgDelimiter))

	arrLen, err := strconv.Atoi(lenRaw)
	if err != nil {
		return nRead, fmt.Errorf("convert length: %w", err)
	}
	*a = make([]any, 0, arrLen)

	for i := 0; i < arrLen; i++ {
		rawItem, err := rdr.ReadLine()
		if err != nil {
			return nRead + int64(len(rawItem)) + int64(len(MsgDelimiter)), fmt.Errorf("read item: %w", err)
		}

		nRead += int64(len(rawItem) + len(MsgDelimiter))
		typ := TypeByte(rawItem[0])
		switch typ {
		case ByteBulkString:
			content, err := rdr.ReadLine()
			if err != nil {
				return nRead, fmt.Errorf("read string content: %w", err)
			}
			nRead += int64(len(content) + len(MsgDelimiter))

			*a = append(*a, TypeBulkString(content))
		default:
			return nRead, fmt.Errorf("unknown type: %s", string(typ))
		}
	}

	return nRead, nil
}

func (t *TypeBulkString) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return io.EOF
	}
	if data[0] != byte(ByteBulkString) {
		return fmt.Errorf("expected type %q, got %q", ByteBulkString, data[0])
	}

	*t = TypeBulkString(data[1:])
	return nil
}

func (t TypeBulkString) MarshalBinary() ([]byte, error) {
	data := make([]byte, 0, t.serializedLen())
	data = append(data, byte(ByteBulkString))
	data = append(data, fmt.Sprintf("%d", len(t))...)
	data = append(data, []byte(MsgDelimiter)...)
	data = append(data, t...)
	data = append(data, []byte(MsgDelimiter)...)
	return data, nil
}

// len returns the length of value when serialized. This is useful for pre-allocating buffers for marshaling.
func (t TypeBulkString) serializedLen() int {
	headerLen := 1
	lenLen := 1
	return len(t) + headerLen + lenLen + len(MsgDelimiter)*2
}

// msgLen adds 3 to the payload length. One byte for the type header, two bytes for the line-ending ("\r\n").
func msgLen(bodyLen int) int {
	return 1 + bodyLen + len(MsgDelimiter)
}

package codec

import (
	"errors"

	"github.com/panjf2000/gnet/v2"
)

var ErrIncompletePacket = errors.New("incomplete packet")

type Coder interface {
	Encode(buf []byte) ([]byte, error)
	Decode(c gnet.Conn) ([]byte, error)
}

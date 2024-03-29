package codec

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/panjf2000/gnet/v2"
)

const (
	magicNumber     = 1314
	magicNumberSize = 2
	bodySize        = 4
	bodyOffset      = magicNumberSize + bodySize
)

var magicNumberBytes []byte

func init() {
	magicNumberBytes = make([]byte, magicNumberSize)
	binary.BigEndian.PutUint16(magicNumberBytes, uint16(magicNumber))
}

// SimpleCodec Protocol format:
//
// * 0           2                       6
// * +-----------+-----------------------+
// * |   magic   |       body len        |
// * +-----------+-----------+-----------+
// * |                                   |
// * +                                   +
// * |           body bytes              |
// * +                                   +
// * |            ... ...                |
// * +-----------------------------------+
type SimpleCodec struct{}

// Encode 封包
func (codec *SimpleCodec) Encode(buf []byte) ([]byte, error) {
	msgLen := bodyOffset + len(buf)

	data := make([]byte, msgLen)
	copy(data, magicNumberBytes)

	binary.BigEndian.PutUint32(data[magicNumberSize:bodyOffset], uint32(len(buf)))
	copy(data[bodyOffset:msgLen], buf)
	return data, nil
}

// Decode 拆包
func (codec *SimpleCodec) Decode(c gnet.Conn) ([]byte, error) {
	buf, err := c.Peek(bodyOffset)
	if err != nil {
		return nil, err
	} else if len(buf) < bodyOffset {
		return nil, ErrIncompletePacket
	}

	if !bytes.Equal(magicNumberBytes, buf[:magicNumberSize]) {
		return nil, errors.New("invalid magic number")
	}

	bodyLen := binary.BigEndian.Uint32(buf[magicNumberSize:bodyOffset])
	msgLen := bodyOffset + int(bodyLen)
	if c.InboundBuffered() < msgLen {
		return nil, ErrIncompletePacket
	}

	if buf, err = c.Peek(msgLen); err != nil {
		return nil, err
	}

	if _, err = c.Discard(msgLen); err != nil {
		return nil, err
	}

	return buf[bodyOffset:msgLen], nil
}

func (codec *SimpleCodec) Unpack(buf []byte) ([]byte, error) {
	if len(buf) < bodyOffset {
		return nil, ErrIncompletePacket
	}

	if !bytes.Equal(magicNumberBytes, buf[:magicNumberSize]) {
		return nil, errors.New("invalid magic number")
	}

	bodyLen := binary.BigEndian.Uint32(buf[magicNumberSize:bodyOffset])
	msgLen := bodyOffset + int(bodyLen)
	if len(buf) < msgLen {
		return nil, ErrIncompletePacket
	}

	return buf[bodyOffset:msgLen], nil
}

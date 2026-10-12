package prefixer

import (
	"encoding/binary"
	"fmt"
	"math"
)

type binaryPrefixer struct {
	byteLen int
	maxVal  int
}

/*
Binary returns a binary length prefixer (Big Endian) of byteLen (1, 2, or 4 bytes).
*/
func Binary(byteLen int) Prefixer {
	var maxVal int
	switch byteLen {
	case 1:
		maxVal = math.MaxUint8
	case 2:
		maxVal = math.MaxUint16
	case 4:
		maxVal = math.MaxInt32
	default:
		panic("prefixer: binary length only supports 1, 2, or 4 bytes")
	}
	return &binaryPrefixer{byteLen: byteLen, maxVal: maxVal}
}

/*
Pre-configured binary length prefixer singletons.
*/
var (
	Binary1   = Binary(1)
	Binary2BE = Binary(2)
	Binary4BE = Binary(4)
)

func (b *binaryPrefixer) PackedLength() int {
	return b.byteLen
}

func (b *binaryPrefixer) EncodeLength(length int) ([]byte, error) {
	if length < 0 || length > b.maxVal {
		return nil, fmt.Errorf("%w: length %d exceeds max %d for %d bytes binary", ErrLengthTooLong, length, b.maxVal, b.byteLen)
	}
	buf := make([]byte, b.byteLen)
	switch b.byteLen {
	case 1:
		buf[0] = byte(length)
	case 2:
		binary.BigEndian.PutUint16(buf, uint16(length))
	case 4:
		binary.BigEndian.PutUint32(buf, uint32(length))
	}
	return buf, nil
}

func (b *binaryPrefixer) DecodeLength(data []byte) (int, int, error) {
	if len(data) < b.byteLen {
		return 0, 0, fmt.Errorf("%w: need %d bytes, got %d", ErrBufferTooShort, b.byteLen, len(data))
	}
	var val int
	switch b.byteLen {
	case 1:
		val = int(data[0])
	case 2:
		val = int(binary.BigEndian.Uint16(data[:2]))
	case 4:
		val = int(binary.BigEndian.Uint32(data[:4]))
	}
	return val, b.byteLen, nil
}

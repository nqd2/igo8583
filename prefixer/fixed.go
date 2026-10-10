package prefixer

import "fmt"

type fixedPrefixer struct {
	length int
}

// Fixed returns a Prefixer for fixed-length fields (stores 0 bytes on wire).
func Fixed(length int) Prefixer {
	return &fixedPrefixer{length: length}
}

func (f *fixedPrefixer) EncodeLength(length int) ([]byte, error) {
	if length != f.length {
		return nil, fmt.Errorf("%w: fixed length must be %d, got %d", ErrLengthTooLong, f.length, length)
	}
	return nil, nil
}

func (f *fixedPrefixer) DecodeLength(data []byte) (int, int, error) {
	return f.length, 0, nil
}

func (f *fixedPrefixer) PackedLength() int {
	return 0
}

package padder

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDataTooLong = errors.New("padder: data length exceeds target length")
)

// Padder defines how a string field is padded or unpadded.
type Padder interface {
	// Pad pads the input string to the specified target length.
	Pad(data string, targetLen int) (string, error)
	// Unpad removes padding from the input string.
	Unpad(data string) string
}

type leftPadder struct {
	padRune rune
}

// Left returns a Padder that prepends padRune to reach target length.
func Left(padRune rune) Padder {
	return &leftPadder{padRune: padRune}
}

func (p *leftPadder) Pad(data string, targetLen int) (string, error) {
	currLen := len(data)
	if currLen > targetLen {
		return "", fmt.Errorf("%w: current length %d > target length %d", ErrDataTooLong, currLen, targetLen)
	}
	if currLen == targetLen {
		return data, nil
	}
	padding := strings.Repeat(string(p.padRune), targetLen-currLen)
	return padding + data, nil
}

func (p *leftPadder) Unpad(data string) string {
	return strings.TrimLeft(data, string(p.padRune))
}

type rightPadder struct {
	padRune rune
}

// Right returns a Padder that appends padRune to reach target length.
func Right(padRune rune) Padder {
	return &rightPadder{padRune: padRune}
}

func (p *rightPadder) Pad(data string, targetLen int) (string, error) {
	currLen := len(data)
	if currLen > targetLen {
		return "", fmt.Errorf("%w: current length %d > target length %d", ErrDataTooLong, currLen, targetLen)
	}
	if currLen == targetLen {
		return data, nil
	}
	padding := strings.Repeat(string(p.padRune), targetLen-currLen)
	return data + padding, nil
}

func (p *rightPadder) Unpad(data string) string {
	return strings.TrimRight(data, string(p.padRune))
}

type nonePadder struct{}

// None returns a Padder that performs no padding.
func None() Padder {
	return nonePadder{}
}

func (p nonePadder) Pad(data string, targetLen int) (string, error) {
	if targetLen > 0 && len(data) > targetLen {
		return "", fmt.Errorf("%w: current length %d exceeds max length %d", ErrDataTooLong, len(data), targetLen)
	}
	return data, nil
}

func (p nonePadder) Unpad(data string) string {
	return data
}

// Common pre-configured padders
var (
	LeftZero   = Left('0')
	LeftSpace  = Left(' ')
	RightSpace = Right(' ')
	RightZero  = Right('0')
	NoPadding  = None()
)

package padder

import (
	"errors"
	"fmt"
	"strings"
)

/* ============================================================================
 * Errors
 * ============================================================================ */

var (
	ErrDataTooLong = errors.New("padder: data length exceeds target length")
)

/* ============================================================================
 * Padder Interface
 * ============================================================================ */

/*
Padder defines strategies for aligning, padding, and unpadding field string values
to reach required fixed or boundary widths.
*/
type Padder interface {
	Pad(data string, targetLen int) (string, error)
	Unpad(data string) string
}

/* ============================================================================
 * Strategy Factories
 * ============================================================================ */

/*
Left returns a Padder that prepends padRune to reach target length.
*/
func Left(padRune rune) Padder {
	return &leftPadder{padRune: padRune}
}

type leftPadder struct {
	padRune rune
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

/*
Right returns a Padder that appends padRune to reach target length.
*/
func Right(padRune rune) Padder {
	return &rightPadder{padRune: padRune}
}

type rightPadder struct {
	padRune rune
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

/*
None returns a Padder that performs no padding, passing values through unchanged.
*/
func None() Padder {
	return nonePadder{}
}

type nonePadder struct{}

func (p nonePadder) Pad(data string, targetLen int) (string, error) {
	if targetLen > 0 && len(data) > targetLen {
		return "", fmt.Errorf("%w: current length %d exceeds max length %d", ErrDataTooLong, len(data), targetLen)
	}
	return data, nil
}

func (p nonePadder) Unpad(data string) string {
	return data
}

/* ============================================================================
 * Standard Instances
 * ============================================================================ */

/*
Pre-configured padder singletons for standard financial message encodings.
*/
var (
	LeftZero   = Left('0')
	LeftSpace  = Left(' ')
	RightSpace = Right(' ')
	RightZero  = Right('0')
	NoPadding  = None()
)

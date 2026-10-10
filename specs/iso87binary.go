package specs

import "github.com/nqd2/igo8583/packager"

// ISO87Binary builds the 128-position ISO 8583:1987 profile with a binary
// bitmap, BCD MTI and numeric data, and packed BCD variable-length prefixes.
func ISO87Binary() *packager.Codec {
	return newISO87(true)
}

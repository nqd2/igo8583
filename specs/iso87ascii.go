package specs

import "github.com/nqd2/igo8583/packager"

// ISO87ASCII creates the 128-position ISO 8583:1987 ASCII packager.
func ISO87ASCII() *packager.Codec {
	return newISO87(false)
}

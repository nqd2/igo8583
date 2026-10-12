package specs

import "github.com/nqd2/igo8583/packager"

/* ISO87ASCII creates the ISO 8583:1987 ASCII codec. */
func ISO87ASCII() *packager.Codec {
	return newISO87(false)
}

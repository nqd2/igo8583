package retail

import (
	_ "embed"

	"github.com/nqd2/igo8583"
)

//go:embed layout.json
var layoutJSON []byte

/* Codec compiles and returns the retail switch's wire layout. */
func Codec() (*igo8583.Codec, error) {
	return igo8583.FromJSON(layoutJSON)
}

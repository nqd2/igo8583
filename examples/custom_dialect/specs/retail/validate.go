package retail

import (
	"fmt"

	"github.com/nqd2/igo8583"
)

/*
Validate applies the sample retail switch's transaction rules to a message.
The required fields here illustrate application-owned validation; they are
not rules supplied by igo8583 or a real payment network.
*/
func Validate(msg *igo8583.Message) error {
	if msg == nil {
		return fmt.Errorf("retail: nil message")
	}
	var required []int
	switch msg.GetMTI() {
	case "0200":
		required = []int{3, 4, 11, 41}
	case "0210":
		required = []int{3, 4, 11, 39, 41}
	default:
		return fmt.Errorf("retail: unsupported MTI %q", msg.GetMTI())
	}
	for _, id := range required {
		if !msg.HasField(id) {
			return fmt.Errorf("retail: DE %d is required for MTI %s", id, msg.GetMTI())
		}
	}
	return nil
}

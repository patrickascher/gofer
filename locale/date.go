package locale

import (
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/controller"
	"time"
)

// DateFormat will return the defined user date format.
// Drawback: fixed to auth.Claim at the moment. This has to get fixed to be more flexible.
func DateFormat(t time.Time, c controller.Interface) string {
	if c.Context().Request.JWTClaim().(*auth.Claim).Options["DateFormat"] == "DD.MM.YYYY" {
		return t.Format("02.01.2006 15:04")
	}
	return t.Format("2006-01-02 15:04")
}

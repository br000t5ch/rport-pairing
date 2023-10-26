package retrieve

import (
	"github.com/openrport/rport-pairing/deposit"
	"net/http"
)

type UpdateHandler struct {
	StaticDeposit deposit.Deposit
}

// Handle the request for a client update.
// No client data is needed
func (rh *UpdateHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	renderUpdate(rw, clientOs(r))
}

func renderUpdate(rw http.ResponseWriter, os string) {
	switch os {
	case "windows":
		rw.Header().Add("Content-Disposition", "attachment; filename=\"rport-update.ps1\"")
		includeFileRaw(rw, "templates/windows/update_init.ps1")
		includeFile(rw, "templates/header.txt")
		includeFile(rw, "templates/windows/functions.ps1")
		includeFile(rw, "templates/windows/update.ps1")
	default:
		rw.Header().Add("Content-Disposition", "attachment; filename=\"rport-update.sh\"")
		includeFileRaw(rw, "templates/linux/init.sh")
		includeFile(rw, "templates/header.txt")
		includeFile(rw, "templates/linux/vars.sh")
		includeFile(rw, "templates/linux/functions.sh")
		includeFile(rw, "templates/linux/update.sh")
	}
}

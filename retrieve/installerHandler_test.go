package retrieve_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/openrport/rport-pairing/deposit"
	"github.com/openrport/rport-pairing/internal/cache"
	"github.com/openrport/rport-pairing/retrieve"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TestInstallerWith struct {
	userAgent   string
	pairingCode string
}
type ExpectedInstallerResults struct {
	httpStatus int
	keyword    string
	variable   string
}

func TestInstallerHandler_ServeHTTP(t *testing.T) {
	c := cache.New()
	demoDeposit := deposit.Deposit{
		ConnectUrl:  "https://rport.example.com",
		Fingerprint: "2a:c1:71:09:80:ba:7c:10:05:e5:2c:99:6d:15:56:24",
		ClientId:    "client1\";exit",
		Password:    "foobaz",
		Code:        "cZ1ZhsG",
	}
	var tests = []struct {
		tw TestInstallerWith
		er ExpectedInstallerResults
	}{
		{
			TestInstallerWith{"curl/7.79.1", "cZ1ZhsG"},
			ExpectedInstallerResults{200, "BEGINNING of templates/linux/install.sh", strings.ReplaceAll(demoDeposit.ClientId, "\"", "\\\"")},
		},
		{
			TestInstallerWith{"curl/7.79.1", "C6esANp"},
			ExpectedInstallerResults{200, "/bin/sh -e", strings.ReplaceAll(demoDeposit.ClientId, "\"", "\\\"")},
		},
		{
			TestInstallerWith{"Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.20348.1", "cZ1ZhsG"},
			ExpectedInstallerResults{200, "function Expand-Zip", strings.ReplaceAll(demoDeposit.ClientId, "\"", "`\"")},
		},

		{
			TestInstallerWith{"go-test", "abcdefg"},
			ExpectedInstallerResults{404, "#No pairing found by pairing code abcdefg", ""},
		},
	}
	// Store pairing data in the cache
	c.Set("C6esANp", demoDeposit, 10*time.Second)

	// Create the handler to be tested
	installerHandler := &retrieve.InstallerHandler{
		StaticDeposit: demoDeposit,
		Cache:         c,
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("User-Agent='%s' PairingCode=%s", tc.tw.userAgent, tc.tw.pairingCode), func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/"+tc.tw.pairingCode, nil)
			// Simulate a URL like /0000000
			vars := map[string]string{
				"pairingCode": tc.tw.pairingCode,
			}
			request.Header.Set("User-Agent", tc.tw.userAgent)
			request = mux.SetURLVars(request, vars)
			recorder := httptest.NewRecorder()
			installerHandler.ServeHTTP(recorder, request)
			assert.Equal(t, tc.er.httpStatus, recorder.Result().StatusCode)
			assert.Contains(t, recorder.Body.String(), tc.er.keyword, fmt.Sprintf("Expexted key word '%s' missing.", tc.er.keyword))
			assert.Contains(t, recorder.Body.String(), tc.er.variable, "Variable not found in body:\n"+recorder.Body.String())
			if recorder.Result().StatusCode == 200 {
				assert.Contains(t, recorder.Header().Get("Content-Disposition"), "attachment; filename=\"rport-installer", "Content-Disposition Header wrong or missing")
			}
			t.Log("Got HTTP status code", recorder.Result().StatusCode)
		})
	}
}

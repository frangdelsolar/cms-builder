package server_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/server_test"
	"github.com/stretchr/testify/assert"
)

func TestSecurity(t *testing.T) {
	bed := SetupServerTestBed()
	apiBaseUrl := "http://localhost:8080"

	// Start the server in a goroutine
	go func() {
		bed.Server.Run(bed.Mgr.GetRoutes, apiBaseUrl)
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	t.Run("SQL Injection Prevention", testSQLInjection(&bed, apiBaseUrl))
	t.Run("XSS Prevention", testXSSPrevention(&bed, apiBaseUrl))

}

func testSQLInjection(bed *TestUtils, apiBaseUrl string) func(t *testing.T) {
	return func(t *testing.T) {
		// Assuming you have an endpoint that interacts with a database
		// and is potentially vulnerable to SQL injection.

		payload := `{"field1": "value' OR '1'='1", "field2": "value2"}` // Attempted SQL injection
		rr := HitEndpoint(
			t,
			bed.Server.Root.ServeHTTP,
			"POST",
			apiBaseUrl+"/private/api/mock-structs/new", // adjust to a vulnerable endpoint
			payload,
			true,
			bed.AdminUser,
			bed.Logger,
			"192.168.1.1",
		)

		if rr.Code == http.StatusOK {
			// Check if the response indicates an error or unexpected data due to the injection.
			if strings.Contains(rr.Body.String(), "unexpected data") || strings.Contains(rr.Body.String(), "internal error") {
				t.Errorf("SQL injection vulnerability detected: %s", rr.Body.String())
			}
		}
		// ideally, you would check the database directly to confirm.
	}
}

func testXSSPrevention(bed *TestUtils, apiBaseUrl string) func(t *testing.T) {
	return func(t *testing.T) {
		payload := `{"field1": "<script>alert('XSS')</script>", "field2": "value2"}`
		rr := HitEndpoint(
			t,
			bed.Server.Root.ServeHTTP,
			"POST",
			apiBaseUrl+"/private/api/mock-structs/new", // adjust to an endpoint that returns data
			payload,
			true,
			bed.AdminUser,
			bed.Logger,
			"192.168.1.1",
		)

		t.Log(rr.Body.String())
		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotContains(t, rr.Body.String(), "<script>alert('XSS')</script>")
	}
}

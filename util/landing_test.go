package util

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
)

// TestLandingPage verifies that a Falcore RequestFilter generated by
// NewLandingFilter responds appropriately to a request. Specifically, this test
// verifies that no error occurs in the response.
func TestLandingPage(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	l := new(LandingFilter)
	l.ServeHTTP(w, request)
	response := w.Result()
	assert.Equal(t, 200, response.StatusCode)
}

// TestAnotherLandingPage verifies that a Falcore RequestFilter generated by
// NewLandingFilter responds appropriately to a request for an unexpected URI.
// Specifically, this test verifies that no error occurs in the response.
func TestAnotherLandingPage(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/other/page/somewhere/else.php",
		nil,
	)
	w := httptest.NewRecorder()
	l := new(LandingFilter)
	l.ServeHTTP(w, request)
	response := w.Result()
	assert.Equal(t, 200, response.StatusCode)
}

// TestPostLandingPage verifies that a Falcore RequestFilter generated by
// NewLandingFilter responds appropriately to a POST request. Specifically, this
// test verifies that no error occurs in the response.
func TestPostLandingPage(t *testing.T) {
	request := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	l := new(LandingFilter)
	l.ServeHTTP(w, request)
	response := w.Result()
	assert.Equal(t, 200, response.StatusCode)
}

func TestLandingFilterFromConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "landingfilter",
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        "",
				Explanation: "empty config",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data:        "{}",
				Explanation: "empty object",
			},
			{
				Data:        "null",
				Explanation: "null config",
			},
			{
				Data:        "42",
				Explanation: "numeric config",
			},
			{
				Data:        `"anything goes"`,
				Explanation: "string config",
			},
		},
	}
	test.Run(t)
}

package test

import (
	. ".."
	"github.com/fitstar/falcore"
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLandingPage(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	_, response := falcore.TestWithRequest(request, NewLandingFilter(), nil)
	assert.Equal(t, 200, response.StatusCode)
}

func TestAnotherLandingPage(t *testing.T) {
	request, err := http.NewRequest("GET", "/other/page/somewhere/else.php", nil)
	assert.NoError(t, err)
	_, response := falcore.TestWithRequest(request, NewLandingFilter(), nil)
	assert.Equal(t, 200, response.StatusCode)
}

func TestPostLandingPage(t *testing.T) {
	request, err := http.NewRequest("POST", "/", nil)
	assert.NoError(t, err)
	_, response := falcore.TestWithRequest(request, NewLandingFilter(), nil)
	assert.Equal(t, 200, response.StatusCode)
}


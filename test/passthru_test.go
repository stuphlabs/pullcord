package test

import (
	. ".."
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func TestPassthru(t *testing.T) {
	request, err := http.NewRequest("GET", "http://localhost", nil)
	assert.NoError(t, err)

	landingPipeline := falcore.NewPipeline()
	landingPipeline.Upstream.PushBack(NewLandingFilter())

	landingServer := falcore.NewServer(58080, landingPipeline)

	regex, err := regexp.Compile("Pullcord Landing Page")
	assert.NoError(t, err)

	go serveLandingPage(t, landingServer)

	_, response := falcore.TestWithRequest(request, NewPassthruFilter("localhost", 58080), nil)
	landingServer.StopAccepting()
	assert.Equal(t, 200, response.StatusCode)
	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, regex.Match(contents))
}

func serveLandingPage(t *testing.T, landingServer *falcore.Server) {
	if err := landingServer.ListenAndServe(); err != nil {
		assert.NoError(t, err)
	}
}

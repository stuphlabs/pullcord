package scaledown

import (
	"github.com/fitstar/falcore"
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
	"regexp"
	"io/ioutil"
)

func TestPassthru(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	landingPipeline := falcore.NewPipeline()
	landingPipeline.Upstream.PushBack(NewLandingFilter())

	landingServer := falcore.NewServer(58080, landingPipeline)

	passthruPipeline := falcore.NewPipeline()
	passthruPipeline.Upstream.PushBack(NewPassthruFilter("localhost", 58080))

	regex, err := regexp.Compile("Pullcord Landing Page")
	assert.NoError(t, err)

	go serveLandingPage(t, landingServer)

	_, response := falcore.TestWithRequest(request, NewLandingFilter(), nil)
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


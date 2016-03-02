package pullcord

import (
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func TestConditionalFilter(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	condition := true
	regexTrue, err := regexp.Compile(".*condition was true.*")
	assert.NoError(t, err)
	regexFalse, err := regexp.Compile(".*condition was false.*")
	assert.NoError(t, err)
	conditionalFilter := NewConditionalFilter(
		func(req *falcore.Request) bool {
			return condition
		},
		falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					200,
					nil,
					"<html><body><p>condition was true</p></body></html>",
				)
			},
		),
		falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					200,
					nil,
					"<html><body><p>condition was false</p></body></html>",
				)
			},
		),
	)

	condition = false
	_, responseFalse := falcore.TestWithRequest(request, conditionalFilter, nil)
	assert.Equal(t, 200, responseFalse.StatusCode)
	contentsFalse, err := ioutil.ReadAll(responseFalse.Body)
	assert.NoError(t, err)
	assert.True(t, regexFalse.Match(contentsFalse))

	condition = true
	_, responseTrue := falcore.TestWithRequest(request, conditionalFilter, nil)
	assert.Equal(t, 200, responseTrue.StatusCode)
	contentsTrue, err := ioutil.ReadAll(responseTrue.Body)
	assert.NoError(t, err)
	assert.True(t, regexTrue.Match(contentsTrue))
}

package test

import (
	. ".."
	"errors"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func testableConditionalFilter(condition *bool, err *error) falcore.RequestFilter {
	return NewConditionalFilter(
		func(req *falcore.Request) (bool, error) {
			return *condition, *err
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
		falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					500,
					nil,
					"<html><body><p>an error occurred</p></body></html>",
				)
			},
		),
	)
}

func TestConditionalFilterTrue(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), "condition was true"))
	assert.False(t, strings.Contains(string(contents), "condition was false"))
	assert.False(t, strings.Contains(string(contents), "an error occurred"))
}

func TestConditionalFilterFalse(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := false
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), "condition was true"))
	assert.True(t, strings.Contains(string(contents), "condition was false"))
	assert.False(t, strings.Contains(string(contents), "an error occurred"))
}

func TestConditionalFilterError(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = errors.New("test error")
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), "condition was true"))
	assert.False(t, strings.Contains(string(contents), "condition was false"))
	assert.True(t, strings.Contains(string(contents), "an error occurred"))
}

func TestConditionalFilterTrueFalse(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(request, conditional_filter, nil)
	condition = false
	_, response2 := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents1), "condition was true"))
	assert.False(t, strings.Contains(string(contents1), "condition was false"))
	assert.False(t, strings.Contains(string(contents1), "an error occurred"))
	assert.False(t, strings.Contains(string(contents2), "condition was true"))
	assert.True(t, strings.Contains(string(contents2), "condition was false"))
	assert.False(t, strings.Contains(string(contents2), "an error occurred"))
}

func TestConditionalFilterFalseTrue(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := false
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(request, conditional_filter, nil)
	condition = true
	_, response2 := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents1), "condition was true"))
	assert.True(t, strings.Contains(string(contents1), "condition was false"))
	assert.False(t, strings.Contains(string(contents1), "an error occurred"))
	assert.True(t, strings.Contains(string(contents2), "condition was true"))
	assert.False(t, strings.Contains(string(contents2), "condition was false"))
	assert.False(t, strings.Contains(string(contents2), "an error occurred"))
}

func TestConditionalFilterTrueError(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(request, conditional_filter, nil)
	test_err = errors.New("test error")
	condition = false
	_, response2 := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 500, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents1), "condition was true"))
	assert.False(t, strings.Contains(string(contents1), "condition was false"))
	assert.False(t, strings.Contains(string(contents1), "an error occurred"))
	assert.False(t, strings.Contains(string(contents2), "condition was true"))
	assert.False(t, strings.Contains(string(contents2), "condition was false"))
	assert.True(t, strings.Contains(string(contents2), "an error occurred"))
}

func TestConditionalFilterErrorFalse(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = errors.New("test error")
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(request, conditional_filter, nil)
	condition = false
	test_err = nil
	_, response2 := falcore.TestWithRequest(request, conditional_filter, nil)

	/* check */
	assert.Equal(t, 500, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents1), "condition was true"))
	assert.False(t, strings.Contains(string(contents1), "condition was false"))
	assert.True(t, strings.Contains(string(contents1), "an error occurred"))
	assert.False(t, strings.Contains(string(contents2), "condition was true"))
	assert.True(t, strings.Contains(string(contents2), "condition was false"))
	assert.False(t, strings.Contains(string(contents2), "an error occurred"))
}

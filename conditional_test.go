package pullcord

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
)

// testableConditionalFilter is a testing helper function that gets a Falcore
// RequestFilter created by NewConditionalFilter which should display one of
// three static web pages based on the current value of an external bool
// variable and an external error variable, both initially passed by reference
// to this function.
func testableConditionalFilter(
	condition *bool,
	err *error,
) falcore.RequestFilter {
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
					"<html><body><p>"+
						"condition was true"+
						"</p></body></html>",
				)
			},
		),
		falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					200,
					nil,
					"<html><body><p>"+
						"condition was false"+
						"</p></body></html>",
				)
			},
		),
		falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					500,
					nil,
					"<html><body><p>"+
						"an error occurred"+
						"</p></body></html>",
				)
			},
		),
	)
}

// TestConditionalFilterTrue verifies that a NewConditionalFilter will
// correctly route a request to the appropriate Falcore ResponseFilter based on
// the return value of a callback function which will return true in this case.
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
	assert.False(
		t,
		strings.Contains(string(contents), "condition was false"),
	)
	assert.False(t, strings.Contains(string(contents), "an error occurred"))
}

// TestConditionalFilterFalse verifies that a NewConditionalFilter will
// correctly route a request to the appropriate Falcore ResponseFilter based on
// the return value of a callback function which will return false in this case.
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
	assert.False(
		t,
		strings.Contains(string(contents), "condition was true"),
	)
	assert.True(
		t,
		strings.Contains(string(contents), "condition was false"),
	)
	assert.False(t, strings.Contains(string(contents), "an error occurred"))
}

// TestConditionalFilterError verifies that a NewConditionalFilter will
// correctly route a request to the appropriate Falcore ResponseFilter based on
// the return value of a callback function which will throw an error in this
// case.
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
	assert.False(
		t,
		strings.Contains(string(contents), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents), "condition was false"),
	)
	assert.True(t, strings.Contains(string(contents), "an error occurred"))
}

// TestConditionalFilterTrueFalse verifies that a NewConditionalFilter will
// correctly route consecutive requests to the appropriate Falcore
// ResponseFilters based on the return values of a callback function which,
// in this case, will return true during the first request but return false
// during the second request.
func TestConditionalFilterTrueFalse(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)
	condition = false
	_, response2 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(contents1), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "an error occurred"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "condition was true"),
	)
	assert.True(
		t,
		strings.Contains(string(contents2), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "an error occurred"),
	)
}

// TestConditionalFilterFalseTrue verifies that a NewConditionalFilter will
// correctly route consecutive requests to the appropriate Falcore
// ResponseFilters based on the return values of a callback function which,
// in this case, will return false during the first request but return true
// during the second request.
func TestConditionalFilterFalseTrue(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := false
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)
	condition = true
	_, response2 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.False(
		t,
		strings.Contains(string(contents1), "condition was true"),
	)
	assert.True(
		t,
		strings.Contains(string(contents1), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "an error occurred"),
	)
	assert.True(
		t,
		strings.Contains(string(contents2), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "an error occurred"),
	)
}

// TestConditionalFilterTrueError verifies that a NewConditionalFilter will
// correctly route consecutive requests to the appropriate Falcore
// ResponseFilters based on the return values of a callback function which,
// in this case, will return true during the first request but throw an error
// during the second request.
func TestConditionalFilterTrueError(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = nil
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)
	test_err = errors.New("test error")
	condition = false
	_, response2 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response1.StatusCode)
	assert.Equal(t, 500, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(contents1), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "an error occurred"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "condition was false"),
	)
	assert.True(t, strings.Contains(string(contents2), "an error occurred"))
}

// TestConditionalFilterErrorFalse verifies that a NewConditionalFilter will
// correctly route consecutive requests to the appropriate Falcore
// ResponseFilters based on the return values of a callback function which,
// in this case, will throw an error during the first request but will return
// false during the second request.
func TestConditionalFilterErrorFalse(t *testing.T) {
	/* setup */
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	condition := true
	var test_err error
	test_err = errors.New("test error")
	conditional_filter := testableConditionalFilter(&condition, &test_err)

	/* run */
	_, response1 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)
	condition = false
	test_err = nil
	_, response2 := falcore.TestWithRequest(
		request,
		conditional_filter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response1.StatusCode)
	assert.Equal(t, 200, response2.StatusCode)

	contents1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	contents2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.False(
		t,
		strings.Contains(string(contents1), "condition was true"),
	)
	assert.False(
		t,
		strings.Contains(string(contents1), "condition was false"),
	)
	assert.True(t, strings.Contains(string(contents1), "an error occurred"))
	assert.False(
		t,
		strings.Contains(string(contents2), "condition was true"),
	)
	assert.True(
		t,
		strings.Contains(string(contents2), "condition was false"),
	)
	assert.False(
		t,
		strings.Contains(string(contents2), "an error occurred"),
	)
}

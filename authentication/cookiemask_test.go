package authentication

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dustin/randbo"
	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
)

var exampleCookieValueRegex = regexp.MustCompile(
	"^[0-9A-Fa-f]{" +
		strconv.Itoa(minSessionCookieValueRandSize*2) +
		"}$",
)
var randgen = randbo.New()

// gostring is a testing helper function that serializes any object.
func gostring(i interface{}) string {
	return fmt.Sprintf("%#v", i)
}

var cookieMaskTestPage = falcore.NewRequestFilter(
	func(req *falcore.Request) *http.Response {
		var content = "<html><body><h1>cookies</h1><ul>"
		for _, cke := range req.HttpRequest.Cookies() {
			content +=
				"<li class=\"cke\">" +
					cke.String() +
					"</li>"
		}
		content += "</ul><h1>context</h1><ul>"
		sesh := req.Context["session"].(*MinSession)
		for key, val := range sesh.GetValues() {
			content +=
				"<li class=\"sesh\">" +
					key +
					": " +
					gostring(val) +
					"</li>"
		}
		content += "</ul></body></html>"
		return falcore.StringResponse(
			req.HttpRequest,
			200,
			nil,
			content,
		)
	},
)

// testCookieGen is a testing helper function that generates a randomized
// cookie name and value pair but with the given variant string being in a
// predictable location in the cookie name so it will be easy to identify with
// regular expressions.
func testCookieGen(variant string) (result http.Cookie) {
	nbytes := make([]byte, minSessionCookieNameRandSize)
	vbytes := make([]byte, minSessionCookieValueRandSize)
	randgen.Read(nbytes)
	randgen.Read(vbytes)
	result.Name = variant + "-" + hex.EncodeToString(nbytes)
	result.Value = hex.EncodeToString(vbytes)
	return result
}

type alwaysErrorHandler struct {
}

func (handler alwaysErrorHandler) GetSession() (Session, error) {
	return nil, errors.New(
		"Ran GetSession on an instance of alwaysErrorSessionHandler",
	)
}

type alwaysErrorSession struct {
}

func (sesh alwaysErrorSession) GetValue(key string) (interface{}, error) {
	return nil, errors.New(
		"Ran GetValue on an instance of alwaysErrorSession",
	)
}

func (sesh alwaysErrorSession) SetValue(key string, val interface{}) error {
	return errors.New(
		"Ran SetValue on an instance of alwaysErrorSession",
	)
}

func (sesh alwaysErrorSession) CookieMask(inc []*http.Cookie) (
	[]*http.Cookie,
	[]*http.Cookie,
	error,
) {
	return nil, nil, errors.New(
		"Ran CookieMask on an instance of alwaysErrorSession",
	)
}

type alwaysErrorSessionsHandler struct {
}

func (handler alwaysErrorSessionsHandler) GetSession() (Session, error) {
	var sesh alwaysErrorSession
	return sesh, nil
}

var handlerAlwaysErrorsOnSessionRequest alwaysErrorHandler
var handlerSessionsAlwaysError alwaysErrorSessionsHandler

// TestCookiemaskCookieless verifies that a Falcore RequestFilter generated by
// NewCookiemaskFilter behaves as expected when given a request that contains no
// cookies. Specifically, this function verifies that no error occurred, that
// the next RequestFilter in the chain did not receive an indication that a
// maskable cookie was found, and that a new maskable cookie of the expected
// variant is being set.
func TestCookiemaskCookieless(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile(
		"^test-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	filter := CookiemaskFilter{
		NewMinSessionHandler("test", "/", "example.com"),
		cookieMaskTestPage,
	}
	_, response := falcore.TestWithRequest(
		request,
		&filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test-",
		),
		"contents are "+string(contents),
	)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.True(
		t,
		new_cookie_set,
		"regex didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestCookiemaskNoMasking verifies that a Falcore RequestFilter generated by
// NewCookiemaskFilter behaves as expected when given a request that contains
// only a cookie that is not supposed to be masked. Specifically, this function
// verifies that no error occurred, that the submitted cookie which was not
// supposed to be masked did in fact make it safely to the next RequestFilter in
// the chain, that the next RequestFilter in the chain did not receive an
// indication that a maskable cookie was found, and that a new maskable cookie
// of the expected variant is being set.
func TestCookiemaskNoMasking(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile(
		"^test-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := testCookieGen("foo")
	request.AddCookie(&cookie)

	/* run */
	filter := CookiemaskFilter{
		NewMinSessionHandler("test", "/", "example.com"),
		cookieMaskTestPage,
	}
	_, response := falcore.TestWithRequest(
		request,
		&filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie.String()))
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test-",
		),
		"contents are "+string(contents),
	)

	assert.True(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo-",
		),
		"contents are "+string(contents),
	)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.True(
		t,
		new_cookie_set,
		"regex didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestCookiemaskHandlerError verifies that a Falcore RequestFilter generated
// by NewCookiemaskFilter behaves as expected when the session handler throws
// an error. Specifically, this function verifies that an internal server error
// is sent in the response, and that no cookie is being set.
func TestCookiemaskHandlerError(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile(
		"^test-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := testCookieGen("foo")
	request.AddCookie(&cookie)

	/* run */
	filter := CookiemaskFilter{
		handlerAlwaysErrorsOnSessionRequest,
		cookieMaskTestPage,
	}
	_, response := falcore.TestWithRequest(
		request,
		&filter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.False(
		t,
		new_cookie_set,
		"regex matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestCookiemaskError verifies that a Falcore RequestFilter generated by
// NewCookiemaskFilter behaves as expected when the cookie mask function throws
// an error. Specifically, this function verifies that an internal server error
// is sent in the response, and that no cookie is being set.
func TestCookiemaskError(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile(
		"^test-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := testCookieGen("foo")
	request.AddCookie(&cookie)

	/* run */
	filter := CookiemaskFilter{
		handlerSessionsAlwaysError,
		cookieMaskTestPage,
	}
	_, response := falcore.TestWithRequest(
		request,
		&filter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.False(
		t,
		new_cookie_set,
		"regex matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestCookiemaskMasking verifies that a Falcore RequestFilter generated by
// NewCookiemaskFilter behaves as expected when given a request that contains
// a cookie that is supposed to be masked. Specifically, this function verifies
// that no error occurred, that the submitted cookie was not forwarded to the
// next RequestFilter in the chain, that the next RequestFilter in the chain did
// indeed receive an indication that a maskable cookie was found, and that no
// new maskable cookie of the expected variant is being set.
func TestCookiemaskMasking(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile(
		"^test-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	handler := NewMinSessionHandler("test", "/", "example.com")
	sesh, err := handler.GetSession()
	assert.NoError(t, err)
	fwd, ckes, err := sesh.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	cookie := ckes[0]
	request.AddCookie(cookie)

	/* run */
	filter := CookiemaskFilter{
		handler,
		cookieMaskTestPage,
	}
	_, response := falcore.TestWithRequest(
		request,
		&filter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(
		t,
		strings.Contains(string(contents), cookie.String()),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test-",
		),
		"contents are "+string(contents),
	)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.False(
		t,
		new_cookie_set,
		"regex matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskNoMasking verifies that a chain of two Falcore
// RequestFilters each generated by NewCookiemaskFilter behaves as expected when
// given a request that contains two cookies that are not supposed to be masked.
// Specifically, this function verifies that no error occurred, that both the
// submitted cookies which were not supposed to be masked did in fact make it
// safely to the last RequestFilter in the chain, that the last RequestFilter in
// the chain did not receive an indication that any maskable cookies were found,
// and that a new maskable cookie of each of the two expected variants are being
// set.
func TestDoubleCookiemaskNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^test2-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := testCookieGen("foo1")
	cookie2 := testCookieGen("foo2")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		NewMinSessionHandler(
			"test2",
			"/",
			"example.com",
		),
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		NewMinSessionHandler("test1", "/", "example.com"),
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie1.String()))
	assert.True(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo1-",
		),
		"contents are "+string(contents),
	)
	assert.True(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo2-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(
		t,
		new_cookie1_set,
		"regex1 didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.True(
		t,
		new_cookie2_set,
		"regex2 didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskTopMasking verifies that a chain of two Falcore
// RequestFilters each generated by NewCookiemaskFilter behaves as expected when
// given a request that contains one cookie that is to be masked by the first
// RequestFilter in the chain and a second cookie that is not to be masked by
// either RequestFilter. Specifically, this function verifies that no error
// occurred, that only the submitted cookie which was not supposed to be masked
// did in fact make it safely to the last RequestFilter in the chain, that the
// last RequestFilter in the chain received an indication that the particular
// maskable cookie was found, and that a new maskable cookie of the expected
// variant for the second RequestFilter in the chain is being set.
func TestDoubleCookiemaskTopMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^test2-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	handler1 := NewMinSessionHandler("test1", "/", "example.com")
	sesh1, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd, ckes, err := sesh1.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	cookie1 := ckes[0]
	cookie2 := testCookieGen("foo")
	request.AddCookie(cookie1)
	request.AddCookie(&cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		NewMinSessionHandler(
			"test2",
			"/",
			"example.com",
		),
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		handler1,
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.True(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test1-",
		),
		"contents are "+string(contents),
	)
	assert.True(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(
		t,
		new_cookie1_set,
		"regex1 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.True(
		t,
		new_cookie2_set,
		"regex2 didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskBottomMasking verifies that a chain of two Falcore
// RequestFilters each generated by NewCookiemaskFilter behaves as expected when
// given a request that contains one cookie that is to be masked by the second
// RequestFilter in the chain and a second cookie that is not to be masked by
// either RequestFilter. Specifically, this function verifies that no error
// occurred, that only the submitted cookie which was not supposed to be masked
// did in fact make it safely to the last RequestFilter in the chain, that the
// last RequestFilter in the chain received an indication that the particular
// maskable cookie was found, and that a new maskable cookie of the expected
// variant for the first RequestFilter in the chain is being set.
func TestDoubleCookiemaskBottomMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^test2-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	handler2 := NewMinSessionHandler("test2", "/", "example.com")
	sesh2, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd, ckes, err := sesh2.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	cookie1 := testCookieGen("foo")
	cookie2 := ckes[0]
	request.AddCookie(&cookie1)
	request.AddCookie(cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		handler2,
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		NewMinSessionHandler("test1", "/", "example.com"),
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo-",
		),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test2-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(
		t,
		new_cookie1_set,
		"regex1 didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.False(
		t,
		new_cookie2_set,
		"regex2 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskBothMasking verifies that a chain of two Falcore
// RequestFilters each generated by NewCookiemaskFilter behaves as expected when
// given a request that contains two cookies that are each supposed to be
// masked. Specifically, this function verifies that no error occurred, that
// neither of the submitted cookies which were supposed to be masked did in fact
// make it safely to the last RequestFilter in the chain, that the last
// RequestFilter the chain received indications that both of the maskable
// cookies were found, and that no new maskable cookie is being set.
func TestDoubleCookiemaskBothMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^test2-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	handler1 := NewMinSessionHandler("test1", "/", "example.com")
	sesh1, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd, ckes1, err := sesh1.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	handler2 := NewMinSessionHandler("test2", "/", "example.com")
	sesh2, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd, ckes2, err := sesh2.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	cookie1 := ckes1[0]
	cookie2 := ckes2[0]
	request.AddCookie(cookie1)
	request.AddCookie(cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		handler2,
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		handler1,
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test1-",
		),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test2-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(
		t,
		new_cookie1_set,
		"regex1 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.False(
		t,
		new_cookie2_set,
		"regex2 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskBottomErrorTopNoMasking verifies that a chain of two
// Falcore RequestFilters each generated by NewCookiemaskFilter behaves as
// expected when given a request that contains one cookie that is not to be
// masked and a second cookie that is associated with the second RequestFilter,
// though that RequestFilter will throw an error. Specifically, this function
// verifies that the expected error occurred, that the submitted cookie which
// was not supposed to be masked did in fact make it to the onError
// RequestFilter, and that a new maskable cookie of the expected variant for the
// first RequestFilter in the chain is being set.
func TestDoubleCookiemaskBottomErrorTopNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^error-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := testCookieGen("foo")
	cookie2 := testCookieGen("error")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		handlerSessionsAlwaysError,
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		NewMinSessionHandler("test1", "/", "example.com"),
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test1-",
		),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">error-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(
		t,
		new_cookie1_set,
		"regex1 didn't match any of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.False(
		t,
		new_cookie2_set,
		"regex2 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskBottomErrorTopMasking verifies that a chain of two
// Falcore RequestFilters each generated by NewCookiemaskFilter behaves as
// expected when given a request that contains one cookie that is to be masked
// by the first RequestFilter and a second cookie that is associated with the
// second RequestFilter, though that RequestFilter will throw an error.
// Specifically, this function verifies that the expected error occurred, that
// the submitted cookie which was supposed to be masked did not in fact make it
// to the onError RequestFilter, and that no new maskable cookie is being set.
func TestDoubleCookiemaskBottomErrorTopMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^test1-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^error-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	handler1 := NewMinSessionHandler("test1", "/", "example.com")
	sesh1, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd, ckes1, err := sesh1.CookieMask(nil)
	assert.NoError(t, err)
	assert.Empty(t, fwd)
	cookie1 := ckes1[0]
	cookie2 := testCookieGen("error")
	request.AddCookie(cookie1)
	request.AddCookie(&cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		handlerSessionsAlwaysError,
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		handler1,
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	// BUG(proidiot) issue-#45: context should have been preserved
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">test1-",
		),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">error-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(
		t,
		new_cookie1_set,
		"regex1 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.False(
		t,
		new_cookie2_set,
		"regex2 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

// TestDoubleCookiemaskBottomErrorTopMasking verifies that a chain of two
// Falcore RequestFilters each generated by NewCookiemaskFilter behaves as
// expected when given a request that contains one cookie that is not to be
// masked and a second cookie that is associated with the first RequestFilter,
// though that RequestFilter will throw an error. Specifically, this function
// verifies that the expected error occurred, that the submitted cookie which
// was not supposed to be masked did in fact make it to the onError
// RequestFilter, and that no new maskable cookie is being set.
func TestDoubleCookiemaskTopErrorBottomNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile(
		"^error-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	cookieRegex2 := regexp.MustCompile(
		"^test2-[0-9A-Fa-f]+=[0-9A-Fa-f]*[1-9A-Fa-f][0-9A-Fa-f]+",
	)
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := testCookieGen("error")
	cookie2 := testCookieGen("foo")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	innerFilter := CookiemaskFilter{
		NewMinSessionHandler(
			"test2",
			"/",
			"example.com",
		),
		cookieMaskTestPage,
	}
	outerFilter := CookiemaskFilter{
		handlerSessionsAlwaysError,
		&innerFilter,
	}
	_, response := falcore.TestWithRequest(
		request,
		&outerFilter,
		nil,
	)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">error-",
		),
		"contents are "+string(contents),
	)
	assert.False(
		t,
		strings.Contains(
			string(contents),
			"<li class=\"cke\">foo-",
		),
		"contents are "+string(contents),
	)

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(
		t,
		new_cookie1_set,
		"regex1 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
	assert.False(
		t,
		new_cookie2_set,
		"regex2 matched one of "+
			gostring(response.Header["Set-Cookie"]),
	)
}

func TestCookiemaskFromConfig(t *testing.T) {
	util.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "cookiemaskfilter",
		IsValid: func(i json.Unmarshaler) error {
			var c *CookiemaskFilter

			switch i := i.(type) {
			case *CookiemaskFilter:
				c = i
			default:
				return errors.New(
					fmt.Sprintf(
						"Expecting"+
							" unmarsheled"+
							" resource to be a"+
							" cookiemaskfilter,"+
							" but instead got: %v",
						i,
					),
				)
			}

			switch h := c.Handler.(type) {
			case *MinSessionHandler:
				// do nothing
			default:
				return errors.New(
					fmt.Sprintf(
						"Expecting"+
							" handler to be a"+
							" minsessionhandler,"+
							" but instead got: %v",
						h,
					),
				)
			}

			switch m := c.Masked.(type) {
			case *util.LandingFilter:
				// do nothing
			default:
				return errors.New(
					fmt.Sprintf(
						"Expecting"+
							" masked to be a"+
							" landingfilter,"+
							" but instead got: %v",
						m,
					),
				)
			}

			return nil
		},
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        ``,
				Explanation: "empty config",
			},
			{
				Data:        `99`,
				Explanation: "numeric config",
			},
			{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			{
				Data: `{
					"handler": "notgonnadoit",
					"masked": "wouldntbeprudent"
				}`,
				Explanation: "invalid nested resources",
			},
			{
				Data: `{
					"handler": {
						"type": "minsessionhandler",
						"data": {
							"name": "cmtest",
							"path": "/",
							"domain": "example.com"
						}
					},
					"masked": {
						"type": "minsessionhandler",
						"data": {
							"name": "cmtest",
							"path": "/",
							"domain": "example.com"
						}
					}
				}`,
				Explanation: "masked has wrong type",
			},
			{
				Data: `{
					"handler": {
						"type": "landingfilter",
						"data": {}
					},
					"masked": {
						"type": "landingfilter",
						"data": {}
					}
				}`,
				Explanation: "handler has wrong type",
			},
			{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			{
				Data: `{
				}`,
				Explanation: "empty object",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data: `{
					"handler": {
						"type": "minsessionhandler",
						"data": {
							"name": "cmtest",
							"path": "/",
							"domain": "example.com"
						}
					},
					"masked": {
						"type": "landingfilter",
						"data": {}
					}
				}`,
				Explanation: "valid",
			},
		},
	}
	test.Run(t)
}

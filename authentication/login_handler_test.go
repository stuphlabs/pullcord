package authentication

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"github.com/stuphlabs/pullcord/config"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func getXsrfToken(n *html.Node, xsrfName string) (string, error) {
	if n.Type == html.ElementNode && n.Data == "input" {
		correctType := false
		correctName := false
		xsrfToken := ""

		for _, a := range n.Attr {
			if a.Key == "type" {
				if a.Val == "hidden" {
					correctType = true
				} else {
					break
				}
			} else if a.Key == "name" {
				if a.Val == xsrfName {
					correctName = true
				} else {
					break
				}
			} else if a.Key == "value" {
				xsrfToken = a.Val
				break
			}
		}

		if correctType && correctName {
			if xsrfToken != "" {
				return xsrfToken, nil
			} else if n.FirstChild.Type == html.TextNode &&
				n.FirstChild.Data != ""{
				return n.FirstChild.Data, nil
			} else {
				return "", errors.New(
					"Received empty XSRF token",
				)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if s, e := getXsrfToken(c, xsrfName); e != nil || s != "" {
			return s, e
		}
	}

	return "", nil
}

func TestInitialLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}
	_, response := falcore.TestWithRequest(request, filter, nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	content, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content), "xsrf-testLoginHandler"),
		"content is: " + string(content),
	)
	assert.False(
		t,
		strings.Contains(string(content), "error"),
		"content is: " + string(content),
	)

	assert.NotEmpty(t, response.Header["Set-Cookie"])
}

func TestNoXsrfLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])
	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content), "Invalid credentials"),
		"content is: " + string(content),
	)
}

func TestBadXsrfLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-testLoginHandler", "tacos")
	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}


	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])
	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content), "Invalid credentials"),
		"content is: " + string(content),
	)
}

func TestNoUsernameLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)
	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "Invalid credentials"),
		"content is: " + string(content2),
	)
}

func TestNoPasswordLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)
	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser)
	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "Invalid credentials"),
		"content is: " + string(content2),
	)
}

func TestUsernameArrayLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser)
	postdata2.Add("username-" + handler.Identifier, testUser + "-number2")
	postdata2.Add("password-" + handler.Identifier, testPassword)

	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "Bad request"),
		"content is: " + string(content2),
	)
}

func TestBadUsernameLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser + "-bad")
	postdata2.Add("password-" + handler.Identifier, testPassword)

	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "Invalid credentials"),
		"content is: " + string(content2),
	)
}

func TestBadPasswordLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser)
	postdata2.Add("password-" + handler.Identifier, testPassword + "-bad")

	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "Invalid credentials"),
		"content is: " + string(content2),
	)
}

func TestGoodLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser)
	postdata2.Add("password-" + handler.Identifier, testPassword)

	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	/* check */
	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "logged in"),
		"content is: " + string(content2),
	)
}

func TestPassthruLoginPage(t *testing.T) {
	/* setup */
	testUser := "testUser"
	testPassword := "P@ssword1"

	downstreamFilter := falcore.NewRequestFilter(
		func (request *falcore.Request) *http.Response {
			return falcore.StringResponse(
				request.HttpRequest,
				200,
				nil,
				"<html><body><p>logged in</p></body></html>",
			)
		},
	)
	sessionHandler := NewMinSessionHandler(
		"testSessionHandler",
		"/",
		"example.com",
	)
	hash, err := GetPbkdf2Hash(testPassword, Pbkdf2MinIterations)
	assert.NoError(t, err)
	passwordChecker := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			testUser: hash,
		},
	}

	request1, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	handler := &LoginHandler{
		"testLoginHandler",
		&passwordChecker,
		downstreamFilter,
	}
	filter := &CookiemaskFilter{
		sessionHandler,
		handler,
	}

	_, response1 := falcore.TestWithRequest(request1, filter, nil)
	assert.Equal(t, 200, response1.StatusCode)
	assert.NotEmpty(t, response1.Header["Set-Cookie"])

	content1, err := ioutil.ReadAll(response1.Body)
	assert.NoError(t, err)
	htmlRoot, err := html.Parse(bytes.NewReader(content1))
	assert.NoError(t, err)
	xsrfToken, err := getXsrfToken(htmlRoot, "xsrf-" + handler.Identifier)
	assert.NoError(t, err)

	postdata2 := url.Values{}
	postdata2.Add("xsrf-" + handler.Identifier, xsrfToken)
	postdata2.Add("username-" + handler.Identifier, testUser)
	postdata2.Add("password-" + handler.Identifier, testPassword)

	request2, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(postdata2.Encode()),
	)
	request2.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	assert.NoError(t, err)

	for _, cke := range response1.Cookies() {
		request2.AddCookie(cke)
	}

	_, response2 := falcore.TestWithRequest(request2, filter, nil)

	assert.Equal(t, 200, response2.StatusCode)

	content2, err := ioutil.ReadAll(response2.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content2), "logged in"),
		"content is: " + string(content2),
	)

	request3, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	for _, cke := range response1.Cookies() {
		request3.AddCookie(cke)
	}

	_, response3 := falcore.TestWithRequest(request3, filter, nil)


	/* check */
	assert.Equal(t, 200, response3.StatusCode)

	content3, err := ioutil.ReadAll(response3.Body)
	assert.NoError(t, err)
	assert.True(
		t,
		strings.Contains(string(content3), "logged in"),
		"content is: " + string(content3),
	)
}

func TestLoginHandlerFromConfig(t *testing.T) {
	type testStruct struct {
		validator func(json.Unmarshaler) error
		data string
		serverValidate func(*falcore.Server, error)
	}

	testData := []testStruct {
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			``,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an incomplete validator" +
					" resource should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "loginhandler",
				"data": {}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" loginhandler resource should" +
					" produce an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an incomplete" +
					" loginhandler resource should" +
					" be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "loginhandler",
				"data": {
					"passwordchecker": "notgonnadoit",
					"masked: "wouldntbeprudent"
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an invalid nested" +
					" resource should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				var c *LoginHandler

				switch i := i.(type) {
				case *LoginHandler:
					c = i
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" loginhandler," +
							" but instead got: %v",
							i,
						),
					)
				}

				switch h := c.PasswordChecker.(type) {
				case PasswordChecker:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" handler to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							h,
						),
					)
				}

				switch m := c.Downstream.(type) {
				case falcore.RequestFilter:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" masked to be a" +
							" landingfilter," +
							" but instead got: %v",
							m,
						),
					)
				}

				return nil
			},
			`{
				"type": "loginhandler",
				"data": {
					"passwordchecker": {
						"type": "inmempwdstore",
						"data": {
							"test_user": {
								"salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
								"iterations": 4096,
								"hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
							}
						}
					},
					"downstream": {
						"type": "landingfilter",
						"data": {}
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing only a passing" +
					" validator resource should not" +
					" produce an error. The most likely" +
					" explanation is that the validator" +
					" resource is not passing.",
				)
				assert.NotNil(
					t,
					s,
					"A server created from a config" +
					" containing only a" +
					" passing validator resource should" +
					" not be nil. The most likely" +
					" explanation is that the validator" +
					" resource is not passing.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				var c *LoginHandler

				switch i := i.(type) {
				case *LoginHandler:
					c = i
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" loginhandler," +
							" but instead got: %v",
							i,
						),
					)
				}

				switch h := c.PasswordChecker.(type) {
				case PasswordChecker:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" handler to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							h,
						),
					)
				}

				switch m := c.Downstream.(type) {
				case falcore.RequestFilter:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" masked to be a" +
							" landingfilter," +
							" but instead got: %v",
							m,
						),
					)
				}

				return nil
			},
			`{
				"type": "loginhandler",
				"data": {
					"passwordchecker": {
						"type": "inmempwdstore",
						"data": {
							"test_user": {
								"salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
								"iterations": 4096,
								"hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		                                        }
						}
					},
					"downstream": {
						"type": "inmempwdstore",
						"data": {
							"test_user": {
								"salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
								"iterations": 4096,
								"hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		                                        }
						}
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing a nested" +
					" resource with the wrong type" +
					" should produce an error.",
				)
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing a nested resource with" +
					" the wrong type should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				var c *LoginHandler

				switch i := i.(type) {
				case *LoginHandler:
					c = i
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" loginhandler," +
							" but instead got: %v",
							i,
						),
					)
				}

				switch h := c.PasswordChecker.(type) {
				case PasswordChecker:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" handler to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							h,
						),
					)
				}

				switch m := c.Downstream.(type) {
				case falcore.RequestFilter:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" masked to be a" +
							" landingfilter," +
							" but instead got: %v",
							m,
						),
					)
				}

				return nil
			},
			`{
				"type": "loginhandler",
				"data": {
					"passwordchecker": {
						"type": "landingfilter",
						"data": {}
					},
					"downstream": {
						"type": "landingfilter",
						"data": {}
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing a nested" +
					" resource with the wrong type" +
					" should produce an error.",
				)
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing a nested resource with" +
					" the wrong type should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "loginhandler",
				"data": 42
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" loginhandler resource should" +
					" produce an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an invalid" +
					" loginhandler resource should" +
					" be nil.",
				)
			},
		},
	}

	for _, v := range testData {
		n, e := configutil.GenerateValidator(v.validator)
		assert.NoError(
			t,
			e,
			"Generating a validator resource type should not" +
			" produce an error.",
		)
		assert.NotEqual(
			t,
			n,
			"",
			"A generated validator resource type should not have" +
			" an empty resource type name.",
		)

		s, e := config.ServerFromReader(
			strings.NewReader(
				fmt.Sprintf(
					`{
						"resources": {
							"validator": {
								"type": "%s",
								"data": %s
							}
						},
						"pipeline": ["validator"],
						"port": 80
					}`,
					n,
					v.data,
				),
			),
		)
		v.serverValidate(s, e)
	}
}

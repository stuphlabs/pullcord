package authentication

import (
	"bytes"
	"errors"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
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
	test := configutil.ConfigTest{
		ResourceType: "loginhandler",
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "",
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: "{}",
				Explanation: "empty object",
			},
			configutil.ConfigTestData{
				Data: "null",
				Explanation: "null config",
			},
			configutil.ConfigTestData{
				Data: "42",
				Explanation: "numeric config",
			},
			configutil.ConfigTestData{
				Data: `{
					"passwordchecker": "notgonnadoit",
					"masked: "wouldntbeprudent"
				}`,
				Explanation: "invalid subtypes",
			},
			configutil.ConfigTestData{
				Data: `{
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
				}`,
				Explanation: "bad downstream",
			},
			configutil.ConfigTestData{
				Data: `{
					"passwordchecker": {
						"type": "landingfilter",
						"data": {}
					},
					"downstream": {
						"type": "landingfilter",
						"data": {}
					}
				}`,
				Explanation: "bad passwordchecker",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: `{
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
				}`,
				Explanation: "good config",
			},
		},
	}

	test.Run(t)
}

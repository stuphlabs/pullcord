package pullcord

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dustin/randbo"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const cookieNameRandSize = 16
const cookieValueRandSize = 64

var exampleCookieValueRegex = regexp.MustCompile("^[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}$")
var randgen = randbo.New()

func gostring(i interface{}) string {
	return fmt.Sprintf("%#v", i)
}

var cookieMaskTestPage = falcore.NewRequestFilter(
	func(req *falcore.Request) *http.Response {
		var content = "<html><body><h1>cookies</h1><ul>"
		for _, cke := range req.HttpRequest.Cookies() {
			content += "<li>" + cke.String() + "</li>"
		}
		content += "</ul><h1>context</h1><ul>"
		for name, ctx := range req.Context {
			content += "<li>" + name + ": " + gostring(ctx)
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

var errorPage = falcore.NewRequestFilter(
	func(req *falcore.Request) *http.Response {
		return falcore.StringResponse(
			req.HttpRequest,
			500,
			nil,
			"<html><body><p>internal server error</p></body></html>",
		)
	},
)

func exampleCookieGen(variant string) (result http.Cookie) {
	nbytes := make([]byte, cookieNameRandSize)
	vbytes := make([]byte, cookieValueRandSize)
	randgen.Read(nbytes)
	randgen.Read(vbytes)
	result.Name = "example-" + variant + "-" + hex.EncodeToString(nbytes)
	result.Value = hex.EncodeToString(vbytes)
	return result
}

func exampleCookieMaskGen(variant string) func([]*http.Cookie) ([]*http.Cookie, []*http.Cookie, map[string]interface{}, error) {
	var exampleCookieNameRegex = regexp.MustCompile("^example-" + variant + "-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}$")

	return func(in_ckes []*http.Cookie) (passthru_ckes []*http.Cookie, set_ckes []*http.Cookie, ctx map[string]interface{}, err error) {
		if variant == "error" {
			err = errors.New("got an error")
		}
		ctx = make(map[string]interface{})
		found := false
		for _, cke := range in_ckes {
			if exampleCookieNameRegex.MatchString(cke.Name) && exampleCookieValueRegex.MatchString(cke.Value) {
				found = true
				ctx["example-"+variant+"-cookie-found"] = "true"
			} else {
				passthru_ckes = append(passthru_ckes, cke)
			}
		}
		if !found {
			set_ckes = make([]*http.Cookie, 1)
			sub_cke := exampleCookieGen(variant)
			set_ckes[0] = &sub_cke
			ctx["example-"+variant+"-cookie-found"] = "false"
		}

		return passthru_ckes, set_ckes, ctx, err
	}
}

func TestCookiemaskCookieless(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile("^example-test-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test"), cookieMaskTestPage, errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), "example-test-cookie-found: \"false\""), "contents are "+string(contents))

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.True(t, new_cookie_set, "regex didn't match any of "+gostring(response.Header["Set-Cookie"]))
}

func TestCookiemaskNoMasking(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile("^example-test-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := exampleCookieGen("foo")
	request.AddCookie(&cookie)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test"), cookieMaskTestPage, errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie.String()))
	assert.True(t, strings.Contains(string(contents), "example-test-cookie-found: \"false\""), "contents are "+string(contents))

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.True(t, new_cookie_set, "regex didn't match any of "+gostring(response.Header["Set-Cookie"]))
}

func TestCookiemaskError(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile("^example-test-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := exampleCookieGen("foo")
	request.AddCookie(&cookie)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("error"), cookieMaskTestPage, errorPage), nil)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.False(t, new_cookie_set, "regex didn't match any of "+gostring(response.Header["Set-Cookie"]))
}

func TestCookiemaskMasking(t *testing.T) {
	/* setup */
	cookieRegex := regexp.MustCompile("^example-test-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie := exampleCookieGen("test")
	request.AddCookie(&cookie)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test"), cookieMaskTestPage, errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie.String()))
	assert.True(t, strings.Contains(string(contents), "example-test-cookie-found: \"true\""), "contents are "+string(contents))

	new_cookie_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex.MatchString(cke_str) {
			new_cookie_set = true
		}
	}
	assert.False(t, new_cookie_set, "regex matched one of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-test2-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("foo1")
	cookie2 := exampleCookieGen("foo2")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("test2"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie1.String()))
	assert.True(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(t, strings.Contains(string(contents), "example-test1-cookie-found: \"false\""), "contents are "+string(contents))
	assert.True(t, strings.Contains(string(contents), "example-test2-cookie-found: \"false\""), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(t, new_cookie1_set, "regex1 didn't match any of "+gostring(response.Header["Set-Cookie"]))
	assert.True(t, new_cookie2_set, "regex2 didn't match any of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskTopMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-test2-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("test1")
	cookie2 := exampleCookieGen("foo")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("test2"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.True(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(t, strings.Contains(string(contents), "example-test1-cookie-found: \"true\""), "contents are "+string(contents))
	assert.True(t, strings.Contains(string(contents), "example-test2-cookie-found: \"false\""), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(t, new_cookie1_set, "regex1 matched one of "+gostring(response.Header["Set-Cookie"]))
	assert.True(t, new_cookie2_set, "regex2 didn't match any of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskBottomMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-test2-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("foo")
	cookie2 := exampleCookieGen("test2")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("test2"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(t, strings.Contains(string(contents), "example-test1-cookie-found: \"false\""), "contents are "+string(contents))
	assert.True(t, strings.Contains(string(contents), "example-test2-cookie-found: \"true\""), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(t, new_cookie1_set, "regex1 didn't match any of "+gostring(response.Header["Set-Cookie"]))
	assert.False(t, new_cookie2_set, "regex2 matched one of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskBothMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-test2-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("test1")
	cookie2 := exampleCookieGen("test2")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("test2"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 200, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.True(t, strings.Contains(string(contents), "example-test1-cookie-found: \"true\""), "contents are "+string(contents))
	assert.True(t, strings.Contains(string(contents), "example-test2-cookie-found: \"true\""), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(t, new_cookie1_set, "regex1 matched one of "+gostring(response.Header["Set-Cookie"]))
	assert.False(t, new_cookie2_set, "regex2 matched one of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskBottomErrorTopNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-error-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("foo")
	cookie2 := exampleCookieGen("error")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("error"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(t, strings.Contains(string(contents), "example-test1-cookie-found"), "contents are "+string(contents))
	assert.False(t, strings.Contains(string(contents), "example-error-cookie-found"), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.True(t, new_cookie1_set, "regex1 didn't match any of "+gostring(response.Header["Set-Cookie"]))
	assert.False(t, new_cookie2_set, "regex2 matched one of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskBottomErrorTopMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-test1-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-error-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("test1")
	cookie2 := exampleCookieGen("error")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("test1"), NewCookiemaskFilter(exampleCookieMaskGen("error"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(t, strings.Contains(string(contents), "example-test1-cookie-found"), "contents are "+string(contents))
	assert.False(t, strings.Contains(string(contents), "example-error-cookie-found"), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(t, new_cookie1_set, "regex1 matched one of "+gostring(response.Header["Set-Cookie"]))
	assert.False(t, new_cookie2_set, "regex2 matched one of "+gostring(response.Header["Set-Cookie"]))
}

func TestDoubleCookiemaskTopErrorBottomNoMasking(t *testing.T) {
	/* setup */
	cookieRegex1 := regexp.MustCompile("^example-error-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	cookieRegex2 := regexp.MustCompile("^example-test2-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}=[0-9A-Fa-f]{" + strconv.Itoa(cookieValueRandSize*2) + "}")
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	cookie1 := exampleCookieGen("error")
	cookie2 := exampleCookieGen("foo")
	request.AddCookie(&cookie1)
	request.AddCookie(&cookie2)

	/* run */
	_, response := falcore.TestWithRequest(request, NewCookiemaskFilter(exampleCookieMaskGen("error"), NewCookiemaskFilter(exampleCookieMaskGen("test2"), cookieMaskTestPage, errorPage), errorPage), nil)

	/* check */
	assert.Equal(t, 500, response.StatusCode)

	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(contents), cookie1.String()))
	assert.False(t, strings.Contains(string(contents), cookie2.String()))
	assert.False(t, strings.Contains(string(contents), "example-test1-cookie-found"), "contents are "+string(contents))
	assert.False(t, strings.Contains(string(contents), "example-error-cookie-found"), "contents are "+string(contents))

	new_cookie1_set := false
	new_cookie2_set := false
	for _, cke_str := range response.Header["Set-Cookie"] {
		if cookieRegex1.MatchString(cke_str) {
			new_cookie1_set = true
		} else if cookieRegex2.MatchString(cke_str) {
			new_cookie2_set = true
		}
	}
	assert.False(t, new_cookie1_set, "regex1 matched one of "+gostring(response.Header["Set-Cookie"]))
	assert.False(t, new_cookie2_set, "regex2 matched one of "+gostring(response.Header["Set-Cookie"]))
}

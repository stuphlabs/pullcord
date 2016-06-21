package pullcord

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"
)

const minSessionCookieNameRandSize = 32
const minSessionCookieValueRandSize = 128
const minSessionCookieMaxAge = 2 * 60 * 60

type session struct {
	cvalue string
	data   map[string]interface{}
}

// MinSessionHandler is a somewhat minimalist form of a SessionHandler.
//
// The cookies are held in a map from the cookie name to an internal session
// structure that includes the expected cookie value and the current session
// data (visible to the RequestFilter chain as the context). Any time a
// RequestFilter attempts to write to their view of the context, they will
// actually be writing on the MinSessionHandler's original copy of this data. As
// such, this implementation is not particularly safe and should not be seen as
// a viable long-term solution for a SessionHandler.
type MinSessionHandler struct {
	table  map[string]session
	name   string
	path   string
	domain string
}

// CookieMask for the MinSessionHandler is an implementation of the CookieMask
// function required by all SessionHandler derivatives.
func (handler *MinSessionHandler) CookieMask(incomingCookies []*http.Cookie) (
	forwardedCookies []*http.Cookie,
	setCookies []*http.Cookie,
	context map[string]interface{},
	err error,
) {
	new_cookie_needed := true
	cookieNameRegex := regexp.MustCompile(
		"^" +
		handler.name +
		"-[0-9A-Fa-f]{" +
		strconv.Itoa(minSessionCookieNameRandSize * 2) +
		"}$",
	)

	for _, cookie := range incomingCookies {
		if cookieNameRegex.MatchString(cookie.Name) {
			session, cookie_name_legit := handler.table[cookie.Name]
			if cookie_name_legit &&
			    len(cookie.Value) > 0 &&
			    cookie.Value == session.cvalue {
				new_cookie_needed = false
				context = session.data
			} else {
				if cookie_name_legit {
					delete(handler.table, cookie.Name)
				}

				var bad_cookie http.Cookie
				bad_cookie.Name = cookie.Name
				bad_cookie.Value = cookie.Value
				bad_cookie.MaxAge = -1
				setCookies = append(setCookies, &bad_cookie)
			}
		} else {
			forwardedCookies = append(forwardedCookies, cookie)
		}
	}

	if new_cookie_needed {
		nbytes := make([]byte, minSessionCookieNameRandSize)
		vbytes := make([]byte, minSessionCookieValueRandSize)

		cookie_name := ""
		name_gen_needed := true
		for name_gen_needed {
			_, err = rand.Read(nbytes)
			if err != nil {
				return nil, setCookies, nil, err
			}

			cookie_name = handler.name +
				"-" +
				hex.EncodeToString(nbytes)

			_, collides_legit := handler.table[cookie_name]
			if !collides_legit {
				name_gen_needed = false

				for _, cookie := range setCookies {
					if cookie.Name == cookie_name {
						name_gen_needed = true
						break
					}
				}
			}
		}

		_, err = rand.Read(vbytes)
		if err != nil {
			return nil, setCookies, nil, err
		}

		var new_cookie http.Cookie
		new_cookie.Name = cookie_name
		new_cookie.Value = hex.EncodeToString(vbytes)
		new_cookie.Path = handler.path
		new_cookie.Domain = handler.domain
		new_cookie.MaxAge = minSessionCookieMaxAge
		new_cookie.Secure = true
		new_cookie.HttpOnly = true
		setCookies = append(setCookies, &new_cookie)

		handler.table[new_cookie.Name] = session{
			cvalue: new_cookie.Value,
			data: make(map[string]interface{}),
		}
		session := handler.table[new_cookie.Name]

		context = session.data
	}

	return forwardedCookies, setCookies, context, nil
}

// NewMinSessionHandler constructs a new MinSessionHandler given a unique name
// (which will be given to all the cookies), and a path and domain (the two of
// which will simply be sent to the browser along with the cookie, and otherwise
// have no bearing on functionality).
func NewMinSessionHandler(
	handlerName string,
	handlerPath string,
	handlerDomain string,
) *MinSessionHandler {
	var result MinSessionHandler
	result.table = make(map[string]session)
	result.name = handlerName
	result.path = handlerPath
	result.domain = handlerDomain

	return &result
}
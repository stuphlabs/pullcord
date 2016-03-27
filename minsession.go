package pullcord

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"
)

const cookieNameRandSize = 32
const cookieValueRandSize = 128
const cookieMaxAge = 2 * 60 * 60

type session struct {
	cvalue string
	data   map[string]interface{}
}

type MinSessionHandler struct {
	table  map[string]session
	name   string
	path   string
	domain string
}

func (handler MinSessionHandler) CookieMask(incoming_cookies []*http.Cookie) (forwarded_cookies []*http.Cookie, set_cookies []*http.Cookie, context map[string]interface{}, err error) {
	new_cookie_needed := true
	//initialized := false
	//context = make(map[string]interface{})
	cookieNameRegex := regexp.MustCompile("^" + handler.name + "-[0-9A-Fa-f]{" + strconv.Itoa(cookieNameRandSize*2) + "}$")

	for _, cookie := range incoming_cookies {
		if cookieNameRegex.MatchString(cookie.Name) {
			session, cookie_name_legit := handler.table[cookie.Name]
			if cookie_name_legit && len(cookie.Value) > 0 && cookie.Value == session.cvalue {
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
				set_cookies = append(set_cookies, &bad_cookie)
			}
		} else {
			forwarded_cookies = append(forwarded_cookies, cookie)
		}
	}

	if new_cookie_needed {
		nbytes := make([]byte, cookieNameRandSize)
		vbytes := make([]byte, cookieValueRandSize)

		cookie_name := ""
		name_gen_needed := true
		for name_gen_needed {
			_, err = rand.Read(nbytes)
			if err != nil {
				return nil, set_cookies, nil, err
			}

			cookie_name = handler.name + "-" + hex.EncodeToString(nbytes)

			_, collides_legit := handler.table[cookie_name]
			if !collides_legit {
				name_gen_needed = false

				for _, cookie := range set_cookies {
					if cookie.Name == cookie_name {
						name_gen_needed = true
						break
					}
				}
			}
		}

		_, err = rand.Read(vbytes)
		if err != nil {
			return nil, set_cookies, nil, err
		}

		var new_cookie http.Cookie
		new_cookie.Name = cookie_name
		new_cookie.Value = hex.EncodeToString(vbytes)
		new_cookie.Path = handler.path
		new_cookie.Domain = handler.domain
		new_cookie.MaxAge = cookieMaxAge
		new_cookie.Secure = true
		new_cookie.HttpOnly = true
		set_cookies = append(set_cookies, &new_cookie)

		handler.table[new_cookie.Name] = session{cvalue: new_cookie.Value, data: make(map[string]interface{})}
		session := handler.table[new_cookie.Name]

		//context[new_cookie.Name] = &(session.data)
		context = session.data
	}

	return forwarded_cookies, set_cookies, context, nil
}

func NewMinSessionHandler(handler_name string, handler_path string, handler_domain string) SessionHandler {
	var result MinSessionHandler
	result.table = make(map[string]session)
	result.name = handler_name
	result.path = handler_path
	result.domain = handler_domain

	return result
}

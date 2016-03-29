package pullcord

import (
	"net/http"
)

type SessionHandler interface {
	CookieMask(inCookies []*http.Cookie) (fwdCookies, setCookies []*http.Cookie, context map[string]interface{}, err error)
}

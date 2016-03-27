package pullcord

import (
	"net/http"
)

type SessionHandler interface {
	CookieMask([]*http.Cookie) ([]*http.Cookie, []*http.Cookie, map[string]interface{}, error)
}

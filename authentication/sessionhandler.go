package authentication

import (
	// "github.com/stuphlabs/pullcord"
	"net/http"
)

// SessionHandler is an abstract interface describing any system which tracks a
// user's state across requests using a combination of cookies, some kind of
// data store, and a way for various components of the RequestFilter chain to
// add data to that state information.
//
// CookieMask is a function that takes a list of cookies received as part of a
// request and masks any of those cookies which may belong to the SessionHandler
// from all later RequestFilters and external services. All cookies which are
// not being masked will be returned as part of the forward cookies. If any
// cookies which belong to the SessionHandler are found, then any session data
// will be available in the returned context map. It is possible that the
// SessionHandler will also create new cookies which are to be added as
// "Set-cookie" headers in response to be sent back to the requester.
type SessionHandler interface {
	CookieMask(inCookies []*http.Cookie) (
		fwdCookies []*http.Cookie,
		setCookies []*http.Cookie,
		context map[string]interface{},
		err error,
	)
}

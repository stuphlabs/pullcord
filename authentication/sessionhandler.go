package authentication

import (
	"github.com/proidiot/gone/errors"
	// "github.com/stuphlabs/pullcord"
	"net/http"
)

// NoSuchSessionValueError is the error constant that would be returned in the
// event that a given key requested from a Session using its GetValue function
// is not present in the session.
const NoSuchSessionValueError = errors.New(
	"The session does not have a the requested value.",
)

// Session is an abstract interface that allows session-level data to not only
// be read at any point, but also be set at any point as well.
//
// GetValue is a function that gives read access to the implicit key-value
// store of the Session.
//
// SetValue is a function that gives write access to the implicit key-value
// store of the Session.
//
// CookieMask is a function that takes a list of cookies received as part of a
// request and masks any of those cookies which may belong to the SessionHandler
// from all later RequestFilters and external services. All cookies which are
// not being masked will be returned as part of the forward cookies. If any
// cookies which belong to the SessionHandler are found, then the returned
// Session will have the data which had been set on the equivalent Session
// previously (as one would expect from a session). It is possible that the
// SessionHandler will also create new cookies which are to be added as
// "Set-cookie" headers in response to be sent back to the requester.
type Session interface {
	GetValue(key string) (value interface{}, err error)
	SetValue(key string, value interface{}) (err error)
	CookieMask(inCookies []*http.Cookie) (
		fwdCookies []*http.Cookie,
		setCookies []*http.Cookie,
		err error,
	)
}

// SessionHandler is an abstract interface describing any system which tracks a
// user's state across requests using a combination of cookies, some kind of
// data store, and a way for various components of the RequestFilter chain to
// add data to that state information.
//
// GetSession is a function that returns a Session implementation that will be
// managed by the SessionHandler. As the details of determining whether or not
// two Sessions are equivalent is the purview of a SessionHandler, the details
// of such possible comparisons and their implications are abstracted away by
// the Session's CookieMask call.
type SessionHandler interface {
	GetSession() (sesh Session, err error)
}

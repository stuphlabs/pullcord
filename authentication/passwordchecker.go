package authentication

import (
	// "github.com/stuphlabs/pullcord"
)

// PasswordChecker is an abstract interface describing any system which can
// check password (and presumably also store them in some way?).
//
// CheckPassword is a function that takes some kind of identifier (likely a
// username, but perhaps something else) and a raw password and determines
// whether this combination is for an authorized entity, in which case an error
// is not returned. What other errors may be returned is not specified at this
// level, but it is possible that different errors indicating a bad password, a
// bad username, too many tries, or even a broken password checker could be
// returned. It is often not advisable to indicate to a user which particular
// reason a login failed, but it makes sense that an interface providing such
// specificity is a good idea, especially given that the granularity of errors
// could potentially be controlled through some configuration option.
type PasswordChecker interface {
	CheckPassword(identifier string, password string) (err error)
}

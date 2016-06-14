package pullcord

import (
	"github.com/fitstar/falcore"
	"net/http"
)

// NewConditionalFilter generates a Falcore RequesFilter that will apply one of
// two other RequestFilters based on the boolean return value of some function,
// or else applies a third RequestFilter in the event of an error.
//
// In other words, this function allows different web page generation logic to
// be triggered based on the result of some test. For example, if a remote
// service is currently inaccessible, a web page could be generated to allow
// certain configurable actions to be taken, while an opaque web forwarder
// could be used if the remote service is accessible.
func NewConditionalFilter(test func(req *falcore.Request) (bool, error), onTrue falcore.RequestFilter, onFalse falcore.RequestFilter, onError falcore.RequestFilter) falcore.RequestFilter {
	return falcore.NewRequestFilter(
		func(req *falcore.Request) *http.Response {
			test_true, err := test(req)
			if err != nil {
				return onError.FilterRequest(req)
			} else if test_true {
				return onTrue.FilterRequest(req)
			} else {
				return onFalse.FilterRequest(req)
			}
		},
	)
}

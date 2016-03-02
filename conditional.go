package pullcord

import (
	"github.com/fitstar/falcore"
	"net/http"
)

func NewConditionalFilter(test func(req *falcore.Request) bool, onTrue falcore.RequestFilter, onFalse falcore.RequestFilter) falcore.RequestFilter {
	return falcore.NewRequestFilter(
		func(req *falcore.Request) *http.Response {
			if test(req) {
				return onTrue.FilterRequest(req)
			} else {
				return onFalse.FilterRequest(req)
			}
		},
	)
}

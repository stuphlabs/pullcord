package pullcord

import (
	"github.com/fitstar/falcore"
	"net/http"
)

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

package pullcord

import (
	"fmt"
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
func NewConditionalFilter(
	test func(req *falcore.Request) (bool, error),
	onTrue falcore.RequestFilter,
	onFalse falcore.RequestFilter,
	onError falcore.RequestFilter,
) falcore.RequestFilter {
	log().Debug("registering a new conditional filter")

	return falcore.NewRequestFilter(
		func(req *falcore.Request) *http.Response {
			log().Debug("running conditional filter")

			test_true, err := test(req)
			log().Debug("conditional filter test completed")

			if err != nil {
				log().Info(
					fmt.Sprintf(
						"conditional filter's test" +
						" returned an error: %v",
						err,
					),
				)

				return onError.FilterRequest(req)
			} else if test_true {
				log().Info(
					"conditional filter's test returned" +
					" true",
				)

				return onTrue.FilterRequest(req)
			} else {
				log().Info(
					"conditional filter's test returned" +
					" false",
				)

				return onFalse.FilterRequest(req)
			}
		},
	)
}

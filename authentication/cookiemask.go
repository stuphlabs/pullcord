package authentication

import (
	"bytes"
	"fmt"
	"github.com/fitstar/falcore"
	// "github.com/stuphlabs/pullcord"
	"net/http"
	"strings"
)

// NewCookiemaskFilter generates a Falcore RequestFilter that will apply a
// "cookie mask" function before forwarding the (possibly modified) request to
// the next RequestFilter in the chain (which may ultimately lead to a proxy).
//
// As HTTP clients typically have internal cookie managers sophisticated enough
// to send multiple cookies (with possibly different scopes or properties) with
// each request, we should be able to use our own cookies to keep track of
// state without exposing those cookies to any external application to which we
// may be proxying (since we have no guarantee that the presence of such
// cookies would not effect the behavior of such an external application). To
// that end, this function is given a cookie masking function which checks for
// any cookies intended to be received by the particular function, and
// potentially retrieves session information from some session handler to be
// forwarded down the RequestFilter chain as part of the context map. The
// cookie filter function returns the array of cookies which are not masked
// (and are to be allowed down the RequestFilter chain along with the rest of
// the original request), plus an array of any new cookies associated with this
// particular filter which are to be sent to the browser for storage (which
// should happen seamlessly as part of the next response), plus any session data
// which is to be forwarded down the RequestFilter chain as part of the context
// map, and of course a possible error (which will cause the onError
// RequestFilter chain to receive the context instead, with any new cookies
// still being added to the response, even though the onError chain will
// receive no cookies as part of the request).
func NewCookiemaskFilter(
	maskCookies func(inCookies []*http.Cookie) (
		fwdCookies []*http.Cookie,
		setCookies []*http.Cookie,
		ctx map[string]interface{},
		err error,
	),
	masked falcore.RequestFilter,
	onError falcore.RequestFilter,
) falcore.RequestFilter {
	log().Debug("registering a new cookiemask filter")

	return falcore.NewRequestFilter(
		func(req *falcore.Request) *http.Response {
			log().Debug("running cookiemask filter")

			passthru_ckes, set_ckes, ctx_ovrd, err := maskCookies(
				req.HttpRequest.Cookies(),
			)

			ctx_keys_buffer := new(bytes.Buffer)
			for key, val := range ctx_ovrd {
				ctx_keys_buffer.WriteString("\"" + key + "\",")
				req.Context[key] = val
			}
			log().Debug(
				fmt.Sprintf(
					"cookiemask has set these context" +
					" keys: [%s]",
					ctx_keys_buffer.String(),
				),
			)


			var resp *http.Response
			if err != nil {
				log().Err(
					fmt.Sprintf(
						"cookiemask filter's" +
						" maskCookies call returned" +
						" an error: %v",
						err,
					),
				)

				resp = onError.FilterRequest(req)
			} else {
				cke_keys_buffer := new(bytes.Buffer)
				ckes_str := make([]string, len(passthru_ckes))
				for n, cke := range passthru_ckes {
					cke_keys_buffer.WriteString(
						"\"" + cke.Name + "\",",
					)
					ckes_str[n] = cke.String()
				}
				log().Debug(
					fmt.Sprintf(
						"cookiemask forwarding" +
						" cookies with these keys:" +
						" [%s]",
						cke_keys_buffer.String(),
					),
				)

				req.HttpRequest.Header.Set(
					"Cookie",
					strings.Join(ckes_str, "; "),
				)

				log().Info(
					"request has run through cookiemask," +
					" now forwarding to next filter",
				)
				resp = masked.FilterRequest(req)
			}

			set_cke_keys_buffer := new(bytes.Buffer)
			for _, cke := range set_ckes {
				set_cke_keys_buffer.WriteString(
					"\"" + cke.Name + "\",",
				)
				resp.Header.Add("Set-Cookie", cke.String())
			}
			log().Debug(
				fmt.Sprintf(
					"cookiemask sending back with the" +
					" response new cookies with these" +
					" keys: [%s]",
					set_cke_keys_buffer.String(),
				),
			)

			return resp
		},
	)
}

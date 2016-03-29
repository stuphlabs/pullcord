package pullcord

import (
	"github.com/fitstar/falcore"
	"net/http"
	"strings"
)

func NewCookiemaskFilter(maskCookies func(inCookies []*http.Cookie) (fwdCookies, setCookies []*http.Cookie, ctx map[string]interface{}, err error), masked falcore.RequestFilter, onError falcore.RequestFilter) falcore.RequestFilter {
	return falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
		passthru_ckes, set_ckes, ctx_ovrd, err := maskCookies(req.HttpRequest.Cookies())

		for key, val := range ctx_ovrd {
			req.Context[key] = val
		}

		var resp *http.Response
		if err != nil {
			resp = onError.FilterRequest(req)
		} else {
			ckes_str := make([]string, len(passthru_ckes))
			for n, cke := range passthru_ckes {
				ckes_str[n] = cke.String()
			}

			req.HttpRequest.Header.Set("Cookie", strings.Join(ckes_str, "; "))

			resp = masked.FilterRequest(req)
		}

		for _, cke := range set_ckes {
			resp.Header.Add("Set-Cookie", cke.String())
		}

		return resp
	})
}

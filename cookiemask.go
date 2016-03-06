package pullcord

import (
	/*"bytes"*/
	"github.com/fitstar/falcore"
	/*"math"*/
	/*"math/rand"*/
	"net/http"
	/*"regexp"*/
	"strings"
)

/*
var hex = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}

const cookieMaskTime = 5 * 60

const cookieNameLength = 16
const cookieValueLength = 32

var cookieNameRegex = regexp.MustCompile("/^cookieMasked-[0-9A-F]+$/")
var cookieValueRegex = regexp.MustCompile("/^[0-9A-F]+$/")

func genCookieMask() string {
	var cke http.Cookie
	var nbuf bytes.Buffer
	var vbuf bytes.Buffer

	for i := 0; i < int(math.Ceil(cookieNameLength/8.0)); i++ {
		temp := rand.Uint32()
		nbuf.WriteString(hex[temp>>30])
		nbuf.WriteString(hex[(temp>>28)&0x0F])
		nbuf.WriteString(hex[(temp>>26)&0x0F])
		nbuf.WriteString(hex[(temp>>24)&0x0F])
		nbuf.WriteString(hex[(temp>>22)&0x0F])
		nbuf.WriteString(hex[(temp>>20)&0x0F])
		nbuf.WriteString(hex[(temp>>18)&0x0F])
		nbuf.WriteString(hex[(temp>>16)&0x0F])
		nbuf.WriteString(hex[(temp>>14)&0x0F])
		nbuf.WriteString(hex[(temp>>12)&0x0F])
		nbuf.WriteString(hex[(temp>>10)&0x0F])
		nbuf.WriteString(hex[(temp>>8)&0x0F])
		nbuf.WriteString(hex[(temp>>6)&0x0F])
		nbuf.WriteString(hex[(temp>>4)&0x0F])
		nbuf.WriteString(hex[(temp>>2)&0x0F])
		nbuf.WriteString(hex[temp&0x0F])
	}

	for i := 0; i < int(math.Ceil(cookieValueLength/8.0)); i++ {
		temp := rand.Uint32()
		vbuf.WriteString(hex[temp>>30])
		vbuf.WriteString(hex[(temp>>28)&0x0F])
		vbuf.WriteString(hex[(temp>>26)&0x0F])
		vbuf.WriteString(hex[(temp>>24)&0x0F])
		vbuf.WriteString(hex[(temp>>22)&0x0F])
		vbuf.WriteString(hex[(temp>>20)&0x0F])
		vbuf.WriteString(hex[(temp>>18)&0x0F])
		vbuf.WriteString(hex[(temp>>16)&0x0F])
		vbuf.WriteString(hex[(temp>>14)&0x0F])
		vbuf.WriteString(hex[(temp>>12)&0x0F])
		vbuf.WriteString(hex[(temp>>10)&0x0F])
		vbuf.WriteString(hex[(temp>>8)&0x0F])
		vbuf.WriteString(hex[(temp>>6)&0x0F])
		vbuf.WriteString(hex[(temp>>4)&0x0F])
		vbuf.WriteString(hex[(temp>>2)&0x0F])
		vbuf.WriteString(hex[temp&0x0F])
	}

	cke.Name = "cookieMasked-" + nbuf.String()
	cke.Value = vbuf.String()
	cke.Secure = true
	cke.MaxAge = cookieMaskTime

	return cke.String()
}

func isValidCookieMask(cookie *http.Cookie) bool {
	return cookieNameRegex.MatchString(cookie.Name) && cookieValueRegex.MatchString(cookie.Value)
}

func NewCookiemask(masked falcore.RequestFilter) falcore.RequestFilter {
	return falcore.NewRequestFilter(func(req *falcore.Request) *http.Response {
		var buf bytes.Buffer
		first := true
		found := false
		ckes := req.HttpRequest.Cookies()

		for _, cke := range ckes {
			if isValidCookieMask(cke) {
				found = true
			} else {
				if !first {
					buf.WriteString("; ")
				}
				buf.WriteString(cke.String())
				first = false
			}
		}

		req.HttpRequest.Header.Set("Cookie", buf.String())

		resp := masked.FilterRequest(req)

		if !found {
			resp.Header.Add("Set-Cookie", genCookieMask())
		}

		return resp
	})
}
*/

func NewCookiemaskFilter(maskCookies func(ckes []*http.Cookie) (passthru_cookies []*http.Cookie, set_cookies []*http.Cookie, ctx map[string]interface{}, err error), masked falcore.RequestFilter, onError falcore.RequestFilter) falcore.RequestFilter {
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

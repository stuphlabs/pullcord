package authentication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/util"
)

// CookiemaskFilter is a falcore.RequestFilter that will apply a "cookie mask"
// function before forwarding the (possibly modified) request to the next
// RequestFilter in the chain (which may ultimately lead to a proxy).
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
type CookiemaskFilter struct {
	Handler SessionHandler
	Masked  http.Handler
}

func init() {
	config.RegisterResourceType(
		"cookiemaskfilter",
		func() json.Unmarshaler {
			return new(CookiemaskFilter)
		},
	)
}

func (f *CookiemaskFilter) UnmarshalJSON(input []byte) error {
	var t struct {
		Handler config.Resource
		Masked  config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	} else {
		h := t.Handler.Unmarshaled
		switch h := h.(type) {
		case SessionHandler:
			f.Handler = h
		default:
			log.Err(
				"Resource described by \"Handler\" is not a" +
					" SessionHandler",
			)
			return config.UnexpectedResourceType
		}

		m := t.Masked.Unmarshaled
		switch m := m.(type) {
		case http.Handler:
			f.Masked = m
		default:
			log.Err(
				"Resource described by \"Masked\" is not a" +
					" net/http.Handler",
			)
			return config.UnexpectedResourceType
		}

		return nil
	}
}

type cookieAppender struct {
	ckes []*http.Cookie
	w http.ResponseWriter
	hdrs http.Header
	started bool
}

func (ca cookieAppender) writeHeaders() {
	for key, vals := range ca.hdrs {
		for _, val := range vals {
			ca.w.Header().Add(key, val)
		}
	}

	for _, cke := range ca.ckes {
		ca.w.Header().Add("Set-Cookie", cke.String())
	}
}

func (ca cookieAppender) writeTrailers() {
	for key, vals := range ca.hdrs {
		if strings.HasPrefix(key, http.TrailerPrefix) {
			for _, val := range vals {
				ca.w.Header().Add(key, val)
			}
		}
	}

	if keys, present := ca.hdrs["Trailer"]; present {
		for _, key := range keys {
			for _, val := range ca.hdrs[key] {
				ca.w.Header().Add(key, val)
			}
		}
	}
}

func (ca cookieAppender) Header() http.Header {
	return ca.hdrs
}

func (ca cookieAppender) Write(d []byte) (int, error) {
	if !ca.started {
		ca.started = true
		ca.writeHeaders()
	}
	return ca.w.Write(d)
}

func (ca cookieAppender) WriteHeader(statusCode int) {
	if !ca.started {
		ca.started = true
		ca.writeHeaders()
	}
	ca.w.WriteHeader(statusCode)
}

// FilterRequest implements the required function to allow CookiemaskFilter to
// be a falcore.RequestFilter.
func (filter *CookiemaskFilter) ServeHTTP(
	w http.ResponseWriter,
	req *http.Request,
) {
	log.Debug("running cookiemask filter")

	//TODO remove
	log.Debug(fmt.Sprintf("handler is: %v", filter))

	sesh, err := filter.Handler.GetSession()
	if err != nil {
		log.Err(
			fmt.Sprintf(
				"cookiemask filter was unable to get"+
					" a new session from the session"+
					" handler: %v",
				err,
			),
		)

		util.InternalServerError.ServeHTTP(w, req)
		return
	}

	// TODO remove
	log.Debug(fmt.Sprintf("sesh is: %v", sesh))

	req = req.WithContext(context.WithValue(req.Context(), "session", sesh))

	fwd_ckes, set_ckes, err := sesh.CookieMask(
		req.Cookies(),
	)
	req.Header.Del("Cookie")
	for _, cke := range fwd_ckes {
		req.AddCookie(cke)
	}

	ca := cookieAppender{
		ckes: set_ckes,
		w: w,
		hdrs: make(map[string][]string),
		started: false,
	}
	defer ca.writeTrailers()

	if err != nil {
		log.Err(
			fmt.Sprintf(
				"cookiemask filter's call to the"+
					" session handler's CookieMask"+
					" function returned an error: %v",
				err,
			),
		)

		util.InternalServerError.ServeHTTP(ca, req)
		return
	} else {
		log.Info(
			"request has run through cookiemask, now" +
				" forwarding to next filter",
		)
		filter.Masked.ServeHTTP(ca, req)
		return
	}
}

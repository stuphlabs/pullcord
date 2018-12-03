package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/stuphlabs/pullcord/config"
)

// PassthruFilter provides a mechanism by which a net/http/httputil.ReverseProxy
// can be configured. This reverse proxy allows requests received by Pullcord to
// be sent to a remote service.
type PassthruFilter httputil.ReverseProxy

func init() {
	config.RegisterResourceType(
		"passthrufilter",
		func() json.Unmarshaler {
			return new(PassthruFilter)
		},
	)
}

// NewPassthruFilter creates a PassthruFilter using a single host reverse proxy
// pointing at the given url.URL.
func NewPassthruFilter(u *url.URL) *PassthruFilter {
	return (*PassthruFilter)(httputil.NewSingleHostReverseProxy(u))
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (f *PassthruFilter) UnmarshalJSON(input []byte) error {
	var t struct {
		Host string
		Port int
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	} else {
		u, e := url.Parse(fmt.Sprintf("http://%s:%d", t.Host, t.Port))
		if e != nil {
			return e
		}

		*f = PassthruFilter(*httputil.NewSingleHostReverseProxy(u))

		return nil
	}
}

func (f *PassthruFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	(*httputil.ReverseProxy)(f).ServeHTTP(w, r)
}

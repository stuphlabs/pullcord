package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

// ExactPathRouter is a net/http.Handler that routes to other handlers based on
// an exactly matching path, or an optional default handler.
type ExactPathRouter struct {
	Routes  map[string]http.Handler
	Default http.Handler
}

func init() {
	config.MustRegisterResourceType(
		"exactpathrouter",
		func() json.Unmarshaler {
			return new(ExactPathRouter)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (r *ExactPathRouter) UnmarshalJSON(input []byte) error {
	var t struct {
		Routes  map[string]*config.Resource
		Default *config.Resource
	}
	t.Routes = make(map[string]*config.Resource)

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	r.Routes = make(map[string]http.Handler)
	for path, rsc := range t.Routes {
		switch f := rsc.Unmarshalled.(type) {
		case http.Handler:
			r.Routes[path] = f
		default:
			_ = log.Err(
				fmt.Sprintf(
					"Registry value is not a"+
						" http.Handler: %s",
					f,
				),
			)
			return config.UnexpectedResourceType
		}
	}
	if t.Default != nil {
		switch f := t.Default.Unmarshalled.(type) {
		case http.Handler:
			r.Default = f
		default:
			_ = log.Err(
				fmt.Sprintf(
					"Registry value is not a"+
						" http.Handler: %s",
					f,
				),
			)
			return config.UnexpectedResourceType
		}
	}

	return nil
}

func (r *ExactPathRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = log.Info(fmt.Sprintf("Request received for path: %s", req.URL.Path))
	_ = log.Debug(fmt.Sprintf("Path router: %#v", r))
	if f, present := r.Routes[req.URL.Path]; present {
		f.ServeHTTP(w, req)
	} else if r.Default != nil {
		r.Default.ServeHTTP(w, req)
	} else {
		StandardResponse(404).ServeHTTP(w, req)
	}
}

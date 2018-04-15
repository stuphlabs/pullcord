package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

type ExactPathRouter struct {
	Routes map[string]http.Handler
	Default http.Handler
}

func init() {
	config.RegisterResourceType(
		"exactpathrouter",
		func() json.Unmarshaler {
			return new(ExactPathRouter)
		},
	)
}

func (r *ExactPathRouter) UnmarshalJSON(input []byte) error {
	var t struct {
		Routes map[string]*config.Resource
		Default *config.Resource
	}
	t.Routes = make(map[string]*config.Resource)

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	r.Routes = make(map[string]http.Handler)
	for path, rsc := range t.Routes {
		switch f := rsc.Unmarshaled.(type) {
		case http.Handler:
			r.Routes[path] = f
		default:
			log.Err(
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
		switch f := t.Default.Unmarshaled.(type) {
		case http.Handler:
			r.Default = f
		default:
			log.Err(
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
	log.Info(fmt.Sprintf("Request received for path: %s", req.URL.Path))
	log.Debug(fmt.Sprintf("Path router: %#v", r))
	if f, present := r.Routes[req.URL.Path]; present {
		f.ServeHTTP(w, req)
	} else if r.Default != nil {
		r.Default.ServeHTTP(w, req)
	} else {
		StandardResponse(404).ServeHTTP(w, req)
	}
}

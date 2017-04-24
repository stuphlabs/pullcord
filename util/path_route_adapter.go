package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stuphlabs/pullcord/config"
)

type ExactPathRouter struct {
	Routes map[string]*falcore.RequestFilter
}

func init() {
	config.RegisterResourceType(
		"exactpathrouter",
		func() json.Unmarshaler {
			return new(ExactPathRouter)
		},
	)
}

func (r *ExactPathRouter) UnmarshalJSON(input []byte) (error) {
	var t struct {
		Routes map[string]config.Resource
	}
	t.Routes = make(map[string]config.Resource)

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	r.Routes = make(map[string]*falcore.RequestFilter)
	for path, rsc := range t.Routes {
		switch f := rsc.Unmarshaled.(type) {
		case falcore.RequestFilter:
			r.Routes[path] = &f
		default:
			log().Err(
				fmt.Sprintf(
					"Registry value is not a" +
					" RequestFilter: %s",
					f,
				),
			)
			return config.UnexpectedResourceType
		}
	}

	return nil
}

func (r *ExactPathRouter) SelectPipeline(
	req *falcore.Request,
) (falcore.RequestFilter) {
	if f, present := r.Routes[req.HttpRequest.URL.Path]; present {
		return *f
	} else {
		return nil
	}
}


package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stuphlabs/pullcord/registry"
)

type ExactPathRouter struct {
	Routes map[string]*falcore.RequestFilter
}

func (r *ExactPathRouter) UnmarshalJSON(input []byte) (error) {
	var t struct {
		Routes map[string]string
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	for path, filterName := range t.Routes {
		if f, e := registry.Get(filterName); e != nil {
			return e
		} else {
			switch f := f.(type) {
			case *falcore.RequestFilter:
				r.Routes[path] = f
			default:
				log().Err(
					fmt.Sprintf(
						"Registry value is not a" +
						" RequestFilter: %s",
						filterName,
					),
				)
				return registry.UnexpectedType
			}
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


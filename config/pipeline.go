package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fitstar/falcore"
	"net/http"
)

type ConfigPipeline struct {
	Server *falcore.Server
}

func init() {
	e := RegisterResourceType(
		"pipeline",
		func() json.Unmarshaler {
			return &ConfigPipeline{
				Server: falcore.NewServer(
					0,
					falcore.NewPipeline(),
				),
			}
		},
	)

	if e != nil {
		panic(e)
	}
}

func (p *ConfigPipeline) UnmarshalJSON(d []byte) error {
	var t struct {
		Upstream []Resource
		Downstream []Resource
	}
	t.Upstream = make([]Resource, 0)
	t.Downstream = make([]Resource, 0)

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	for _, r := range t.Upstream {
		u := r.Unmarshaled
		switch u := u.(type) {
		case falcore.Router:
		case falcore.RequestFilter:
			p.Server.Pipeline.Upstream.PushBack(u)
		default:
			return errors.New(
				fmt.Sprintf(
					"The pipeline resource is not a" +
					" RequestFilter: %v",
					u,
				),
			)
		}
	}

	for _, r := range t.Downstream {
		u := r.Unmarshaled
		switch u := u.(type) {
		case falcore.Router:
		case falcore.RequestFilter:
			p.Server.Pipeline.Downstream.PushBack(u)
		default:
			return errors.New(
				fmt.Sprintf(
					"The pipeline resource is not a" +
					" RequestFilter: %v",
					u,
				),
			)
		}
	}

	return nil
}

func (p *ConfigPipeline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Server.ServeHTTP(w, r)
}

package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/filter"
	"github.com/stuphlabs/pullcord/config"
	"net/http"
)

type PassthruFilter struct {
	Host string
	Port int
	upstreamFilter *filter.Upstream
}

func init() {
	config.RegisterResourceType(
		"passthrufilter",
		func() json.Unmarshaler {
			return new(PassthruFilter)
		},
	)
}

func NewPassthruFilter(host string, port int) (*PassthruFilter) {
	return &PassthruFilter{
		host,
		port,
		filter.NewUpstream(
			filter.NewUpstreamTransport(
				host,
				port,
				0,
				nil,
			),
		),
	}
}

func (f *PassthruFilter) UnmarshalJSON(input []byte) (error) {
	var t struct {
		Host string
		Port int
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	} else {
		f.Host = t.Host
		f.Port = t.Port
		f.upstreamFilter = filter.NewUpstream(
			filter.NewUpstreamTransport(
				f.Host,
				f.Port,
				0,
				nil,
			),
		)

		return nil
	}
}

// NewPassthruFilter generates a Falcore RequestFilter that proxies all requests
// that reach it back and forth to a given host and port.
//
// As such, this is the core of the proxying system.
func (f *PassthruFilter) FilterRequest(
	req *falcore.Request,
) (*http.Response) {
	if f.upstreamFilter == nil {
		f.upstreamFilter = filter.NewUpstream(
			filter.NewUpstreamTransport(
				f.Host,
				f.Port,
				0,
				nil,
			),
		)
	}

	return f.upstreamFilter.FilterRequest(req)
}


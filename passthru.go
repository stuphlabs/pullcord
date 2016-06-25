package pullcord

import (
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/filter"
)

// NewPassthruFilter generates a Falcore RequestFilter that proxies all requests
// that reach it back and forth to a given host and port.
//
// As such, this is the core of the proxying system.
func NewPassthruFilter(host string, port int) falcore.RequestFilter {
	return filter.NewUpstream(
		filter.NewUpstreamTransport(host, port, 0, nil),
	)
}


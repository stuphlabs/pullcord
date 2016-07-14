package proxy

import (
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/filter"
	// "github.com/stuphlabs/pullcord"
)

// NewPassthruFilter generates a Falcore RequestFilter that proxies all requests
// that reach it back and forth to a given host and port.
//
// As such, this is the core of the proxying system.
func NewPassthruFilter(host string, port int) falcore.RequestFilter {
	log().Debug("registering a new passthru filter")

	return filter.NewUpstream(
		filter.NewUpstreamTransport(host, port, 0, nil),
	)
}


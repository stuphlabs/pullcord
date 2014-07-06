package scaledown

import (
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/filter"
)

func NewPassthruFilter(host string, port int) falcore.RequestFilter {
	return filter.NewUpstream(filter.NewUpstreamTransport(host, port, 0, nil))
}


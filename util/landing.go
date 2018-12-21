package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

// LandingHandler is a net/http.Handler that acts as a default landing page for
// a (presumably not-yet-production) Pullcord instance.
type LandingHandler struct {
}

func init() {
	config.MustRegisterResourceType(
		"landinghandler",
		func() json.Unmarshaler {
			return new(LandingHandler)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (l *LandingHandler) UnmarshalJSON(data []byte) error {
	return nil
}

// NewLandingHandler generates an http.Handler that produces a simple landing
// page.
func (l *LandingHandler) ServeHTTP(
	w http.ResponseWriter,
	req *http.Request,
) {
	_ = log.Info("running landing handler")

	w.WriteHeader(200)
	_, err := w.Write(
		[]byte(
			"<html><head><title>" +
				"Pullcord Landing Page" +
				"</title></head><body><h1>" +
				"Pullcord Landing Page" +
				"</h1><p>" +
				"This is the landing page for Pullcord, " +
				"a reverse proxy for cloud-based web apps " +
				"that allows the servers the web apps run on " +
				"to be turned off when not in use." +
				"</p><p>" +
				"If you are unsure of how to proceed, " +
				"please contact the site administrator." +
				"</p></body></html>",
		),
	)
	if err != nil {
		_ = log.Error(
			fmt.Sprintf(
				"error while writing landing page: %s",
				err.Error(),
			),
		)
	}
}

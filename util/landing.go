package util

import (
	"encoding/json"
	"github.com/fitstar/falcore"
	"github.com/stuphlabs/pullcord/config"
	"net/http"
)

type LandingFilter struct {
}

func init() {
	config.RegisterResourceType(
		"landingfilter",
		func() json.Unmarshaler {
			return new(LandingFilter)
		},
	)
}

func (l *LandingFilter) UnmarshalJSON(data []byte) error {
	return nil
}

// NewLandingFilter generates a Falcore RequestFilter that produces a simple
// landing page.
func (filter *LandingFilter) FilterRequest(
	req *falcore.Request,
) (*http.Response) {
	log().Info("running landing filter")

	return falcore.StringResponse(
		req.HttpRequest,
		200,
		nil,
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
	)
}


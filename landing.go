package pullcord

import (
	"github.com/fitstar/falcore"
	"net/http"
)

func NewLandingFilter() falcore.RequestFilter {
	return falcore.NewRequestFilter(
		func(req *falcore.Request) *http.Response {
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
		},
	)
}


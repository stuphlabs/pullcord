package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

// StandardResponse implements a net/http.Handler that gives a canned version of
// the appropriate response for some HTTP code. For example, rather than going
// to the trouble of creating another net/http.Handler to deal with a request
// for a non-existent page, you could instead cast the literal integer 404 as a
// StandardResponse and the appropriate action will be taken (so long as the
// canned response this implementation provides will suffice).
type StandardResponse int

// MinimumStandardResponse is the lower limit of acceptable HTTP response codes
// for which a StandardResponse can be created.
const MinimumStandardResponse = 100

func init() {
	config.MustRegisterResourceType(
		"standardresponse",
		func() json.Unmarshaler {
			return new(StandardResponse)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (s *StandardResponse) UnmarshalJSON(data []byte) error {
	var t int
	if e := json.Unmarshal(data, &t); e != nil {
		return e
	}

	if t < MinimumStandardResponse {
		return fmt.Errorf(
			"StandardResponse must be a valid HTTP status code (an"+
				" integer greater than %d), but was given: %d",
			MinimumStandardResponse,
			t,
		)
	}

	*s = StandardResponse(t)
	return nil
}

const (
	// Forbidden is a canned StandardResponse for an HTTP 403
	Forbidden = StandardResponse(403)
	// NotFound is a canned StandardResponse for an HTTP 404
	NotFound = StandardResponse(404)
	// InternalServerError is a canned StandardResponse for an HTTP 500
	InternalServerError = StandardResponse(500)
	// NotImplemented is a canned StandardResponse for an HTTP 501
	NotImplemented = StandardResponse(501)
)

var responseTitle = map[StandardResponse]string{
	Forbidden:           "Forbidden",
	NotFound:            "Not Found",
	InternalServerError: "Internal Server Error",
	NotImplemented:      "Not Implemented",
}

var responseText = map[StandardResponse]string{
	Forbidden:           "You are not authorized to make this request.",
	NotFound:            "The requested page was not found.",
	InternalServerError: "An internal server error occurred.",
	NotImplemented: "The requested behavior has not yet been" +
		" implemented.",
}

var responseContact = map[StandardResponse]bool{
	Forbidden:           false,
	NotFound:            false,
	InternalServerError: true,
	NotImplemented:      true,
}

var responseStringTemplate = template.Must(
	template.New("standardResponse").Parse(`<!DOCTYPE html>
<html>
 <head>
  <title>
   {{.Title}}
  </title>
 </head>
 <body>
  <h1>
   {{.Title}}
  </h1>
  <p>
   {{.Message}}
   {{- if .ShouldContact}}Please contact your system administrator.{{end}}
  </p>
 </body>
</html>`))

var responseContactString = "Please contact your system administrator."

func (s StandardResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		Title         string
		Message       string
		ShouldContact bool
	}{}

	rs := s
	if rs < MinimumStandardResponse {
		rs = 500
	}

	if v, present := responseContact[rs]; present && v {
		values.ShouldContact = v
	}

	if v, present := responseTitle[rs]; present {
		values.Title = v
	}

	if v, present := responseText[rs]; present {
		values.Message = v
	}

	w.WriteHeader(int(rs))
	err := responseStringTemplate.Execute(w, values)
	if err != nil {
		_ = log.Error(
			fmt.Sprintf(
				"error while writing standard response"+
					" template: %s",
				err.Error(),
			),
		)
	}
}

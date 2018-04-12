package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/stuphlabs/pullcord/config"
)

type StandardResponse int

const MinimumStandardResponse = 100

func init() {
	config.RegisterResourceType(
		"standardresponse",
		func() json.Unmarshaler {
			return new(StandardResponse)
		},
	)
}

func (s *StandardResponse) UnmarshalJSON(data []byte) error {
	var t int
	if e := json.Unmarshal(data, &t); e != nil {
		return e
	}

	if t < MinimumStandardResponse {
		return errors.New(
			fmt.Sprintf(
				"StandardResponse must be a valid HTTP status"+
					" code (an integer greater than %d),"+
					"but was given: %d",
				MinimumStandardResponse,
				t,
			),
		)
	}

	*s = StandardResponse(t)
	return nil
}

const (
	NotFound            = StandardResponse(404)
	InternalServerError = StandardResponse(500)
	NotImplemented      = StandardResponse(501)
)

var responseTitle = map[StandardResponse]string{
	NotFound:            "Not Found",
	InternalServerError: "Internal Server Error",
	NotImplemented:      "Not Implemented",
}

var responseText = map[StandardResponse]string{
	NotFound:            "The requested page was not found.",
	InternalServerError: "An internal server error occured.",
	NotImplemented: "The requested behavior has not yet been" +
		" implemented.",
}

var responseContact = map[StandardResponse]bool{
	NotFound:            false,
	InternalServerError: true,
	NotImplemented:      true,
}

var responseStringTemplate = template.Must(
	template.New("standardResponse").Parse(
`<!DOCTYPE html>
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
   {{if .ShouldContact}}Please contact your system administrator.{{end}}
  </p>
 </body>
</html>`))

var responseContactString = "Please contact your system administrator."

func (s StandardResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct{
		Title string
		Message string
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
	responseStringTemplate.Execute(w, values)
}

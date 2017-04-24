package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stuphlabs/pullcord/config"
	"net/http"
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
				"StandardResponse must be a valid HTTP" +
				" status code (an integer greater than %d)," +
				" but was given: %d",
				MinimumStandardResponse,
				t,
			),
		)
	}

	*s = StandardResponse(t)
	return nil
}

const (
	NotFound = StandardResponse(404)
	InternalServerError = StandardResponse(500)
	NotImplemented = StandardResponse(501)
)

var responseTitle = map[StandardResponse]string{
	NotFound: "Not Found",
	InternalServerError: "Internal Server Error",
	NotImplemented: "Not Implemented",
}

var responseText = map[StandardResponse]string{
	NotFound: "The requested page was not found.",
	InternalServerError: "An internal server error occured.",
	NotImplemented: "The requested behavior has not yet been implemented.",
}

var responseContact = map[StandardResponse]bool{
	NotFound: false,
	InternalServerError: true,
	NotImplemented: true,
}

var responseStringFormat = `<!DOCTYPE html>
<html>
 <head>
  <title>
   %1$s
  </title>
 </head>
 <body>
  <h1>
   %1$s
  </h1>
  <p>
   %2$s
   %3$s
  </p>
 </body>
</html>`

var responseContactString = "Please contact your system administrator."

func (s StandardResponse) FilterRequest(
	request *falcore.Request,
) (*http.Response) {
	var title, text, contact string

	rs := s
	if rs < MinimumStandardResponse {
		rs = 500
	}

	if v, present := responseContact[rs]; present && v {
		contact = responseContactString
	}

	if v, present := responseTitle[rs]; present {
		title = v
	}

	if v, present := responseText[rs]; present {
		text = v
	}

	return falcore.StringResponse(
		request.HttpRequest,
		int(rs),
		nil,
		fmt.Sprintf(
			responseStringFormat,
			title,
			text,
			contact,
		),
	)
}



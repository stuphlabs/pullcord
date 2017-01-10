package util

import (
	"fmt"
	"github.com/fitstar/falcore"
	"net/http"
)

type StandardResponse int

const (
	NotFound = StandardResponse(404)
	InternalServerError = StandardResponse(500)
	NotImplementedError = StandardResponse(501)
)

var responseTitle = map[StandardResponse]string{
	NotFound: "Not Found",
	InternalServerError: "Internal Server Error",
	NotImplementedError: "Not Implemented",
}

var responseText = map[StandardResponse]string{
	NotFound: "The requested page was not found.",
	InternalServerError: "An internal server error occured.",
	NotImplementedError: "The requested behavior has not yet been" +
				" implemented.",
}

var responseContact = map[StandardResponse]bool{
	NotFound: false,
	InternalServerError: true,
	NotImplementedError: true,
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
	contactString := ""
	if responseContact[s] {
		contactString = responseContactString
	}

	return falcore.StringResponse(
		request.HttpRequest,
		int(s),
		nil,
		fmt.Sprintf(
			responseStringFormat,
			responseTitle[s],
			responseText[s],
			contactString,
		),
	)
}



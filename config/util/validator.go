package util

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/fitstar/falcore"
	"github.com/stuphlabs/pullcord/config"
	"net/http"
)

type validation struct {
	unmarshaled json.Unmarshaler
	validate func(json.Unmarshaler) error
}

func (v *validation) UnmarshalJSON(input []byte) error {
	var r config.Resource

	dec := json.NewDecoder(bytes.NewReader(input))
	if e:= dec.Decode(&r); e != nil {
		return e
	} else {
		v.unmarshaled = r.Unmarshaled
		return v.validate(v.unmarshaled)
	}
}

var responseString = `<!DOCTYPE html>
<html>
 <head>
  <title>
   Pullcord validator output
  </title>
 </head>
 <body>
  <h1>
   Pullcord validator output
  </h1>
  <p>
   This page is the result of a successful config validation resource being
   created. This is probably the result of a unit test passing.
  </p>
 </body>
</html>
`

func (v *validation) FilterRequest(req *falcore.Request) *http.Response {
	return falcore.StringResponse(
		req.HttpRequest,
		200,
		nil,
		responseString,
	)
}

func GenerateValidator(
	validate func(json.Unmarshaler) error,
) (string, error) {
	rbytes := make([]byte, 8) // we just don't want cheating in tests
	if _, e := rand.Read(rbytes); e!= nil {
		return "", e
	}
	validatorName := "validator-" + hex.EncodeToString(rbytes)
	return validatorName, config.RegisterResourceType(
		validatorName,
		func() json.Unmarshaler {
			return &validation{
				unmarshaled: nil,
				validate: validate,
			}
		},
	)
}


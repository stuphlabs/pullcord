package util

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/stuphlabs/pullcord/config"
)

type validation struct {
	unmarshaled json.Unmarshaler
	validate    func(json.Unmarshaler) error
}

func (v *validation) UnmarshalJSON(input []byte) error {
	var r config.Resource

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&r); e != nil {
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

func (v *validation) Accept() (net.Conn, error) {
	return nil, errors.New(
		"Accept was called on a validator, presumably after an" +
			" otherwise passing unit test.",
	)
}

func (v *validation) Close() error {
	return errors.New(
		"Close was called on a validator, presumably after an" +
			" otherwise passing unit test.",
	)
}

func (v *validation) Addr() net.Addr {
	return nil
}

func (v *validation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic(
		errors.New(
			"ServeHTTP was called on a validator, presumably" +
				" after an otherwise passing unit test.",
		),
	)
}

func GenerateValidator(
	validate func(json.Unmarshaler) error,
) (string, error) {
	rbytes := make([]byte, 8) // we just don't want cheating in tests
	if _, e := rand.Read(rbytes); e != nil {
		return "", e
	}
	validatorName := "validator-" + hex.EncodeToString(rbytes)
	return validatorName, config.RegisterResourceType(
		validatorName,
		func() json.Unmarshaler {
			return &validation{
				unmarshaled: nil,
				validate:    validate,
			}
		},
	)

}

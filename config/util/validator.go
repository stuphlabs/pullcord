package util

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"github.com/stuphlabs/pullcord/config"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
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

type ConfigTestData struct {
	Data string
	Explanation string
}

func constructConfigReader(
	listenerTest bool,
	validatorName,
	resourceType,
	data string,
) io.Reader {
	if listenerTest {
		return strings.NewReader(
			fmt.Sprintf(
				`{
					"resources": {
						"handler": {
							"type": "testhandler",
							"data": null
						},
						"listener": {
							"type": "%s",
							"data": {
								"type": "%s",
								"data": %s
							}
						}
					},
					"handler": "handler",
					"listener": "listener"
				}`,
				validatorName,
				resourceType,
				data,
			),
		)
	} else {
		return strings.NewReader(
			fmt.Sprintf(
				`{
					"resources": {
						"handler": {
							"type": "pipeline",
							"data": {
								"upstream": [{
									"type": "%s",
									"data": {
										"type": "%s",
										"data": %s
									}
								}]
							}
						},
						"listener": {
							"type": "testlistener",
							"data": null
						}
					},
					"handler": "handler",
					"listener": "listener"
				}`,
				validatorName,
				resourceType,
				data,
			),
		)
	}
}

type ConfigTest struct {
	ResourceType string
	IsValid func(json.Unmarshaler) error
	SyntacticallyBad []ConfigTestData
	SemanticallyBad []ConfigTestData
	Good []ConfigTestData
	ListenerTest bool
}

func (c *ConfigTest) Run(t *testing.T) {
	isValidWasRun := false

	validatorName, e := GenerateValidator(
		func(i json.Unmarshaler) error {
			isValidWasRun = true
			if c.IsValid != nil {
				return c.IsValid(i)
			} else {
				return nil
			}
		},
	)
	assert.NoError(
		t,
		e,
		"Generating a validator resource type should not produce an" +
		" error.",
	)
	assert.NotEqual(
		t,
		validatorName,
		"",
		"A generated validator resource type should not have an" +
		" empty resource type name.",
	)

	for _, d := range c.SyntacticallyBad {
		isValidWasRun = false

		s, e := config.ServerFromReader(
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		)

		assert.Error(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with syntax errors should produce an" +
				" error, but no error was generated for a" +
				" config with explanation: %s",
				d.Explanation,
			),
		)

		assert.False(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with syntactic errors should not run" +
				" IsValid, but IsValid was run for a config" +
				" with description: %s",
				d.Explanation,
			),
		)

		assert.Nil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with syntax errors should produce a nil" +
				" server, but a non-nil server was generated" +
				" for a config with explanation: %s",
				d.Explanation,
			),
		)
	}

	for _, d := range c.SemanticallyBad {
		isValidWasRun = false

		s, e := config.ServerFromReader(
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		)

		assert.Error(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with semantic errors should produce an" +
				" error, but no error was generated for a" +
				" config with explanation: %s",
				d.Explanation,
			),
		)

		assert.True(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with semantic errors should run IsValid," +
				" but IsValid was not run for a config with" +
				" description: %s",
				d.Explanation,
			),
		)

		assert.Nil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a config" +
				" with semantic errors should produce a nil" +
				" server, but a non-nil server was generated" +
				" for a config with explanation: %s",
				d.Explanation,
			),
		)
	}

	for _, d := range c.Good {
		isValidWasRun = false

		s, e := config.ServerFromReader(
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		)

		assert.NoError(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a valid" +
				" config should produce no errors, but an" +
				" error was generated for a config with" +
				" explanation: %s",
				d.Explanation,
			),
		)

		assert.True(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a valid" +
				" config should run IsValid, but IsValid was" +
				" not run for a config with description: %s",
				d.Explanation,
			),
		)

		assert.NotNil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a valid" +
				" config should produce a non-nil server," +
				" but a nil server was generated for a" +
				" config with explanation: %s",
				d.Explanation,
			),
		)
	}
}

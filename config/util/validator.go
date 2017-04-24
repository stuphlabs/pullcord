package util

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	"github.com/stuphlabs/pullcord/config"
	"io"
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

func constructConfigReader(validatorName, resourceType, data string) io.Reader {
	return strings.NewReader(
		fmt.Sprintf(
			`{
				"resources": {
					"validator": {
						"type": "%s",
						"data": {
							"type": "%s",
							"data": %s
						}
					}
				},
				"pipeline": ["validator"],
				"port": 80
			}`,
			validatorName,
			resourceType,
			data,
		),
	)
}

type ConfigTest struct {
	ResourceType string
	IsValid func(json.Unmarshaler) error
	SyntacticallyBad []ConfigTestData
	SemanticallyBad []ConfigTestData
	Good []ConfigTestData
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

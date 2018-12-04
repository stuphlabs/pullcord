package util

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stuphlabs/pullcord/config"
)

// ConfigTestData contains the data for a specific test case for config
// behavior, plus the corresponding explanation string.
type ConfigTestData struct {
	Data        string
	Explanation string
}

// ConfigTest represents an entire suite of tests for config behavior of a
// particular resource type.
type ConfigTest struct {
	ResourceType     string
	IsValid          func(json.Unmarshaler) error
	SyntacticallyBad []ConfigTestData
	SemanticallyBad  []ConfigTestData
	Good             []ConfigTestData
	ListenerTest     bool
}

// Run executes the suite of config behavior tests.
func (c *ConfigTest) Run(t *testing.T) {
	isValidWasRun := false

	validatorName, e := GenerateValidator(
		func(i json.Unmarshaler) error {
			isValidWasRun = true
			if c.IsValid != nil {
				return c.IsValid(i)
			}
			return nil
		},
	)
	assert.NoError(
		t,
		e,
		"Generating a validator resource type should not produce an"+
			" error.",
	)
	assert.NotEqual(
		t,
		validatorName,
		"",
		"A generated validator resource type should not have an"+
			" empty resource type name.",
	)

	for _, d := range c.SyntacticallyBad {
		isValidWasRun = false

		parser := config.Parser{
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		}
		s, e := parser.Server()

		assert.Error(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with syntax errors should produce an"+
					" error, but no error was generated for a"+
					" config with explanation: %s",
				d.Explanation,
			),
		)

		assert.False(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with syntactic errors should not run"+
					" IsValid, but IsValid was run for a config"+
					" with description: %s",
				d.Explanation,
			),
		)

		assert.Nil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with syntax errors should produce a nil"+
					" server, but a non-nil server was generated"+
					" for a config with explanation: %s",
				d.Explanation,
			),
		)
	}

	for _, d := range c.SemanticallyBad {
		isValidWasRun = false

		parser := config.Parser{
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		}
		s, e := parser.Server()

		assert.Error(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with semantic errors should produce an"+
					" error, but no error was generated for a"+
					" config with explanation: %s",
				d.Explanation,
			),
		)

		assert.True(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with semantic errors should run IsValid,"+
					" but IsValid was not run for a config with"+
					" description: %s",
				d.Explanation,
			),
		)

		assert.Nil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a config"+
					" with semantic errors should produce a nil"+
					" server, but a non-nil server was generated"+
					" for a config with explanation: %s",
				d.Explanation,
			),
		)
	}

	for _, d := range c.Good {
		isValidWasRun = false

		parser := config.Parser{
			constructConfigReader(
				c.ListenerTest,
				validatorName,
				c.ResourceType,
				d.Data,
			),
		}
		s, e := parser.Server()

		assert.NoError(
			t,
			e,
			fmt.Sprintf(
				"Attempting to create a server from a valid"+
					" config should produce no errors, but an"+
					" error was generated for a config with"+
					" explanation: %s",
				d.Explanation,
			),
		)

		assert.True(
			t,
			isValidWasRun,
			fmt.Sprintf(
				"Attempting to create a server from a valid"+
					" config should run IsValid, but IsValid was"+
					" not run for a config with description: %s",
				d.Explanation,
			),
		)

		assert.NotNil(
			t,
			s,
			fmt.Sprintf(
				"Attempting to create a server from a valid"+
					" config should produce a non-nil server,"+
					" but a nil server was generated for a"+
					" config with explanation: %s",
				d.Explanation,
			),
		)
	}
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
					"server": {
						"type": "httpserver",
						"data": {
							"handler": {
								"type": "ref",
								"data": "handler"
							},
							"listener": {
								"type": "ref",
								"data": "listener"
							}
						}
					}
				}`,
				validatorName,
				resourceType,
				data,
			),
		)
	}
	return strings.NewReader(
		fmt.Sprintf(
			`{
				"resources": {
					"handler": {
						"type": "%s",
						"data": {
							"type": "%s",
							"data": %s
						}
					},
					"listener": {
						"type": "testlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			validatorName,
			resourceType,
			data,
		),
	)
}

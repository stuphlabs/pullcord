package util

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stuphlabs/pullcord/config"
)

type ConfigTest struct {
	ResourceType     string
	IsValid          func(json.Unmarshaler) error
	SyntacticallyBad []ConfigTestData
	SemanticallyBad  []ConfigTestData
	Good             []ConfigTestData
	ListenerTest     bool
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

package trigger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
	"testing"
)

type counterTriggerHandler struct {
	count int
}

func (th *counterTriggerHandler) Trigger() error {
	if th.count >= 0 {
		th.count += 1
		return nil
	} else {
		return errors.New("this trigger always errors")
	}
}

func TestCompoundTriggerNoErrors(t *testing.T) {
	th1 := &counterTriggerHandler{}
	th2 := &counterTriggerHandler{}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.NoError(t, err)

	err = ct.Trigger()
	assert.NoError(t, err)

	assert.Equal(t, 2, th1.count)
	assert.Equal(t, 2, th2.count)
}

func TestCompoundTriggerAllErrors(t *testing.T) {
	th1 := &counterTriggerHandler{-1}
	th2 := &counterTriggerHandler{-1}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.Error(t, err)

	assert.Equal(t, -1, th1.count)
	assert.Equal(t, -1, th2.count)
}

func TestCompoundTriggerSomeErrors(t *testing.T) {
	th1 := &counterTriggerHandler{}
	th2 := &counterTriggerHandler{-1}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.Error(t, err)

	assert.Equal(t, 1, th1.count)
	assert.Equal(t, -1, th2.count)
}

func TestCompoundTriggerFromConfig(t *testing.T) {
	util.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "compoundtrigger",
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "",
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: `{
					"triggers": 7
				}`,
				Explanation: "numeric triggers",
			},
			configutil.ConfigTestData{
				Data: `{
					"triggers": [
						7,
						42
					]
				}`,
				Explanation: "numeric array triggers",
			},
			configutil.ConfigTestData{
				Data: `{
					"triggers": [{
						"type": "landingfilter",
						"data": {}
					}]
				}`,
				Explanation: "non-trigger in triggers",
			},
			configutil.ConfigTestData{
				Data: "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "{}",
				Explanation: "empty object",
			},
			configutil.ConfigTestData{
				Data: "null",
				Explanation: "null config",
			},
			configutil.ConfigTestData{
				Data: `{
					"triggers": []
				}`,
				Explanation: "empty triggers",
			},
			configutil.ConfigTestData{
				Data: `{
					"triggers": [{
						"type": "compoundtrigger",
						"data": {}
					}]
				}`,
				Explanation: "basic valid compound trigger",
			},
		},
	}
	test.Run(t)
}

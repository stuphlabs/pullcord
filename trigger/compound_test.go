package trigger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
)

type counterTriggerrer struct {
	count int
}

func (th *counterTriggerrer) Trigger() error {
	if th.count >= 0 {
		th.count++
		return nil
	}

	return errors.New("this trigger always errors")
}

func TestCompoundTriggerNoErrors(t *testing.T) {
	th1 := &counterTriggerrer{}
	th2 := &counterTriggerrer{}

	ct := CompoundTrigger{[]Triggerrer{th1, th2}}

	err := ct.Trigger()
	assert.NoError(t, err)

	err = ct.Trigger()
	assert.NoError(t, err)

	assert.Equal(t, 2, th1.count)
	assert.Equal(t, 2, th2.count)
}

func TestCompoundTriggerAllErrors(t *testing.T) {
	th1 := &counterTriggerrer{-1}
	th2 := &counterTriggerrer{-1}

	ct := CompoundTrigger{[]Triggerrer{th1, th2}}

	err := ct.Trigger()
	assert.Error(t, err)

	assert.Equal(t, -1, th1.count)
	assert.Equal(t, -1, th2.count)
}

func TestCompoundTriggerSomeErrors(t *testing.T) {
	th1 := &counterTriggerrer{}
	th2 := &counterTriggerrer{-1}

	ct := CompoundTrigger{[]Triggerrer{th1, th2}}

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
			{
				Data:        "",
				Explanation: "empty config",
			},
			{
				Data: `{
					"triggers": 7
				}`,
				Explanation: "numeric triggers",
			},
			{
				Data: `{
					"triggers": [
						7,
						42
					]
				}`,
				Explanation: "numeric array triggers",
			},
			{
				Data: `{
					"triggers": [{
						"type": "landingfilter",
						"data": {}
					}]
				}`,
				Explanation: "non-trigger in triggers",
			},
			{
				Data:        "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data:        "{}",
				Explanation: "empty object",
			},
			{
				Data:        "null",
				Explanation: "null config",
			},
			{
				Data: `{
					"triggers": []
				}`,
				Explanation: "empty triggers",
			},
			{
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

package trigger

import (
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
	"testing"
	"time"
)

func TestDelayTriggerSingleDelay(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(cth, time.Second)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerDoubleDelay(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerErrorMasking(t *testing.T) {
	cth := &counterTriggerHandler{-1}

	dt := NewDelayTrigger(
		cth,
		time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, -1, cth.count)
}

func TestDelayTriggerReplaceError(t *testing.T) {
	cth := &counterTriggerHandler{-1}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, -1, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, -1, cth.count)

	// removes error situation
	cth.count = 0
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerIntroduceError(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	// introduces error situation
	cth.count = -1
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, -1, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, -1, cth.count)
}

func TestDelayTriggerFromConfig(t *testing.T) {
	util.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "delaytrigger",
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "",
				Explanation: "empty config",
			},
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
					"delayedtrigger": 7,
					"delay": "42s"
				}`,
				Explanation: "numeric trigger",
			},
			configutil.ConfigTestData{
				Data: `{
					"delayedtrigger": {
						"type": "landingfilter",
						"data": {}
					},
					"delay": "42s"
				}`,
				Explanation: "non-trigger as trigger",
			},
			configutil.ConfigTestData{
				Data: `{
					"delayedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"delay": 42
				}`,
				Explanation: "numeric delay",
			},
			configutil.ConfigTestData{
				Data: `{
					"delayedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"delay": "42q"
				}`,
				Explanation: "nonsensical delay",
			},
			configutil.ConfigTestData{
				Data: `{
					"delayedtrigger": {},
					"delay": "42s"
				}`,
				Explanation: "empty trigger",
			},
			configutil.ConfigTestData{
				Data: "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: `{
					"delayedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"delay": "42s"
				}`,
				Explanation: "valid delay trigger",
			},
		},
	}
	test.Run(t)
}

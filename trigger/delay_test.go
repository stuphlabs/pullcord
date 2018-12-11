package trigger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
)

func TestDelayTriggerSingleDelay(t *testing.T) {
	cth := &counterTriggerrer{}

	dt := NewDelayTrigger(cth, time.Second)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)

	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerDoubleDelay(t *testing.T) {
	cth := &counterTriggerrer{}

	dt := NewDelayTrigger(
		cth,
		3*time.Second,
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

	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerErrorMasking(t *testing.T) {
	cth := &counterTriggerrer{-1}

	dt := NewDelayTrigger(
		cth,
		time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)

	assert.Equal(t, -1, cth.count)
}

func TestDelayTriggerReplaceError(t *testing.T) {
	cth := &counterTriggerrer{-1}

	dt := NewDelayTrigger(
		cth,
		3*time.Second,
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

	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerIntroduceError(t *testing.T) {
	cth := &counterTriggerrer{}

	dt := NewDelayTrigger(
		cth,
		3*time.Second,
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

	time.Sleep(2 * time.Second)
	assert.Equal(t, -1, cth.count)
}

func TestDelayTriggerFromConfig(t *testing.T) {
	util.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "delaytrigger",
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        "",
				Explanation: "empty config",
			},
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
					"delayedtrigger": 7,
					"delay": "42s"
				}`,
				Explanation: "numeric trigger",
			},
			{
				Data: `{
					"delayedtrigger": {
						"type": "landingfilter",
						"data": {}
					},
					"delay": "42s"
				}`,
				Explanation: "non-trigger as trigger",
			},
			{
				Data: `{
					"delayedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"delay": 42
				}`,
				Explanation: "numeric delay",
			},
			{
				Data: `{
					"delayedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"delay": "42q"
				}`,
				Explanation: "nonsensical delay",
			},
			{
				Data: `{
					"delayedtrigger": {},
					"delay": "42s"
				}`,
				Explanation: "empty trigger",
			},
			{
				Data:        "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			{
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

package trigger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
)

func TestRateLimit(t *testing.T) {
	cth := &counterTriggerrer{}

	rlt := NewRateLimitTrigger(cth, 1, time.Second)

	err := rlt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 1, cth.count)

	err = rlt.Trigger()
	assert.Error(t, err)
	assert.Equal(t, ErrRateLimitExceeded, err)
	assert.Equal(t, 1, cth.count)

	time.Sleep(time.Second)
	err = rlt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 2, cth.count)
}

func TestRateLimitTriggerFromConfig(t *testing.T) {
	util.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "ratelimittrigger",
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
					"guardedtrigger": 7,
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "numeric trigger",
			},
			{
				Data: `{
					"guardedtrigger": {
						"type": "landingfilter",
						"data": {}
					},
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "non-trigger as trigger",
			},
			{
				Data: `{
					"guardedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"maxallowed": 42,
					"period": 42
				}`,
				Explanation: "numeric delay",
			},
			{
				Data: `{
					"guardedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"maxallowed": 42,
					"period": "42q"
				}`,
				Explanation: "nonsensical delay",
			},
			{
				Data: `{
					"guardedtrigger": {},
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "empty trigger",
			},
			{
				Data:        "42",
				Explanation: "numeric config",
			},
			{
				Data: `{
					"guardedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"maxallowed": -42,
					"period": "42s"
				}`,
				Explanation: "negative max allowed",
			},
			{
				Data: `{
					"guardedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"maxallowed": "forty-two",
					"period": "42s"
				}`,
				Explanation: "string max allowed",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data: `{
					"guardedtrigger": {
						"type": "compoundtrigger",
						"data": {}
					},
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "valid rate limit trigger",
			},
		},
	}
	test.Run(t)
}

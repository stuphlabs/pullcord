package trigger

import (
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/util"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	cth := &counterTriggerHandler{}

	rlt := NewRateLimitTrigger(cth, 1, time.Second)

	err := rlt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 1, cth.count)

	err = rlt.Trigger()
	assert.Error(t, err)
	assert.Equal(t, RateLimitExceededError, err)
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
					"guardedtrigger": 7,
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "numeric trigger",
			},
			configutil.ConfigTestData{
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
			configutil.ConfigTestData{
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
			configutil.ConfigTestData{
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
			configutil.ConfigTestData{
				Data: `{
					"guardedtrigger": {},
					"maxallowed": 42,
					"period": "42s"
				}`,
				Explanation: "empty trigger",
			},
			configutil.ConfigTestData{
				Data: "42",
				Explanation: "numeric config",
			},
			configutil.ConfigTestData{
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
			configutil.ConfigTestData{
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
			configutil.ConfigTestData{
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

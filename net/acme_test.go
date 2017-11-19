package net

import (
	"testing"

	configutil"github.com/stuphlabs/pullcord/config/util"
)

func TestAcmeConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "acme",
		ListenerTest: true,
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: ``,
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: `42`,
				Explanation: "numeric config",
			},
			configutil.ConfigTestData{
				Data: `"test"`,
				Explanation: "string config",
			},
			configutil.ConfigTestData{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			configutil.ConfigTestData{
				Data: `{
					"accepttos": false
				}`,
				Explanation: "unaccepted tos",
			},
			configutil.ConfigTestData{
				Data: `{
					"accepttos": true,
					"domains": 7
				}`,
				Explanation: "bad domain list",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: `{
					"AcceptTOS": true,
					"domains": [
						"localhost",
						"127.0.0.1",
						"::1"
					]
				}`,
				Explanation: "good config",
			},
			configutil.ConfigTestData{
				Data: `{
					"accepttos": true
				}`,
				Explanation: "missing domains",
			},
		},
	}

	test.Run(t)
}

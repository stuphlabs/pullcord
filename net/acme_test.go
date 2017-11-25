package net

import (
	"testing"

	configutil "github.com/stuphlabs/pullcord/config/util"
)

func TestAcmeConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "acme",
		ListenerTest: true,
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        ``,
				Explanation: "empty config",
			},
			{
				Data:        `42`,
				Explanation: "numeric config",
			},
			{
				Data:        `"test"`,
				Explanation: "string config",
			},
			{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			{
				Data: `{
					"accepttos": false
				}`,
				Explanation: "unaccepted tos",
			},
			{
				Data: `{
					"accepttos": true,
					"domains": 7
				}`,
				Explanation: "bad domain list",
			},
		},
		Good: []configutil.ConfigTestData{
			{
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
			{
				Data: `{
					"accepttos": true
				}`,
				Explanation: "missing domains",
			},
		},
	}

	test.Run(t)
}

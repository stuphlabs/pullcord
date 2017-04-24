package util

import (
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"net/http"
	"testing"
)

func TestStandardResponseFilterRequest(t *testing.T) {
	type testCase struct {
		s StandardResponse
		req *http.Request
		check func(*testing.T, *http.Response)
	}

	testCases := []testCase{
		testCase {
			s: NotFound,
		},
		testCase {
			s: InternalServerError,
		},
		testCase {
			s: NotImplemented,
		},
		testCase {
			s: 200,
		},
		testCase {
			s: -5,
			check: func(t *testing.T, r *http.Response) {
				assert.Equal(t, 500, r.StatusCode)
			},
		},
	}

	for _, c := range testCases {
		var r *http.Response
		if c.req == nil {
			req, e := http.NewRequest("GET", "/", nil)
			assert.NoError(t, e)
			_, r = falcore.TestWithRequest(req, c.s, nil)
		} else {
			_, r = falcore.TestWithRequest(c.req, c.s, nil)
		}
		if c.check == nil {
			assert.Equal(t, int(c.s), r.StatusCode)
		} else {
			c.check(t, r)
		}
	}
}

func TestStandardResponseFromConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "standardresponse",
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
				Data: "-5",
				Explanation: "negative config",
			},
			configutil.ConfigTestData{
				Data: "99",
				Explanation: "low config",
			},
			configutil.ConfigTestData{
				Data: "null",
				Explanation: "null (zero) config",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "404",
				Explanation: "not found config",
			},
			configutil.ConfigTestData{
				Data: `700`,
				Explanation: "odd config",
			},
		},
	}
	test.Run(t)
}

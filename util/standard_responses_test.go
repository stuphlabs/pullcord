package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
)

func TestStandardResponseFilterRequest(t *testing.T) {
	type testCase struct {
		s     StandardResponse
		req   *http.Request
		check func(*testing.T, *http.Response)
	}

	testCases := []testCase{
		{
			s: NotFound,
		},
		{
			s: InternalServerError,
		},
		{
			s: NotImplemented,
		},
		{
			s: 200,
		},
		{
			s: -5,
			check: func(t *testing.T, r *http.Response) {
				assert.Equal(t, 500, r.StatusCode)
			},
		},
	}

	for _, c := range testCases {
		var req *http.Request
		if c.req != nil {
			req = c.req
		} else {
			req = httptest.NewRequest("GET", "/", nil)
		}
		w := httptest.NewRecorder()
		c.s.ServeHTTP(w, req)
		res := w.Result()
		if c.check == nil {
			assert.Equal(t, int(c.s), res.StatusCode)
		} else {
			c.check(t, res)
		}
	}
}

func TestStandardResponseFromConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "standardresponse",
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
				Data:        "-5",
				Explanation: "negative config",
			},
			{
				Data:        "99",
				Explanation: "low config",
			},
			{
				Data:        "null",
				Explanation: "null (zero) config",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data:        "404",
				Explanation: "not found config",
			},
			{
				Data:        `700`,
				Explanation: "odd config",
			},
		},
	}
	test.Run(t)
}

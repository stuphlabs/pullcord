package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proidiot/gone/log"
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/trigger"
)

type stringFilter string

func (s stringFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, err := w.Write([]byte(s))
	if err != nil {
		_ = log.Debug(
			fmt.Sprintf(
				"in path_route_adapter_test, error writing to"+
					" stringFilter: %s",
				err.Error(),
			),
		)
	}
}

func TestExactPathRouterHandler(t *testing.T) {
	type testCase struct {
		p     *ExactPathRouter
		req   *http.Request
		check func(*testing.T, *http.Response)
	}

	genReq := func(method, path string, body io.Reader) *http.Request {
		if r, e := http.NewRequest(method, path, body); e != nil {
			panic(
				fmt.Sprintf(
					"Unrecoverable error while"+
						" generating http request for unit"+
						" test, likely due to a bad url: %v",
					e,
				),
			)
		} else {
			return r
		}
	}

	testCases := []testCase{
		{
			p: &ExactPathRouter{
				Routes: map[string]http.Handler{
					"/foo": stringFilter("foo"),
					"/bar": stringFilter("bar"),
				},
			},
			req: genReq("GET", "/foo", nil),
			check: func(t *testing.T, r *http.Response) {
				assert.Equal(
					t,
					200,
					r.StatusCode,
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to return a 200 Okay, but that is"+
						" not what was received.",
				)
				content, e := ioutil.ReadAll(r.Body)
				assert.NoError(
					t,
					e,
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to produce a response body that"+
						" can be successfully converted into"+
						" a string, but an attempt at string"+
						" conversion resulted in an error.",
				)
				assert.Equal(
					t,
					"foo",
					string(content),
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to have a specific string as the"+
						" response body, but the response"+
						" body did not match the string.",
				)
			},
		},
		{
			p: &ExactPathRouter{
				Routes: map[string]http.Handler{
					"/foo": stringFilter("foo"),
					"/bar": stringFilter("bar"),
				},
			},
			req: genReq("GET", "/bar", nil),
			check: func(t *testing.T, r *http.Response) {
				assert.Equal(
					t,
					200,
					r.StatusCode,
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to return a 200 Okay, but that is"+
						" not what was received.",
				)
				content, e := ioutil.ReadAll(r.Body)
				assert.NoError(
					t,
					e,
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to produce a response body that"+
						" can be successfully converted into"+
						" a string, but an attempt at string"+
						" conversion resulted in an error.",
				)
				assert.Equal(
					t,
					"bar",
					string(content),
					"A request for a path present in an"+
						" exact path router should route to"+
						" the underlying filter request. For"+
						" this test, the router is expected"+
						" to have a specific string as the"+
						" response body, but the response"+
						" body did not match the string.",
				)
			},
		},
		{
			p: &ExactPathRouter{
				Routes: map[string]http.Handler{
					"/foo": stringFilter("foo"),
					"/bar": stringFilter("bar"),
				},
			},
			req: genReq("GET", "/baz", nil),
			check: func(t *testing.T, r *http.Response) {
				assert.Equal(
					t,
					404,
					r.StatusCode,
					"A request for a path that is not"+
						" present in an exact path router"+
						" should fall back to the next"+
						" filter in the pipeline's upstream."+
						" In this case, we are expecting a"+
						" 404 Not Found, but that is not"+
						" what we received.",
				)
			},
		},
	}

	for _, c := range testCases {
		w := httptest.NewRecorder()
		c.p.ServeHTTP(w, c.req)
		r := w.Result()
		c.check(t, r)
	}
}

func TestPathRouteAdapterFromConfig(t *testing.T) {
	trigger.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "exactpathrouter",
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        "",
				Explanation: "empty config",
			},
			{
				Data:        "99",
				Explanation: "numeric config",
			},
			{
				Data: `{
					"routes": 42
				}`,
				Explanation: "numeric routes",
			},
			{
				Data: `{
					"routes": "forty-two"
				}`,
				Explanation: "string routes",
			},
			{
				Data: `{
					"routes": [
						42
					]
				}`,
				Explanation: "numeric array routes",
			},
			{
				Data: `{
					"routes": [
						"forty-two"
					]
				}`,
				Explanation: "string array routes",
			},
			{
				Data: `{
					"routes": {
						"forty-two": 42
					}
				}`,
				Explanation: "numeric map routes",
			},
			{
				Data: `{
					"routes": {
						"/index.html": {
							"type": "compoundtrigger",
							"data": {}
						}
					}
				}`,
				Explanation: "basic valid config",
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
					"routes": {
						"/index.html": {
							"type": "landinghandler",
							"data": {}
						}
					}
				}`,
				Explanation: "basic valid config",
			},
		},
	}
	test.Run(t)
}

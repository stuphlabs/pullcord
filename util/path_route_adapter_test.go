package util

import (
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"github.com/stuphlabs/pullcord/trigger"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestExactPathRouterWithinPipeline(t *testing.T) {
	type testCase struct {
		p *ExactPathRouter
		req *http.Request
		check func(*testing.T, *http.Response)
	}

	stringFilter := func(s string) *falcore.RequestFilter {
		f := falcore.NewRequestFilter(
			func(req *falcore.Request) *http.Response {
				return falcore.StringResponse(
					req.HttpRequest,
					200,
					nil,
					s,
				)
			},
		)
		return &f
	}

	genReq := func(method, path string, body io.Reader) *http.Request {
		if r, e := http.NewRequest(method, path, body); e !=nil {
			panic(
				fmt.Sprintf(
					"Unrecoverable error while" +
					" generating http request for unit" +
					" test, likely due to a bad url: %v",
					e,
				),
			)
		} else {
			return r
		}
	}

	testCases := []testCase{
		testCase {
			p: &ExactPathRouter{
				Routes: map[string]*falcore.RequestFilter{
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
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to return a 200 Okay, but that is" +
					" not what was received.",
				)
				content, e := ioutil.ReadAll(r.Body)
				assert.NoError(
					t,
					e,
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to produce a response body that" +
					" can be successfully converted into" +
					" a string, but an attempt at string" +
					" conversion resulted in an error.",
				)
				assert.Equal(
					t,
					"foo",
					string(content),
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to have a specific string as the" +
					" response body, but the response" +
					" body did not match the string.",
				)
			},
		},
		testCase {
			p: &ExactPathRouter{
				Routes: map[string]*falcore.RequestFilter{
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
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to return a 200 Okay, but that is" +
					" not what was received.",
				)
				content, e := ioutil.ReadAll(r.Body)
				assert.NoError(
					t,
					e,
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to produce a response body that" +
					" can be successfully converted into" +
					" a string, but an attempt at string" +
					" conversion resulted in an error.",
				)
				assert.Equal(
					t,
					"bar",
					string(content),
					"A request for a path present in an" +
					" exact path router should route to" +
					" the underlying filter request. For" +
					" this test, the router is expected" +
					" to have a specific string as the" +
					" response body, but the response" +
					" body did not match the string.",
				)
			},
		},
		testCase {
			p: &ExactPathRouter{
				Routes: map[string]*falcore.RequestFilter{
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
					"A request for a path that is not" +
					" present in an exact path router" +
					" should fall back to the next" +
					" filter in the pipeline's upstream." +
					" In this case, we are expecting a" +
					" 404 Not Found, but that is not" +
					" what we received.",
				)
			},
		},
	}

	for _, c := range testCases {
		pipeline := falcore.NewPipeline()
		_ = pipeline.Upstream.PushBack(falcore.Router(c.p))
		_ = pipeline.Upstream.PushBack(NotFound)
		_, r := falcore.TestWithRequest(c.req, pipeline, nil)
		c.check(t, r)
	}
}

func TestPathRouteAdapterFromConfig(t *testing.T) {
	trigger.LoadPlugin()
	test := configutil.ConfigTest{
		ResourceType: "exactpathrouter",
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "",
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: "99",
				Explanation: "numeric config",
			},
			configutil.ConfigTestData{
				Data: `{
					"routes": 42
				}`,
				Explanation: "numeric routes",
			},
			configutil.ConfigTestData{
				Data: `{
					"routes": "forty-two"
				}`,
				Explanation: "string routes",
			},
			configutil.ConfigTestData{
				Data: `{
					"routes": [
						42
					]
				}`,
				Explanation: "numeric array routes",
			},
			configutil.ConfigTestData{
				Data: `{
					"routes": [
						"forty-two"
					]
				}`,
				Explanation: "string array routes",
			},
			configutil.ConfigTestData{
				Data: `{
					"routes": {
						"forty-two": 42
					}
				}`,
				Explanation: "numeric map routes",
			},
			configutil.ConfigTestData{
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
					"routes": {
						"/index.html": {
							"type": "landingfilter",
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

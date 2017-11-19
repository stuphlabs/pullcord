package config

import (
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"strings"
	"testing"
)

// we need an umarshaler, but we won't actually be using it, so...
type dummyType struct {}
func (s *dummyType) UnmarshalJSON(i []byte) error {
	if string(i) == "\"error\"" {
		return errors.New(`*dummyType.UnmarshalJSON("error")`)
	} else {
		return nil
	}
}
func (s *dummyType) FilterRequest(*falcore.Request) *http.Response {
	return nil
}
func newDummy() json.Unmarshaler {
	return new(dummyType)
}

type randomType int
func (s *randomType) UnmarshalJSON([]byte) error {
	return nil
}
func newRandom() json.Unmarshaler {
	var r randomType = 4
	return &r
}

type dummyRouter struct {}
func (r *dummyRouter) UnmarshalJSON([]byte) error {
	return nil
}
func (r *dummyRouter) SelectPipeline(*falcore.Request) falcore.RequestFilter {
	return new(dummyType)
}
func newDummyRouter() json.Unmarshaler {
	return new(dummyRouter)
}

func TestRegisterResourceType(t *testing.T) {
	type testStruct struct {
		validate func(error)
		name string
		newFunc func() json.Unmarshaler
	}

	testData := []testStruct {
		testStruct {
			func(e error) {
				assert.NoError(
					t,
					e,
					"RegisterResourceType should not" +
					" return an error when a new valid" +
					" type is registerred for the first" +
					" time.",
				)
			},
			"dummyType",
			newDummy,
		},
		testStruct {
			func(e error) {
				assert.NoError(
					t,
					e,
					"RegisterResourceType should not" +
					" return an error when a new valid" +
					" type is registerred for the first" +
					" time.",
				)
			},
			"randomType",
			newRandom,
		},
		testStruct {
			func(e error) {
				assert.Error(
					t,
					e,
					"RegisterResourceType should return" +
					" an error when an attempt is made" +
					" to register a constructor for the" +
					" second time with the same type" +
					" name.",
				)
			},
			"dummyType",
			newDummy,
		},
		testStruct {
			func(e error) {
				assert.Error(
					t,
					e,
					"RegisterResourceType should return" +
					" an error if an attempt is made to" +
					" register a constructor with the" +
					" reserved type name %s.",
					ReferenceResourceTypeName,
				)
			},
			ReferenceResourceTypeName,
			newDummy,
		},
	}

	for _, v := range testData {
		e := RegisterResourceType(
			v.name,
			v.newFunc,
		)
		v.validate(e)
	}
}

func TestResourceUnmarshalJSON(t *testing.T) {
	type testStruct struct {
		validate func(*Resource, error)
		input []byte
	}

	testData := []testStruct {
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"An error should be generated if an" +
					" attempt is made to unmarshal into" +
					" a Resource non-JSON input as JSON.",
				)
				assert.False(
					t,
					r.complete,
					"Invalid input should not allow" +
					" Resource unmarshalling to complete.",
				)
			},
			[]byte("not json"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" attempting to unmarshal a JSON" +
					" null should not produce an error.",
				)
				assert.True(
					t,
					r.complete,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" attempting to unmarshal a JSON" +
					" null should allow the Resource" +
					" unmarshalling to ultimately" +
					" complete.",
				)
				assert.Nil(
					t,
					r.Unmarshaled,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" unmarshalling a JSON null should" +
					" result in a nil valued object.",
				)
			},
			[]byte("null"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" attempting to unmarshal an empty" +
					" JSON object (which could" +
					" reasonably be interpereted as a" +
					" null value) should not produce an" +
					" error.",
				)
				assert.True(
					t,
					r.complete,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" attempting to unmarshal an empty" +
					" JSON object (which could" +
					" reasonably be interpereted as a" +
					" null value) should allow the" +
					" Resource unmarshalling to" +
					" ultimately complete.",
				)
				assert.Nil(
					t,
					r.Unmarshaled,
					"As a null Resource value could make" +
					" sense in some circumstances," +
					" attempting to unmarshal an empty" +
					" JSON object (which could" +
					" reasonably be interpereted as a" +
					" null value) should result in a nil" +
					" valued object.",
				)
			},
			[]byte("{}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a Resource" +
					" with a non-string type name should" +
					" produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a Resource" +
					" with a non-string type name should" +
					" not produce a Resource that has" +
					" completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":7}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a" +
					" partially defined Resource should" +
					" produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a" +
					" partially defined Resource should" +
					" not produce a Resource that has" +
					" completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":\"dummyType\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference" +
					" Resource with a non-string data" +
					" section should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference" +
					" Resource with a non-string data" +
					" section should not produce a" +
					" Resource that has completed its" +
					" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":7}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference" +
					" Resource with a non-string data" +
					" section should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference" +
					" Resource with a non-string data" +
					" section should not produce a" +
					" Resource that has completed its" +
					" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":null}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference" +
					" Resource with an unregisterred" +
					" name should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference" +
					" Resource with an unregisterred" +
					" name should not produce a Resource" +
					" that has completed its" +
					" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":\"taco\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a Resource" +
					" with an unregisterred type name" +
					" should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a Resource" +
					" with an unregisterred type name" +
					" should not produce a Resource that" +
					" has completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":\"mytype\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"An attempt to unmarshal a Resource" +
					" given a valid JSON input should" +
					" not produce an error.",
				)
				assert.True(
					t,
					r.complete,
					"An attempt to unmarshal a Resource" +
					" given a valid JSON input should" +
					" produce a Resource that has" +
					" completed its unmarshalling.",
				)
				var expectedType *dummyType
				assert.IsType(
					t,
					expectedType,
					r.Unmarshaled,
				)
			},
			[]byte("{\"type\":\"dummyType\",\"data\":{}}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"An attempt to unmarshal a Resource" +
					" given valid JSON input which would" +
					" cause the particular Resource" +
					" type's unmarshaler to produce an" +
					" error should result in an error" +
					" coming from the top-level" +
					" json.Unmarshal call as well.",
				)
				assert.False(
					t,
					r.complete,
					"An attempt to unmarshal a Resource" +
					" given valid JSON input which would" +
					" cause the particular Resource" +
					" type's unmarshaler to produce an" +
					" error should result in a Resource" +
					" that has not completed its" +
					" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"dummyType\",\"data\":\"error\"}"),
		},
	}

	RegisterResourceType("dummyType", newDummy)

	for _, v := range testData {
		var r Resource
		e := json.Unmarshal(v.input, &r)
		v.validate(&r, e)
	}
}

type TestHandler struct {}

func (h *TestHandler) UnmarshalJSON([]byte) error {
	return nil
}

func (h *TestHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic(
		errors.New(
			"A testHandler can't actually serve HTTP, but an" +
			" attempt was made to do so.",
		),
	)
}

type TestListener struct {}

func (l *TestListener) UnmarshalJSON([]byte) error {
	return nil
}

func (l *TestListener) Accept() (net.Conn, error) {
	return nil, errors.New(
		"A testListener can't actually accept connections, but an" +
		" attempt was made to do so.",
	)
}

func (l *TestListener) Close() error {
	return errors.New(
		"A testListener can't actually close, but an attempt was" +
		" made to do so.",
	)
}

func (l *TestListener) Addr() net.Addr {
	return nil
}

func TestServerFromReader(t *testing.T) {
	type testStruct struct {
		config string
		reason string
	}

	RegisterResourceType(
		"internaltesthandler",
		func() json.Unmarshaler {
			return new(TestHandler)
		},
	)

	RegisterResourceType(
		"internaltestlistener",
		func() json.Unmarshaler {
			return new(TestListener)
		},
	)

	syntacticallyBad := []testStruct {
		testStruct {
			config: `not JSON`,
			reason: "bad JSON",
		},
		testStruct {
			config: `null`,
			reason: "null config",
		},
		testStruct {
			config: `{
			}`,
			reason: "empty config",
		},
		testStruct {
			config: `{
				"resources": {
					"handler": {
						"type": "dummyType",
						"data": null
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "invalid handler type",
		},
		testStruct {
			config: `{
				"resources": {
					"handler": {
						"type": "internaltesthandler",
						"data": null
					},
					"listener": {
						"type": "dummyType",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "invalid listener type",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource": {
						"type": "randomType",
						"data": null
					},
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "non-filter Resource specified as an" +
			" upstream for a Pipeline",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource": {
						"type": "randomType",
						"data": null
					},
					"handler": {
						"type": "pipeline",
						"data": {
							"downstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "non-filter Resource specified as a" +
			" downstream for a Pipeline",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource2": {
						"type": "dummyType",
						"data": null
					},
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "non-existant Resource specified as an" +
			" upstream for a Pipeline",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource": {
						"type": "dummyType",
						"data": "error"
					},
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "error during Resource creation",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource": null,
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "null Resource specified as an upstream for" +
			" a Pipeline",
		},
		testStruct {
			config: `{
				"resources": {
					"testResource": {
						"type": "ref",
						"data": "testResource"
					},
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "self-referential Resource specified as an" +
			" upstream for a Pipeline",
		},
	}

	good := []testStruct{
		testStruct{
			config: `{
				"resources": {
					"handler": {
						"type": "internaltesthandler",
						"data": null
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "test resources",
		},
		testStruct {
			config:`{
				"resources": {
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"testResource": {
						"type": "dummyType",
						"data": "foo"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "valid config",
		},
		testStruct {
			config:`{
				"resources": {
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"testResource2": {
						"type": "dummyType",
						"data": "foo"
					},
					"testResource": {
						"type": "ref",
						"data": "testResource2"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "double reference",
		},
		testStruct {
			config:`{
				"resources": {
					"handler": {
						"type": "pipeline",
						"data": {
							"upstream": [{
								"type": "ref",
								"data": "testResource"
							}]
						}
					},
					"testResource": {
						"type": "dummyRouter",
						"data": null
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"handler": "handler",
				"listener": "listener"
			}`,
			reason: "router in pipeline",
		},
	}

	RegisterResourceType("dummyType", newDummy)
	RegisterResourceType("dummyRouter", newDummyRouter)

	for _, d := range syntacticallyBad {
		d := d
		t.Run(d.reason, func(t *testing.T) {
			t.Parallel()
			s, e := ServerFromReader(strings.NewReader(d.config))
			assert.Nil(
				t,
				s,
				fmt.Sprintf(
					"ServerFromReader is expected to return a" +
					" nil config.Server when called with a" +
					" syntactically bad config, but a non-nil" +
					" config.Server was returned for such a" +
					" config: %s",
					d.reason,
				),
			)
			assert.Error(
				t,
				e,
				fmt.Sprintf(
					"ServerFromReader is expected to return an" +
					" error when called with a syntactically bad" +
					" config, but no error was returned for such" +
					" a config: %s",
					d.reason,
				),
			)
		})
	}

	for _, d := range good {
		d := d
		t.Run(d.reason, func(t *testing.T) {
			t.Parallel()
			s, e := ServerFromReader(strings.NewReader(d.config))
			assert.NotNil(
				t,
				s,
				fmt.Sprintf(
					"ServerFromReader is expected to return a" +
					" non-nil config.Server when called with a" +
					" good config, but a nil config.Server was" +
					" returned for such a config: %s",
					d.reason,
				),
			)
			assert.NoError(
				t,
				e,
				fmt.Sprintf(
					"ServerFromReader is expected to not return" +
					" an error when called with a good config," +
					" but an error was returned for such a" +
					" config: %s",
					d.reason,
				),
			)
		})
	}
}

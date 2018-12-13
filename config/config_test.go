package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
)

// we need an umarshaler, but we won't actually be using it, so...
type dummyType struct{}

func (s *dummyType) UnmarshalJSON(i []byte) error {
	if string(i) == "\"error\"" {
		return errors.New(`*dummyType.UnmarshalJSON("error")`)
	}

	return nil
}
func (s *dummyType) ServeHTTP(http.ResponseWriter, *http.Request) {
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

func TestRegisterResourceType(t *testing.T) {
	type testStruct struct {
		validate func(error)
		name     string
		newFunc  func() json.Unmarshaler
	}

	testData := []testStruct{
		{
			func(e error) {
				assert.NoError(
					t,
					e,
					"RegisterResourceType should not"+
						" return an error when a new valid"+
						" type is registerred for the first"+
						" time.",
				)
			},
			"dummyType",
			newDummy,
		},
		{
			func(e error) {
				assert.NoError(
					t,
					e,
					"RegisterResourceType should not"+
						" return an error when a new valid"+
						" type is registerred for the first"+
						" time.",
				)
			},
			"randomType",
			newRandom,
		},
		{
			func(e error) {
				assert.Error(
					t,
					e,
					"RegisterResourceType should return"+
						" an error when an attempt is made"+
						" to register a constructor for the"+
						" second time with the same type"+
						" name.",
				)
			},
			"dummyType",
			newDummy,
		},
		{
			func(e error) {
				assert.Error(
					t,
					e,
					"RegisterResourceType should return"+
						" an error if an attempt is made to"+
						" register a constructor with the"+
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
		input    []byte
	}

	testData := []testStruct{
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"An error should be generated if an"+
						" attempt is made to unmarshal into"+
						" a Resource non-JSON input as JSON.",
				)
				assert.False(
					t,
					r.complete,
					"Invalid input should not allow"+
						" Resource unmarshalling to complete.",
				)
			},
			[]byte("not json"),
		},
		{
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" attempting to unmarshal a JSON"+
						" null should not produce an error.",
				)
				assert.True(
					t,
					r.complete,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" attempting to unmarshal a JSON"+
						" null should allow the Resource"+
						" unmarshalling to ultimately"+
						" complete.",
				)
				assert.Nil(
					t,
					r.Unmarshaled,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" unmarshalling a JSON null should"+
						" result in a nil valued object.",
				)
			},
			[]byte("null"),
		},
		{
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" attempting to unmarshal an empty"+
						" JSON object (which could"+
						" reasonably be interpereted as a"+
						" null value) should not produce an"+
						" error.",
				)
				assert.True(
					t,
					r.complete,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" attempting to unmarshal an empty"+
						" JSON object (which could"+
						" reasonably be interpereted as a"+
						" null value) should allow the"+
						" Resource unmarshalling to"+
						" ultimately complete.",
				)
				assert.Nil(
					t,
					r.Unmarshaled,
					"As a null Resource value could make"+
						" sense in some circumstances,"+
						" attempting to unmarshal an empty"+
						" JSON object (which could"+
						" reasonably be interpereted as a"+
						" null value) should result in a nil"+
						" valued object.",
				)
			},
			[]byte("{}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a Resource"+
						" with a non-string type name should"+
						" produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a Resource"+
						" with a non-string type name should"+
						" not produce a Resource that has"+
						" completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":7}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a"+
						" partially defined Resource should"+
						" produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a"+
						" partially defined Resource should"+
						" not produce a Resource that has"+
						" completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":\"dummyType\"}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference"+
						" Resource with a non-string data"+
						" section should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference"+
						" Resource with a non-string data"+
						" section should not produce a"+
						" Resource that has completed its"+
						" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":7}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference"+
						" Resource with a non-string data"+
						" section should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference"+
						" Resource with a non-string data"+
						" section should not produce a"+
						" Resource that has completed its"+
						" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":null}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a reference"+
						" Resource with an unregisterred"+
						" name should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a reference"+
						" Resource with an unregisterred"+
						" name should not produce a Resource"+
						" that has completed its"+
						" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"ref\",\"data\":\"taco\"}"),
		},
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"Attempting to unmarshal a Resource"+
						" with an unregisterred type name"+
						" should produce an error.",
				)
				assert.False(
					t,
					r.complete,
					"Attempting to unmarshal a Resource"+
						" with an unregisterred type name"+
						" should not produce a Resource that"+
						" has completed its unmarshalling.",
				)
			},
			[]byte("{\"type\":\"mytype\"}"),
		},
		{
			func(r *Resource, e error) {
				assert.NoError(
					t,
					e,
					"An attempt to unmarshal a Resource"+
						" given a valid JSON input should"+
						" not produce an error.",
				)
				assert.True(
					t,
					r.complete,
					"An attempt to unmarshal a Resource"+
						" given a valid JSON input should"+
						" produce a Resource that has"+
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
		{
			func(r *Resource, e error) {
				assert.Error(
					t,
					e,
					"An attempt to unmarshal a Resource"+
						" given valid JSON input which would"+
						" cause the particular Resource"+
						" type's unmarshaler to produce an"+
						" error should result in an error"+
						" coming from the top-level"+
						" json.Unmarshal call as well.",
				)
				assert.False(
					t,
					r.complete,
					"An attempt to unmarshal a Resource"+
						" given valid JSON input which would"+
						" cause the particular Resource"+
						" type's unmarshaler to produce an"+
						" error should result in a Resource"+
						" that has not completed its"+
						" unmarshalling.",
				)
			},
			[]byte("{\"type\":\"dummyType\",\"data\":\"error\"}"),
		},
	}

	err := RegisterResourceType("dummyType", newDummy)
	assert.NoError(t, err)

	for _, v := range testData {
		var r Resource
		e := json.Unmarshal(v.input, &r)
		v.validate(&r, e)
	}
}

type TestHandler struct{}

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

type TestListener struct{}

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

	var e error
	e = RegisterResourceType(
		"internaltesthandler",
		func() json.Unmarshaler {
			return new(TestHandler)
		},
	)
	assert.NoError(t, e)

	e = RegisterResourceType(
		"internaltestlistener",
		func() json.Unmarshaler {
			return new(TestListener)
		},
	)
	assert.NoError(t, e)

	syntacticallyBad := []testStruct{
		{
			config: `not JSON`,
			reason: "bad JSON",
		},
		{
			config: `null`,
			reason: "null config",
		},
		{
			config: `{
			}`,
			reason: "empty config",
		},
		{
			config: `{
				"resources": {
					"handler": {
						"type": "randomType",
						"data": null
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "invalid handler type",
		},
		{
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
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "invalid listener type",
		},
		{
			config: `{
				"resources": {
					"testResource": {
						"type": "randomType",
						"data": null
					},
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "non-filter Resource specified as an" +
				" upstream for a Pipeline",
		},
		{
			config: `{
				"resources": {
					"testResource": {
						"type": "randomType",
						"data": null
					},
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "non-filter Resource specified as a" +
				" downstream for a Pipeline",
		},
		{
			config: `{
				"resources": {
					"testResource2": {
						"type": "dummyType",
						"data": null
					},
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "non-existant Resource specified as an" +
				" upstream for a Pipeline",
		},
		{
			config: `{
				"resources": {
					"testResource": {
						"type": "dummyType",
						"data": "error"
					},
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "error during Resource creation",
		},
		{
			config: `{
				"resources": {
					"testResource": null,
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "null Resource specified as an upstream for" +
				" a Pipeline",
		},
		{
			config: `{
				"resources": {
					"testResource": {
						"type": "ref",
						"data": "testResource"
					},
					"handler": {
						"type": "ref",
						"data": "testResource"
					},
					"listener": {
						"type": "internaltestlistener",
						"data": null
					}
				},
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "self-referential Resource specified as an" +
				" upstream for a Pipeline",
		},
		{
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
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handlerr"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "typo in handler specifier",
		},
	}

	good := []testStruct{
		{
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
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "test resources",
		},
		{
			config: `{
				"resources": {
					"handler": {
						"type": "ref",
						"data": "testResource"
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
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "valid config",
		},
		{
			config: `{
				"resources": {
					"handler": {
						"type": "ref",
						"data": "testResource"
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
				"server": {
					"type": "httpserver",
					"data": {
						"handler": {
							"type": "ref",
							"data": "handler"
						},
						"listener": {
							"type": "ref",
							"data": "listener"
						}
					}
				}
			}`,
			reason: "double reference",
		},
	}

	err := RegisterResourceType("dummyType", newDummy)
	assert.NoError(t, err)

	for _, d := range syntacticallyBad {
		d := d
		t.Run(d.reason, func(t *testing.T) {
			t.Parallel()
			parser := Parser{strings.NewReader(d.config)}
			s, e := parser.Server()
			assert.Nil(
				t,
				s,
				fmt.Sprintf(
					"ServerFromReader is expected to return a"+
						" nil config.Server when called with a"+
						" syntactically bad config, but a non-nil"+
						" config.Server was returned for such a"+
						" config: %s",
					d.reason,
				),
			)
			assert.Error(
				t,
				e,
				fmt.Sprintf(
					"ServerFromReader is expected to return an"+
						" error when called with a syntactically bad"+
						" config, but no error was returned for such"+
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
			parser := Parser{strings.NewReader(d.config)}
			s, e := parser.Server()
			assert.NotNil(
				t,
				s,
				fmt.Sprintf(
					"ServerFromReader is expected to return a"+
						" non-nil config.Server when called with a"+
						" good config, but a nil config.Server was"+
						" returned for such a config: %s",
					d.reason,
				),
			)
			assert.NoError(
				t,
				e,
				fmt.Sprintf(
					"ServerFromReader is expected to not return"+
						" an error when called with a good config,"+
						" but an error was returned for such a"+
						" config: %s",
					d.reason,
				),
			)
		})
	}
}

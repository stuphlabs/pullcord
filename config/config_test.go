package config

import (
	"encoding/json"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
	"io"
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

func TestServerFromReader(t *testing.T) {
	type testStruct struct {
		validate func(*falcore.Server, error)
		r io.Reader
	}

	testData := []testStruct {
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" given bad JSON.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server if the config is clearly" +
					" invalid.",
				)
			},
			strings.NewReader("not json"),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" given a null config.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server if the config is null.",
				)
			},
			strings.NewReader("null"),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" an empty JSON object (which could" +
					" reasonably be interpereted as a" +
					" null value) is given for the" +
					" config.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when an empty JSON object" +
					" (which could reasonably be" +
					" interpereted as a null value) is" +
					" given for the config.",

				)
			},
			strings.NewReader("{}"),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" an empty JSON object (which could" +
					" reasonably be interpereted as a" +
					" null value) is given for the" +
					" config.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when an empty JSON object" +
					" (which could reasonably be" +
					" interpereted as a null value) is" +
					" given for the config.",

				)
			},
			strings.NewReader("{}"),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" no Resource is specified for a" +
					" Pipeline.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when no Resource is" +
					" specified for a Pipeline.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": {
						"type": "dummyType",
						"data": null
					}
				},
				"pipeline": [],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" the Pipeline tries to include an" +
					" undefined Resource.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when the Pipeline tries" +
					" to include an undefined Resource.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource2": {
						"type": "dummyType",
						"data": null
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" Resource creation produces an" +
					" error.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when Resource creation" +
					" produces an error.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": {
						"type": "dummyType",
						"data": "error"
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" the Pipeline tries to use a" +
					" Resource of the wrong type.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when the Pipeline tries" +
					" to use a Resource of the wrong" +
					" type.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": null
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"ServerFromReader should fail when" +
					" a self-referrential Resource is" +
					" present in the config.",
				)
				assert.Nil(
					t,
					s,
					"ServerFromReader should not create" +
					" a server when a self-referrential" +
					" Resource is present in the config.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": {
						"type": "ref",
						"data": "testResource"
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"ServerFromReader should not fail" +
					" when given a valid config.",
				)
				assert.NotNil(
					t,
					s,
					"ServerFromReader should" +
					" successfully create a server when" +
					" given a valid config.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": {
						"type": "dummyType",
						"data": "foo"
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"ServerFromReader should not fail" +
					" when given a valid config with a" +
					" reference Resource.",
				)
				assert.NotNil(
					t,
					s,
					"ServerFromReader should" +
					" successfully create a server when" +
					" given a valid config with a" +
					" reference Resource.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource2": {
						"type": "dummyType",
						"data": "foo"
					},
					"testResource": {
						"type": "ref",
						"data": "testResource2"
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"ServerFromReader should not fail" +
					" when given a valid config with a" +
					" double reference Resource.",
				)
				assert.NotNil(
					t,
					s,
					"ServerFromReader should" +
					" successfully create a server when" +
					" given a valid config with a" +
					" double reference Resource.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource3": {
						"type": "dummyType",
						"data": "foo"
					},
					"testResource2": {
						"type": "ref",
						"data": "testResource3"
					},
					"testResource": {
						"type": "ref",
						"data": "testResource2"
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
		testStruct {
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"ServerFromReader should not fail" +
					" when given a valid config with a" +
					" Router Resource.",
				)
				assert.NotNil(
					t,
					s,
					"ServerFromReader should" +
					" successfully create a server when" +
					" given a valid config with a" +
					" Router Resource.",
				)
			},
			strings.NewReader(`{
				"resources": {
					"testResource": {
						"type": "dummyRouter",
						"data": null
					}
				},
				"pipeline": ["testResource"],
				"port": 80
			}`),
		},
	}

	RegisterResourceType("dummyType", newDummy)
	RegisterResourceType("dummyRouter", newDummyRouter)

	for _, v := range testData {
		s, e := ServerFromReader(v.r)
		v.validate(s, e)
	}
}

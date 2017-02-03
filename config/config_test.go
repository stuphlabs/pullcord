package config

import (
	"encoding/json"
	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
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
func newDummy() json.Unmarshaler {
	return new(dummyType)
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
				assert.NoError(t, e)
			},
			"dummyType",
			newDummy,
		},
		testStruct {
			func(e error) {
				assert.Error(t, e)
			},
			"dummyType",
			newDummy,
		},
		testStruct {
			func(e error) {
				assert.Error(t, e)
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
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("not json"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(t, e)
				assert.True(t, r.complete)
				assert.Nil(t, r.Unmarshaled)
			},
			[]byte("null"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(t, e)
				assert.True(t, r.complete)
				assert.Nil(t, r.Unmarshaled)
			},
			[]byte("{}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":7}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\",\"data\":7}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\",\"data\":\"taco\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"mytype\"}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.NoError(t, e)
				assert.True(t, r.complete)
				var expectedType *dummyType
				assert.IsType(t, expectedType, r.Unmarshaled)
			},
			[]byte("{\"type\":\"dummyType\",\"data\":{}}"),
		},
		testStruct {
			func(r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
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


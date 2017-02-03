package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

// we need an umarshaler, but we won't actually be using it, so...
type dummyStruct struct {}
func (s *dummyStruct) UnmarshalJSON(i []byte) error {
	return nil
}

func TestRegisterResourceType(t *testing.T) {
	type testStruct struct {
		validate func(*testing.T, error)
		name string
		newFunc func() json.Unmarshaler
	}

	testData := []testStruct {
		testStruct {
			func(tt *testing.T, e error) {
				assert.NoError(t, e)
			},
			"name1",
			func() json.Unmarshaler {
				return new(dummyStruct)
			},
		},
		testStruct {
			func(tt *testing.T, e error) {
				assert.Error(t, e)
			},
			"name1",
			func() json.Unmarshaler {
				return new(dummyStruct)
			},
		},
		testStruct {
			func(tt *testing.T, e error) {
				assert.Error(t, e)
			},
			ReferenceResourceTypeName,
			func() json.Unmarshaler {
				return new(dummyStruct)
			},
		},
	}

	for _, v := range testData {
		v.validate(
			t,
			RegisterResourceType(
				v.name,
				v.newFunc,
			),
		)
	}
}

func TestResourceUnmarshalJSON(t *testing.T) {
	type testStruct struct {
		validate func(*testing.T, *Resource, error)
		input []byte
	}

	testData := []testStruct {
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("not json"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.NoError(t, e)
				assert.True(t, r.complete)
				assert.Nil(t, r.Unmarshaled)
			},
			[]byte("null"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.NoError(t, e)
				assert.True(t, r.complete)
				assert.Nil(t, r.Unmarshaled)
			},
			[]byte("{}"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":7}"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\"}"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\",\"data\":7}"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"ref\",\"data\":\"taco\"}"),
		},
		testStruct {
			func(tt *testing.T, r *Resource, e error) {
				assert.Error(t, e)
				assert.False(t, r.complete)
			},
			[]byte("{\"type\":\"mytype\"}"),
		},
	}

	for _, v := range testData {
		var r Resource
		v.validate(
			t,
			&r,
			json.Unmarshal(v.input, &r),
		)
	}
}


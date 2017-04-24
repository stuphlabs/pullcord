package authentication

import (
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stuphlabs/pullcord/config"
	configutil "github.com/stuphlabs/pullcord/config/util"
	"strings"
	"testing"
)

func TestBadIdentifier(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"

	store := InMemPwdStore{}
	err := store.CheckPassword(identifier, password)

	assert.Error(t, err)
	assert.Equal(t, NoSuchIdentifierError, err)
}

func TestGoodPassword(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"

	hashStruct, err := GetPbkdf2Hash(password, Pbkdf2MinIterations)
	assert.NoError(t, err)
	store := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			identifier: hashStruct,
		},
	}

	err = store.CheckPassword(identifier, password)
	assert.NoError(t, err)
}

func TestBadPassword(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	badPassword := "someOtherPassword"

	hashStruct, err := GetPbkdf2Hash(password, Pbkdf2MinIterations)
	assert.NoError(t, err)
	store := InMemPwdStore{
		map[string]*Pbkdf2Hash{
			identifier: hashStruct,
		},
	}

	err = store.CheckPassword(identifier, badPassword)
	assert.Error(t, err)
	assert.Equal(t, BadPasswordError, err)
}

func TestGoodPasswordFromHash(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	jsonData := `{
		"test_user": {
			"Salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
			"Iterations" : 4096,
			"Hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, password)
	assert.NoError(t, err)
}

func TestBadPasswordFromHash(t *testing.T) {
	identifier := "test_user"
	password := "someOtherPassword"
	jsonData := `{
		"test_user": {
			"Salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
			"Iterations": 4096,
			"Hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, password)
	assert.Error(t, err)
	assert.Equal(t, BadPasswordError, err)
}

func TestInsufficientIterationsHash(t *testing.T) {
	jsonData := `{
		"test_user": {
			"Salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
			"Iterations": 4095,
			"Hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.Error(t, err)
	assert.Equal(t, InsufficientIterationsError, err)
}

func TestInsufficientIterations(t *testing.T) {
	//identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	iterations := Pbkdf2MinIterations - 1

	_, err := GetPbkdf2Hash(password, iterations)
	assert.Error(t, err)
	assert.Equal(t, InsufficientIterationsError, err)
}

func TestIncorrectSaltLengthError(t *testing.T) {
	jsonData := `{
		"test_user": {
			"Salt": "WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
			"Iterations": 4096,
			"Hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.Error(t, err)
	assert.Equal(t, IncorrectSaltLengthError, err)
}

func TestIncorrectHashLengthError(t *testing.T) {
	jsonData := `{
		"test_user": {
			"Salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
			"Iterations": 4096,
			"Hash": "0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.Error(t, err)
	assert.Equal(t, IncorrectHashLengthError, err)
}

func TestBadBase64Error(t *testing.T) {
	//identifier := "test_user"
	jsonData := `{
		"test_user": {
			"Salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ",
			"Iterations": 4096,
			"Hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
		}
	}`

	var store InMemPwdStore
	err := json.Unmarshal([]byte(jsonData), &store)
	assert.Error(t, err)
}

func TestInMemPwdStoreFromConfig(t *testing.T) {
	type testStruct struct {
		validator func(json.Unmarshaler) error
		data string
		serverValidate func(*falcore.Server, error)
	}

	testData := []testStruct {
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			``,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an incomplete validator" +
					" resource should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "inmempwdstore",
				"data": ["test_user"]
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" inmempwdstore resource should" +
					" produce an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an incomplete" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an incomplete" +
					" inmempwdstore resource should" +
					" be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "inmempwdstore",
				"data": {
					"test_user": {}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an invalid nested" +
					" resource should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				switch i := i.(type) {
				case *InMemPwdStore:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							i,
						),
					)
				}

				return nil
			},
			`{
				"type": "inmempwdstore",
				"data": {
					"test_user": {
						"salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
						"iterations": 4096,
						"hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing only a passing" +
					" validator resource should not" +
					" produce an error. The most likely" +
					" explanation is that the validator" +
					" resource is not passing.",
				)
				assert.NotNil(
					t,
					s,
					"A server created from a config" +
					" containing only a" +
					" passing validator resource should" +
					" not be nil. The most likely" +
					" explanation is that the validator" +
					" resource is not passing.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				switch i := i.(type) {
				case *InMemPwdStore:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							i,
						),
					)
				}

				return nil
			},
			`{
				"type": "inmempwdstore",
				"data": {
					"test_user": {
						"salt": 7,
						"iterations": 4096,
						"hash": -5
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing a nested" +
					" resource with the wrong type" +
					" should produce an error.",
				)
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing a nested resource with" +
					" the wrong type should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				switch i := i.(type) {
				case *InMemPwdStore:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							i,
						),
					)
				}

				return nil
			},
			`{
				"type": "inmempwdstore",
				"data": {
					"test_user": {
						"salt": "RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==",
						"iterations": "Four thousand ninety six",,
						"hash": "3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2CyYR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg=="
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing a nested" +
					" resource with the wrong type" +
					" should produce an error.",
				)
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing a nested resource with" +
					" the wrong type should be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "inmempwdstore",
				"data": 42
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" inmempwdstore resource should" +
					" produce an error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" validator resource should produce" +
					" an error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an invalid" +
					" inmempwdstore resource should" +
					" be nil.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				switch i := i.(type) {
				case *InMemPwdStore:
					// do nothing
				default:
					return errors.New(
						fmt.Sprintf(
							"Expecting" +
							" unmarsheled" +
							" resource to be a" +
							" inmempwdstore," +
							" but instead got: %v",
							i,
						),
					)
				}

				return nil
			},
			`{
				"type": "inmempwdstore",
				"data": {}
			}`,
			func(s *falcore.Server, e error) {
				assert.NoError(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing only a passing" +
					" validator resource (even if that" +
					" resource is an empty" +
					" inmempwdstore) should not  produce" +
					" an error. The most likely" +
					" explanation is that the validator" +
					" resource is not passing.",
				)
				assert.NotNil(
					t,
					s,
					"A server created from a config" +
					" containing only a passing" +
					" validator resource (even if that" +
					" resource is an empty" +
					" inmempwdstore) should  not be nil." +
					" The most likely  explanation is" +
					" that the validator resource is not" +
					" passing.",
				)
			},
		},
		testStruct {
			func(i json.Unmarshaler) error {
				return errors.New(
					fmt.Sprintf(
						"Not expecting validator to" +
						" actually run, but it was" +
						" run with: %v",
						i,
					),
				)
			},
			`{
				"type": "inmempwdstore",
				"data": {
					"test_user": {
						"hash": "hey does this look base64 to you?",
						"iterations": 4096,
						"salt": "maybe it's base65?"
					}
				}
			}`,
			func(s *falcore.Server, e error) {
				assert.Error(
					t,
					e,
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error.",
				)
				if e != nil {
				assert.False(
					t,
					strings.HasPrefix(
						e.Error(),
						"Not expecting validator to" +
						" actually run",
					),
					"Attempting to create a server from" +
					" a config containing an invalid" +
					" nested resource should produce an" +
					" error apart from any created by" +
					" the validator.",
				)
				}
				assert.Nil(
					t,
					s,
					"A server created from a config" +
					" containing an invalid nested" +
					" resource should be nil.",
				)
			},
		},
	}

	for _, v := range testData {
		n, e := configutil.GenerateValidator(v.validator)
		assert.NoError(
			t,
			e,
			"Generating a validator resource type should not" +
			" produce an error.",
		)
		assert.NotEqual(
			t,
			n,
			"",
			"A generated validator resource type should not have" +
			" an empty resource type name.",
		)

		s, e := config.ServerFromReader(
			strings.NewReader(
				fmt.Sprintf(
					`{
						"resources": {
							"validator": {
								"type": "%s",
								"data": %s
							}
						},
						"pipeline": ["validator"],
						"port": 80
					}`,
					n,
					v.data,
				),
			),
		)
		v.serverValidate(s, e)
	}
}

package authentication

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	// "github.com/stuphlabs/pullcord"
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


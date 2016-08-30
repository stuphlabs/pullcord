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

	store := NewInMemPwdStore()
	err := store.CheckPassword(identifier, password)

	assert.Error(t, err)
	assert.Equal(t, NoSuchIdentifierError, err)
}

func TestGoodPassword(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"

	store := NewInMemPwdStore()
	err := store.SetPassword(identifier, password, Pbkdf2MinIterations)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, password)
	assert.NoError(t, err)
}

func TestBadPassword(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	badPassword := "someOtherPassword"

	store := NewInMemPwdStore()
	err := store.SetPassword(identifier, password, Pbkdf2MinIterations)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, badPassword)
	assert.Error(t, err)
	assert.Equal(t, BadPasswordError, err)
}

func TestGoodPasswordFromHash(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	jsonHashStruct := "{\"Salt\":\"RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==\",\"Ite" +
		"rations\":4096,\"Hash\":\"3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, password)
	assert.NoError(t, err)
}

func TestBadPasswordFromHash(t *testing.T) {
	identifier := "test_user"
	password := "someOtherPassword"
	jsonHashStruct := "{\"Salt\":\"RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==\",\"Ite" +
		"rations\":4096,\"Hash\":\"3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.NoError(t, err)

	err = store.CheckPassword(identifier, password)
	assert.Error(t, err)
	assert.Equal(t, BadPasswordError, err)
}

func TestInsufficientIterationsHash(t *testing.T) {
	identifier := "test_user"
	jsonHashStruct := "{\"Salt\":\"RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==\",\"Ite" +
		"rations\":4096,\"Hash\":\"3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)
	hashStruct.Iterations = Pbkdf2MinIterations - 1

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.Error(t, err)
	assert.Equal(t, InsufficientIterationsError, err)
}

func TestInsufficientIterations(t *testing.T) {
	identifier := "test_user"
	password := "SuperAwes0meP@ssword"
	iterations := Pbkdf2MinIterations - 1

	store := NewInMemPwdStore()
	err := store.SetPassword(identifier, password, iterations)
	assert.Error(t, err)
	assert.Equal(t, InsufficientIterationsError, err)
}

func TestIncorrectSaltLengthError(t *testing.T) {
	identifier := "test_user"
	jsonHashStruct := "{\"Salt\":\"WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==\",\"Ite" +
		"rations\":4096,\"Hash\":\"3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.Error(t, err)
	assert.Equal(t, IncorrectSaltLengthError, err)
}

func TestIncorrectHashLengthError(t *testing.T) {
	identifier := "test_user"
	jsonHashStruct := "{\"Salt\":\"RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ==\",\"Ite" +
		"rations\":4096,\"Hash\":\"0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.Error(t, err)
	assert.Equal(t, IncorrectHashLengthError, err)
}

func TestBadBase64Error(t *testing.T) {
	identifier := "test_user"
	jsonHashStruct := "{\"Salt\":\"RMM0WEV4s0vxZWb9Yvw0ooBU1Bs9louzqNsa+" +
		"/E/SVzZg+ez72TLoXL8pFOOzk2aOFO5XLtbSECYKUK7XtF+ZQ\",\"Ite" +
		"rations\":4096,\"Hash\":\"3Ezu0RAlDXNhkvnVq0H4z/0dUrItfd2Cy" +
		"YR06u/arA6f9XAeAA0UWWB/9y/0fQOVmZi7XxyiePtR/hC33tNWXg==\"}"
	var hashStruct Pbkdf2HashArg
	err := json.Unmarshal([]byte(jsonHashStruct), &hashStruct)
	assert.NoError(t, err)

	store := NewInMemPwdStore()
	err = store.SetPbkdf2Hash(identifier, hashStruct)
	assert.Error(t, err)
}


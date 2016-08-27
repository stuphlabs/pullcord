package authentication

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	// "github.com/stuphlabs/pullcord"
	"golang.org/x/crypto/pbkdf2"
)

// Pbkdf2KeyLength is the length (in bytes) of the generated PBKDF2 hashes.
const Pbkdf2KeyLength = 64

// Pbkdf2MinIterations is the minimum number of iterations allowed for PBKDF2
// hashes.
const Pbkdf2MinIterations = 4096

type errorString string

func (str errorString) Error() string {
	return string(str)
}

// InsufficientIterationsError is the error object that is returned if the
// requested number of iterations for a new PBKDF2 hash is less than
// Pbkdf2MinIterations.
const InsufficientIterationsError = errorString(
	"The number of iterations must be at least Pbkdf2MinIterations",
)

// InsufficientIterationsError is the error object that is returned if the
// operating system does not have enough entropy to generated a random salt of
// length Pbkdf2KeyLength.
const InsufficientEntropyError = errorString(
	"The amount of entropy available from the operating system was not" +
	" enough to generate a salt of length Pbkdf2KeyLength",
)

// NoSuchIdentifierError is the error object that is returned if the given
// identifier (probably a username) does not exist in the password store.
//
// It is considered best practice to not indicate to a possible attacker
// whether an authentication attempt failed due to a bad password or due to
// a non-existent user. However, while this implementation makes a few very
// modest attempts to reduce time-based information leakage, the way the
// identifier lookup process is implemented is likely to leak information about
// the presence of a user. Perhaps that issue will be fixed at a later time,
// but it is worth at least knowing for the time being.
const NoSuchIdentifierError = errorString(
	"The given identifier does not have an entry in the password store",
)

// BadPasswordError is the error object that is returned if the given
// identifier (probably a username) does exist in the password store, but the
// given password does not generate a matching hash.
const BadPasswordError = errorString(
	"The hash generated from the given password does not match the hash" +
	" associated with the given identifier in the password store",
)

type pbkdf2Hash struct {
	hash []byte
	salt []byte
	iterations uint16
}

// InMemPwdStore is a basic password store where all the identifiers and hash
// information are stored in memory. This would likely not be a useful password
// store implementation in a production environment, but it can be useful in
// testing. All passwords are hashed using PBKDF2 with SHA-256.
type InMemPwdStore struct {
	table map[string]*pbkdf2Hash
}

// InMemPwdStore generates a new InMemPwdStore instance.
func NewInMemPwdStore() *InMemPwdStore {
	log().Info("initializing in-memory password store")

	var store InMemPwdStore
	store.table = make(map[string]*pbkdf2Hash)

	return &store
}

// SetPassword is a function that allows a password to be hashed and added to
// an InMemPwdStore instance.
func (store *InMemPwdStore) SetPassword(
	identifier string,
	password string,
	iterations uint16,
) error {
	if iterations < Pbkdf2MinIterations {
		return InsufficientIterationsError
	}

	var hashStruct pbkdf2Hash
	hashStruct.salt = make([]byte, Pbkdf2KeyLength)

	randCount, err := rand.Read(hashStruct.salt)
	if err != nil {
		return err
	} else if randCount != Pbkdf2KeyLength {
		return InsufficientEntropyError
	}

	hashStruct.iterations = iterations

	hashStruct.hash = pbkdf2.Key(
		[]byte(password),
		hashStruct.salt,
		int(hashStruct.iterations),
		Pbkdf2KeyLength,
		sha256.New,
	)

	store.table[identifier] = &hashStruct

	return nil
}

// SetPbkdf2Hash is a function that allows a PBKDF2 hash to be explicitly added
// to an InMemPwdStore instance. This could be useful in order to create an
// immutable password store based on a static configuration file (i.e. by
// adding the hashes one at a time to an InMemPwdStore instance).
func (store *InMemPwdStore) SetPbkdf2Hash(
	identifier string,
	hash [Pbkdf2KeyLength]byte,
	salt [Pbkdf2KeyLength]byte,
	iterations uint16,
) error {
	if iterations < Pbkdf2MinIterations {
		return InsufficientIterationsError
	}

	hashStruct := pbkdf2Hash{
		hash[:],
		salt[:],
		iterations,
	}

	store.table[identifier] = &hashStruct

	return nil
}

// CheckPassword implements the required password checking function to make
// InMemPwdStore a PasswordChecker implementation.
func (store *InMemPwdStore) CheckPassword(
	identifier string,
	password string,
) error {
	hashStruct, present := store.table[identifier]
	if ! present {
		return NoSuchIdentifierError
	}

	genHash := pbkdf2.Key(
		[]byte(password),
		hashStruct.salt,
		int(hashStruct.iterations),
		Pbkdf2KeyLength,
		sha256.New,
	)

	if 1 == subtle.ConstantTimeCompare(hashStruct.hash, genHash) {
		return nil
	} else {
		return BadPasswordError
	}
}


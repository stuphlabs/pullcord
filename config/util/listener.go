package util

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/stuphlabs/pullcord/config"
)

// TestListener is a dummy net.Listener that can be used for testing.
type TestListener struct {
}

func init() {
	e := config.RegisterResourceType(
		"testlistener",
		func() json.Unmarshaler {
			return new(TestListener)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (t *TestListener) UnmarshalJSON([]byte) error {
	return nil
}

// Accept implements net.Listener.
func (t *TestListener) Accept() (net.Conn, error) {
	return nil, errors.New(
		"TestListener is not intended to be used as a real listener",
	)
}

// Close implements net.Listener.
func (t *TestListener) Close() error {
	return errors.New(
		"TestListener is not intended to be used as a real listener",
	)
}

// Addr implements net.Listener.
func (t *TestListener) Addr() net.Addr {
	return nil
}

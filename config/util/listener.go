package util

import (
	"encoding/json"
	"errors"
	"github.com/stuphlabs/pullcord/config"
	"net"
)

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

func (t *TestListener) UnmarshalJSON([]byte) error {
	return nil
}

func (t *TestListener) Accept() (net.Conn, error) {
	return nil, errors.New(
		"TestListener is not intended to be used as a real listener.",
	)
}

func (t *TestListener) Close() error {
	return errors.New(
		"TestListener is not intended to be used as a real listener.",
	)
}

func (t *TestListener) Addr() net.Addr {
	return nil
}

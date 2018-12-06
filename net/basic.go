package net

import (
	"encoding/json"
	"net"

	"github.com/stuphlabs/pullcord/config"
)

// BasicListener augments the functionality of a net.Listen call by wrapping it
// in an encoding/json.Unmarshaler (and thereby making it be configurable).
type BasicListener struct {
	Listener net.Listener
}

func init() {
	e := config.RegisterResourceType(
		"basiclistener",
		func() json.Unmarshaler {
			return new(BasicListener)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (b *BasicListener) UnmarshalJSON(d []byte) error {
	var t struct {
		Proto string
		Laddr string
	}

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	l, e := net.Listen(t.Proto, t.Laddr)
	if e != nil {
		return e
	}

	b.Listener = l
	return nil
}

// Accept implements net.Listener.
func (b *BasicListener) Accept() (net.Conn, error) {
	return b.Listener.Accept()
}

// Close implements net.Listener.
func (b *BasicListener) Close() error {
	return b.Listener.Close()
}

// Addr implements net.Listener.
func (b *BasicListener) Addr() net.Addr {
	return b.Listener.Addr()
}

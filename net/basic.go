package net

import (
	"encoding/json"
	"github.com/stuphlabs/pullcord/config"
	"net"
)

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

func (b *BasicListener) UnmarshalJSON(d []byte) error {
	var t struct {
		Proto string
		Laddr string
	}

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	if l, e := net.Listen(t.Proto, t.Laddr); e != nil {
		return e
	} else {
		b.Listener = l
		return nil
	}
}

func (b *BasicListener) Accept() (net.Conn, error) {
	return b.Listener.Accept()
}

func (b *BasicListener) Close() error {
	return b.Listener.Close()
}

func (b *BasicListener) Addr() net.Addr {
	return b.Listener.Addr()
}

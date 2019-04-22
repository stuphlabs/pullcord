package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"

	"github.com/proidiot/gone/log"

	"github.com/stuphlabs/pullcord/config"
)

// TLSCertificateGetter provides a mechanism by which the assignment of the
// crypto/tls.Config.GetCertificate method can be abstracted independently from
// other aspects of the creation of the net.Listener returned from a call to
// crypto/tls.NewListener.
type TLSCertificateGetter interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

// BasicTLSListener combines the certificate retrieval process abstracted by a
// TLSCertificateGetter with a supplied net.Listener to give a simple
// configurable wrapper around a call to crypto/tls.NewListener.
type BasicTLSListener struct {
	Listener       net.Listener
	CertGetter     TLSCertificateGetter
	actualListener net.Listener
}

func init() {
	e := config.RegisterResourceType(
		"basictlslistener",
		func() json.Unmarshaler {
			return new(BasicTLSListener)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (b *BasicTLSListener) UnmarshalJSON(d []byte) error {
	var t struct {
		Listener   config.Resource
		CertGetter config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	var ok bool
	b.Listener, ok = t.Listener.Unmarshalled.(net.Listener)
	if !ok {
		_ = log.Debug(
			fmt.Sprintf(
				"Resource is not a net.Listener: %#v",
				t.Listener.Unmarshalled,
			),
		)
		return config.UnexpectedResourceType
	}

	b.CertGetter, ok = t.CertGetter.Unmarshalled.(TLSCertificateGetter)
	if !ok {
		_ = log.Debug(
			fmt.Sprintf(
				"Resource is not a TLSCertificateGetter: %#v",
				t.CertGetter.Unmarshalled,
			),
		)
		return config.UnexpectedResourceType
	}

	b.assureActualListenerCreated()

	return nil
}

func (b *BasicTLSListener) assureActualListenerCreated() {
	if b.actualListener == nil {
		b.actualListener = tls.NewListener(
			b.Listener,
			&tls.Config{
				GetCertificate: b.CertGetter.GetCertificate,
			},
		)
	}
}

// Accept implements net.Listener.
func (b *BasicTLSListener) Accept() (net.Conn, error) {
	b.assureActualListenerCreated()
	return b.actualListener.Accept()
}

// Close implements net.Listener.
func (b *BasicTLSListener) Close() error {
	b.assureActualListenerCreated()
	return b.actualListener.Close()
}

// Addr implements net.Listener.
func (b *BasicTLSListener) Addr() net.Addr {
	b.assureActualListenerCreated()
	return b.actualListener.Addr()
}

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

// TlsCertificateGetter provides a mechanism by which the assignment of the
// crypto/tls.Config.GetCertificate method can be abstracted independently from
// other aspects of the creation of the net.Listener returned from a call to
// crypto/tls.NewListener.
type TlsCertificateGetter interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

// BasicTlsListener combines the certificate retrieval process abstracted by a
// TlsCertificateGetter with a supplied net.Listener to give a simple
// configurable wrapper around a call to crypto/tls.NewListener.
type BasicTlsListener struct {
	Listener       net.Listener
	CertGetter     TlsCertificateGetter
	actualListener net.Listener
}

func init() {
	e := config.RegisterResourceType(
		"basictlslistener",
		func() json.Unmarshaler {
			return new(BasicTlsListener)
		},
	)

	if e != nil {
		panic(e)
	}
}

func (b *BasicTlsListener) UnmarshalJSON(d []byte) error {
	var t struct {
		Listener   config.Resource
		CertGetter config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	var ok bool
	b.Listener, ok = t.Listener.Unmarshaled.(net.Listener)
	if !ok {
		log.Debug(
			fmt.Sprintf(
				"Resource is not a net.Listener: %#v",
				t.Listener.Unmarshaled,
			),
		)
		return config.UnexpectedResourceType
	}

	b.CertGetter, ok = t.CertGetter.Unmarshaled.(TlsCertificateGetter)
	if !ok {
		log.Debug(
			fmt.Sprintf(
				"Resource is not a TlsCertificateGetter: %#v",
				t.CertGetter.Unmarshaled,
			),
		)
		return config.UnexpectedResourceType
	}

	b.assureActualListenerCreated()

	return nil
}

func (b *BasicTlsListener) assureActualListenerCreated() {
	if b.actualListener == nil {
		b.actualListener = tls.NewListener(
			b.Listener,
			&tls.Config{
				GetCertificate: b.CertGetter.GetCertificate,
			},
		)
	}
}

func (b *BasicTlsListener) Accept() (net.Conn, error) {
	b.assureActualListenerCreated()
	return b.actualListener.Accept()
}

func (b *BasicTlsListener) Close() error {
	b.assureActualListenerCreated()
	return b.actualListener.Close()
}

func (b *BasicTlsListener) Addr() net.Addr {
	b.assureActualListenerCreated()
	return b.actualListener.Addr()
}

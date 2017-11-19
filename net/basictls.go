package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/stuphlabs/pullcord/config"
	"net"
)

type TlsCertificateGetter interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

type BasicTlsListener struct {
	Listener net.Listener
	TlsConfig TlsCertificateGetter
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
		Listener config.Resource
		CertGetter config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	var ok bool
	b.Listener, ok = t.Listener.Unmarshaled.(net.Listener)
	if !ok {
		return config.UnexpectedResourceType
	}

	b.TlsConfig, ok = t.CertGetter.Unmarshaled.(TlsCertificateGetter)
	if !ok {
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
				GetCertificate: b.TlsConfig.GetCertificate,
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

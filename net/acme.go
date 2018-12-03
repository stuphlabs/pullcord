package net

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"

	"github.com/stuphlabs/pullcord/config"

	"golang.org/x/crypto/acme/autocert"
)

func init() {
	config.MustRegisterResourceType(
		"acme",
		func() json.Unmarshaler {
			return new(AcmeConfig)
		},
	)
}

// AcmeConfig represents the configuration details to be used with
// x/crypto/acme/autocert. It can also act as a wrapper around the
// autocert.Manager and net.Listener it generates.
type AcmeConfig struct {
	AcceptTOS bool
	Domains   []string
	mgr       *autocert.Manager
	lsr       net.Listener
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (a *AcmeConfig) UnmarshalJSON(d []byte) error {
	var t struct {
		AcceptTOS bool
		Domains   []string
	}

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	if !t.AcceptTOS {
		return errors.New(
			"The terms of service must be accepted in order to" +
				" use the default ACME setup.",
		)
	}

	a.AcceptTOS = t.AcceptTOS
	a.Domains = t.Domains

	return nil
}

// GetManager retrieves an x/crypto/acme/autocert.Manager with the previously
// defined config values. Subsequent calls return the same autocert.Manager
// object which has been preserved.
func (a *AcmeConfig) GetManager() (*autocert.Manager, error) {
	if a.mgr != nil {
		return a.mgr, nil
	}

	if !a.AcceptTOS {
		return nil, errors.New(
			"The terms of service must be accepted in order to" +
				" use the default ACME setup.",
		)
	}

	a.mgr = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Domains...),
	}

	return a.mgr, nil
}

// Listener retrieves a net.Listener from the autocert.Manager given by
// AcmeConfig.GetManager. Subsequent calls return the same net.Listener object
// which has been preserved.
func (a *AcmeConfig) Listener() (net.Listener, error) {
	if a.lsr != nil {
		return a.lsr, nil
	}

	mgr, e := a.GetManager()
	if e != nil {
		return nil, e
	}

	a.lsr = mgr.Listener()

	return a.lsr, nil
}

// GetCertificate implements TlsCertificateGetter so that this object can be
// used as part of another net.Listener in case behavior other than what is
// provided by AcmeConfig.Listener is desired.
func (a *AcmeConfig) GetCertificate(
	hello *tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	mgr, e := a.GetManager()
	if e != nil {
		return nil, e
	}

	return mgr.GetCertificate(hello)
}

// Accept implements net.Listener.
func (a *AcmeConfig) Accept() (net.Conn, error) {
	lsr, e := a.Listener()
	if e != nil {
		return nil, e
	}

	return lsr.Accept()
}

// Close implements net.Listener.
func (a *AcmeConfig) Close() error {
	lsr, e := a.Listener()
	if e != nil {
		return e
	}

	return lsr.Close()
}

// Addr implements net.Listener.
func (a *AcmeConfig) Addr() net.Addr {
	lsr, e := a.Listener()
	if e != nil {
		return nil
	}

	return lsr.Addr()
}

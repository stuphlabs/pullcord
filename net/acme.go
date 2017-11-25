package net

import (
	"crypto/tls"
	"encoding/json"
	"errors"

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

type AcmeConfig struct {
	AcceptTOS bool
	Domains   []string
}

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

func (a *AcmeConfig) GetManager() (*autocert.Manager, error) {
	if !a.AcceptTOS {
		return nil, errors.New(
			"The terms of service must be accepted in order to" +
				" use the default ACME setup.",
		)
	}

	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Domains...),
	}, nil
}

func (a *AcmeConfig) GetCertificate(
	hello *tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	mgr, e := a.GetManager()
	if e != nil {
		return nil, e
	}

	return mgr.GetCertificate(hello)
}

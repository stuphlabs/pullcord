package net

import (
	"golang.org/x/crypto/acme/autocert"
)

type AcmeConfig struct {
	AcceptTOS bool
	Domains []string
	acmeMgr autocert.Manager
}

func (a *AcmeConfig) UnmarshalJSON(d []byte) error {
	var t struct {
		AcceptTOS bool
		Domains []string
	}

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	if e := a.assureSetup(); e != nil {
		return e
	}

	return nil
}

func (a *AcmeConfig) assureSetup() error {
	if !a.AcceptTOS {
		return errors.New(
			"The terms of service must be accepted in order to use"+
				" the default ACME setup.",
		)
	}

	a.acmeMgr = autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Domains...),
	}

	return nil
}

func (a *AcmeConfig) GetCertificate() (Certificate, error) {
	if e := a.assureSetup(); e != nil {
		return nil, e
	}

	return acmeMgr.GetCertificate()
}

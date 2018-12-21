package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/proidiot/gone/log"
)

// HTTPServer implements the Pullcord server interface with an HTTP handler.
type HTTPServer struct {
	Listener net.Listener
	Handler  http.Handler
}

func init() {
	e := RegisterResourceType(
		"httpserver",
		func() json.Unmarshaler {
			return new(HTTPServer)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (s *HTTPServer) UnmarshalJSON(d []byte) error {
	var t struct {
		Listener Resource
		Handler  Resource
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	var ok bool
	s.Listener, ok = t.Listener.Unmarshaled.(net.Listener)
	if !ok {
		return UnexpectedResourceType
	}

	s.Handler, ok = t.Handler.Unmarshaled.(http.Handler)
	if !ok {
		return UnexpectedResourceType
	}

	return nil
}

// Serve implements .../pullcord.Server.
func (s *HTTPServer) Serve() error {
	_ = log.Debug(
		fmt.Sprintf(
			"Serving with listener %#v and handler %#v",
			s.Listener,
			s.Handler,
		),
	)

	e := log.Notice(
		fmt.Sprintf("Starting server at %s...", s.Listener.Addr()),
	)
	if e != nil {
		e = http.Serve(s.Listener, s.Handler)
	}
	return e
}

// Close implements .../pullcord.Server.
func (s *HTTPServer) Close() error {
	_ = log.Info(fmt.Sprintf("Closing server at %s...", s.Listener.Addr()))
	return s.Listener.Close()
}

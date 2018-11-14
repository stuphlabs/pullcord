package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/proidiot/gone/log"
)

// HttpServer implements the Pullcord server interface with an HTTP handler.
type HttpServer struct {
	Listener net.Listener
	Handler  http.Handler
}

func init() {
	e := RegisterResourceType(
		"httpserver",
		func() json.Unmarshaler {
			return new(HttpServer)
		},
	)

	if e != nil {
		panic(e)
	}
}

func (s *HttpServer) UnmarshalJSON(d []byte) error {
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

func (s *HttpServer) Serve() error {
	log.Debug(
		fmt.Sprintf(
			"Serving with listener %#v and handler %#v",
			s.Listener,
			s.Handler,
		),
	)

	log.Notice(fmt.Sprintf("Starting server at %s...", s.Listener.Addr()))
	return http.Serve(s.Listener, s.Handler)
}

func (s *HttpServer) Close() error {
	log.Info(fmt.Sprintf("Closing server at %s...", s.Listener.Addr()))
	return s.Listener.Close()
}

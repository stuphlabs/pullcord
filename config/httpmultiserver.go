package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/proidiot/gone/log"
)

// HTTPMultiServer implements the Pullcord server interface with an HTTP handler
// and multiple listeners.
type HTTPMultiServer struct {
	Listeners []net.Listener
	Handler   http.Handler
}

func init() {
	e := RegisterResourceType(
		"httpmultiserver",
		func() json.Unmarshaler {
			return new(HTTPServer)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (s *HTTPMultiServer) UnmarshalJSON(d []byte) error {
	var t struct {
		Listeners []Resource
		Handler   Resource
	}

	dec := json.NewDecoder(bytes.NewReader(d))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	s.Listeners = make([]net.Listener, len(t.Listeners))

	var ok bool
	for _, r := range t.Listeners {
		l, ok := r.Unmarshaled.(net.Listener)
		if !ok {
			return UnexpectedResourceType
		}
		s.Listeners = append(s.Listeners, l)
	}

	s.Handler, ok = t.Handler.Unmarshaled.(http.Handler)
	if !ok {
		return UnexpectedResourceType
	}

	return nil
}

// Serve implements .../pullcord/Server.
func (s *HTTPMultiServer) Serve() error {
	log.Debug(
		fmt.Sprintf(
			"Serving with listeners %#v and handler %#v",
			s.Listeners,
			s.Handler,
		),
	)
	defer func() {
		_ = s.Close()
	}()

	errChan := make(chan error)
	for _, l := range s.Listeners {
		go func(gl net.Listener, gErrChan chan<- error) {
			log.Notice(
				fmt.Sprintf(
					"Starting server at %s...",
					gl.Addr(),
				),
			)
			gErrChan <- http.Serve(gl, s.Handler)
		}(l, errChan)
	}

	err := <-errChan

	go func(gErrChan <-chan error, eat int) {
		for ; eat > 0; eat-- {
			_ = <-gErrChan
		}
	}(errChan, len(s.Listeners)-1)

	return err
}

// Close implements .../pullcord/Server.
func (s *HTTPMultiServer) Close() error {
	var err error
	for _, l := range s.Listeners {
		log.Info(fmt.Sprintf("Closing server at %s...", l.Addr()))
		thisErr := l.Close()
		if err == nil && thisErr != nil {
			err = thisErr
		}
	}
	return err
}

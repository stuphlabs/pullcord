package util

import (
	"encoding/json"
	"net/http"

	"github.com/stuphlabs/pullcord/config"
)

// TestHandler is a dummy net/http.Handler implementation that can be used for
// testing.
type TestHandler struct {
}

func init() {
	e := config.RegisterResourceType(
		"testhandler",
		func() json.Unmarshaler {
			return new(TestHandler)
		},
	)

	if e != nil {
		panic(e)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (t *TestHandler) UnmarshalJSON([]byte) error {
	return nil
}

// ServeHTTP implements net/http.Handler.
func (t *TestHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
}

package util

import (
	"encoding/json"
	"github.com/stuphlabs/pullcord/config"
	"net/http"
)

type TestHandler struct {
}

func init() {
	e :=config.RegisterResourceType(
		"testhandler",
		func() json.Unmarshaler {
			return new(TestHandler)
		},
	)

	if e != nil {
		panic(e)
	}
}

func (t *TestHandler) UnmarshalJSON([]byte) error {
	return nil
}

func (t *TestHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
}

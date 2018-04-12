package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/proidiot/gone/errors"
	"github.com/proidiot/gone/log"
)

const UnexpectedResourceType = errors.New(
	"The requested resource does not have the expected type.",
)
const ReferenceResourceTypeName = "ref"

var typeRegistry = make(map[string]func() json.Unmarshaler)

var registry map[string]*Resource
var unregisterredResources map[string]json.RawMessage
var registrationMutex sync.Mutex

func RegisterResourceType(
	typeName string,
	newFunc func() json.Unmarshaler,
) error {
	_, present := typeRegistry[typeName]
	if present || typeName == ReferenceResourceTypeName {
		return errors.New(
			fmt.Sprintf(
				"More than one resource type has registerred"+
					" the same name: %s",
				typeName,
			),
		)
	}

	typeRegistry[typeName] = newFunc
	return nil
}

func MustRegisterResourceType(
	typeName string,
	newFunc func() json.Unmarshaler,
) {
	e := RegisterResourceType(typeName, newFunc)
	if e != nil {
		panic(e)
	}
}

type Resource struct {
	Unmarshaled json.Unmarshaler
	complete    bool
}

func (rsc *Resource) UnmarshalJSON(input []byte) error {
	var newRscDef struct {
		Type string
		Data json.RawMessage
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&newRscDef); e != nil {
		log.Crit(
			fmt.Sprintf(
				"Unable to decode resource definition: %s",
				e.Error(),
			),
		)
		return e
	}

	if newRscDef.Type == "" && newRscDef.Data == nil {
		rsc.Unmarshaled = nil
		rsc.complete = true
		return nil
	}

	if newRscDef.Type == ReferenceResourceTypeName {
		var name string
		if e := json.Unmarshal(newRscDef.Data, &name); e != nil {
			log.Crit(
				fmt.Sprintf(
					"Unable to decode the resource"+
						" string: %s",
					e.Error(),
				),
			)
			return e
		}

		d, present := registry[name]
		if present {
			if d.complete {
				rsc.Unmarshaled = d.Unmarshaled
				rsc.complete = true
				return nil
			} else {
				e := errors.New(
					fmt.Sprintf(
						"The resource depenency was"+
							" already under"+
							" construction"+
							" (implying a cyclic"+
							" dependency): %s",
						name,
					),
				)
				log.Crit(e.Error())
				return e
			}
		}

		return rsc.unmarshalByName(name)
	}

	newFunc, present := typeRegistry[newRscDef.Type]
	if !present {
		return errors.New(
			fmt.Sprintf(
				"The specified resource type is not a"+
					" registerred resource type: %s",
				newRscDef.Type,
			),
		)
	}

	u := newFunc()
	if e := json.Unmarshal(newRscDef.Data, u); e != nil {
		return e
	}
	rsc.Unmarshaled = u
	rsc.complete = true
	return nil
}

func (r *Resource) unmarshalByName(name string) error {
	if d, present := unregisterredResources[name]; !present {
		return errors.New(
			fmt.Sprintf(
				"No resource specified with name: %s",
				name,
			),
		)
	} else {
		return json.Unmarshal(d, r)
	}
}

type Server struct {
	Listener net.Listener
	Handler  http.Handler
}

func (s *Server) Serve() error {
	log.Debug(
		fmt.Sprintf(
			"Serving with listener %#v and handler %#v",
			s.Listener,
			s.Handler,
		),
	)
	return http.Serve(s.Listener, s.Handler)
}

func ServerFromReader(r io.Reader) (*Server, error) {
	registrationMutex.Lock()
	defer registrationMutex.Unlock()

	var config struct {
		Resources map[string]json.RawMessage
		Listener  string
		Handler   string
	}

	dec := json.NewDecoder(r)
	registry = make(map[string]*Resource)

	if e := dec.Decode(&config); e != nil {
		log.Crit(
			fmt.Sprintf(
				"Unable to decode server config: %#v",
				e,
			),
		)
		return nil, e
	}

	if _, present := config.Resources[config.Listener]; !present {
		e := errors.New(
			fmt.Sprintf(
				"A config must specify the name of a" +
					" network listener resource.",
			),
		)
		log.Crit(e.Error())
		return nil, e
	}

	if _, present := config.Resources[config.Handler]; !present {
		e := errors.New(
			fmt.Sprintf(
				"A config must specify the name of an" +
					" HTTP handler resource.",
			),
		)
		log.Crit(e.Error())
		return nil, e
	}

	unregisterredResources = config.Resources
	for name := range config.Resources {
		if _, present := registry[name]; !present {
			r := new(Resource)
			registry[name] = r
			if e := r.unmarshalByName(name); e != nil {
				return nil, e
			} else {
				r.complete = true
				log.Debug(
					fmt.Sprintf(
						"Saved resource to"+
							" registry: %s: %#v",
						name,
						r.Unmarshaled,
					),
				)
			}
		}
	}

	server := new(Server)

	l := registry[config.Listener].Unmarshaled
	switch l := l.(type) {
	case net.Listener:
		server.Listener = l
	default:
		e := errors.New(
			fmt.Sprintf(
				"The specified resource is not a"+
					" net.Listener: %s",
				config.Listener,
			),
		)
		log.Crit(e.Error())
		log.Debug(
			errors.New(
				fmt.Sprintf(
					"not a listener: %#v",
					l,
				),
			).Error(),
		)
		return nil, e
	}
	log.Debug(
		fmt.Sprintf(
			"Processed listener: %#v",
			l,
		),
	)

	h := registry[config.Handler].Unmarshaled
	switch h := h.(type) {
	case http.Handler:
		server.Handler = h
	default:
		e := errors.New(
			fmt.Sprintf(
				"The specified resource is not an"+
					" http.Handler: %s",
				config.Handler,
			),
		)
		log.Crit(e.Error())
		// TODO remove
		log.Debug(
			fmt.Sprintf(
				"handler %s has type %s",
				config.Handler,
				h,
			),
		)
		return nil, e
	}
	log.Debug(
		fmt.Sprintf(
			"Processed handler: %#v",
			h,
		),
	)

	return server, nil
}

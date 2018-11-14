package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/proidiot/gone/errors"
	"github.com/proidiot/gone/log"

	"github.com/stuphlabs/pullcord"
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

func ServerFromReader(r io.Reader) (pullcord.Server, error) {
	registrationMutex.Lock()
	defer registrationMutex.Unlock()

	var config struct {
		Resources map[string]json.RawMessage
		Server    json.RawMessage
	}

	dec := json.NewDecoder(r)
	registry = make(map[string]*Resource)

	if e := dec.Decode(&config); e != nil {
		log.Crit(
			fmt.Sprintf(
				"Unable to decode resource: %#v",
				e,
			),
		)
		return nil, e
	}

	unregisterredResources = config.Resources
	for name := range config.Resources {
		log.Debug(fmt.Sprintf("Assessing resource: %s", name))
		if _, present := registry[name]; !present {
			log.Debug(
				fmt.Sprintf(
					"Resource does not already exist in"+
						" the registry: %s",
					name,
				),
			)
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
		} else {
			log.Debug(
				fmt.Sprintf(
					"Resource already exists in the"+
						" registry: %s",
					name,
				),
			)
		}
	}

	rserver := new(Resource)
	if e := json.Unmarshal(config.Server, rserver); e != nil {
		return nil, e
	}

	if server, ok := rserver.Unmarshaled.(pullcord.Server); ok {
		return server, nil
	}

	//err := errors.New("The given server has the wrong type")
	err := errors.New(
		fmt.Sprintf(
			"not a server: %s - %#v",
			config.Server,
			rserver.Unmarshaled,
		),
	)
	log.Crit(err.Error())

	return nil, err
}

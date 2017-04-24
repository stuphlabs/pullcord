package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/proidiot/gone/errors"
	"io"
	"sync"
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
				"More than one resource type has registerred" +
				" the same name: %s",
				typeName,
			),
		)
	}

	typeRegistry[typeName] = newFunc
	return nil
}

type Resource struct {
	Unmarshaled json.Unmarshaler
	complete bool
}

func (rsc *Resource) UnmarshalJSON(input []byte) error {
	var newRscDef struct {
		Type string
		Data json.RawMessage
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&newRscDef); e != nil {
		log().Crit(
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
			log().Crit(
				fmt.Sprintf(
					"Unable to decode the resource" +
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
						"The resource depenency was" +
						" already under construction" +
						" (implying a cyclic" +
						" dependency): %s",
						name,
					),
				)
				log().Crit(e.Error())
				return e
			}
		}

		return rsc.unmarshalByName(name)
	}

	newFunc, present := typeRegistry[newRscDef.Type]
	if !present {
		return errors.New(
			fmt.Sprintf(
				"The specified resource type is not a" +
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

func ServerFromReader(r io.Reader) (*falcore.Server, error) {
	registrationMutex.Lock()

	var config struct {
		Resources map[string]json.RawMessage
		Pipeline []string
		Port int
	}

	dec := json.NewDecoder(r)
	registry = make(map[string]*Resource)

	if e := dec.Decode(&config); e != nil {
		log().Crit(
			fmt.Sprintf(
				"Unable to decode server config: %v",
				e,
			),
		)
		registrationMutex.Unlock()
		return nil, e
	}

	if config.Pipeline == nil || len(config.Pipeline) == 0 {
		e := errors.New(
			fmt.Sprintf(
				"A config must specify at least one named" +
				" resource to use as a filter in the" +
				" pipeline.",
			),
		)
		log().Crit(e.Error())
		registrationMutex.Unlock()
		return nil, e
	}

	unregisterredResources = config.Resources
	for name, _ := range config.Resources {
		if _, present := registry[name]; !present {
			r := new(Resource)
			registry[name] = r
			if e := r.unmarshalByName(name); e != nil {
				registrationMutex.Unlock()
				return nil, e
			} else {
				r.complete = true
				log().Debug(
					fmt.Sprintf(
						"Saved resource to" +
						" registry: %s: %v",
						name,
						r.Unmarshaled,
					),
				)
			}
		}
	}

	pipeline := falcore.NewPipeline()

	for _, upstream := range config.Pipeline {
		r, present := registry[upstream]
		if !present {
			e := errors.New(
				fmt.Sprintf(
					"The requested resource has not been" +
					" registerred: %s",
					upstream,
				),
			)
			log().Crit(e.Error())
			registrationMutex.Unlock()
			return nil, e
		}
		u := r.Unmarshaled
		switch u := u.(type) {
		case falcore.Router:
		case falcore.RequestFilter:
			pipeline.Upstream.PushBack(u)
		default:
			e := errors.New(
				fmt.Sprintf(
					"The requested pipeline resource is" +
					" not a RequestFilter: %s",
					upstream,
				),
			)
			log().Crit(e.Error())
			registrationMutex.Unlock()
			return nil, e
		}
		log().Debug(
			fmt.Sprintf(
				"Added upstream resource: %s",
				upstream,
			),
		)
	}

	server := falcore.NewServer(config.Port, pipeline)

	registry = nil

	registrationMutex.Unlock()

	return server, nil
}


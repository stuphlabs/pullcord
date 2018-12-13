package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/proidiot/gone/errors"
	"github.com/proidiot/gone/log"
)

// UnexpectedResourceType indicates that a sub-resource did not have an
// expected type.
const UnexpectedResourceType = errors.New(
	"The requested resource does not have the expected type",
)

// ReferenceResourceTypeName is the reserved name to indicate a placeholder for
// an already registered resource instead of giving the type for for a new
// resource being defined.
const ReferenceResourceTypeName = "ref"

// Resource represents a configurable object which allows an arbitrary JSON blob
// to be turned into an implementation of whatever type it specifies. This
// abstraction is the key to allowing instantiation of structs which have
// interface members instead of just non-abstracted data types.
type Resource struct {
	Unmarshaled json.Unmarshaler
	complete    bool
}

// UnmarshalJSON implements encoding/json.Unmarshaler, which is the core
// requirement which allows these resources to be instantiated at all.
func (rsc *Resource) UnmarshalJSON(input []byte) error {
	var newRscDef struct {
		Type string
		Data json.RawMessage
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&newRscDef); e != nil {
		_ = log.Crit(
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
			_ = log.Crit(
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
			}

			e := fmt.Errorf(
				"The resource depenency was already under"+
					" construction (implying a cyclic"+
					" dependency): %s",
				name,
			)
			_ = log.Crit(e.Error())
			return e
		}

		return rsc.unmarshalByName(name)
	}

	newFunc, present := typeRegistry[newRscDef.Type]
	if !present {
		return fmt.Errorf(
			"The specified resource type is not a registerred"+
				" resource type: %s",
			newRscDef.Type,
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

func (rsc *Resource) unmarshalByName(name string) error {
	d, present := unregisterredResources[name]

	if !present {
		return fmt.Errorf(
			"No resource specified with name: %s",
			name,
		)
	}

	return json.Unmarshal(d, rsc)
}

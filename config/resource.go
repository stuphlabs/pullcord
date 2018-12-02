package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/proidiot/gone/errors"
	"github.com/proidiot/gone/log"
)

const UnexpectedResourceType = errors.New(
	"The requested resource does not have the expected type.",
)
const ReferenceResourceTypeName = "ref"

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

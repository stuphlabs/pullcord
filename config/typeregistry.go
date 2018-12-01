package config

import (
	"encoding/json"
	"fmt"

	"github.com/proidiot/gone/errors"
)

var typeRegistry = make(map[string]func() json.Unmarshaler)

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

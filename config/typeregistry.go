package config

import (
	"encoding/json"
	"fmt"
)

var typeRegistry = make(map[string]func() json.Unmarshaler)

// RegisterResourceType is an infrastructure hack that allows new config
// Resource types to be specified at run-time. It needs to be run before a
// Parser is used, presumably in an Init function of the package of a config
// plugin.
func RegisterResourceType(
	typeName string,
	newFunc func() json.Unmarshaler,
) error {
	_, present := typeRegistry[typeName]
	if present || typeName == ReferenceResourceTypeName {
		return fmt.Errorf(
			"More than one resource type has registered the same"+
				" name: %s",
			typeName,
		)
	}

	typeRegistry[typeName] = newFunc
	return nil
}

// MustRegisterResourceType is a convenience function around
// RegisterResourceType that panics on error.
func MustRegisterResourceType(
	typeName string,
	newFunc func() json.Unmarshaler,
) {
	e := RegisterResourceType(typeName, newFunc)
	if e != nil {
		panic(e)
	}
}

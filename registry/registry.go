package registry

import (
	"fmt"
	"github.com/proidiot/gone/errors"
	"runtime/debug"
)

const OutsideConfigProcess = errors.New(
	"An attempt was made to get or set a resource in the config registry" +
	" outside of the initial configuration process,",
)

// TODO remove?
const NoSuchRegistryValue = errors.New(
	"The named resource was not found in the registry.",
)

const ExistingRegistryValue = errors.New(
	"There is already a resource with the given name in the registry.",
)

const UnexpectedType = errors.New(
	"The specified registry value does not have the desired type.",
)

var registry map[string]interface{}

func Init() (error) {
	registry = make(map[string]interface{})
	return nil
}

func Get(name string) (interface{}, error) {
	if registry == nil {
		// TODO remove
		debug.PrintStack()
		return nil, OutsideConfigProcess
	} else if r, present := registry[name]; ! present {
		// TODO remove
		debug.PrintStack()
		return nil, errors.New(
			fmt.Sprintf(
				"The requested resource was not found in the" +
				" registry: %s",
				name,
			),
		)
	} else {
		return r, nil
	}
}

func Set(name string, value interface{}) (error) {
	if registry == nil {
		// TODO remove
		debug.PrintStack()
		return OutsideConfigProcess
	} else if _, present := registry[name]; present {
		// TODO remove
		debug.PrintStack()
		return ExistingRegistryValue
	} else {
		registry[name] = value
		return nil
	}
}

func Wipe() (error) {
	registry = nil
	return nil
}


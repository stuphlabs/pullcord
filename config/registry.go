package config

import (
	"encoding/json"
	"sync"
)

var registry map[string]*Resource
var unregisterredResources map[string]json.RawMessage
var registrationMutex sync.Mutex

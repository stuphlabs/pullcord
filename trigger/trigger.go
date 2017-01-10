package trigger

import (
	// "github.com/stuphlabs/pullcord"
)

// TriggerHandler is an abstract interface describing a system which provides
// triggers that can be called based on certain events (like a service being
// detected as down, an amount of time passing without a service being
// accessed, etc.). For the moment, the only trigger mechanism provided is a
// simple string being passed in, and although additional trigger mechanisms
// could be added later, it is not yet clear that others would be necessary.
//
// TriggerString is a function that provides a trigger mechanism based on an
// arbitrary string. A minimal amount of additional structure could be achieved
// through the use of serialization format like JSON.
type TriggerHandler interface {
	Trigger() (err error)
}


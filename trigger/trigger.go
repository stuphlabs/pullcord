package trigger

// Triggerrer is an abstract interface describing a system which provides
// triggers that can be called based on certain events (like a service being
// detected as down, an amount of time passing without a service being
// accessed, etc.).
type Triggerrer interface {
	Trigger() (err error)
}

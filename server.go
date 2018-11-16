package pullcord

// Server represents a Pullcord server in the most generalized form.
type Server interface {
	Serve() error
	Close() error
}

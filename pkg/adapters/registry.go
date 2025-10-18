package adapters

// Registry holds available adapter definitions.
type Registry struct{}

// NewRegistry returns a registry with built-in adapters.
func NewRegistry() *Registry {
	return &Registry{}
}

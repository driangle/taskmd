package sync

import (
	"fmt"
	"sort"
)

var sources = map[string]Source{}

// Register adds a source to the global registry.
func Register(s Source) {
	sources[s.Name()] = s
}

// GetSource returns a registered source by name.
func GetSource(name string) (Source, error) {
	s, ok := sources[name]
	if !ok {
		return nil, fmt.Errorf("unknown source: %q (registered: %v)", name, RegisteredNames())
	}
	return s, nil
}

// Unregister removes a source from the registry. Intended for testing.
func Unregister(name string) {
	delete(sources, name)
}

// RegisteredNames returns the sorted names of all registered sources.
func RegisteredNames() []string {
	names := make([]string, 0, len(sources))
	for n := range sources {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

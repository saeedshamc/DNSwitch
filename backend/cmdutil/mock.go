package cmdutil

import (
	"context"
	"fmt"
	"strings"
)

// Call is a recorded Runner invocation, used by tests.
type Call struct {
	Name string
	Args []string
}

// MockRunner is an in-memory Runner for unit tests.
type MockRunner struct {
	Lookups map[string]string
	// Handlers maps "name arg1 arg2" (space-joined) prefixes to results.
	Handlers map[string]Result
	Default  Result
	Calls    []Call
}

// Run records the call and returns a configured result.
func (m *MockRunner) Run(name string, args ...string) Result {
	return m.RunContext(context.Background(), name, args...)
}

// RunContext records the call and returns a configured result.
func (m *MockRunner) RunContext(_ context.Context, name string, args ...string) Result {
	m.Calls = append(m.Calls, Call{Name: name, Args: append([]string{}, args...)})

	key := name
	if len(args) > 0 {
		key = name + " " + strings.Join(args, " ")
	}
	if m.Handlers != nil {
		if res, ok := m.Handlers[key]; ok {
			return res
		}
		for prefix, res := range m.Handlers {
			if strings.HasPrefix(key, prefix) {
				return res
			}
		}
	}
	return m.Default
}

// LookPath returns a fake path when configured.
func (m *MockRunner) LookPath(name string) (string, error) {
	if m.Lookups != nil {
		if p, ok := m.Lookups[name]; ok {
			if p == "" {
				return "", fmt.Errorf("not found: %s", name)
			}
			return p, nil
		}
	}
	return "", fmt.Errorf("not found: %s", name)
}

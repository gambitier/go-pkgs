package domainerr

import (
	"errors"

	cerrors "github.com/cockroachdb/errors"
)

// CollectFields merges Fields from every *Error in the unwrap chain.
// Inner layers are applied first; outer layers override duplicate keys.
func CollectFields(err error) map[string]any {
	chain := domainErrorsInChain(err)
	if len(chain) == 0 {
		return nil
	}

	merged := make(map[string]any)
	for i := len(chain) - 1; i >= 0; i-- {
		for key, value := range chain[i].Fields {
			merged[key] = value
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

// domainErrorsInChain returns domain errors from outermost to innermost.
func domainErrorsInChain(err error) []*Error {
	if err == nil {
		return nil
	}

	var chain []*Error
	seen := map[*Error]struct{}{}
	for current := err; current != nil; {
		if de, ok := domainErrorAt(current); ok {
			if _, exists := seen[de]; !exists {
				seen[de] = struct{}{}
				chain = append(chain, de)
			}
		}
		next := cerrors.Unwrap(current)
		if next == nil {
			next = errors.Unwrap(current)
		}
		current = next
	}
	return chain
}

// domainErrorAt returns the domain error only when err is (or directly holds) one.
// errors.As is intentionally not used here because it would always resolve the
// outermost domain error and skip inner layers in the chain.
func domainErrorAt(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	if de, ok := err.(*Error); ok && de != nil {
		return de, true
	}
	if fe, ok := err.(*fluentError); ok && fe != nil && fe.value != nil {
		return fe.value, true
	}
	var fe *fluentError
	if errors.As(err, &fe) && fe != nil && fe.value != nil {
		return fe.value, true
	}
	return nil, false
}

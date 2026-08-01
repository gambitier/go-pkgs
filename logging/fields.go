package logging

// Fields stores structured logging attributes.
type Fields map[string]any

func cloneFields(input Fields) logrusFields {
	if len(input) == 0 {
		return nil
	}
	out := make(logrusFields, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

package v1alpha1

// Any will return true if any callbacks return true. It follows the same logic
// as the any() function in Python.
//
// If the list is empty then false is always returned.
func (ss Sources) Any(fn func(value Source) bool) bool {
	for _, value := range ss {
		if fn(value) {
			return true
		}
	}

	return false
}

package graph

// FindFirst returns the first element in the slice that matches the predicate function.
// If no element matches, it returns the zero value for the type and false.
func FindFirst[T any](slice []T, match func(T) bool) (T, bool) {
	for _, item := range slice {
		if match(item) {
			return item, true
		}
	}
	var zero T // Default zero value of type T
	return zero, false
}

// LastElement returns the last element of a slice of any type.
// It returns the zero value and false if the slice is empty.
func LastElement[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T // Default zero value of type T
		return zero, false
	}
	return s[len(s)-1], true
}

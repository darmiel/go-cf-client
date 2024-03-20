package util

import "cmp"

// IFTTT returns `a` if `this` is true, otherwise `b`
func IFTTT[T any](this bool, a, b T) T {
	if this {
		return a
	}
	return b
}

// Clamp returns `this` if it is between `min` and `max`, otherwise the closest bound
func Clamp[T cmp.Ordered](this, min, max T) T {
	if this < min {
		return min
	}
	if this > max {
		return max
	}
	return this
}

package util

import (
	"cmp"
	"strings"
)

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

type KV map[string]any

// AsMap converts a KV to a map[string]any
// This is useful for long keys with dots, e.g. "a.b.c" which will be converted to {"a": {"b": {"c": ...}}}
func (kv KV) AsMap() map[string]any {
	result := make(map[string]any)
	for k, v := range kv {
		if strings.Contains(k, ".") {
			keys := strings.Split(k, ".")

			m := result
			for i, key := range keys {
				if i == len(keys)-1 {
					// last key
					m[key] = v
				} else {
					// create map if not exists
					if _, ok := m[key]; !ok {
						m[key] = make(map[string]any)
					}
					// set map
					m = m[key].(map[string]any)
				}
			}
		} else {
			result[k] = v
		}
	}
	return result
}

func DataGUID(guid string) KV {
	return KV{"data": KV{"guid": guid}}
}

func Data(data KV) KV {
	return KV{"data": data}
}

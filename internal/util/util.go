package util

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// MaxItemsPerPage is the maximum number of items that can be requested per page
const MaxItemsPerPage = 5000

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
type Query map[string]string

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

func IsValueEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch t := value.(type) {
	case string:
		return t == ""
	case []string:
		return t == nil || len(t) == 0
	case map[string]string:
		return t == nil || len(t) == 0
	case KV:
		return t == nil || len(t) == 0
	case Query:
		return t == nil || len(t) == 0
	default:
		panic("unknown type checked for empty: " + fmt.Sprintf("%T", t))
	}
}

func CreateQueryParams(values Query, perPage ...int) Query {
	queryParams := make(Query)
	for key, value := range values {
		if !IsValueEmpty(value) {
			queryParams[key] = value
		}
	}
	if len(perPage) > 0 {
		perPage := perPage[0]
		if perPage > 0 {
			queryParams["per_page"] = strconv.Itoa(Clamp(perPage, 1, MaxItemsPerPage))
		} else {
			queryParams["per_page"] = strconv.Itoa(MaxItemsPerPage)
		}
	}
	return queryParams
}

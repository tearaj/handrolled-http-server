package utils

import "maps"

func MergeMaps[K comparable, V any](
	a, b map[K]V,
	resolver func(aVal, bVal V) V,
) map[K]V {
	merged := make(map[K]V, len(a)+len(b))

	maps.Copy(merged, a)

	for k, v := range b {
		if existing, ok := merged[k]; ok {
			merged[k] = resolver(existing, v)
		} else {
			merged[k] = v
		}
	}
	return merged
}

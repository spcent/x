package sign

import "hash/fnv"

func Sum64(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))

	return h.Sum64()
}

func Sum64s(ss []string) uint64 {
	h := fnv.New64a()
	for _, s := range ss {
		h.Write([]byte(s))
		h.Write([]byte{0})
	}

	return h.Sum64()
}

package slices

func Map[I any, O any](ins []I, f func(I) O) []O {
	out := make([]O, 0, len(ins))

	for _, in := range ins {
		out = append(out, f(in))
	}

	return out
}

func MapErr[I any, O any](ins []I, f func(I) (O, error)) ([]O, error) {
	out := make([]O, 0, len(ins))

	for _, in := range ins {
		v, err := f(in)
		if err != nil {
			return []O{}, err
		}
		out = append(out, v)
	}

	return out, nil
}

func ToMap[I any, K comparable](ins []I, f func(I) K) map[K]I {
	out := make(map[K]I, len(ins))

	for _, in := range ins {
		k := f(in)
		out[k] = in
	}

	return out
}

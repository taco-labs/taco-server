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

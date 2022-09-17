package slices

func Map[I any, O any](ins []I, f func(I) O) []O {
	out := make([]O, 0, len(ins))

	for _, in := range ins {
		out = append(out, f(in))
	}

	return out
}

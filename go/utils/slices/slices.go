package slices

func Map[I any, O any](ins []I, f func(I) O) []O {
	out := make([]O, 0, len(ins))

	for _, in := range ins {
		out = append(out, f(in))
	}

	return out
}

func Filter[I any](ins []I, f func(I) bool) []I {
	out := make([]I, 0, len(ins))

	for _, in := range ins {
		if f(in) {
			out = append(out, in)
		}
	}

	return out
}

func Foreach[I any](ins []I, f func(I)) {
	for idx := range ins {
		f(ins[idx])
	}
}

func ForeachErr[I any](ins []I, f func(I) error) error {
	for idx := range ins {
		if err := f(ins[idx]); err != nil {
			return err
		}
	}
	return nil
}

func ForeachRef[I any](ins []I, f func(*I)) {
	for idx := range ins {
		f(&ins[idx])
	}
}

func ForeachErrRef[I any](ins []I, f func(*I) error) error {
	for idx := range ins {
		if err := f(&ins[idx]); err != nil {
			return err
		}
	}
	return nil
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

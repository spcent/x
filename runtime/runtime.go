package runtime

// Must panics if err is not nil.
// It is useful for handling errors in initialization code where recovery is not possible.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Must1 is like Must but returns the value if err is nil.
// It is useful for handling errors in initialization code where recovery is not possible
// and a value needs to be returned.
func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// Must2 is like Must but returns two values if err is nil.
// It is useful for handling errors in initialization code where recovery is not possible
// and two values need to be returned.
func Must2[T1 any, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}

	return v1, v2
}

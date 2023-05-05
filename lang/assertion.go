package lang

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

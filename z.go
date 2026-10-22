package cardrank

// init is in z.go so that it runs after the init funcs of the other files.
func init() {
	if err := registerDefaultTypes(); err != nil {
		panic(err)
	}
}

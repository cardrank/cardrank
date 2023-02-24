//go:build !noinit

package cardrank

func init() {
	Init()
	if err := RegisterDefaultTypes(); err != nil {
		panic(err)
	}
}

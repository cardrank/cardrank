//go:build !noinit

package paycalc

func init() {
	if err := Init(); err != nil {
		panic(err)
	}
}

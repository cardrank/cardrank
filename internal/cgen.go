//go:build ignore

package main

import (
	"flag"
	"fmt"

	"gonum.org/v1/gonum/stat/combin"
)

func main() {
	var a, b int
	flag.IntVar(&a, "a", 7, "a")
	flag.IntVar(&b, "b", 5, "b")
	flag.Parse()
	c := combin.Combinations(a, b)
	fmt.Printf("// t%dc%d is used for taking %d, choosing %d.\n", a, b, a, b)
	fmt.Printf("var t%dc%d = [%d][%d]uint8{\n", a, b, len(c), a)
	for _, v := range c {
		m := make(map[int]bool)
		fmt.Print("\t{")
		for i := 0; i < b; i++ {
			if i != 0 {
				fmt.Print(", ")
			}
			m[v[i]] = true
			fmt.Printf("%d", v[i])
		}
		for i := 0; i < a; i++ {
			if !m[i] {
				fmt.Printf(", %d", i)
			}
		}
		fmt.Println("},")
	}
	fmt.Println("}")
}

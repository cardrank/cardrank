//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/cardrank/cardrank"
)

func main() {
	var n, k int
	flag.IntVar(&n, "n", 7, "n")
	flag.IntVar(&k, "k", 5, "k")
	flag.Parse()
	count, g, v := 0, cardrank.NewBinGen(n, k), make([]int, n)
	buf := new(bytes.Buffer)
	for g.NextUnused(v) {
		fmt.Fprint(buf, "\t{")
		for i := 0; i < n; i++ {
			if i != 0 {
				fmt.Fprint(buf, ", ")
			}
			fmt.Fprintf(buf, "%d", v[i])
		}
		fmt.Fprintln(buf, "},")
		count++
	}
	fmt.Printf("// t%dc%d is used for taking %d, choosing %d.\n", n, k, n, k)
	fmt.Printf("var t%dc%d = [%d][%d]uint8{\n", n, k, count, n)
	os.Stdout.Write(buf.Bytes())
	fmt.Println("}")
}

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
	n := flag.Int("n", 7, "n")
	k := flag.Int("k", 5, "k")
	flag.Parse()
	u := make([]int, *n)
	for i := range *n {
		u[i] = i
	}
	buf, count := new(bytes.Buffer), 0
	for g, v := cardrank.NewCombinUnusedGen(u, *k); g.Next(); {
		fmt.Fprint(buf, "\t{")
		for i := range *n {
			if i != 0 {
				fmt.Fprint(buf, ", ")
			}
			fmt.Fprintf(buf, "%d", v[i])
		}
		fmt.Fprintln(buf, "},")
		count++
	}
	fmt.Printf("// t%dc%d is used for taking %d, choosing %d.\n", *n, *k, *n, *k)
	fmt.Printf("var t%dc%d = [%d][%d]uint8{\n", *n, *k, count, *n)
	os.Stdout.Write(buf.Bytes())
	fmt.Println("}")
}

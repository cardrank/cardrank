package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	r := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.OmahaHiLo.Deal(r, 3, players)
	evs := cardrank.OmahaHiLo.EvalPockets(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		hi, lo := evs[i].Desc(false), evs[i].Desc(true)
		fmt.Printf("  %d: %b %b %b %s\n", i+1, pockets[i], hi.Best, hi.Unused, hi)
		fmt.Printf("                   %b %b %s\n", lo.Best, lo.Unused, lo)
	}
	hiOrder, hiPivot := cardrank.Order(evs, false)
	loOrder, loPivot := cardrank.Order(evs, true)
	hi := evs[hiOrder[0]].Desc(false)
	if hiPivot == 1 {
		fmt.Printf("Result: %d wins with %s, %b\n", hiOrder[0]+1, hi, hi.Best)
	} else {
		var s []string
		for i := 0; i < hiPivot; i++ {
			s = append(s, strconv.Itoa(hiOrder[i]+1))
		}
		fmt.Printf("Result: %s push with %s\n", strings.Join(s, ", "), hi)
	}
	if loPivot == 0 {
		fmt.Printf("        None\n")
	} else if loPivot == 1 {
		lo := evs[loOrder[0]].Desc(true)
		fmt.Printf("        %d wins with %s %b\n", loOrder[0]+1, lo, lo.Best)
	} else {
		var s []string
		for j := 0; j < loPivot; j++ {
			s = append(s, strconv.Itoa(loOrder[j]+1))
		}
		lo := evs[loOrder[0]].Desc(true)
		fmt.Printf("        %s push with %s\n", strings.Join(s, ", "), lo)
	}
}

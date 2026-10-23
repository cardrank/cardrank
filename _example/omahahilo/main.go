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
	pockets, board := cardrank.OmahaHiLo.Deal(r, players)
	evs := cardrank.OmahaHiLo.EvalPockets(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := range players {
		hi, lo := evs[i].Hi, evs[i].Lo
		fmt.Printf("  %d: %b %b %b %s\n", i+1, pockets[i], hi.Best, hi.Unused, hi)
		fmt.Printf("                   %b %b %s\n", lo.Best, lo.Unused, lo)
	}
	hiWin, loWin := cardrank.HiOrder(evs), cardrank.LoOrder(evs)
	hi := evs[hiWin.Order[0]].Hi
	if hiWin.Pivot == 1 {
		fmt.Printf("Result: %d wins with %s, %b\n", hiWin.Order[0]+1, hi, hi.Best)
	} else {
		fmt.Printf("Result: %s push with %s\n", positions(hiWin.Winners()), hi)
	}
	switch lo := evs[loWin.Order[0]].Lo; loWin.Pivot {
	case 0:
		fmt.Printf("        None\n")
	case 1:
		fmt.Printf("        %d wins with %s %b\n", loWin.Order[0]+1, lo, lo.Best)
	default:
		fmt.Printf("        %s push with %s\n", positions(loWin.Winners()), lo)
	}
}

// positions returns the 1-based positions as a string.
func positions(v []int) string {
	var s []string
	for _, i := range v {
		s = append(s, strconv.Itoa(i+1))
	}
	return strings.Join(s, ", ")
}

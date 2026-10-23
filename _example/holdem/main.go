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
	pockets, board := cardrank.Holdem.Deal(r, players)
	evs := cardrank.Holdem.EvalPockets(pockets, board)
	fmt.Printf("------ Holdem %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := range players {
		hi := evs[i].Hi
		fmt.Printf("  %d: %b %b %b %s\n", i+1, pockets[i], hi.Best, hi.Unused, hi)
	}
	win := cardrank.HiOrder(evs)
	hi := evs[win.Order[0]].Hi
	if win.Pivot == 1 {
		fmt.Printf("Result: %d wins with %s %b\n", win.Order[0]+1, hi, hi.Best)
	} else {
		var s, b []string
		for _, i := range win.Winners() {
			s = append(s, strconv.Itoa(i+1))
			b = append(b, fmt.Sprintf("%b", evs[i].Hi.Best))
		}
		fmt.Printf("Result: %s push with %s %s\n", strings.Join(s, ", "), hi, strings.Join(b, ", "))
	}
}

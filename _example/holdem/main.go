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
	pockets, board := cardrank.Holdem.Deal(r, 3, players)
	evs := cardrank.Holdem.EvalPockets(pockets, board)
	fmt.Printf("------ Holdem %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		desc := evs[i].Desc(false)
		fmt.Printf("  %d: %b %b %b %s\n", i+1, pockets[i], desc.Best, desc.Unused, desc)
	}
	order, pivot := cardrank.Order(evs, false)
	desc := evs[order[0]].Desc(false)
	if pivot == 1 {
		fmt.Printf("Result: %d wins with %s %b\n", order[0]+1, desc, desc.Best)
	} else {
		var s, b []string
		for j := 0; j < pivot; j++ {
			s = append(s, strconv.Itoa(order[j]+1))
			b = append(b, fmt.Sprintf("%b", evs[order[j]].HiBest))
		}
		fmt.Printf("Result: %s push with %s %s\n", strings.Join(s, ", "), desc, strings.Join(b, ", "))
	}
}

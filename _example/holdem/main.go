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
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.Holdem.Deal(rnd, players)
	hands := cardrank.Holdem.RankHands(pockets, board)
	fmt.Printf("------ Holdem %d ------\n", seed)
	fmt.Printf("Board:    %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b %s %b %b\n", i+1, hands[i].Pocket(), hands[i].Description(), hands[i].Best(), hands[i].Unused())
	}
	h, pivot := cardrank.Order(hands)
	if pivot == 1 {
		fmt.Printf("Result:   Player %d wins with %s %b\n", h[0]+1, hands[h[0]].Description(), hands[h[0]].Best())
	} else {
		var s, b []string
		for j := 0; j < pivot; j++ {
			s = append(s, strconv.Itoa(h[j]+1))
			b = append(b, fmt.Sprintf("%b", hands[h[j]].Best()))
		}
		fmt.Printf("Result:   Players %s push with %s %s\n", strings.Join(s, ", "), hands[h[0]].Description(), strings.Join(b, ", "))
	}
}

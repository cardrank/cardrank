package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cardrank.io/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.OmahaHiLo.Deal(rnd.Shuffle, players)
	highs := cardrank.OmahaHiLo.RankHands(pockets, board)
	lows := cardrank.OmahaHiLo.LowRankHands(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b\n", i+1, pockets[i])
		fmt.Printf("  Hi: %s %b %b\n", highs[i].Description(), highs[i].Best(), highs[i].Unused())
		if lows[i].Rank() < 31 {
			fmt.Printf("  Lo: %s %b %b\n", lows[i].LowDescription(), lows[i].LowBest(), lows[i].LowUnused())
		} else {
			fmt.Printf("  Lo: None\n")
		}
	}
	h, hPivot := cardrank.OrderHands(highs)
	l, lPivot := cardrank.LowOrderHands(lows)
	typ := "wins"
	if lPivot == 0 {
		typ = "scoops"
	}
	if hPivot == 1 {
		fmt.Printf("Result (Hi): Player %d %s with %s %b\n", h[0]+1, typ, highs[h[0]].Description(), highs[h[0]].Best())
	} else {
		var s, b []string
		for i := 0; i < hPivot; i++ {
			s = append(s, strconv.Itoa(h[i]+1))
			b = append(b, fmt.Sprintf("%b", highs[h[i]].Best()))
		}
		fmt.Printf("Result (Hi): Players %s push with %s %s\n", strings.Join(s, ", "), highs[h[0]].Description(), strings.Join(b, ", "))
	}
	if lPivot == 1 {
		fmt.Printf("Result (Lo): Player %d wins with %s %b\n", l[0]+1, lows[l[0]].LowDescription(), lows[l[0]].LowBest())
	} else if lPivot > 1 {
		var s, b []string
		for j := 0; j < lPivot; j++ {
			s = append(s, strconv.Itoa(l[j]+1))
			b = append(b, fmt.Sprintf("%b", lows[l[j]].LowBest()))
		}
		fmt.Printf("Result (Lo): Players %s push with %s %s\n", strings.Join(s, ", "), lows[l[0]].LowDescription(), strings.Join(b, ", "))
	} else {
		fmt.Printf("Result (Lo): no player made a low hand\n")
	}
}

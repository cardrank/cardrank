package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 3
	seed := time.Now().UnixNano()
	for _, typ := range []cardrank.Type{cardrank.Royal, cardrank.CourchevelHiLo, cardrank.Badugi} {
		// note: use a better pseudo-random number generator
		r := rand.New(rand.NewSource(seed))
		fmt.Printf("------ %s %d ------\n", typ, seed)
		// setup dealer
		d := typ.Dealer(r, 3)
		// display cards in deck
		all := d.All()
		fmt.Printf("Deck (%s, %d):\n", typ.DeckType(), len(all))
		for i := 0; i < len(all); i += 8 {
			fmt.Printf("  %v\n", all[i:min(i+8, len(all))])
		}
		// need to know if type supports double or supports lows as it
		// determines the output display below
		double, low := typ.Double(), typ.Low()
		var pockets [][]cardrank.Card
		var hiBoard, loBoard []cardrank.Card
		for d.Next() {
			fmt.Printf("%s:\n", d)
			// collect pockets and hi board
			pockets, hiBoard = d.Deal(pockets, hiBoard, players)
			// display pockets
			if 0 < d.Pocket() {
				for i := 0; i < players; i++ {
					fmt.Printf("  % 2d: %v\n", i, pockets[i])
				}
			}
			if 0 < d.Board() {
				// display hiboard
				fmt.Printf("  Board: %v\n", hiBoard)
				if double {
					// collect and display lo board, if any
					loBoard = d.DealBoard(loBoard, false)
					fmt.Printf("         %v\n", loBoard)
				}
			}
		}
		// rank the hands
		hi := typ.RankHands(pockets, hiBoard)
		var lo []*cardrank.Hand
		if double {
			lo = typ.RankHands(pockets, loBoard)
		}
		fmt.Printf("Showdown:\n")
		for i := 0; i < len(hi); i++ {
			// display hi hand eval
			fmt.Printf(" % 2d: %04d %v %v %s\n", i, hi[i].HiRank, hi[i].HiBest, hi[i].HiUnused, hi[i].Description())
			// display lo hand eval
			switch {
			case double:
				fmt.Printf("     %04d %v %v %s\n", lo[i].HiRank, lo[i].HiBest, lo[i].HiUnused, lo[i].Description())
			case low:
				fmt.Printf("     %04d %v %v %s\n", hi[i].LoRank, hi[i].LoBest, hi[i].LoUnused, hi[i].LowDescription())
			}
		}
		fmt.Printf("Result:\n")
		// display winner(s)
		win := cardrank.NewWin(hi, lo, low)
		fmt.Printf("  %s\n", win.HiDesc(func(_, i int) string {
			return strconv.Itoa(i)
		}))
		if !win.Scoop() && (double || low) {
			fmt.Printf("  %s\n", win.LoDesc(func(_, i int) string {
				return strconv.Itoa(i)
			}))
		}
	}
}

// ordered is the ordered constraint.
type ordered interface {
	~float32 | ~float64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// min returns the min of a, b.
func min[T ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

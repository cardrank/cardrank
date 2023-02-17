package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 4
	seed := time.Now().UnixNano()
	for _, typ := range []cardrank.Type{cardrank.Royal, cardrank.Double, cardrank.OmahaDouble, cardrank.CourchevelHiLo, cardrank.Razz, cardrank.Badugi} {
		// note: use a better pseudo-random number generator
		r := rand.New(rand.NewSource(seed))
		fmt.Printf("------ %s %d ------\n", typ, seed)
		// setup dealer and display
		d := typ.Dealer(r, 3, players)
		desc := typ.TypeDesc()
		fmt.Printf("Eval: %s\n", desc.Eval)
		fmt.Printf("HiComp: %s LoComp: %s\n", desc.HiComp, desc.LoComp)
		fmt.Printf("HiDesc: %s LoDesc: %s\n", desc.HiDesc, desc.LoDesc)
		// display deck
		deck := d.Deck.All()
		fmt.Printf("Deck: %s [%d]\n", desc.Deck, len(deck))
		for i := 0; i < len(deck); i += 8 {
			fmt.Printf("  %v\n", deck[i:min(i+8, len(deck))])
		}
		// iterate deal streets
		for d.Next() {
			fmt.Printf("%s\n", d)
			// display pockets
			if d.HasPocket() {
				for i := 0; i < players; i++ {
					fmt.Printf("  %d: %v\n", i, d.Pockets[i])
				}
			}
			// display discarded cards
			if v := d.Discarded(); len(v) != 0 {
				fmt.Printf("  Discard: %v\n", v)
			}
			// display board
			if d.HasBoard() {
				for i := 0; i < len(d.Boards); i++ {
					fmt.Printf("  Run %d: %v\n", i, d.Boards[i].Hi)
					if d.Double {
						fmt.Printf("         %v\n", d.Boards[i].Lo)
					}
				}
			}
			// change runs to 2, after the flop
			if d.Id() == 'f' {
				d.Runs(2)
			}
		}
		// iterate eval results
		fmt.Printf("Showdown:\n")
		for d.NextResult() {
			run, res := d.Result()
			fmt.Printf("  Run %d:\n", run)
			for i := 0; i < players; i++ {
				if d.Active[i] {
					hi := res.Evals[i].Desc(false)
					fmt.Printf("    %d: %v %v %s\n", i, hi.Best, hi.Unused, hi)
					if d.Low || d.Double {
						lo := res.Evals[i].Desc(true)
						fmt.Printf("       %v %v %s\n", lo.Best, lo.Unused, lo)
					}
				} else {
					fmt.Printf("    %d: inactive\n", i)
				}
			}
			hi, lo := res.Win()
			fmt.Printf("    Result: %s with %S\n", hi, hi)
			if lo != nil {
				fmt.Printf("            %s with %S\n", lo, lo)
			}
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

package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 4
	seed := time.Now().UnixNano()
	for _, typ := range []cardrank.Type{
		cardrank.Royal,
		cardrank.Double,
		cardrank.OmahaDouble,
		cardrank.CourchevelHiLo,
		cardrank.FusionHiLo,
		cardrank.Razz,
		cardrank.Badugi,
	} {
		// note: use a better pseudo-random number generator
		r := rand.New(rand.NewSource(seed))
		fmt.Printf("------ %s %d ------\n", typ, seed)
		// setup dealer and display
		d := typ.Dealer(r, players)
		desc := typ.Desc()
		fmt.Printf("Eval: %l\n", typ)
		fmt.Printf("Desc: %s/%s\n", desc.HiDesc, desc.LoDesc)
		// display deck
		deck := slices.Collect(d.Deck().All())
		fmt.Printf("Deck: %s [%d]\n", desc.Deck, len(deck))
		for i := 0; i < len(deck); i += 8 {
			fmt.Printf("  %v\n", deck[i:min(i+8, len(deck))])
		}
		// iterate deal streets
		last := -1
		for i, run := range d.Deal() {
			if last != i {
				fmt.Printf("Run %d:\n", i)
			}
			last = i
			fmt.Printf("  %s\n", d)
			street := d.Street()
			// display pockets
			if street.Pocket != 0 {
				for i := range players {
					fmt.Printf("    %d: %v\n", i, run.Pockets[i])
				}
			}
			// display discarded cards
			if v := d.Discarded(); len(v) != 0 {
				fmt.Printf("    Discard: %v\n", v)
			}
			// display board
			if street.Board != 0 {
				fmt.Printf("    Board: %v\n", run.Hi)
				if typ.Double() {
					fmt.Printf("           %v\n", run.Lo)
				}
			}
			// change runs to 3, after the flop
			if street.ID == 'f' && i == 0 {
				if err := d.ChangeRuns(3); err != nil {
					fmt.Println("unable to change runs:", err)
					return
				}
			}
		}
		// iterate eval results
		fmt.Printf("Showdown:\n")
		for run, res := range d.Results() {
			fmt.Printf("  Run %d:\n", run)
			for i := range players {
				if d.Active(i) {
					hi := res.Evals[i].Hi
					fmt.Printf("    %d: %v %v %s\n", i, hi.Best, hi.Unused, hi)
					if typ.Low() || typ.Double() {
						lo := res.Evals[i].Lo
						fmt.Printf("       %v %v %s\n", lo.Best, lo.Unused, lo)
					}
				} else {
					fmt.Printf("    %d: inactive\n", i)
				}
			}
			fmt.Printf("    Result: %d with %s\n", res.Hi, res.Hi)
			if res.Lo.Valid() {
				fmt.Printf("            %d with %s\n", res.Lo, res.Lo)
			}
		}
	}
}

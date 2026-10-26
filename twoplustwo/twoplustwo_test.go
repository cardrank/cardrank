package twoplustwo_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/cardrank/cardrank"
	"github.com/cardrank/cardrank/twoplustwo"
)

func TestNew(t *testing.T) {
	f, cactus := twoplustwo.New(), cardrank.RankEvaluator(cardrank.CactusFast)
	r := rand.New(rand.NewSource(0x2b2))
	for range 10000 {
		for n := 5; n <= 7; n++ {
			v := cardrank.DeckFrench.Shuffle(r).Draw(n)
			ev := cardrank.EvalOf(cardrank.Holdem)
			cactus.Rank(ev, v[:2], v[2:])
			if rank := f(v); rank != ev.Hi.Rank {
				t.Fatalf("%v: expected %d, got: %d", v, ev.Hi.Rank, rank)
			}
		}
	}
}

func Example() {
	// importing the twoplustwo package registers the Two-Plus-Two evaluator,
	// which the built-in types use for 7 card evals
	pocket, board := cardrank.Must("Ah Kh"), cardrank.Must("Qh Jh Th 2s 3s")
	ev := cardrank.Holdem.Eval(pocket, board)
	fmt.Printf("%s\n", ev)
	// Output:
	// Straight Flush, Ace-high, Royal Flush [Ah Kh Qh Jh Th]
}

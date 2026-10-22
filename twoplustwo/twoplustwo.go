// Package twoplustwo provides the Two-Plus-Two poker hand evaluator, which
// uses a large (~124 MiB) lookup table for extremely fast 5, 6, and 7 card
// evaluation.
//
// Importing the package registers the evaluator with [cardrank], which then
// uses it for the 7 card evals of the built-in types:
//
//	import _ "github.com/cardrank/cardrank/twoplustwo"
//
// The lookup table is embedded in the package, so it is only included in
// binaries that import the package. The table is decoded into memory on first
// use (about 140 ms).
package twoplustwo

import (
	"github.com/cardrank/cardrank"
	"github.com/cardrank/cardrank/internal/tpt"
)

func init() {
	cardrank.RegisterTwoPlusTwo(New())
}

// New creates a Two-Plus-Two rank eval func for 5, 6, or 7 cards.
func New() func([]cardrank.Card) cardrank.EvalRank {
	return cardrank.NewTwoPlusTwoEval(tpt.Table)
}

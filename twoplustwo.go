package cardrank

// NewTwoPlusTwoEval creates a Two-Plus-Two rank eval func using the lookup
// table returned by table, a version of the 2+2 poker forum rank evaluator,
// providing extremely fast 5, 6, and 7 card lookup.
//
// The table func is called on every eval, and should cache the table (such as
// with [sync.OnceValue]). Use the [twoplustwo] package for a ready to use
// eval func backed by an embedded table.
//
// See: https://github.com/tangentforks/TwoPlusTwoHandEvaluator
//
// [twoplustwo]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo
func NewTwoPlusTwoEval(table func() []uint32) func([]Card) EvalRank {
	// card suit and rank bits to table offset
	var card [256]uint32
	for i, r := uint32(0), Two; r <= Ace; r++ {
		for _, s := range []Suit{Spade, Heart, Club, Diamond} {
			card[byte(New(r, s)>>8)] = i + 1
			i++
		}
	}
	ranks := [10]uint32{
		uint32(Invalid),
		uint32(HighCard),
		uint32(Pair),
		uint32(TwoPair),
		uint32(ThreeOfAKind),
		uint32(Straight),
		uint32(Flush),
		uint32(FullHouse),
		uint32(FourOfAKind),
		uint32(StraightFlush),
	}
	return func(v []Card) EvalRank {
		t, i := table(), uint32(53)
		for _, c := range v {
			i = t[i+card[byte(c>>8)]]
		}
		if len(v) < 7 {
			i = t[i]
		}
		return EvalRank(ranks[i>>12] - i&0xfff + 1)
	}
}

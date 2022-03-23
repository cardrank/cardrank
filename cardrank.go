// Package cardrank provides types and utilities for working with playing cards
// and evaluating poker hands.
package cardrank

import (
	"fmt"
)

// Ranker are the types of hand rankers.
type Ranker int

// Ranker values.
const (
	Cactus Ranker = iota
	CactusFast
	TwoPlus
	Hybrid
	CactusFastSixPlus
	EightOrBetter
)

// rankers are the available card rankers.
var rankers = [...]func([]Card) HandRank{nil, nil, nil, nil, nil, nil}

// String satisfies the fmt.Stringer interface.
func (r Ranker) String() string {
	switch r {
	case Cactus:
		return "Cactus"
	case CactusFast:
		return "CactusFast"
	case TwoPlus:
		return "TwoPlus"
	case Hybrid:
		return "Hybrid"
	case CactusFastSixPlus:
		return "ShortDeckFast"
	case EightOrBetter:
		return "EightOrBetter"
	}
	return fmt.Sprintf("Ranker(%d)", r)
}

// Available indicates if the ranker is available.
func (r Ranker) Available() bool {
	return rankers[r] != nil
}

// Rank ranks the hand.
func (r Ranker) Rank(hand []Card) HandRank {
	return rankers[r](hand)
}

// HandRank is a poker hand rank.
type HandRank uint16

// Poker hand rank values.
const (
	StraightFlush HandRank = 10
	FourOfAKind   HandRank = 166
	FullHouse     HandRank = 322
	Flush         HandRank = 1599
	Straight      HandRank = 1609
	ThreeOfAKind  HandRank = 2467
	TwoPair       HandRank = 3325
	Pair          HandRank = 6185
	Nothing       HandRank = 7462
	HighCard      HandRank = Nothing
	Invalid                = HandRank(^uint16(0))
)

// Fixed converts a relative poker rank to a fixed hand rank.
func (r HandRank) Fixed() HandRank {
	switch {
	case r <= StraightFlush:
		return StraightFlush
	case r <= FourOfAKind:
		return FourOfAKind
	case r <= FullHouse:
		return FullHouse
	case r <= Flush:
		return Flush
	case r <= Straight:
		return Straight
	case r <= ThreeOfAKind:
		return ThreeOfAKind
	case r <= TwoPair:
		return TwoPair
	case r <= Pair:
		return Pair
	case r != Invalid:
		return Nothing
	}
	return Invalid
}

// String satisfies the fmt.Stringer interface.
func (r HandRank) String() string {
	switch r.Fixed() {
	case StraightFlush:
		return "Straight Flush"
	case FourOfAKind:
		return "Four of a Kind"
	case FullHouse:
		return "Full House"
	case Flush:
		return "Flush"
	case Straight:
		return "Straight"
	case ThreeOfAKind:
		return "Three of a Kind"
	case TwoPair:
		return "Two Pair"
	case Pair:
		return "Pair"
	case Nothing:
		return "Nothing"
	}
	return "Invalid"
}

// Name returns the hand rank name.
func (r HandRank) Name() string {
	switch r.Fixed() {
	case StraightFlush:
		return "StraightFlush"
	case FourOfAKind:
		return "FourOfAKind"
	case FullHouse:
		return "FullHouse"
	case Flush:
		return "Flush"
	case Straight:
		return "Straight"
	case ThreeOfAKind:
		return "ThreeOfAKind"
	case TwoPair:
		return "TwoPair"
	case Pair:
		return "Pair"
	}
	return "Nothing"
}

// RankerFunc is a wrapper for ranking funcs.
type RankerFunc func(c0, c1, c2, c3, c4 Card) uint16

// Rank satisfies the Ranker interface.
func (f RankerFunc) Rank(hand []Card) HandRank {
	switch n := len(hand); {
	case n == 5:
		return HandRank(f(hand[0], hand[1], hand[2], hand[3], hand[4]))
	case n == 6:
		r := f(hand[0], hand[1], hand[2], hand[3], hand[4])
		r = min(r, f(hand[0], hand[1], hand[2], hand[3], hand[5]))
		r = min(r, f(hand[0], hand[1], hand[2], hand[4], hand[5]))
		r = min(r, f(hand[0], hand[1], hand[3], hand[4], hand[5]))
		r = min(r, f(hand[0], hand[2], hand[3], hand[4], hand[5]))
		r = min(r, f(hand[1], hand[2], hand[3], hand[4], hand[5]))
		return HandRank(r)
	}
	r, rank := uint16(0), uint16(9999)
	for i := 0; i < 21; i++ {
		if r = f(
			hand[h7c5[i][0]],
			hand[h7c5[i][1]],
			hand[h7c5[i][2]],
			hand[h7c5[i][3]],
			hand[h7c5[i][4]],
		); r < rank {
			rank = r
		}
	}
	return HandRank(rank)
}

// h7c5 is used to choose 5 cards from a hand of 7.
var h7c5 = [21][5]uint8{
	{0, 1, 2, 3, 4},
	{0, 1, 2, 3, 5},
	{0, 1, 2, 3, 6},
	{0, 1, 2, 4, 5},
	{0, 1, 2, 4, 6},
	{0, 1, 2, 5, 6},
	{0, 1, 3, 4, 5},
	{0, 1, 3, 4, 6},
	{0, 1, 3, 5, 6},
	{0, 1, 4, 5, 6},
	{0, 2, 3, 4, 5},
	{0, 2, 3, 4, 6},
	{0, 2, 3, 5, 6},
	{0, 2, 4, 5, 6},
	{0, 3, 4, 5, 6},
	{1, 2, 3, 4, 5},
	{1, 2, 3, 4, 6},
	{1, 2, 3, 5, 6},
	{1, 2, 4, 5, 6},
	{1, 3, 4, 5, 6},
	{2, 3, 4, 5, 6},
}

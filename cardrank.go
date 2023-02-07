// Package github.com/cardrank/cardrank is a library of types, utilities, and
// interfaces for working with playing cards, card decks, and evaluating poker
// hand ranks.
//
// Supports Texas Holdem, Texas Holdem Short (6+), Texas Holdem Royal (10+),
// Omaha, Omaha Hi/Lo, Stud, Stud Hi/Lo, Razz, and Badugi.
package cardrank

import (
	"sort"
)

// HandRank is a poker hand rank.
//
// Ranks are ordered low-to-high.
type HandRank uint16

// Poker hand rank values.
//
// See: https://archive.is/G6GZg
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

// SixPlusRanker creates a 6-plus hand ranker.
func SixPlusRanker(f RankFiveFunc) RankFiveFunc {
	return func(c0, c1, c2, c3, c4 Card) uint16 {
		r := f(c0, c1, c2, c3, c4)
		switch r {
		case 747: // Straight Flush, 9, 8, 7, 6, Ace
			return 6
		case 6610: // Straight, 9, 8, 7, 6, Ace
			return 1605
		}
		return r
	}
}

// EightOrBetterRanker is a 8-or-better low hand ranker. Aces are low,
// straights and flushes do not count. Any card with rank 8 or higher will
// cause
func EightOrBetterRanker(c0, c1, c2, c3, c4 Card) uint16 {
	return aceFiveLow(0xff00, c0, c1, c2, c3, c4)
}

// RazzRanker is a Razz (Ace-to-Five) low hand ranker. Aces are low, straights
// and flushes do not count.
//
// When there is a pair (or higher) of matching ranks, will be the inverted
// value of the regular hand rank.
func RazzRanker(c0, c1, c2, c3, c4 Card) uint16 {
	if r := aceFiveLow(0, c0, c1, c2, c3, c4); r < lowMaxRank {
		return r
	}
	return ^uint16(0) - DefaultCactus(c0, c1, c2, c3, c4)
}

// aceFiveLow is a Ace-to-Five low hand ranker.
func aceFiveLow(mask uint16, c0, c1, c2, c3, c4 Card) uint16 {
	rank := uint16(0)
	// c0
	r := uint16(c0>>8&0xf+1) % 13
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c1
	r = uint16(c1>>8&0xf+1) % 13
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c2
	r = uint16(c2>>8&0xf+1) % 13
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c3
	r = uint16(c3>>8&0xf+1) % 13
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c4
	r = uint16(c4>>8&0xf+1) % 13
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	return rank
}

// twoSixLow is a Two-Six low hand ranker.
func twoSixLow(mask uint16, c0, c1, c2, c3, c4 Card) uint16 {
	rank := uint16(0)
	// c0
	r := uint16(c0 >> 8 & 0xf)
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c1
	r = uint16(c1 >> 8 & 0xf)
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c2
	r = uint16(c2 >> 8 & 0xf)
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c3
	r = uint16(c3 >> 8 & 0xf)
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c4
	r = uint16(c4 >> 8 & 0xf)
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	return rank
}

// BadugiRanker is a Badugi hand ranker.
func BadugiRanker(hand []Card) (HandRank, []Card, []Card) {
	s := make([][]Card, 4)
	for i := 0; i < len(hand) && i < 4; i++ {
		idx := hand[i].SuitIndex()
		s[idx] = append(s[idx], hand[i])
	}
	sort.SliceStable(s, func(i, j int) bool {
		a, b := len(s[i]), len(s[j])
		switch {
		case a != b:
			return a < b
		case a == 0:
			return true
		case b == 0:
			return false
		}
		return uint16(s[i][0]>>8&0xf+1)%13 < uint16(s[j][0]>>8&0xf+1)%13
	})
	count, rank := 4, 0
	var best, unused []Card
	for i := 0; i < 4; i++ {
		sort.Slice(s[i], func(j, k int) bool {
			return uint16(s[i][j]>>8&0xf+1)%13 < uint16(s[i][k]>>8&0xf+1)%13
		})
		captured := false
		for j := 0; j < len(s[i]); j++ {
			if r := 1 << (uint16(s[i][j]>>8&0xf+1) % 13); rank&r == 0 && !captured {
				captured, best = true, append(best, s[i][j])
				rank |= r
				count--
			} else {
				unused = append(unused, s[i][j])
			}
		}
	}
	sort.Slice(best, func(i, j int) bool {
		return uint16(best[i]>>8&0xf+1)%13 > uint16(best[j]>>8&0xf+1)%13
	})
	sort.Slice(unused, func(i, j int) bool {
		a, b := uint16(unused[i]>>8&0xf+1)%13, uint16(unused[j]>>8&0xf+1)%13
		if a != b {
			return a > b
		}
		return unused[i].Suit() < unused[j].Suit()
	})
	return HandRank(count<<13 | rank), best, unused
}

// HandRanker creates a new hand ranker for 5, 6, or 7 cards using f.
func HandRanker(f RankFiveFunc) RankerFunc {
	return func(hand []Card) HandRank {
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
				hand[t7c5[i][0]],
				hand[t7c5[i][1]],
				hand[t7c5[i][2]],
				hand[t7c5[i][3]],
				hand[t7c5[i][4]],
			); r < rank {
				rank = r
			}
		}
		return HandRank(rank)
	}
}

// HybridRanker creates a hybrid ranker using f5 for hands with 5 and 6 cards,
// and f7 for hands with 7 cards.
func HybridRanker(f5 RankFiveFunc, f7 RankerFunc) RankerFunc {
	return func(hand []Card) HandRank {
		switch len(hand) {
		case 5:
			return HandRank(f5(hand[0], hand[1], hand[2], hand[3], hand[4]))
		case 6:
			r := f5(hand[0], hand[1], hand[2], hand[3], hand[4])
			r = min(r, f5(hand[0], hand[1], hand[2], hand[3], hand[5]))
			r = min(r, f5(hand[0], hand[1], hand[2], hand[4], hand[5]))
			r = min(r, f5(hand[0], hand[1], hand[3], hand[4], hand[5]))
			r = min(r, f5(hand[0], hand[2], hand[3], hand[4], hand[5]))
			r = min(r, f5(hand[1], hand[2], hand[3], hand[4], hand[5]))
			return HandRank(r)
		}
		return f7(hand)
	}
}

// RankFiveFunc ranks a hand of 5 cards.
type RankFiveFunc func(c0, c1, c2, c3, c4 Card) uint16

// RankerFunc ranks a hand of 5, 6, or 7 cards.
type RankerFunc func([]Card) HandRank

// DefaultRanker is the default hand ranker.
var DefaultRanker RankerFunc

// DefaultSixPlusRanker is the default 6-plus (short deck) hand ranker.
var DefaultSixPlusRanker RankerFunc

// DefaultCactus is the default Cactus Kev implementation.
var DefaultCactus RankFiveFunc

// Package rankers (set in z.go).
var (
	cactus     RankFiveFunc
	cactusFast RankFiveFunc
	twoPlusTwo RankerFunc
)

// Init inits the package level default variables. Must be manually called
// prior to using this package when built with the `noinit` build tag.
func Init() {
	switch {
	case twoPlusTwo != nil && cactusFast != nil:
		DefaultRanker = HybridRanker(cactusFast, twoPlusTwo)
	case cactusFast != nil:
		DefaultRanker = HandRanker(cactusFast)
	case cactus != nil:
		DefaultRanker = HandRanker(cactus)
	}
	switch {
	case cactusFast != nil:
		DefaultCactus, DefaultSixPlusRanker = cactusFast, HandRanker(SixPlusRanker(cactusFast))
	case cactus != nil:
		DefaultCactus, DefaultSixPlusRanker = cactus, HandRanker(SixPlusRanker(cactus))
	}
}

// Error is a error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidCard is the invalid card error.
	ErrInvalidCard Error = "invalid card"
	// ErrInvalidType is the invalid type error.
	ErrInvalidType Error = "invalid type"
)

// min returns the min of a, b.
func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

// t4c2 is used for taking 4, choosing 2.
var t4c2 = [6][4]int{
	{0, 1, 2, 3},
	{0, 2, 1, 3},
	{0, 3, 1, 2},
	{1, 2, 0, 3},
	{1, 3, 0, 2},
	{2, 3, 0, 1},
}

// t5c3 is used for taking 5, choosing 3.
var t5c3 = [10][5]int{
	{0, 1, 2, 3, 4},
	{0, 1, 3, 2, 4},
	{0, 1, 4, 2, 3},
	{0, 2, 3, 1, 4},
	{0, 2, 4, 1, 3},
	{0, 3, 4, 1, 2},
	{1, 2, 3, 0, 4},
	{1, 2, 4, 0, 3},
	{1, 3, 4, 0, 2},
	{2, 3, 4, 0, 1},
}

// t7c5 is used for taking 7, choosing 5.
var t7c5 = [21][7]uint8{
	{0, 1, 2, 3, 4, 5, 6},
	{0, 1, 2, 3, 5, 4, 6},
	{0, 1, 2, 3, 6, 4, 5},
	{0, 1, 2, 4, 5, 3, 6},
	{0, 1, 2, 4, 6, 3, 5},
	{0, 1, 2, 5, 6, 3, 4},
	{0, 1, 3, 4, 5, 2, 6},
	{0, 1, 3, 4, 6, 2, 5},
	{0, 1, 3, 5, 6, 2, 4},
	{0, 1, 4, 5, 6, 2, 3},
	{0, 2, 3, 4, 5, 1, 6},
	{0, 2, 3, 4, 6, 1, 5},
	{0, 2, 3, 5, 6, 1, 4},
	{0, 2, 4, 5, 6, 1, 3},
	{0, 3, 4, 5, 6, 1, 2},
	{1, 2, 3, 4, 5, 0, 6},
	{1, 2, 3, 4, 6, 0, 5},
	{1, 2, 3, 5, 6, 0, 4},
	{1, 2, 4, 5, 6, 0, 3},
	{1, 3, 4, 5, 6, 0, 2},
	{2, 3, 4, 5, 6, 0, 1},
}

// primes are the first 13 prime numbers (one per card rank).
var primes = [...]uint8{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}

const (
	// eightOrBetterMaxRank is the eight-or-better max rank for a qualifying
	// low hand.
	eightOrBetterMaxRank = 512
	// lowMaxRank is the low max rank for a qualifying low hand.
	lowMaxRank = 16384
)

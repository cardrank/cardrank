// Package github.com/cardrank/cardrank is a library of types, utilities, and
// interfaces for working with playing cards, card decks, and evaluating poker
// hand ranks.
package cardrank

// HandRank is a poker hand rank.
//
// Ranks are ordered low-to-high.
type HandRank uint16

// Poker hand rank values.
//
// See: https://archive.is/G6GZg
const (
	StraightFlush        HandRank = 10
	FourOfAKind          HandRank = 166
	FullHouse            HandRank = 322
	Flush                HandRank = 1599
	Straight             HandRank = 1609
	ThreeOfAKind         HandRank = 2467
	TwoPair              HandRank = 3325
	Pair                 HandRank = 6185
	Nothing              HandRank = 7462
	HighCard             HandRank = Nothing
	Invalid                       = ^HandRank(0)
	rankMax                       = Nothing + 1
	rankEightOrBetterMax HandRank = 512
	rankLowMax           HandRank = 16384
)

// Fixed converts a relative poker rank to a fixed rank.
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

// RankFunc ranks a hand of 5 cards.
type RankFunc func(c0, c1, c2, c3, c4 Card) HandRank

// RankEightOrBetter is a 8-or-better low hand rank func. Aces are low,
// straights and flushes do not count.
func RankEightOrBetter(c0, c1, c2, c3, c4 Card) HandRank {
	return RankLowAceFive(0xff00, c0, c1, c2, c3, c4)
}

// RankRazz is a Razz (Ace-to-Five) low hand rank func. Aces are low, straights
// and flushes do not count.
//
// When there is a pair (or higher) of matching ranks, will be the inverted
// value of the regular hand rank.
func RankRazz(c0, c1, c2, c3, c4 Card) HandRank {
	if r := RankLowAceFive(0, c0, c1, c2, c3, c4); r < rankLowMax {
		return r
	}
	return Invalid - DefaultCactus(c0, c1, c2, c3, c4)
}

// RankLowAceFive is a Ace-to-Five low hand rank func.
func RankLowAceFive(mask HandRank, c0, c1, c2, c3, c4 Card) HandRank {
	rank := HandRank(0)
	// c0
	r := c0.AceIndex()
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c1
	r = c1.AceIndex()
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c2
	r = c2.AceIndex()
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c3
	r = c3.AceIndex()
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	mask |= 1 << r
	// c4
	r = c4.AceIndex()
	rank |= 1<<r | ((mask&(1<<r)>>r)&1)*0x8000
	return rank
}

// RankLowball is a Two-to-Seven low hand rank func.
func RankLowball(c0, c1, c2, c3, c4 Card) HandRank {
	return rankMax - DefaultCactus(c0, c1, c2, c3, c4)
}

// HandRankFunc ranks a hand of 5, 6, or 7 cards.
type HandRankFunc func([]Card) HandRank

// NewRankFunc creates a hand eval for 5, 6, or 7 cards using f.
func NewRankFunc(f RankFunc) HandRankFunc {
	return func(hand []Card) HandRank {
		switch n := len(hand); {
		case n == 5:
			return f(hand[0], hand[1], hand[2], hand[3], hand[4])
		case n == 6:
			r := f(hand[0], hand[1], hand[2], hand[3], hand[4])
			r = min(r, f(hand[0], hand[1], hand[2], hand[3], hand[5]))
			r = min(r, f(hand[0], hand[1], hand[2], hand[4], hand[5]))
			r = min(r, f(hand[0], hand[1], hand[3], hand[4], hand[5]))
			r = min(r, f(hand[0], hand[2], hand[3], hand[4], hand[5]))
			r = min(r, f(hand[1], hand[2], hand[3], hand[4], hand[5]))
			return r
		}
		r, rank := HandRank(0), Invalid
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
		return rank
	}
}

// NewHybrid creates a hybrid rank func using f5 for hands with 5 and 6 cards,
// and f7 for hands with 7 cards.
func NewHybrid(f5 RankFunc, f7 HandRankFunc) HandRankFunc {
	return func(hand []Card) HandRank {
		switch len(hand) {
		case 5:
			return f5(hand[0], hand[1], hand[2], hand[3], hand[4])
		case 6:
			r := f5(hand[0], hand[1], hand[2], hand[3], hand[4])
			r = min(r, f5(hand[0], hand[1], hand[2], hand[3], hand[5]))
			r = min(r, f5(hand[0], hand[1], hand[2], hand[4], hand[5]))
			r = min(r, f5(hand[0], hand[1], hand[3], hand[4], hand[5]))
			r = min(r, f5(hand[0], hand[2], hand[3], hand[4], hand[5]))
			r = min(r, f5(hand[1], hand[2], hand[3], hand[4], hand[5]))
			return r
		}
		return f7(hand)
	}
}

var (
	// DefaultRank is the default hand rank func.
	DefaultRank HandRankFunc
	// DefaultCactus is the default Cactus Kev implementation.
	DefaultCactus RankFunc

	// Package rank funcs (set in z.go).
	cactus     RankFunc
	cactusFast RankFunc
	twoPlusTwo HandRankFunc
)

// Init inits the package level default variables. Must be manually called
// prior to using this package when built with the `noinit` build tag.
func Init() error {
	switch {
	case twoPlusTwo != nil && cactusFast != nil:
		DefaultRank = NewHybrid(cactusFast, twoPlusTwo)
	case cactusFast != nil:
		DefaultRank = NewRankFunc(cactusFast)
	case cactus != nil:
		DefaultRank = NewRankFunc(cactus)
	}
	switch {
	case cactusFast != nil:
		DefaultCactus = cactusFast
	case cactus != nil:
		DefaultCactus = cactus
	}
	return RegisterDefaultTypes()
}

// Error is a error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidId is the invalid id error.
	ErrInvalidId Error = "invalid id"
	// ErrMismatchedIdAndType is the mismatched id and type error.
	ErrMismatchedIdAndType Error = "mismatched id and type"
	// ErrInvalidCard is the invalid card error.
	ErrInvalidCard Error = "invalid card"
	// ErrInvalidType is the invalid type error.
	ErrInvalidType Error = "invalid type"
)

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

// t4c2 is used for taking 4, choosing 2.
var t4c2 = [6][4]uint8{
	{0, 1, 2, 3},
	{0, 2, 1, 3},
	{0, 3, 1, 2},
	{1, 2, 0, 3},
	{1, 3, 0, 2},
	{2, 3, 0, 1},
}

// t5c2 is used for taking 5, choosing 2.
var t5c2 = [10][5]uint8{
	{0, 1, 2, 3, 4},
	{0, 2, 1, 3, 4},
	{0, 3, 1, 2, 4},
	{0, 4, 1, 2, 3},
	{1, 2, 0, 3, 4},
	{1, 3, 0, 2, 4},
	{1, 4, 0, 2, 3},
	{2, 3, 0, 1, 4},
	{2, 4, 0, 1, 3},
	{3, 4, 0, 1, 2},
}

// t5c3 is used for taking 5, choosing 3.
var t5c3 = [10][5]uint8{
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

// t6c2 is used for taking 6, choosing 2.
var t6c2 = [15][6]uint8{
	{0, 1, 2, 3, 4, 5},
	{0, 2, 1, 3, 4, 5},
	{0, 3, 1, 2, 4, 5},
	{0, 4, 1, 2, 3, 5},
	{0, 5, 1, 2, 3, 4},
	{1, 2, 0, 3, 4, 5},
	{1, 3, 0, 2, 4, 5},
	{1, 4, 0, 2, 3, 5},
	{1, 5, 0, 2, 3, 4},
	{2, 3, 0, 1, 4, 5},
	{2, 4, 0, 1, 3, 5},
	{2, 5, 0, 1, 3, 4},
	{3, 4, 0, 1, 2, 5},
	{3, 5, 0, 1, 2, 4},
	{4, 5, 0, 1, 2, 3},
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
var primes = [...]uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}

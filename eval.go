package cardrank

import (
	"fmt"
	"sort"
)

// EvalRank is a eval rank.
//
// Ranks are ordered low-to-high.
type EvalRank uint16

// Eval ranks.
//
// See: https://archive.is/G6GZg
const (
	StraightFlush     EvalRank = 10
	FourOfAKind       EvalRank = 166
	FullHouse         EvalRank = 322
	Flush             EvalRank = 1599
	Straight          EvalRank = 1609
	ThreeOfAKind      EvalRank = 2467
	TwoPair           EvalRank = 3325
	Pair              EvalRank = 6185
	Nothing           EvalRank = 7462
	HighCard          EvalRank = Nothing
	Invalid           EvalRank = ^EvalRank(0)
	jacksOrBetterMax  EvalRank = 4205
	eightOrBetterMax  EvalRank = 512
	aceFiveMax        EvalRank = 16384
	flushUnder        EvalRank = 156
	flushOver         EvalRank = 1277
	lowballAceFlush   EvalRank = 811
	lowballAceNothing EvalRank = 6678
	sokoFlush         EvalRank = TwoPair + 13*715
	sokoStraight      EvalRank = sokoFlush + 13*10
	sokoNothing       EvalRank = sokoStraight + (Nothing - TwoPair)
)

// Fixed converts a relative eval rank to a fixed eval rank.
func (r EvalRank) Fixed() EvalRank {
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

// String satisfies the [fmt.Stringer] interface.
func (r EvalRank) String() string {
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

// Name returns the eval rank name.
func (r EvalRank) Name() string {
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

// ToFlushOver changes a Cactus rank to a Flush Over a Full House rank.
//
//	FullHouse: FullHouse(322) - FourOfAKind(166) == 156
//	Flush:     Flush(1599)    - FullHouse(322)   == 1277
func (r EvalRank) ToFlushOver() EvalRank {
	switch {
	case FourOfAKind < r && r <= FullHouse:
		return r + flushOver
	case FullHouse < r && r <= Flush:
		return r - flushUnder
	}
	return r
}

// FromFlushOver changes a rank from a Flush Over a Full House rank to a Cactus
// rank.
//
//	FullHouse: FullHouse(322) - FourOfAKind(166) == 156
//	Flush:     Flush(1599)    - FullHouse(322)   == 1277
func (r EvalRank) FromFlushOver() EvalRank {
	switch {
	case FourOfAKind < r && r <= FourOfAKind+flushOver:
		return r + flushUnder
	case FourOfAKind+flushOver < r && r <= Flush:
		return r - flushOver
	}
	return r
}

// ToLowball converts a Cactus rank to a [Lowball] rank, by inverting the rank
// and converting the lowest Straight and Straight Flushes (5-4-3-2-A) to
// different ranks.
//
// Changes the rank as follows:
//
//	Moves lowest Straight Flush (10) to lowest Ace Flush (811)
//	Moves any rank between Straight Flush (10) < r <= lowest Ace Flush (811) down 1 rank
//	Moves lowest Straight (1609) to lowest Ace Nothing (6678)
//	Moves any rank between Straight (1609) < r <= lowest Ace Nothing (6678) down 1 rank
//	Inverts the rank (Nothing - r + 1)
func (r EvalRank) ToLowball() EvalRank {
	switch {
	case r == StraightFlush:
		// change lowest straight flush to lowest ace high flush
		r = lowballAceFlush
	case StraightFlush < r && r <= lowballAceFlush:
		// move everything between 11 and 811 down 1
		r--
	case r == Straight:
		// change lowest ace straight to lowest ace high nothing
		r = lowballAceNothing
	case Straight < r && r <= lowballAceNothing:
		// move everything between 1610 and 6678 down 1
		r--
	}
	return Nothing - r + 1
}

// FromLowball converts a [Lowball] rank to a Cactus rank.
//
// See [EvalRank.ToLowball] for a description of the operations performed.
func (r EvalRank) FromLowball() EvalRank {
	r = Nothing - (r - 1)
	switch {
	case r == lowballAceFlush:
		// change lowest lowest ace high flush to lowest straight flush
		r = StraightFlush
	case StraightFlush <= r && r < lowballAceFlush:
		// move everything between 10 and 812 up 1
		r++
	case r == lowballAceNothing:
		// change lowest ace high nothing to lowest ace straight
		r = Straight
	case Straight <= r && r < lowballAceNothing:
		// move everything between 1609 and 6679 up 1
		r++
	}
	return r
}

// RankFunc returns the eval rank of 5 cards.
type RankFunc func(c0, c1, c2, c3, c4 Card) EvalRank

// RankAceFiveLow is a A-to-5 low rank eval func. [Ace]'s are low, [Straight]'s
// and [Flush]'s do not count.
func RankAceFiveLow(mask EvalRank, c0, c1, c2, c3, c4 Card) EvalRank {
	var rank EvalRank
	// c0
	n := c0.AceRank()
	rank |= 1<<n | ((mask&(1<<n)>>n)&1)*0x8000
	mask |= 1 << n
	// c1
	n = c1.AceRank()
	rank |= 1<<n | ((mask&(1<<n)>>n)&1)*0x8000
	mask |= 1 << n
	// c2
	n = c2.AceRank()
	rank |= 1<<n | ((mask&(1<<n)>>n)&1)*0x8000
	mask |= 1 << n
	// c3
	n = c3.AceRank()
	rank |= 1<<n | ((mask&(1<<n)>>n)&1)*0x8000
	mask |= 1 << n
	// c4
	n = c4.AceRank()
	rank |= 1<<n | ((mask&(1<<n)>>n)&1)*0x8000
	return rank
}

// RankEightOrBetter is a 8-or-better low rank eval func. [Ace]'s are low,
// [Straight]'s and [Flush]'s do not count.
func RankEightOrBetter(c0, c1, c2, c3, c4 Card) EvalRank {
	return RankAceFiveLow(0xff00, c0, c1, c2, c3, c4)
}

// RankShort is a [Short] rank eval func.
func RankShort(c0, c1, c2, c3, c4 Card) EvalRank {
	r := RankCactus(c0, c1, c2, c3, c4)
	switch r {
	case 747:
		// promote to Straight Flush, 9, 8, 7, 6, Ace
		r = 6
	case 6610:
		// promote to Straight, 9, 8, 7, 6, Ace
		r = 1605
	}
	return r.ToFlushOver()
}

// RankManila is a [Manila] rank eval func.
func RankManila(c0, c1, c2, c3, c4 Card) EvalRank {
	r := RankCactus(c0, c1, c2, c3, c4)
	switch r {
	case 691:
		// promote to Straight Flush, T, 9, 8, 7, Ace
		r = 5
	case 6554:
		// promote to Straight, T, 9, 8, 7, Ace
		r = 1604
	}
	return r.ToFlushOver()
}

// RankSpanish is a [Spanish] rank eval func.
func RankSpanish(c0, c1, c2, c3, c4 Card) EvalRank {
	r := RankCactus(c0, c1, c2, c3, c4)
	switch r {
	case 607:
		// promote to Straight Flush, J, 10, 9, 8, Ace
		r = 4
	case 6470:
		// promote to Straight, J, 10, 9, 8, Ace
		r = 1603
	}
	return r.ToFlushOver()
}

// RankRazz is a [Razz] (A-to-5) low rank eval func. [Ace]'s are low,
// [Straight]'s and [Flush]'s do not count.
//
// When there is a [Pair] (or higher) of matching ranks, will be the inverted
// Cactus value.
func RankRazz(c0, c1, c2, c3, c4 Card) EvalRank {
	if r := RankAceFiveLow(0, c0, c1, c2, c3, c4); r < aceFiveMax {
		return r
	}
	return Invalid - RankCactus(c0, c1, c2, c3, c4)
}

// RankLowball is a [Lowball] (2-to-7) low rank eval func. [Ace]'s are high,
// [Straight]'s and [Flush]'s count.
//
// Works by adding 2 additional ranks for [Ace]-high [StraightFlush]'s and
// [Straight]'s.
//
// See [EvalRank.ToLowball].
func RankLowball(c0, c1, c2, c3, c4 Card) EvalRank {
	return RankCactus(c0, c1, c2, c3, c4).ToLowball()
}

// EvalFunc is a eval func.
type EvalFunc func(*Eval, []Card, []Card)

// NewEval returns a eval func that ranks 5, 6, or 7 cards using f. The
// returned eval func will store the results on an eval's Hi.
func NewEval(f RankFunc) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, []Card)
		n, m := len(p), len(b)
		switch n + m {
		default:
			return
		case 5:
			eval = ev.Hi5
		case 6:
			eval = ev.Hi6
		case 7:
			eval = ev.Hi7
		}
		v := make([]Card, n+m)
		copy(v, p)
		copy(v[n:], b)
		eval(f, v)
	}
}

// NewMaxEval returns a eval func that ranks 5, 6, or 7 cards using f and max.
//
// The returned eval func will store results on an eval's Hi only when lower
// than max.
func NewMaxEval(f RankFunc, max EvalRank, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, []Card, EvalRank, bool)
		n, m := len(p), len(b)
		switch n + m {
		default:
			return
		case 5:
			eval = ev.Max5
		case 6:
			eval = ev.Max6
		case 7:
			eval = ev.Max7
		}
		v := make([]Card, n+m)
		copy(v, p)
		copy(v[n:], b)
		eval(f, v, max, low)
	}
}

// NewSplitEval returns a eval func that ranks 5, 6, or 7 cards using hi, lo
// and max.
//
// The returned eval func will store results on an eval's Hi and Lo depending
// on the result of hi and lo, respectively. Will store the Lo value only when
// lower than max.
func NewSplitEval(hi, lo RankFunc, max EvalRank) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, RankFunc, []Card, EvalRank)
		n, m := len(p), len(b)
		switch n + m {
		default:
			return
		case 5:
			eval = ev.HiLo5
		case 6:
			eval = ev.HiLo6
		case 7:
			eval = ev.HiLo7
		}
		v := make([]Card, n+m)
		copy(v, p)
		copy(v[n:], b)
		eval(hi, lo, v, max)
	}
}

// NewHybridEval creates a hybrid Cactus and TwoPlusTwo eval func, using
// [RankCactus] for 5 and 6 cards, and a TwoPlusTwo eval func for 7 cards.
//
// Gives optimal performance when evaluating the best-5 of any 5, 6, or 7 cards
// of a combined pocket and board.
func NewHybridEval(normalize, low bool) EvalFunc {
	var f EvalFunc
	if low {
		f = NewSplitEval(RankCactus, RankEightOrBetter, eightOrBetterMax)
	} else {
		f = NewEval(RankCactus)
	}
	return func(ev *Eval, p, b []Card) {
		switch n, m := len(p), len(b); n + m {
		case 5, 6:
			f(ev, p, b)
			if normalize {
				bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
				if low && ev.LoRank < eightOrBetterMax {
					bestAceLow(ev.LoBest)
					bestAceHigh(ev.LoUnused)
				}
			}
		case 7:
			v := make([]Card, n+m)
			copy(v, p)
			copy(v[n:], b)
			ev.HiRank = twoPlusTwo(v)
			if normalize {
				ev.HiBest, ev.HiUnused = bestCactusSplit(ev.HiRank, v, 0)
			}
			if low {
				u := make([]Card, n+m)
				copy(u, p)
				copy(u[n:], b)
				ev.Max7(RankEightOrBetter, u, eightOrBetterMax, true)
				if normalize && ev.LoRank < eightOrBetterMax {
					bestAceLow(ev.LoBest)
					bestAceHigh(ev.LoUnused)
				}
			}
		}
	}
}

// NewCactusEval creates a Cactus eval func.
func NewCactusEval(normalize, low bool) EvalFunc {
	var f EvalFunc
	switch {
	case twoPlusTwo != nil:
		f = NewHybridEval(normalize, low)
	case low:
		f = NewSplitEval(RankCactus, RankEightOrBetter, eightOrBetterMax)
	default:
		f = NewEval(RankCactus)
	}
	return func(ev *Eval, p, b []Card) {
		f(ev, p, b)
		if normalize && twoPlusTwo == nil {
			bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
			if low {
				bestAceLow(ev.LoBest)
				bestAceHigh(ev.LoUnused)
			}
		}
	}
}

// NewModifiedEval creates a modified Cactus eval.
func NewModifiedEval(hi RankFunc, base Rank, inv func(EvalRank) EvalRank, normalize, low bool) EvalFunc {
	var f EvalFunc
	if low {
		f = NewSplitEval(hi, RankEightOrBetter, eightOrBetterMax)
	} else {
		f = NewEval(hi)
	}
	return func(ev *Eval, p, b []Card) {
		f(ev, p, b)
		if normalize {
			bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, base, inv)
			if low {
				bestAceLow(ev.LoBest)
				bestAceHigh(ev.LoUnused)
			}
		}
	}
}

// NewJacksOrBetterEval creates a JacksOrBetter eval func, used for [Video].
func NewJacksOrBetterEval(normalize bool) EvalFunc {
	hi := NewMaxEval(RankCactus, jacksOrBetterMax, false)
	return func(ev *Eval, p, b []Card) {
		hi(ev, p, b)
		if normalize {
			bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
		}
	}
}

// NewShortEval creates a [Short] eval func.
func NewShortEval(normalize bool) EvalFunc {
	return NewModifiedEval(RankShort, Rank(DeckShort), EvalRank.FromFlushOver, normalize, false)
}

// NewManilaEval creates a [Manila] eval func.
func NewManilaEval(normalize bool) EvalFunc {
	return NewDallasEval(RankManila, Rank(DeckManila), EvalRank.FromFlushOver, normalize, false)
}

// NewSpanishEval creates a [Spanish] eval func.
func NewSpanishEval(normalize bool) EvalFunc {
	return NewDallasEval(RankSpanish, Rank(DeckSpanish), EvalRank.FromFlushOver, normalize, false)
}

// NewDallasEval creates a [Dallas] eval func.
//
// Uses pocket of 2 and any 3 from a board of 3, 4, or 5 to make a best-5.
func NewDallasEval(hi RankFunc, base Rank, inv func(EvalRank) EvalRank, normalize, low bool) EvalFunc {
	lo, max := RankFunc(nil), Invalid
	if low {
		lo, max = RankEightOrBetter, eightOrBetterMax
	}
	return func(ev *Eval, p, b []Card) {
		if len(p) < 2 {
			return
		}
		var f func(RankFunc, RankFunc, Card, Card, []Card, EvalRank)
		switch len(b) {
		case 3:
			f = ev.HiLo23
		case 4:
			f = ev.HiLo24
		case 5:
			f = ev.HiLo25
		}
		f(hi, lo, p[0], p[1], b, max)
		if normalize {
			bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, base, inv)
			if low {
				bestAceLow(ev.LoUnused)
				bestAceHigh(ev.LoUnused)
			}
		}
	}
}

// NewHoustonEval creates a [Houston] eval func.
//
// Uses pocket of any 2 from 3, and any 3 from a board of 3, 4, or 5 to make a
// best-5.
func NewHoustonEval(hi RankFunc, base Rank, inv func(EvalRank) EvalRank, normalize, low bool) EvalFunc {
	f := NewDallasEval(hi, base, inv, normalize, low)
	return func(ev *Eval, p, b []Card) {
		if len(p) != 3 {
			return
		}
		v := make([]Card, 2)
		for i := 0; i < 3; i++ {
			v[0], v[1] = p[i%3], p[(i+1)%3]
			uv := EvalOf(ev.Type)
			f(uv, v, b)
			if uv.HiRank < ev.HiRank {
				ev.HiRank, ev.HiBest, ev.HiUnused = uv.HiRank, uv.HiBest, uv.HiUnused
				ev.HiUnused = append(ev.HiUnused, p[(i+2)%3])
			}
		}
	}
}

// NewOmahaEval creates a [Omaha] eval func.
func NewOmahaEval(normalize, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		ev.Init(5, 4, low)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 6; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = p[t4c2[i][0]], p[t4c2[i][1]] // pocket
				v[2], v[3] = b[t5c3[j][0]], b[t5c3[j][1]] // board
				v[4] = b[t5c3[j][2]]                      // board
				if r = RankCactus(v[0], v[1], v[2], v[3], v[4]); r < ev.HiRank {
					ev.HiRank = r
					copy(ev.HiBest, v)
					ev.HiUnused[0], ev.HiUnused[1] = p[t4c2[i][2]], p[t4c2[i][3]] // pocket
					ev.HiUnused[2], ev.HiUnused[3] = b[t5c3[j][3]], b[t5c3[j][4]] // board
				}
				if low {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < eightOrBetterMax {
						ev.LoRank = r
						copy(ev.LoBest, v)
						ev.LoUnused[0], ev.LoUnused[1] = p[t4c2[i][2]], p[t4c2[i][3]] // pocket
						ev.LoUnused[2], ev.LoUnused[3] = b[t5c3[j][3]], b[t5c3[j][4]] // board
					}
				}
			}
		}
		bestOmaha(ev, normalize, low)
	}
}

// NewOmahaFiveEval creates a [OmahaFive] eval func.
func NewOmahaFiveEval(normalize, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		ev.Init(5, 5, low)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = p[t5c2[i][0]], p[t5c2[i][1]] // pocket
				v[2], v[3] = b[t5c3[j][0]], b[t5c3[j][1]] // board
				v[4] = b[t5c3[j][2]]                      // board
				if r = RankCactus(v[0], v[1], v[2], v[3], v[4]); r < ev.HiRank {
					ev.HiRank = r
					copy(ev.HiBest, v)
					ev.HiUnused[0], ev.HiUnused[1] = p[t5c2[i][2]], p[t5c2[i][3]] // pocket
					ev.HiUnused[2] = p[t5c2[i][4]]                                // pocket
					ev.HiUnused[3], ev.HiUnused[4] = b[t5c3[j][3]], b[t5c3[j][4]] // board
				}
				if low {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < eightOrBetterMax {
						ev.LoRank = r
						copy(ev.LoBest, v)
						ev.LoUnused[0], ev.LoUnused[1] = p[t5c2[i][2]], p[t5c2[i][3]] // pocket
						ev.LoUnused[2] = p[t5c2[i][4]]                                // pocket
						ev.LoUnused[3], ev.LoUnused[4] = b[t5c3[j][3]], b[t5c3[j][4]] // board
					}
				}
			}
		}
		bestOmaha(ev, normalize, low)
	}
}

// NewOmahaSixEval creates a [OmahaSix] eval func.
func NewOmahaSixEval(normalize, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		ev.Init(5, 6, low)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 15; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = p[t6c2[i][0]], p[t6c2[i][1]] // pocket
				v[2], v[3] = b[t5c3[j][0]], b[t5c3[j][1]] // board
				v[4] = b[t5c3[j][2]]                      // board
				if r = RankCactus(v[0], v[1], v[2], v[3], v[4]); r < ev.HiRank {
					ev.HiRank = r
					copy(ev.HiBest, v)
					ev.HiUnused[0], ev.HiUnused[1] = p[t6c2[i][2]], p[t6c2[i][3]] // pocket
					ev.HiUnused[2], ev.HiUnused[3] = p[t6c2[i][4]], p[t6c2[i][5]] // pocket
					ev.HiUnused[4], ev.HiUnused[5] = b[t5c3[j][3]], b[t5c3[j][4]] // board
				}
				if low {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < eightOrBetterMax {
						ev.LoRank = r
						copy(ev.LoBest, v)
						ev.LoUnused[0], ev.LoUnused[1] = p[t6c2[i][2]], p[t6c2[i][3]] // pocket
						ev.LoUnused[2], ev.LoUnused[3] = p[t6c2[i][4]], p[t6c2[i][5]] // pocket
						ev.LoUnused[4], ev.LoUnused[5] = b[t5c3[j][3]], b[t5c3[j][4]] // board
					}
				}
			}
		}
		bestOmaha(ev, normalize, low)
	}
}

// NewSokoEval creates a [Soko] eval func.
func NewSokoEval(normalize, low bool) EvalFunc {
	var f EvalFunc
	if low {
		f = NewSplitEval(RankSoko, RankEightOrBetter, eightOrBetterMax)
	} else {
		f = NewEval(RankSoko)
	}
	return func(ev *Eval, p, b []Card) {
		f(ev, p, b)
		if normalize {
			bestSoko(ev.HiRank, ev.HiBest, ev.HiUnused)
			if low {
				bestAceLow(ev.LoBest)
				bestAceHigh(ev.LoUnused)
			}
		}
	}
}

// NewLowballEval creates a [Lowball] eval func.
func NewLowballEval(normalize bool) EvalFunc {
	f := NewEval(RankLowball)
	return func(ev *Eval, p, b []Card) {
		f(ev, p, b)
		if normalize {
			bestAceHigh(ev.HiBest)
			bestAceHigh(ev.HiUnused)
		}
	}
}

// NewRazzEval creates a [Razz] eval func.
func NewRazzEval(normalize bool) EvalFunc {
	f := NewEval(RankRazz)
	return func(ev *Eval, p, b []Card) {
		f(ev, p, b)
		if normalize {
			if ev.HiRank < aceFiveMax {
				bestAceLow(ev.HiBest)
			} else {
				switch (Invalid - ev.HiRank).Fixed() {
				case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
					bestSet(ev.HiBest)
				}
			}
			bestAceHigh(ev.HiUnused)
		}
	}
}

// NewBadugiEval creates a [Badugi] eval func.
//
//	4 cards, low evaluation of separate suits
//	All 4 face down pre-flop
//	3 rounds of player discards (up to 4)
func NewBadugiEval(normalize bool) EvalFunc {
	return func(ev *Eval, p, _ []Card) {
		s := make([][]Card, 4)
		for i := 0; i < len(p) && i < 4; i++ {
			idx := p[i].SuitIndex()
			s[idx] = append(s[idx], p[i])
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
			return s[i][0].AceRank() < s[j][0].AceRank()
		})
		var best, unused []Card
		count, rank := 4, 0
		for i := 0; i < 4; i++ {
			sort.Slice(s[i], func(j, k int) bool {
				return s[i][j].AceRank() < s[i][k].AceRank()
			})
			captured, r := false, 0
			for j := 0; j < len(s[i]); j++ {
				if r = 1 << s[i][j].AceRank(); rank&r == 0 && !captured {
					captured, best = true, append(best, s[i][j])
					rank |= r
					count--
				} else {
					unused = append(unused, s[i][j])
				}
			}
		}
		if normalize {
			bestAceLow(best)
			bestAceHigh(unused)
		}
		ev.HiRank, ev.HiBest, ev.HiUnused = EvalRank(count<<13|rank), best, unused
	}
}

// Eval contains the eval results of a type's Hi/Lo.
type Eval struct {
	Type     Type
	HiRank   EvalRank
	HiBest   []Card
	HiUnused []Card
	LoRank   EvalRank
	LoBest   []Card
	LoUnused []Card
}

// EvalOf creates a eval for the type.
func EvalOf(typ Type) *Eval {
	return &Eval{
		Type:   typ,
		HiRank: Invalid,
		LoRank: Invalid,
	}
}

// Eval evaluates the pocket, board.
func (ev *Eval) Eval(pocket, board []Card) {
	evals[ev.Type](ev, pocket, board)
}

// Init inits best, unused.
func (ev *Eval) Init(n, m int, low bool) {
	if 0 < n {
		ev.HiBest = make([]Card, n)
	}
	if 0 < m {
		ev.HiUnused = make([]Card, m)
	}
	if low {
		if 0 < n {
			ev.LoBest = make([]Card, n)
		}
		if 0 < m {
			ev.LoUnused = make([]Card, m)
		}
	}
}

// Comp compares the eval's Hi/Lo to b's Hi/Lo.
func (ev *Eval) Comp(b *Eval, low bool) int {
	switch {
	case ev == nil && b == nil:
		return -1
	case ev == nil:
		return +1
	case b == nil:
		return -1
	case !low && ev.HiRank < b.HiRank:
		return -1
	case !low && b.HiRank < ev.HiRank:
		return +1
	case low && ev.LoRank < b.LoRank:
		return -1
	case low && b.LoRank < ev.LoRank:
		return +1
	}
	return 0
}

// Desc returns a descriptior for the eval's Hi/Lo.
func (ev *Eval) Desc(low bool) *EvalDesc {
	switch {
	case ev == nil:
		return nil
	case !low:
		return &EvalDesc{
			Type:   ev.Type.Desc().HiDesc,
			Rank:   ev.HiRank,
			Best:   ev.HiBest,
			Unused: ev.HiUnused,
		}
	}
	return &EvalDesc{
		Type:   ev.Type.Desc().LoDesc,
		Rank:   ev.LoRank,
		Best:   ev.LoBest,
		Unused: ev.LoUnused,
	}
}

// Format satisfies the [fmt.Formatter] interface.
func (ev *Eval) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		fmt.Fprintf(f, "%s %s", ev.Desc(false), ev.HiBest)
	case 'q':
		fmt.Fprintf(f, "\"%s %s\"", ev.Desc(false), ev.HiBest)
	case 'S':
		fmt.Fprintf(f, "%S", ev.Desc(false))
	case 'b':
		fmt.Fprintf(f, "%s %b", ev.Desc(false), ev.HiBest)
	case 'h':
		fmt.Fprintf(f, "%s %h", ev.Desc(false), CardFormatter(ev.HiBest))
	case 'c':
		fmt.Fprintf(f, "%s %c", ev.Desc(false), ev.HiBest)
	case 'C':
		fmt.Fprintf(f, "%s %C", ev.Desc(false), CardFormatter(ev.HiBest))
	case 'f':
		for _, c := range ev.HiBest {
			c.Format(f, 's')
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, Eval<%s>: %s/%s %d)", verb, ev.Type, ev.HiBest, ev.HiUnused, ev.HiRank)
	}
}

// Hi5 evaluates the 5 cards in v, using f.
func (ev *Eval) Hi5(f RankFunc, v []Card) {
	ev.HiRank, ev.HiBest = f(v[0], v[1], v[2], v[3], v[4]), v
}

// HiLo5 evaluates the 5 cards in v, using hi, lo.
func (ev *Eval) HiLo5(hi, lo RankFunc, v []Card, max EvalRank) {
	ev.HiRank, ev.HiBest = hi(v[0], v[1], v[2], v[3], v[4]), v
	if r := lo(v[0], v[1], v[2], v[3], v[4]); r < max {
		ev.LoRank, ev.LoBest = r, v
	}
}

// Max5 evaluates the 5 cards in v, using f, storing only when below max.
func (ev *Eval) Max5(f RankFunc, v []Card, max EvalRank, low bool) {
	if r := f(v[0], v[1], v[2], v[3], v[4]); r < max {
		if !low {
			ev.HiRank, ev.HiBest = r, v
		} else {
			ev.LoRank, ev.LoBest = r, v
		}
	}
}

// Hi6 evaluates the 6 cards in v, using f.
func (ev *Eval) Hi6(f RankFunc, v []Card) {
	ev.HiRank, ev.HiBest, ev.HiUnused = Invalid, make([]Card, 5), make([]Card, 1)
	for i, r := 0, EvalRank(0); i < 6; i++ {
		if r = f(
			v[t6c5[i][0]],
			v[t6c5[i][1]],
			v[t6c5[i][2]],
			v[t6c5[i][3]],
			v[t6c5[i][4]],
		); r < ev.HiRank {
			ev.HiRank = r
			ev.HiBest[0], ev.HiBest[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			ev.HiBest[2], ev.HiBest[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			ev.HiBest[4] = v[t6c5[i][4]]
			ev.HiUnused[0] = v[t6c5[i][5]]
		}
	}
}

// Max6 evaluates the 6 cards in v, using f, storing only when below max.
func (ev *Eval) Max6(f RankFunc, v []Card, max EvalRank, low bool) {
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 1)
	for i, r := 0, EvalRank(0); i < 6; i++ {
		if r = f(
			v[t6c5[i][0]],
			v[t6c5[i][1]],
			v[t6c5[i][2]],
			v[t6c5[i][3]],
			v[t6c5[i][4]],
		); r < rank && r < max {
			rank = r
			best[0], best[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			best[2], best[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			best[4] = v[t6c5[i][4]]
			unused[0] = v[t6c5[i][5]]
		}
	}
	if rank < max {
		if !low {
			ev.HiRank, ev.HiBest, ev.HiUnused = rank, best, unused
		} else {
			ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
		}
	}
}

// HiLo6 evaluates the 6 cards in v, using hi, lo.
func (ev *Eval) HiLo6(hi, lo RankFunc, v []Card, max EvalRank) {
	ev.HiRank, ev.HiBest, ev.HiUnused = Invalid, make([]Card, 5), make([]Card, 1)
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 1)
	for i, r := 0, EvalRank(0); i < 6; i++ {
		if r = hi(
			v[t6c5[i][0]],
			v[t6c5[i][1]],
			v[t6c5[i][2]],
			v[t6c5[i][3]],
			v[t6c5[i][4]],
		); r < ev.HiRank {
			ev.HiRank = r
			ev.HiBest[0], ev.HiBest[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			ev.HiBest[2], ev.HiBest[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			ev.HiBest[4] = v[t6c5[i][4]]
			ev.HiUnused[0] = v[t6c5[i][5]]
		}
		if r = lo(
			v[t6c5[i][0]],
			v[t6c5[i][1]],
			v[t6c5[i][2]],
			v[t6c5[i][3]],
			v[t6c5[i][4]],
		); r < rank && r < max {
			rank = r
			best[0], best[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			best[2], best[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			best[4] = v[t6c5[i][4]]
			unused[0] = v[t6c5[i][5]]
		}
	}
	if rank < max {
		ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
	}
}

// Hi7 evaluates the 7 cards in v, using f.
func (ev *Eval) Hi7(f RankFunc, v []Card) {
	ev.HiRank, ev.HiBest, ev.HiUnused = Invalid, make([]Card, 5), make([]Card, 2)
	for i, r := 0, EvalRank(0); i < 21; i++ {
		if r = f(
			v[t7c5[i][0]],
			v[t7c5[i][1]],
			v[t7c5[i][2]],
			v[t7c5[i][3]],
			v[t7c5[i][4]],
		); r < ev.HiRank {
			ev.HiRank = r
			ev.HiBest[0], ev.HiBest[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			ev.HiBest[2], ev.HiBest[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			ev.HiBest[4] = v[t7c5[i][4]]
			ev.HiUnused[0], ev.HiUnused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
	}
}

// Max7 evaluates the 7 cards in v, using f, storing only when below max.
func (ev *Eval) Max7(f RankFunc, v []Card, max EvalRank, low bool) {
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 2)
	for i, r := 0, EvalRank(0); i < 21; i++ {
		if r = f(
			v[t7c5[i][0]],
			v[t7c5[i][1]],
			v[t7c5[i][2]],
			v[t7c5[i][3]],
			v[t7c5[i][4]],
		); r < rank && r < max {
			rank = r
			best[0], best[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			best[2], best[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			best[4] = v[t7c5[i][4]]
			unused[0], unused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
	}
	if rank < max {
		if !low {
			ev.HiRank, ev.HiBest, ev.HiUnused = rank, best, unused
		} else {
			ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
		}
	}
}

// HiLo7 evaluates the 7 cards in v, using hi, lo.
func (ev *Eval) HiLo7(hi, lo RankFunc, v []Card, max EvalRank) {
	ev.HiRank, ev.HiBest, ev.HiUnused = Invalid, make([]Card, 5), make([]Card, 2)
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 2)
	for i, r := 0, EvalRank(0); i < 21; i++ {
		if r = hi(
			v[t7c5[i][0]],
			v[t7c5[i][1]],
			v[t7c5[i][2]],
			v[t7c5[i][3]],
			v[t7c5[i][4]],
		); r < ev.HiRank {
			ev.HiRank = r
			ev.HiBest[0], ev.HiBest[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			ev.HiBest[2], ev.HiBest[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			ev.HiBest[4] = v[t7c5[i][4]]
			ev.HiUnused[0], ev.HiUnused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
		if r = lo(
			v[t7c5[i][0]],
			v[t7c5[i][1]],
			v[t7c5[i][2]],
			v[t7c5[i][3]],
			v[t7c5[i][4]],
		); r < rank && r < max {
			rank = r
			best[0], best[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			best[2], best[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			best[4] = v[t7c5[i][4]]
			unused[0], unused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
	}
	if rank < max {
		ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
	}
}

// HiLo23 evaluates the 2 cards c0, c1 and the 3 in b, using hi, lo.
func (ev *Eval) HiLo23(hi, lo RankFunc, c0, c1 Card, b []Card, max EvalRank) {
	ev.HiRank, ev.HiBest = hi(c0, c1, b[0], b[1], b[2]), []Card{c0, c1, b[0], b[1], b[2]}
	if lo != nil {
		if r := lo(c0, c1, b[0], b[1], b[2]); r < max {
			ev.LoRank, ev.LoBest = r, ev.HiBest
		}
	}
}

// HiLo24 evaluates the 2 cards c0, c1 and the 4 in b, using hi, lo.
func (ev *Eval) HiLo24(hi, lo RankFunc, c0, c1 Card, b []Card, max EvalRank) {
	ev.HiBest, ev.HiUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 1)
	if lo != nil {
		ev.LoBest, ev.LoUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 1)
	}
	v, r := make([]Card, 3), EvalRank(0)
	for i := 0; i < 4; i++ {
		v[0], v[1], v[2] = b[i%4], b[(i+1)%4], b[(i+2)%4]
		if r = hi(c0, c1, v[0], v[1], v[2]); r < ev.HiRank {
			ev.HiRank = r
			copy(ev.HiBest[2:], v)
			ev.HiUnused[0] = b[(i+3)%4]
		}
		if lo != nil {
			if r = lo(c0, c1, v[0], v[1], v[2]); r < ev.LoRank && r < max {
				ev.LoRank = r
				copy(ev.LoBest[2:], v)
				ev.LoUnused[0] = b[(i+3)%4]
			}
		}
	}
}

// HiLo25 evaluates the 2 cards c0, c1 and the 5 in b, using hi, lo.
func (ev *Eval) HiLo25(hi, lo RankFunc, c0, c1 Card, b []Card, max EvalRank) {
	ev.HiBest, ev.HiUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 2)
	if lo != nil {
		ev.LoBest, ev.LoUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 2)
	}
	v, r := make([]Card, 3), EvalRank(0)
	for i := 0; i < 10; i++ {
		v[0], v[1], v[2] = b[t5c3[i][0]], b[t5c3[i][1]], b[t5c3[i][2]]
		if r = hi(c0, c1, v[0], v[1], v[2]); r < ev.HiRank {
			ev.HiRank = r
			copy(ev.HiBest[2:], v)
			ev.HiUnused[0], ev.HiUnused[1] = b[t5c3[i][3]], b[t5c3[i][4]]
		}
		if lo != nil {
			if r = lo(c0, c1, v[0], v[1], v[2]); r < ev.LoRank && r < max {
				ev.LoRank = r
				copy(ev.LoBest[2:], v)
				ev.LoUnused[0], ev.LoUnused[1] = b[t5c3[i][3]], b[t5c3[i][4]]
			}
		}
	}
}

// EvalDesc describes a Hi/Lo eval.
type EvalDesc struct {
	Type   DescType
	Rank   EvalRank
	Best   []Card
	Unused []Card
}

// Format satisfies the [fmt.Stringer] interface.
func (desc *EvalDesc) Format(f fmt.State, verb rune) {
	desc.Type.Desc(f, verb, desc.Rank, desc.Best, desc.Unused)
}

// Order builds an ordered slice of indices for the provided evals, ordered by
// either Hi or Lo (per [Eval.Comp]), returning the slice of indices and a
// pivot into the indices indicating the winning vs losing position.
//
// Pivot will always be 1 or higher when ordering by Hi's. When ordering by
// Lo's, if there are no valid (ie, qualified) evals, the returned pivot will
// be 0.
func Order(evs []*Eval, low bool) ([]int, int) {
	if len(evs) == 0 {
		return nil, 0
	}
	n := len(evs)
	i, m, v := 0, make(map[int]*Eval, n), make([]int, n)
	// set up
	for ; i < n; i++ {
		m[i], v[i] = evs[i], i
	}
	// sort v based on mapped evals
	sort.SliceStable(v, func(j, k int) bool {
		return m[v[j]].Comp(m[v[k]], low) < 0
	})
	if !low {
		// determine hi pivot
		for i = 1; i < n && m[v[i-1]] != nil && m[v[i]] != nil && m[v[i-1]].HiRank == m[v[i]].HiRank; i++ {
		}
	} else {
		// determine if any qualified low evals
		if m[v[0]] == nil || m[v[0]].LoRank == 0 || m[v[0]].LoRank == Invalid {
			return nil, 0
		}
		// determine lo pivot
		for i = 1; i < n && m[v[i-1]] != nil && m[v[i]] != nil && m[v[i-1]].LoRank == m[v[i]].LoRank; i++ {
		}
	}
	return v, i
}

// bestCactus orders the best and unused cards in v and u, with the specified
// straight base, and inv func to inverse the passed eval rank.
func bestCactus(rank EvalRank, v, u []Card, base Rank, inv func(EvalRank) EvalRank) {
	if inv != nil {
		rank = inv(rank)
	}
	bestAceHigh(v)
	switch rank.Fixed() {
	case StraightFlush:
		bestStraight(v, base)
	case Straight:
		bestStraight(v, base)
		suitNormalize(v, u)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		bestSet(v)
		suitNormalize(v, u)
	case Nothing:
		suitNormalize(v, u)
	}
	bestAceHigh(u)
}

// bestCactusSplit returns the best and unused cards in v.
func bestCactusSplit(rank EvalRank, v []Card, base Rank) ([]Card, []Card) {
	bestAceHigh(v)
	switch rank.Fixed() {
	case StraightFlush:
		bestStraightFlush(v, base)
	case Flush:
		bestFlush(v)
	case Straight:
		bestStraight(v, base)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		bestSet(v)
	case Nothing:
	}
	bestAceHigh(v[5:])
	return v[:5], v[5:]
}

// bestAceHigh orders v by rank, high to low, Aces are high.
func bestAceHigh(v []Card) {
	sort.Slice(v, func(i, j int) bool {
		m, n := v[i].Rank(), v[j].Rank()
		if m == n {
			return v[j].Suit() < v[i].Suit()
		}
		return n < m
	})
}

// bestAceLow orders v by rank, high to low, Aces are low.
func bestAceLow(v []Card) {
	sort.Slice(v, func(i, j int) bool {
		if a, b := v[i].AceRank(), v[j].AceRank(); a != b {
			return b < a
		}
		return v[i].Suit() < v[j].Suit()
	})
}

// bestOmaha sets the best Omaha on the eval.
func bestOmaha(ev *Eval, normalize, low bool) {
	if !normalize {
		return
	}
	bestCactus(ev.HiRank, ev.HiBest, nil, 0, nil)
	bestAceHigh(ev.HiUnused)
	switch {
	case low && ev.LoRank < eightOrBetterMax:
		bestAceLow(ev.LoBest)
		bestAceHigh(ev.LoUnused)
	case low:
		ev.LoBest, ev.LoUnused = nil, nil
	}
}

// bestSoko sets the best Soko in v.
func bestSoko(rank EvalRank, v, u []Card) {
	switch {
	case rank <= TwoPair:
		bestCactus(rank, v, u, 0, nil)
	case rank <= sokoFlush:
		suit := v[0].Suit()
		for i := 0; i < 4; i++ {
			if v[i].Suit() != suit {
				v[i], v[i+1] = v[i+1], v[i]
			}
		}
		if v[0].Suit() != v[1].Suit() {
			c := v[0]
			copy(v, v[1:])
			v[4] = c
		}
		sort.Slice(v[:4], func(i, j int) bool {
			return v[i].Rank() > v[j].Rank()
		})
		bestAceHigh(u)
	case rank <= sokoStraight:
		bestAceHigh(v)
		for i, r := 0, v[0].Rank()+1; i < 4; i++ {
			if v[i].Rank() != r-1 {
				v[i], v[i+1] = v[i+1], v[i]
			}
			r = v[i].Rank()
		}
		bestAceHigh(u)
	default:
		bestCactus(rank-sokoStraight+TwoPair, v, u, 0, nil)
	}
}

// bestStraightFlush sorts v by best straight flush.
func bestStraightFlush(v []Card, base Rank) {
	s := orderSuits(v)
	var b, d []Card
	for _, c := range v {
		switch c.Suit() {
		case s[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	bestStraight(b, base)
	copy(v, b)
	copy(v[len(b):], d)
}

// bestFlush sorts v by best flush.
func bestFlush(v []Card) {
	suits := orderSuits(v)
	var b, d []Card
	for _, c := range v {
		switch c.Suit() {
		case suits[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	copy(v, b)
	copy(v[len(b):], d)
}

// bestStraight sorts v by best-5 straight.
func bestStraight(v []Card, base Rank) {
	m := make(map[Rank][]Card)
	for _, c := range v {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	b := make([]Card, 5)
	for h, i, j, k, l := Ace, King, Queen, Jack, Ten; base+Five <= h; h, i, j, k, l = h-1, i-1, j-1, k-1, l-1 {
		if l == base-1 {
			l = Ace
		}
		if m[h] != nil && m[i] != nil && m[j] != nil && m[k] != nil && m[l] != nil {
			b[0], b[1], b[2], b[3], b[4] = m[h][0], m[i][0], m[j][0], m[k][0], m[l][0]
			m[h], m[i], m[j], m[k], m[l] = m[h][1:], m[i][1:], m[j][1:], m[k][1:], m[l][1:]
			break
		}
		if l == base-1 {
			break
		}
	}
	copy(v, b)
	// collect remaining
	var d []Card
	for i := Ace; i != 255; i-- {
		if _, ok := m[i]; ok && m[i] != nil {
			d = append(d, m[i]...)
		}
	}
	copy(v[5:], d)
}

// bestSet sorts v by best matching sets in v.
func bestSet(v []Card) {
	m := make(map[Rank][]Card)
	for _, c := range v {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	var ranks []Rank
	for rank := range m {
		ranks = append(ranks, rank)
	}
	sort.Slice(ranks, func(i, j int) bool {
		n, m := len(m[ranks[i]]), len(m[ranks[j]])
		if n == m {
			return ranks[j] < ranks[i]
		}
		return m < n
	})
	var i int
	for _, rank := range ranks {
		sort.Slice(m[rank], func(i, j int) bool {
			return m[rank][j].Suit() < m[rank][i].Suit()
		})
		copy(v[i:], m[rank])
		i += len(m[rank])
	}
	i = 5
	j, k := len(m[ranks[0]]), len(m[ranks[1]])
	switch {
	case j == 4:
		i = 4
	case j == 3 && k == 1:
	case j == 2 && k == 2:
		i = 4
	case j == 2:
		i = 2
	}
	bestAceHigh(v[i:])
}

// orderSuits orders v's card suits by count.
func orderSuits(v []Card) []Suit {
	m := make(map[Suit]int)
	var suits []Suit
	for _, c := range v {
		s := c.Suit()
		if _, ok := m[s]; !ok {
			suits = append(suits, s)
		}
		m[s]++
	}
	sort.Slice(suits, func(i, j int) bool {
		if m[suits[i]] == m[suits[j]] {
			return suits[j] < suits[i]
		}
		return m[suits[j]] < m[suits[i]]
	})
	return suits
}

func suitNormalize(v, u []Card) {
	m := make(map[Rank][]Card)
	for _, c := range v {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	for _, c := range u {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	for k := range m {
		sort.Slice(m[k], func(i, j int) bool {
			return m[k][j].Suit() < m[k][i].Suit()
		})
	}
	for i, c := range v {
		r := c.Rank()
		v[i], m[r] = m[r][0], m[r][1:]
	}
	for i, c := range u {
		r := c.Rank()
		u[i], m[r] = m[r][0], m[r][1:]
	}
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

// t6c5 is used for taking 6, choosing 5.
var t6c5 = [6][6]uint8{
	{0, 1, 2, 3, 4, 5},
	{0, 1, 2, 3, 5, 4},
	{0, 1, 2, 4, 5, 3},
	{0, 1, 3, 4, 5, 2},
	{0, 2, 3, 4, 5, 1},
	{1, 2, 3, 4, 5, 0},
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

package cardrank

import (
	"cmp"
	"fmt"
	"slices"
)

// EvalRank is an eval rank.
//
// Ranks are ordered low-to-high.
type EvalRank uint16

// Eval ranks.
//
// See: https://archive.is/G6GZg
const (
	RoyalFlush    EvalRank = 1
	StraightFlush EvalRank = 10
	FourOfAKind   EvalRank = 166
	FullHouse     EvalRank = 322
	Flush         EvalRank = 1599
	Straight      EvalRank = 1609
	ThreeOfAKind  EvalRank = 2467
	TwoPair       EvalRank = 3325
	Pair          EvalRank = 6185
	Nothing       EvalRank = 7462
	HighCard      EvalRank = Nothing

	JacksOrBetter     EvalRank = 4205
	EightOrBetter     EvalRank = 512
	AceFiveLow        EvalRank = 16384
	FlushUnder        EvalRank = 156
	FlushOver         EvalRank = 1277
	LowballAceFlush   EvalRank = 811
	LowballAceNothing EvalRank = 6678

	SokoFlush    EvalRank = TwoPair + 13*715
	SokoStraight EvalRank = SokoFlush + 13*10
	SokoNothing  EvalRank = SokoStraight + (Nothing - TwoPair)

	AceHigh   EvalRank = 6678
	KingHigh  EvalRank = 7007
	QueenHigh EvalRank = 7216
	JackHigh  EvalRank = 7341
	TenHigh   EvalRank = 7410
	NineHigh  EvalRank = 7444
	EightHigh EvalRank = 7458
	SevenHigh EvalRank = 7462 // seven, five, four, three, two all same as Nothing
	SixHigh   EvalRank = 7461
	FiveHigh  EvalRank = 7462
	FourHigh  EvalRank = 7462
	ThreeHigh EvalRank = 7462
	TwoHigh   EvalRank = 7462

	Invalid EvalRank = ^EvalRank(0)
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

// Name returns the eval rank name.
//
// Examples:
//
//	StraightFlush
//	FourOfAKind
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

// Title returns the eval rank title.
//
// Examples:
//
//	Straight Flush
//	Four of a Kind
func (r EvalRank) Title() string {
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

// Format satisfies the [fmt.Formatter] interface.
func (r EvalRank) Format(f fmt.State, verb rune) {
	switch verb {
	case 'd':
		fmt.Fprint(f, uint16(r))
	case 'e', 's', 'S', 'v':
		fmt.Fprint(f, r.Title())
	case 'n':
		fmt.Fprint(f, r.Name())
	}
}

// ToFlushOver changes a Cactus rank to a [FlushOver] a Full House rank.
//
//	FullHouse: FullHouse(322) - FourOfAKind(166) == 156
//	Flush:     Flush(1599)    - FullHouse(322)   == 1277
func (r EvalRank) ToFlushOver() EvalRank {
	switch {
	case FourOfAKind < r && r <= FullHouse:
		return r + FlushOver
	case FullHouse < r && r <= Flush:
		return r - FlushUnder
	}
	return r
}

// FromFlushOver changes a rank from a [FlushOver] a Full House rank to a Cactus
// rank.
//
//	FullHouse: FullHouse(322) - FourOfAKind(166) == 156
//	Flush:     Flush(1599)    - FullHouse(322)   == 1277
func (r EvalRank) FromFlushOver() EvalRank {
	switch {
	case FourOfAKind < r && r <= FourOfAKind+FlushOver:
		return r + FlushUnder
	case FourOfAKind+FlushOver < r && r <= Flush:
		return r - FlushOver
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
		r = LowballAceFlush
	case StraightFlush < r && r <= LowballAceFlush:
		// move everything between 11 and 811 down 1
		r--
	case r == Straight:
		// change lowest ace straight to lowest ace high nothing
		r = LowballAceNothing
	case Straight < r && r <= LowballAceNothing:
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
	case r == LowballAceFlush:
		// change lowest lowest ace high flush to lowest straight flush
		r = StraightFlush
	case StraightFlush <= r && r < LowballAceFlush:
		// move everything between 10 and 812 up 1
		r++
	case r == LowballAceNothing:
		// change lowest ace high nothing to lowest ace straight
		r = Straight
	case Straight <= r && r < LowballAceNothing:
		// move everything between 1609 and 6679 up 1
		r++
	}
	return r
}

// RankFunc returns the eval rank of 5 cards.
type RankFunc func(c0, c1, c2, c3, c4 Card) EvalRank

// RankAceFiveLow is an A-to-5 low rank eval func. [Ace] cards are low,
// [Straight] and [Flush] hands do not count.
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

// RankEightOrBetter is an 8-or-better low rank eval func. [Ace] cards are low,
// [Straight] and [Flush] hands do not count.
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

// RankRazz is a [Razz] (A-to-5) low rank eval func. [Ace] cards are low,
// [Straight] and [Flush] hands do not count.
//
// When there is a [Pair] (or higher) of matching ranks, will be the inverted
// Cactus value.
func RankRazz(c0, c1, c2, c3, c4 Card) EvalRank {
	if r := RankAceFiveLow(0, c0, c1, c2, c3, c4); r < AceFiveLow {
		return r
	}
	return Invalid - RankCactus(c0, c1, c2, c3, c4)
}

// RankLowball is a [Lowball] (2-to-7) low rank eval func. [Ace] cards are high,
// [Straight] and [Flush] hands count.
//
// Works by adding 2 additional ranks for [Ace]-high [StraightFlush] and
// [Straight] hands.
//
// See [EvalRank.ToLowball].
func RankLowball(c0, c1, c2, c3, c4 Card) EvalRank {
	return RankCactus(c0, c1, c2, c3, c4).ToLowball()
}

// EvalFunc is an eval func.
type EvalFunc func(*Eval, []Card, []Card)

// NewEval returns an eval func that ranks 5, 6, or 7 cards using f. The
// returned eval func will store the results on an eval's Hi.
func NewEval(f RankFunc) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, []Card)
		np, nb := len(p), len(b)
		switch np + nb {
		default:
			return
		case 5:
			eval = ev.Hi5
		case 6:
			eval = ev.Hi6
		case 7:
			eval = ev.Hi7
		}
		v := make([]Card, np+nb)
		copy(v, p)
		copy(v[np:], b)
		eval(f, v)
	}
}

// NewMaxEval returns an eval func that ranks 5, 6, or 7 cards using f and max.
//
// The returned eval func will store results on an eval's Hi only when lower
// than max.
func NewMaxEval(f RankFunc, maximum EvalRank, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, []Card, EvalRank, bool)
		np, nb := len(p), len(b)
		switch np + nb {
		default:
			return
		case 5:
			eval = ev.Max5
		case 6:
			eval = ev.Max6
		case 7:
			eval = ev.Max7
		}
		v := make([]Card, np+nb)
		copy(v, p)
		copy(v[np:], b)
		eval(f, v, maximum, low)
	}
}

// NewSplitEval returns an eval func that ranks 5, 6, or 7 cards using hi, lo
// and max.
//
// The returned eval func will store results on an eval's Hi and Lo depending
// on the result of hi and lo, respectively. Will store the Lo value only when
// lower than max.
func NewSplitEval(hi, lo RankFunc, maximum EvalRank) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		var eval func(RankFunc, RankFunc, []Card, EvalRank)
		np, nb := len(p), len(b)
		switch np + nb {
		default:
			return
		case 5:
			eval = ev.HiLo5
		case 6:
			eval = ev.HiLo6
		case 7:
			eval = ev.HiLo7
		}
		v := make([]Card, np+nb)
		copy(v, p)
		copy(v[np:], b)
		eval(hi, lo, v, maximum)
	}
}

// NewHybridEval creates a hybrid Cactus and Two-Plus-Two eval func, using
// [RankCactus] for 5 and 6 cards, and the registered Two-Plus-Two eval func
// (see [RegisterTwoPlusTwo]) for 7 cards. Uses [RankCactus] for 7 cards when
// no Two-Plus-Two eval func is registered.
//
// Gives optimal performance when evaluating the best-5 of any 5, 6, or 7 cards
// of a combined pocket and board.
func NewHybridEval(normalize, low bool) EvalFunc {
	return newHybridEval(normalize, low, twoPlusTwo)
}

// newHybridEval creates a hybrid Cactus and Two-Plus-Two eval func, using
// twoPlusTwo (when not nil) for 7 cards.
func newHybridEval(normalize, low bool, twoPlusTwo func([]Card) EvalRank) EvalFunc {
	var f EvalFunc
	if low {
		f = NewSplitEval(RankCactus, RankEightOrBetter, EightOrBetter)
	} else {
		f = NewEval(RankCactus)
	}
	return func(ev *Eval, p, b []Card) {
		switch np, nb := len(p), len(b); {
		case np+nb == 7 && twoPlusTwo != nil:
			v := make([]Card, np+nb)
			copy(v, p)
			copy(v[np:], b)
			ev.HiRank = twoPlusTwo(v)
			if normalize {
				ev.HiBest, ev.HiUnused = bestCactusSplit(ev.HiRank, v, 0)
			}
			if low {
				u := make([]Card, np+nb)
				copy(u, p)
				copy(u[np:], b)
				ev.Max7(RankEightOrBetter, u, EightOrBetter, true)
				if normalize && ev.LoRank < EightOrBetter {
					bestAceLow(ev.LoBest)
					bestAceHigh(ev.LoUnused)
				}
			}
		case 5 <= np+nb && np+nb <= 7:
			f(ev, p, b)
			if normalize {
				bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
				if low && ev.LoRank < EightOrBetter {
					bestAceLow(ev.LoBest)
					bestAceHigh(ev.LoUnused)
				}
			}
		}
	}
}

// NewCactusEval creates a Cactus eval func.
func NewCactusEval(board int, normalize, low bool) EvalFunc {
	var f EvalFunc
	hybrid := twoPlusTwo != nil
	switch {
	case hybrid:
		f = NewHybridEval(normalize, low)
	case low:
		f = NewSplitEval(RankCactus, RankEightOrBetter, EightOrBetter)
	default:
		f = NewEval(RankCactus)
	}
	return func(ev *Eval, p, b []Card) {
		// fewer than 5 cards (ie, pre-flop), use the pocket's starting rank
		if np, nb := len(p), len(b); np < 3 && np+nb < 5 {
			if r := StartingEvalRank(p); r != 0 && r != Invalid {
				ev.HiRank, ev.HiBest = r, make([]Card, 5)
				copy(ev.HiBest, p)
				copy(ev.HiBest[np:], b)
				bestAceHigh(ev.HiBest)
				return
			}
		}
		f(ev, p, b)
		if normalize && !hybrid {
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
		f = NewSplitEval(hi, RankEightOrBetter, EightOrBetter)
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
	hi := NewMaxEval(RankCactus, JacksOrBetter, false)
	return func(ev *Eval, p, b []Card) {
		hi(ev, p, b)
		if normalize {
			bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
		}
	}
}

// NewOmahaEval creates an [Omaha] eval func.
//
// Uses any 2 from 2, 3, 4, 5, or 6 pocket cards, and any 3 from 3, 4 or 5
// board cards to make a best-5.
func NewOmahaEval(hi RankFunc, base Rank, inv func(EvalRank) EvalRank, normalize, low bool) EvalFunc {
	return func(ev *Eval, p, b []Card) {
		np, nb := len(p), len(b)
		switch {
		case np < 2, 6 < np:
			return
		case nb < 3:
			ev.HiRank, ev.HiBest = StartingEvalRank(p), p
			return
		case 5 < nb:
			return
		}
		pockets, boards := omahaPockets[np], omahaBoards[nb]
		// copy the board combinations, as each is used with every pocket
		var buf [10][3]Card
		v := buf[:len(boards)]
		for j, board := range boards {
			v[j] = [3]Card{b[board[0]], b[board[1]], b[board[2]]}
		}
		hi, hiRank, loRank := hi, ev.HiRank, ev.LoRank
		hiPocket, hiBoard, loPocket, loBoard := -1, -1, -1, -1
		for i, pocket := range pockets {
			c0, c1 := p[pocket[0]], p[pocket[1]]
			if !low {
				for j := range v {
					if r := hi(c0, c1, v[j][0], v[j][1], v[j][2]); r < hiRank {
						hiRank, hiPocket, hiBoard = r, i, j
					}
				}
				continue
			}
			for j := range v {
				c2, c3, c4 := v[j][0], v[j][1], v[j][2]
				if r := hi(c0, c1, c2, c3, c4); r < hiRank {
					hiRank, hiPocket, hiBoard = r, i, j
				}
				if r := RankEightOrBetter(c0, c1, c2, c3, c4); r < EightOrBetter && r < loRank {
					loRank, loPocket, loBoard = r, i, j
				}
			}
		}
		ev.HiRank, ev.LoRank = hiRank, loRank
		if hiPocket != -1 {
			ev.HiBest, ev.HiUnused = omahaBest(p, b, pockets[hiPocket], boards[hiBoard])
		} else {
			ev.HiUnused = make([]Card, np+nb-5)
		}
		switch {
		case low && ev.LoRank != Invalid && loPocket != -1:
			ev.LoBest, ev.LoUnused = omahaBest(p, b, pockets[loPocket], boards[loBoard])
		case low && ev.LoRank != Invalid:
			ev.LoBest, ev.LoUnused = nil, make([]Card, np+nb-5)
		}
		if normalize {
			bestCactus(ev.HiRank, ev.HiBest, nil, base, inv)
			bestAceHigh(ev.HiUnused)
			switch {
			case low && ev.LoRank < EightOrBetter:
				bestAceLow(ev.LoBest)
				bestAceHigh(ev.LoUnused)
			case low:
				ev.LoBest, ev.LoUnused = nil, nil
			}
		}
	}
}

// omahaBest returns the best and unused cards for the pocket and board
// combination.
func omahaBest(p, b []Card, pocket, board []uint8) ([]Card, []Card) {
	v := make([]Card, 0, len(p)+len(b))
	v = append(v, p[pocket[0]], p[pocket[1]], b[board[0]], b[board[1]], b[board[2]])
	for _, i := range pocket[2:] {
		v = append(v, p[i])
	}
	for _, i := range board[3:] {
		v = append(v, b[i])
	}
	return v[:5:5], v[5:]
}

// omahaPockets and omahaBoards are the pocket and board combinations (as
// indexes) for Omaha evals, where the first 2 pocket and the first 3 board
// indexes of each combination are used, and the remaining are unused.
var omahaPockets, omahaBoards [7][][]uint8

func init() {
	omahaPockets[2] = [][]uint8{{0, 1}}
	omahaPockets[3] = [][]uint8{{0, 1, 2}, {1, 2, 0}, {2, 0, 1}}
	for i := range t4c2 {
		omahaPockets[4] = append(omahaPockets[4], t4c2[i][:])
	}
	for i := range t5c2 {
		omahaPockets[5] = append(omahaPockets[5], t5c2[i][:])
	}
	for i := range t6c2 {
		omahaPockets[6] = append(omahaPockets[6], t6c2[i][:])
	}
	omahaBoards[3] = [][]uint8{{0, 1, 2}}
	omahaBoards[4] = [][]uint8{{0, 1, 2, 3}, {1, 2, 3, 0}, {2, 3, 0, 1}, {3, 0, 1, 2}}
	for i := range t5c3 {
		omahaBoards[5] = append(omahaBoards[5], t5c3[i][:])
	}
}

// NewSokoEval creates a [Soko] eval func.
func NewSokoEval(normalize, low bool) EvalFunc {
	var f EvalFunc
	if low {
		f = NewSplitEval(RankSoko, RankEightOrBetter, EightOrBetter)
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
			if ev.HiRank < AceFiveLow {
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
		slices.SortStableFunc(s, func(a, b []Card) int {
			if c := cmp.Compare(len(a), len(b)); c != 0 || len(a) == 0 {
				return c
			}
			return cmp.Compare(a[0].AceRank(), b[0].AceRank())
		})
		var best, unused []Card
		count, rank := 4, 0
		for i := range 4 {
			slices.SortFunc(s[i], func(a, b Card) int {
				return cmp.Compare(a.AceRank(), b.AceRank())
			})
			captured, r := false, 0
			for j := range len(s[i]) {
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

// NewHighEval creates a high card eval func.
func NewHighEval() EvalFunc {
	return func(ev *Eval, p, b []Card) {
		np, nb := len(p), len(b)
		if 0 < np+nb && p[0].Rank() <= Ace {
			v := make([]Card, np+nb)
			copy(v, p)
			copy(v[np:], b)
			bestAceHigh(v)
			ev.HiRank = EvalRank(Ace-v[0].Rank()) + 1
			ev.HiBest = []Card{v[0]}
			if 1 < np+nb {
				ev.HiUnused = v[1:]
			}
		}
	}
}

/*
// NewThreeEval creates a best-3 eval func.
//
//	Straight Flush
//	Three of a Kind
//	Straight
//	Flush
//	One Pair
//	High Card
func NewThreeEval() EvalFunc {
	return func(ev *Eval, p, b []Card) {
	}
}
*/

/*
// NewLeducEval creates a matching high card eval func.
func NewLeducEval() EvalFunc {
	return func(ev *Eval, p, b []Card) {
	}
}
*/

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

// EvalOf creates an eval for the type.
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
	case 'd':
		fmt.Fprintf(f, "%d", uint16(ev.HiRank))
	case 's', 'v':
		fmt.Fprintf(f, "%s %s", ev.Desc(false), ev.HiBest)
	case 'e':
		fmt.Fprintf(f, "%s %e", ev.Desc(false), ev.HiBest)
	case 'a':
		fmt.Fprintf(f, "%s %a", ev.Desc(false), Formatter(ev.HiBest))
	case 'q':
		fmt.Fprintf(f, "\"%s %s\"", ev.Desc(false), ev.HiBest)
	case 'S':
		fmt.Fprintf(f, "%S", ev.Desc(false))
	case 'b':
		fmt.Fprintf(f, "%s %b", ev.Desc(false), ev.HiBest)
	case 'h':
		fmt.Fprintf(f, "%s %h", ev.Desc(false), Formatter(ev.HiBest))
	case 'c':
		fmt.Fprintf(f, "%s %c", ev.Desc(false), ev.HiBest)
	case 'C':
		fmt.Fprintf(f, "%s %C", ev.Desc(false), Formatter(ev.HiBest))
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
func (ev *Eval) HiLo5(hi, lo RankFunc, v []Card, maximum EvalRank) {
	ev.HiRank, ev.HiBest = hi(v[0], v[1], v[2], v[3], v[4]), v
	if r := lo(v[0], v[1], v[2], v[3], v[4]); r < maximum {
		// copy, as Hi and Lo best are ordered separately
		ev.LoRank, ev.LoBest = r, slices.Clone(v)
	}
}

// Max5 evaluates the 5 cards in v, using f, storing only when below max.
func (ev *Eval) Max5(f RankFunc, v []Card, maximum EvalRank, low bool) {
	if r := f(v[0], v[1], v[2], v[3], v[4]); r < maximum {
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
func (ev *Eval) Max6(f RankFunc, v []Card, maximum EvalRank, low bool) {
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 1)
	for i, r := 0, EvalRank(0); i < 6; i++ {
		if r = f(
			v[t6c5[i][0]],
			v[t6c5[i][1]],
			v[t6c5[i][2]],
			v[t6c5[i][3]],
			v[t6c5[i][4]],
		); r < rank && r < maximum {
			rank = r
			best[0], best[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			best[2], best[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			best[4] = v[t6c5[i][4]]
			unused[0] = v[t6c5[i][5]]
		}
	}
	if rank < maximum {
		if !low {
			ev.HiRank, ev.HiBest, ev.HiUnused = rank, best, unused
		} else {
			ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
		}
	}
}

// HiLo6 evaluates the 6 cards in v, using hi, lo.
func (ev *Eval) HiLo6(hi, lo RankFunc, v []Card, maximum EvalRank) {
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
		); r < rank && r < maximum {
			rank = r
			best[0], best[1] = v[t6c5[i][0]], v[t6c5[i][1]]
			best[2], best[3] = v[t6c5[i][2]], v[t6c5[i][3]]
			best[4] = v[t6c5[i][4]]
			unused[0] = v[t6c5[i][5]]
		}
	}
	if rank < maximum {
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
func (ev *Eval) Max7(f RankFunc, v []Card, maximum EvalRank, low bool) {
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 2)
	for i, r := 0, EvalRank(0); i < 21; i++ {
		if r = f(
			v[t7c5[i][0]],
			v[t7c5[i][1]],
			v[t7c5[i][2]],
			v[t7c5[i][3]],
			v[t7c5[i][4]],
		); r < rank && r < maximum {
			rank = r
			best[0], best[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			best[2], best[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			best[4] = v[t7c5[i][4]]
			unused[0], unused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
	}
	if rank < maximum {
		if !low {
			ev.HiRank, ev.HiBest, ev.HiUnused = rank, best, unused
		} else {
			ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
		}
	}
}

// HiLo7 evaluates the 7 cards in v, using hi, lo.
func (ev *Eval) HiLo7(hi, lo RankFunc, v []Card, maximum EvalRank) {
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
		); r < rank && r < maximum {
			rank = r
			best[0], best[1] = v[t7c5[i][0]], v[t7c5[i][1]]
			best[2], best[3] = v[t7c5[i][2]], v[t7c5[i][3]]
			best[4] = v[t7c5[i][4]]
			unused[0], unused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
		}
	}
	if rank < maximum {
		ev.LoRank, ev.LoBest, ev.LoUnused = rank, best, unused
	}
}

// HiLo23 evaluates the 2 cards c0, c1 and the 3 in b, using hi, lo.
func (ev *Eval) HiLo23(hi, lo RankFunc, c0, c1 Card, b []Card, maximum EvalRank) {
	ev.HiRank, ev.HiBest = hi(c0, c1, b[0], b[1], b[2]), []Card{c0, c1, b[0], b[1], b[2]}
	if lo != nil {
		if r := lo(c0, c1, b[0], b[1], b[2]); r < maximum {
			ev.LoRank, ev.LoBest = r, ev.HiBest
		}
	}
}

// HiLo24 evaluates the 2 cards c0, c1 and the 4 in b, using hi, lo.
func (ev *Eval) HiLo24(hi, lo RankFunc, c0, c1 Card, b []Card, maximum EvalRank) {
	ev.HiBest, ev.HiUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 1)
	if lo != nil {
		ev.LoBest, ev.LoUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 1)
	}
	v, r := make([]Card, 3), EvalRank(0)
	for i := range 4 {
		v[0], v[1], v[2] = b[i], b[(i+1)%4], b[(i+2)%4]
		if r = hi(c0, c1, v[0], v[1], v[2]); r < ev.HiRank {
			ev.HiRank = r
			copy(ev.HiBest[2:], v)
			ev.HiUnused[0] = b[(i+3)%4]
		}
		if lo != nil {
			if r = lo(c0, c1, v[0], v[1], v[2]); r < ev.LoRank && r < maximum {
				ev.LoRank = r
				copy(ev.LoBest[2:], v)
				ev.LoUnused[0] = b[(i+3)%4]
			}
		}
	}
}

// HiLo25 evaluates the 2 cards c0, c1 and the 5 in b, using hi, lo.
func (ev *Eval) HiLo25(hi, lo RankFunc, c0, c1 Card, b []Card, maximum EvalRank) {
	ev.HiBest, ev.HiUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 2)
	if lo != nil {
		ev.LoBest, ev.LoUnused = []Card{c0, c1, 0, 0, 0}, make([]Card, 2)
	}
	v, r := make([]Card, 3), EvalRank(0)
	for i := range 10 {
		v[0], v[1], v[2] = b[t5c3[i][0]], b[t5c3[i][1]], b[t5c3[i][2]]
		if r = hi(c0, c1, v[0], v[1], v[2]); r < ev.HiRank {
			ev.HiRank = r
			copy(ev.HiBest[2:], v)
			ev.HiUnused[0], ev.HiUnused[1] = b[t5c3[i][3]], b[t5c3[i][4]]
		}
		if lo != nil {
			if r = lo(c0, c1, v[0], v[1], v[2]); r < ev.LoRank && r < maximum {
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
	slices.SortStableFunc(v, func(a, b int) int {
		return m[a].Comp(m[b], low)
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
		normalizeSuits(v, u)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		bestSet(v)
		normalizeSuits(v, u)
	case Nothing:
		normalizeSuits(v, u)
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

// bestAceHigh orders v by rank, high to low, Aces are high, and then by suit.
func bestAceHigh(v []Card) {
	sortDesc(v)
}

// sortDesc sorts v in descending order. As a card's rank bits are higher than
// its suit bits, cards are ordered by rank and then by suit.
func sortDesc(v []Card) {
	if 12 < len(v) {
		slices.SortFunc(v, func(a, b Card) int {
			return cmp.Compare(b, a)
		})
		return
	}
	// insertion sort is faster for the few cards in a hand
	for i := 1; i < len(v); i++ {
		for j := i; 0 < j && v[j-1] < v[j]; j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
}

// bestAceLow orders v by rank, high to low, Aces are low.
func bestAceLow(v []Card) {
	slices.SortFunc(v, func(a, b Card) int {
		return cmp.Or(cmp.Compare(b.AceRank(), a.AceRank()), cmp.Compare(a.Suit(), b.Suit()))
	})
}

// bestSoko sets the best Soko in v.
func bestSoko(rank EvalRank, v, u []Card) {
	switch {
	case rank <= TwoPair:
		bestCactus(rank, v, u, 0, nil)
	case rank <= SokoFlush:
		suit := v[0].Suit()
		for i := range 4 {
			if v[i].Suit() != suit {
				v[i], v[i+1] = v[i+1], v[i]
			}
		}
		if v[0].Suit() != v[1].Suit() {
			c := v[0]
			copy(v, v[1:])
			v[4] = c
		}
		slices.SortFunc(v[:4], func(a, b Card) int {
			return cmp.Compare(b.Rank(), a.Rank())
		})
		bestAceHigh(u)
	case rank <= SokoStraight:
		bestSokoStraight(v, 0)
		bestAceHigh(u)
		normalizeSuits(v, u)
	default:
		bestCactus(rank-SokoStraight+TwoPair, v, u, 0, nil)
	}
}

// bestStraightFlush sorts v by best straight flush.
func bestStraightFlush(v []Card, base Rank) {
	n := partitionSuit(v, flushSuit(v))
	bestStraight(v[:n], base)
}

// bestFlush sorts v by best flush.
func bestFlush(v []Card) {
	partitionSuit(v, flushSuit(v))
}

// bestStraight sorts v by best-5 straight, followed by the remaining cards
// ordered by rank.
func bestStraight(v []Card, base Rank) {
	// first is the index + 1 of the first card of each rank in v
	var first [16]int
	var ranks uint16
	for i := len(v) - 1; 0 <= i; i-- {
		r := v[i].Rank()
		first[r], ranks = i+1, ranks|1<<r
	}
	has := func(r Rank) bool {
		return ranks&(1<<r) != 0
	}
	var buf [16]Card
	var used uint64
	b := buf[:0]
	for h, i, j, k, l := Ace, King, Queen, Jack, Ten; base+Five <= h; h, i, j, k, l = h-1, i-1, j-1, k-1, l-1 {
		if l == base-1 {
			l = Ace
		}
		if has(h) && has(i) && has(j) && has(k) && has(l) {
			for _, r := range [...]Rank{h, i, j, k, l} {
				b, used = append(b, v[first[r]-1]), used|1<<(first[r]-1)
			}
			break
		}
		if l == base-1 {
			break
		}
	}
	// collect remaining
	for r := Ace; r != InvalidRank; r-- {
		for i, c := range v {
			if c.Rank() == r && used&(1<<i) == 0 {
				b = append(b, c)
			}
		}
	}
	copy(v, b)
}

// bestSokoStraight sorts v by best-4 straight.
func bestSokoStraight(v []Card, base Rank) {
	m := make(map[Rank][]Card)
	for _, c := range v {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	b := make([]Card, 4)
	for h, i, j, k := Ace, King, Queen, Jack; base+Four <= h; h, i, j, k = h-1, i-1, j-1, k-1 {
		if k == base-1 {
			k = Ace
		}
		if m[h] != nil && m[i] != nil && m[j] != nil && m[k] != nil {
			b[0], b[1], b[2], b[3] = m[h][0], m[i][0], m[j][0], m[k][0]
			m[h], m[i], m[j], m[k] = m[h][1:], m[i][1:], m[j][1:], m[k][1:]
			break
		}
		if k == base-1 {
			break
		}
	}
	copy(v, b)
	// collect remaining
	var d []Card
	for i := Ace; i != InvalidRank; i-- {
		if _, ok := m[i]; ok && m[i] != nil {
			d = append(d, m[i]...)
		}
	}
	copy(v[4:], d)
}

// bestSet sorts v by best matching sets in v.
func bestSet(v []Card) {
	var count [16]int
	for _, c := range v {
		count[c.Rank()]++
	}
	// order by set size, and then by rank and suit
	key := func(c Card) uint64 {
		return uint64(count[c.Rank()])<<32 | uint64(c)
	}
	for i := 1; i < len(v); i++ {
		for j := i; 0 < j && key(v[j-1]) < key(v[j]); j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
	// j and k are the sizes of the 2 largest sets
	j, k := count[v[0].Rank()], 0
	if j < len(v) {
		k = count[v[j].Rank()]
	}
	i := 5
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

// flushSuit returns the most common suit in v, preferring higher suits.
func flushSuit(v []Card) Suit {
	var count [16]int
	for _, c := range v {
		count[c.Suit()]++
	}
	var suit Suit
	for s := len(count) - 1; 0 <= s; s-- {
		if count[s] > count[suit] {
			suit = Suit(s)
		}
	}
	return suit
}

// partitionSuit moves the cards in v having suit to the front of v,
// preserving their order, returning the number of cards having suit.
func partitionSuit(v []Card, suit Suit) int {
	var buf [16]Card
	b := buf[:0]
	for _, c := range v {
		if c.Suit() == suit {
			b = append(b, c)
		}
	}
	n := len(b)
	for _, c := range v {
		if c.Suit() != suit {
			b = append(b, c)
		}
	}
	copy(v, b)
	return n
}

// normalizeSuits normalizes the suits in v, swapping cards from v of the same
// rank in u, such that the cards of each rank are ordered by suit across v
// and u.
func normalizeSuits(v, u []Card) {
	n := len(v) + len(u)
	at := func(i int) *Card {
		if i < len(v) {
			return &v[i]
		}
		return &u[i-len(v)]
	}
	var idxBuf [16]int
	var cardBuf [16]Card
	for i := range n {
		r := at(i).Rank()
		// skip ranks already normalized
		seen := false
		for j := 0; j < i && !seen; j++ {
			seen = at(j).Rank() == r
		}
		if seen {
			continue
		}
		idx, cards := idxBuf[:0], cardBuf[:0]
		for j := i; j < n; j++ {
			if c := *at(j); c.Rank() == r {
				idx, cards = append(idx, j), append(cards, c)
			}
		}
		if len(idx) == 1 {
			continue
		}
		// same rank cards, so ordered by suit
		sortDesc(cards)
		for k, j := range idx {
			*at(j) = cards[k]
		}
	}
}

// take2c2 generates the combinations of v.
func take2c2(v []Card) ([][]Card, int) {
	return [][]Card{v}, 1
}

// take3c2 generates the combinations of v.
func take3c2(v []Card) ([][]Card, int) {
	u := make([][]Card, 3)
	for i := range 3 {
		u[i] = []Card{v[i], v[(i+1)%3], v[(i+2)%3]}
	}
	return u, 3
}

// take4c2 generates the combinations of v.
func take4c2(v []Card) ([][]Card, int) {
	u := make([][]Card, 6)
	for i := range 6 {
		u[i] = []Card{
			v[t4c2[i][0]],
			v[t4c2[i][1]],
			v[t4c2[i][2]],
			v[t4c2[i][3]],
		}
	}
	return u, 6
}

// take5c2 generates the combinations of v.
func take5c2(v []Card) ([][]Card, int) {
	u := make([][]Card, 10)
	for i := range 10 {
		u[i] = []Card{
			v[t5c2[i][0]],
			v[t5c2[i][1]],
			v[t5c2[i][2]],
			v[t5c2[i][3]],
			v[t5c2[i][4]],
		}
	}
	return u, 10
}

// take6c2 generates the combinations of v.
func take6c2(v []Card) ([][]Card, int) {
	u := make([][]Card, 15)
	for i := range 15 {
		u[i] = []Card{
			v[t6c2[i][0]],
			v[t6c2[i][1]],
			v[t6c2[i][2]],
			v[t6c2[i][3]],
			v[t6c2[i][4]],
			v[t6c2[i][5]],
		}
	}
	return u, 15
}

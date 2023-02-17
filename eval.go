package cardrank

import (
	"fmt"
	"sort"
)

// EvalRank is a poker eval rank.
//
// Ranks are ordered low-to-high.
type EvalRank uint16

// Poker eval rank values.
//
// See: https://archive.is/G6GZg
const (
	StraightFlush        EvalRank = 10
	FourOfAKind          EvalRank = 166
	FullHouse            EvalRank = 322
	Flush                EvalRank = 1599
	Straight             EvalRank = 1609
	ThreeOfAKind         EvalRank = 2467
	TwoPair              EvalRank = 3325
	Pair                 EvalRank = 6185
	Nothing              EvalRank = 7462
	HighCard             EvalRank = Nothing
	Invalid                       = ^EvalRank(0)
	rankMax                       = Nothing + 1
	rankEightOrBetterMax EvalRank = 512
	rankAceFiveMax       EvalRank = 16384
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

// String satisfies the fmt.Stringer interface.
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

// RankFunc returns the eval rank of 5 cards.
type RankFunc func(c0, c1, c2, c3, c4 Card) EvalRank

// RankEightOrBetter is a 8-or-better low eval rank func. Aces are low,
// straights and flushes do not count.
func RankEightOrBetter(c0, c1, c2, c3, c4 Card) EvalRank {
	return RankLowAceFive(0xff00, c0, c1, c2, c3, c4)
}

// RankRazz is a Razz (Ace-to-Five) low eval rank func. Aces are low, straights
// and flushes do not count.
//
// When there is a pair (or higher) of matching ranks, will be the inverted
// value of the regular eval rank.
func RankRazz(c0, c1, c2, c3, c4 Card) EvalRank {
	if r := RankLowAceFive(0, c0, c1, c2, c3, c4); r < rankAceFiveMax {
		return r
	}
	return Invalid - DefaultCactus(c0, c1, c2, c3, c4)
}

// RankLowAceFive is a Ace-to-Five low eval rank func.
func RankLowAceFive(mask EvalRank, c0, c1, c2, c3, c4 Card) EvalRank {
	var rank EvalRank
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

// RankLowball is a Two-to-Seven (Ace always high) low eval rank func.
func RankLowball(c0, c1, c2, c3, c4 Card) EvalRank {
	r := DefaultCactus(c0, c1, c2, c3, c4)
	return rankMax - r
}

// CactusFunc returns the eval rank of 5, 6, or 7 cards.
type CactusFunc func([]Card) EvalRank

// NewRankFunc creates a rank eval func for 5, 6, or 7 cards using f.
func NewRankFunc(f RankFunc) CactusFunc {
	return func(v []Card) EvalRank {
		switch n := len(v); {
		case n == 5:
			return f(v[0], v[1], v[2], v[3], v[4])
		case n == 6:
			r := f(v[0], v[1], v[2], v[3], v[4])
			r = min(r, f(v[0], v[1], v[2], v[3], v[5]))
			r = min(r, f(v[0], v[1], v[2], v[4], v[5]))
			r = min(r, f(v[0], v[1], v[3], v[4], v[5]))
			r = min(r, f(v[0], v[2], v[3], v[4], v[5]))
			r = min(r, f(v[1], v[2], v[3], v[4], v[5]))
			return r
		}
		r, rank := EvalRank(0), Invalid
		for i := 0; i < 21; i++ {
			if r = f(
				v[t7c5[i][0]],
				v[t7c5[i][1]],
				v[t7c5[i][2]],
				v[t7c5[i][3]],
				v[t7c5[i][4]],
			); r < rank {
				rank = r
			}
		}
		return rank
	}
}

// NewHybrid creates a hybrid eval rank func using f5 for 5 and 6 cards, and f7
// for 7 cards.
func NewHybrid(f5 RankFunc, f7 CactusFunc) CactusFunc {
	return func(v []Card) EvalRank {
		switch len(v) {
		case 5:
			return f5(v[0], v[1], v[2], v[3], v[4])
		case 6:
			r := f5(v[0], v[1], v[2], v[3], v[4])
			r = min(r, f5(v[0], v[1], v[2], v[3], v[5]))
			r = min(r, f5(v[0], v[1], v[2], v[4], v[5]))
			r = min(r, f5(v[0], v[1], v[3], v[4], v[5]))
			r = min(r, f5(v[0], v[2], v[3], v[4], v[5]))
			r = min(r, f5(v[1], v[2], v[3], v[4], v[5]))
			return r
		}
		return f7(v)
	}
}

// EvalFunc is a rank eval func.
type EvalFunc func(*Eval, []Card, []Card)

// NewCactusEval creates a Cactus rank eval func.
func NewCactusEval(f CactusFunc, straightHigh Rank) EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		v := make([]Card, len(pocket)+len(board))
		copy(v, pocket)
		copy(v[len(pocket):], board)
		ev.HiRank = f(v)
		bestHoldem(ev, v, straightHigh)
	}
}

// NewShortEval creates a Short rank eval func.
func NewShortEval() EvalFunc {
	return NewCactusEval(NewRankFunc(func(c0, c1, c2, c3, c4 Card) EvalRank {
		r := DefaultCactus(c0, c1, c2, c3, c4)
		switch r {
		case 747: // Straight Flush, 9, 8, 7, 6, Ace
			return 6
		case 6610: // Straight, 9, 8, 7, 6, Ace
			return 1605
		}
		return r
	}), Nine)
}

// NewManilaEval creates a Manila rank eval func.
func NewManilaEval() EvalFunc {
	return NewCactusEval(NewRankFunc(func(c0, c1, c2, c3, c4 Card) EvalRank {
		r := DefaultCactus(c0, c1, c2, c3, c4)
		switch r {
		case 691: // Straight Flush, 10, 9, 8, 7, Ace
			return 5
		case 6554: // Straight, 10, 9, 8, 7, Ace
			return 1604
		}
		return r
	}), Ten)
}

// NewOmahaEval creates a Omaha rank eval func.
func NewOmahaEval(loMax EvalRank) EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		ev.Init(5, 4, loMax)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 6; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = pocket[t4c2[i][0]], pocket[t4c2[i][1]] // pocket
				v[2], v[3] = board[t5c3[j][0]], board[t5c3[j][1]]   // board
				v[4] = board[t5c3[j][2]]                            // board
				if r = DefaultEval(v); r < ev.HiRank {
					copy(ev.HiBest, v)
					ev.HiRank = r
					ev.HiUnused[0], ev.HiUnused[1] = pocket[t4c2[i][2]], pocket[t4c2[i][3]] // pocket
					ev.HiUnused[2], ev.HiUnused[3] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < loMax {
						copy(ev.LoBest, v)
						ev.LoRank = r
						ev.LoUnused[0], ev.LoUnused[1] = pocket[t4c2[i][2]], pocket[t4c2[i][3]] // pocket
						ev.LoUnused[2], ev.LoUnused[3] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(ev, loMax)
	}
}

// NewOmahaFiveEval creates a new OmahaFive rank eval func.
func NewOmahaFiveEval(loMax EvalRank) EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		ev.Init(5, 5, loMax)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = pocket[t5c2[i][0]], pocket[t5c2[i][1]] // pocket
				v[2], v[3] = board[t5c3[j][0]], board[t5c3[j][1]]   // board
				v[4] = board[t5c3[j][2]]                            // board
				if r = DefaultEval(v); r < ev.HiRank {
					copy(ev.HiBest, v)
					ev.HiRank = r
					ev.HiUnused[0], ev.HiUnused[1] = pocket[t5c2[i][2]], pocket[t5c2[i][3]] // pocket
					ev.HiUnused[2] = pocket[t5c2[i][4]]                                     // pocket
					ev.HiUnused[3], ev.HiUnused[4] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < loMax {
						copy(ev.LoBest, v)
						ev.LoRank = r
						ev.LoUnused[0], ev.LoUnused[1] = pocket[t5c2[i][2]], pocket[t5c2[i][3]] // pocket
						ev.LoUnused[2] = pocket[t5c2[i][4]]                                     // pocket
						ev.LoUnused[3], ev.LoUnused[4] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(ev, loMax)
	}
}

// NewOmahaSixEval creates a new OmahaFive rank eval func.
func NewOmahaSixEval(loMax EvalRank) EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		ev.Init(5, 6, loMax)
		v, r := make([]Card, 5), EvalRank(0)
		for i := 0; i < 15; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = pocket[t6c2[i][0]], pocket[t6c2[i][1]] // pocket
				v[2], v[3] = board[t5c3[j][0]], board[t5c3[j][1]]   // board
				v[4] = board[t5c3[j][2]]                            // board
				if r = DefaultEval(v); r < ev.HiRank {
					copy(ev.HiBest, v)
					ev.HiRank = r
					ev.HiUnused[0], ev.HiUnused[1] = pocket[t6c2[i][2]], pocket[t6c2[i][3]] // pocket
					ev.HiUnused[2], ev.HiUnused[3] = pocket[t6c2[i][4]], pocket[t6c2[i][5]] // pocket
					ev.HiUnused[4], ev.HiUnused[5] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < ev.LoRank && r < loMax {
						copy(ev.LoBest, v)
						ev.LoRank = r
						ev.LoUnused[0], ev.LoUnused[1] = pocket[t6c2[i][2]], pocket[t6c2[i][3]] // pocket
						ev.LoUnused[2], ev.LoUnused[3] = pocket[t6c2[i][4]], pocket[t6c2[i][5]] // pocket
						ev.LoUnused[4], ev.LoUnused[5] = board[t5c3[j][3]], board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(ev, loMax)
	}
}

// NewStudEval creates a Stud rank eval func.
func NewStudEval(loMax EvalRank) EvalFunc {
	hi := NewCactusEval(DefaultEval, Five)
	lo := NewLowEval(RankEightOrBetter, loMax)
	return func(ev *Eval, pocket, board []Card) {
		hi(ev, pocket, board)
		if loMax != Invalid {
			v := EvalOf(StudHiLo)
			lo(v, pocket, board)
			if v.HiRank < loMax {
				ev.LoRank, ev.LoBest, ev.LoUnused = v.HiRank, v.HiBest, v.HiUnused
			}
		}
	}
}

// NewRazzEval creates a Razz rank eval func.
func NewRazzEval() EvalFunc {
	f := NewLowEval(RankRazz, Invalid)
	return func(ev *Eval, pocket, board []Card) {
		f(ev, pocket, board)
		if rankAceFiveMax <= ev.HiRank {
			switch r := Invalid - ev.HiRank; r.Fixed() {
			case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
				ev.HiBest = bestSet(ev.HiBest)
			default:
				panic("bad rank")
			}
		}
	}
}

// NewBadugiEval creates a Badugi rank eval func.
func NewBadugiEval() EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		s := make([][]Card, 4)
		for i := 0; i < len(pocket) && i < 4; i++ {
			idx := pocket[i].SuitIndex()
			s[idx] = append(s[idx], pocket[i])
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
			return s[i][0].AceIndex() < s[j][0].AceIndex()
		})
		count, rank := 4, 0
		for i := 0; i < 4; i++ {
			sort.Slice(s[i], func(j, k int) bool {
				return s[i][j].AceIndex() < s[i][k].AceIndex()
			})
			captured, r := false, 0
			for j := 0; j < len(s[i]); j++ {
				if r = 1 << s[i][j].AceIndex(); rank&r == 0 && !captured {
					captured, ev.HiBest = true, append(ev.HiBest, s[i][j])
					rank |= r
					count--
				} else {
					ev.HiUnused = append(ev.HiUnused, s[i][j])
				}
			}
		}
		sort.Slice(ev.HiBest, func(i, j int) bool {
			return ev.HiBest[j].AceIndex() < ev.HiBest[i].AceIndex()
		})
		sort.Slice(ev.HiUnused, func(i, j int) bool {
			if a, b := ev.HiUnused[i].AceIndex(), ev.HiUnused[j].AceIndex(); a != b {
				return b < a
			}
			return ev.HiUnused[i].Suit() < ev.HiUnused[j].Suit()
		})
		ev.HiRank = EvalRank(count<<13 | rank)
	}
}

// NewLowballEval creates a Lowball rank eval func.
func NewLowballEval() EvalFunc {
	f := NewRankFunc(RankLowball)
	return func(ev *Eval, pocket, board []Card) {
		if len(pocket) != 5 {
			panic("bad pocket")
		}
		v := make([]Card, len(pocket)+len(board))
		copy(v, pocket)
		copy(v[len(pocket):], board)
		ev.HiRank = f(pocket)
		bestLowball(ev, v)
	}
}

// NewSokoEval creates a Soko rank eval func.
func NewSokoEval() EvalFunc {
	f := NewCactusEval(DefaultEval, Five)
	return func(ev *Eval, pocket, board []Card) {
		f(ev, pocket, board)
	}
}

// NewLowEval creates a low rank eval func, using f to determine the best 5 low
// cards out of 7.
func NewLowEval(f RankFunc, loMax EvalRank) EvalFunc {
	return func(ev *Eval, pocket, board []Card) {
		v := make([]Card, len(pocket)+len(board))
		copy(v, pocket)
		copy(v[len(pocket):], board)
		if len(v) != 7 {
			panic("bad pocket or board")
		}
		best, unused := make([]Card, 5), make([]Card, 2)
		rank, r := Invalid, EvalRank(0)
		for i := 0; i < 21; i++ {
			if r = f(
				v[t7c5[i][0]],
				v[t7c5[i][1]],
				v[t7c5[i][2]],
				v[t7c5[i][3]],
				v[t7c5[i][4]],
			); r < rank && r < loMax {
				rank = r
				best[0], best[1] = v[t7c5[i][0]], v[t7c5[i][1]]
				best[2], best[3] = v[t7c5[i][2]], v[t7c5[i][3]]
				best[4] = v[t7c5[i][4]]
				unused[0], unused[1] = v[t7c5[i][5]], v[t7c5[i][6]]
			}
		}
		if loMax <= rank {
			return
		}
		// order
		sort.Slice(best, func(i, j int) bool {
			return best[j].AceIndex() < best[i].AceIndex()
		})
		ev.HiRank, ev.HiBest, ev.HiUnused = rank, best, unused
	}
}

// Eval contains eval info.
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
func (ev *Eval) Eval(pocket, board []Card) *Eval {
	evals[ev.Type](ev, pocket, board)
	return ev
}

// Double evaluates the pocket, board, coping the results to the LoRank,
// LoBest, and LoUnused.
func (ev *Eval) Double(pocket, board []Card) {
	v := EvalOf(ev.Type)
	evals[ev.Type](v, pocket, board)
	ev.LoRank, ev.LoBest, ev.LoUnused = v.HiRank, v.HiBest, v.HiUnused
}

// Init inits best, unused.
func (ev *Eval) Init(n, m int, loMax EvalRank) {
	if 0 < n {
		ev.HiBest = make([]Card, n)
	}
	if 0 < m {
		ev.HiUnused = make([]Card, m)
	}
	if loMax != Invalid {
		if 0 < n {
			ev.LoBest = make([]Card, n)
		}
		if 0 < m {
			ev.LoUnused = make([]Card, m)
		}
	}
}

// HiComp compares the hi eval.
func (ev *Eval) HiComp(b *Eval) int {
	return ev.Type.HiComp()(ev, b)
}

// LoComp compares the lo eval.
func (ev *Eval) LoComp(b *Eval) int {
	return ev.Type.LoComp()(ev, b)
}

// HiDesc returns the desc for the hi eval.
func (ev *Eval) HiDesc() *Desc {
	return &Desc{
		Type:   ev.Type.HiDesc(),
		Rank:   ev.HiRank,
		Best:   ev.HiBest,
		Unused: ev.HiUnused,
	}
}

// LoDesc returns the desc for the lo eval.
func (ev *Eval) LoDesc() *Desc {
	if low := ev.Type.Low(); low || ev.Type.Double() {
		return &Desc{
			Type:   ev.Type.LoDesc(),
			Rank:   ev.LoRank,
			Best:   ev.LoBest,
			Unused: ev.LoUnused,
			Low:    low,
		}
	}
	return nil
}

// Format satisfies the fmt.Formatter interface.
func (ev *Eval) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		fmt.Fprintf(f, "%s %s", ev.HiDesc(), ev.HiBest)
	case 'q':
		fmt.Fprintf(f, "\"%s %s\"", ev.HiDesc(), ev.HiBest)
	case 'S':
		fmt.Fprintf(f, "%s %S", ev.HiDesc(), CardFormatter(ev.HiBest))
	case 'b':
		fmt.Fprintf(f, "%s %b", ev.HiDesc(), ev.HiBest)
	case 'h':
		fmt.Fprintf(f, "%s %h", ev.HiDesc(), CardFormatter(ev.HiBest))
	case 'c':
		fmt.Fprintf(f, "%s %c", ev.HiDesc(), ev.HiBest)
	case 'C':
		fmt.Fprintf(f, "%s %C", ev.HiDesc(), CardFormatter(ev.HiBest))
	case 'f':
		for _, c := range ev.HiBest {
			c.Format(f, 's')
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, Eval<%s>: %s/%s %d)", verb, ev.Type, ev.HiBest, ev.HiUnused, ev.HiRank)
	}
}

// Desc wraps describing results.
type Desc struct {
	Type   DescType
	Rank   EvalRank
	Best   []Card
	Unused []Card
	Low    bool
}

// Format satisfies the fmt.Stringer interface.
func (desc *Desc) Format(f fmt.State, verb rune) {
	desc.Type.Desc(f, verb, desc.Rank, desc.Best, desc.Unused, desc.Low)
}

// HiOrder determines the order for v, low to high, using HiComp. Returns
// indices and pivot of winning vs losing. Pivot will always be 1 or higher.
func HiOrder(evs []*Eval) ([]int, int) {
	if len(evs) == 0 {
		return nil, 0
	}
	i, n := 0, len(evs)
	m, v := make(map[int]*Eval, n), make([]int, n)
	for ; i < n; i++ {
		m[i], v[i] = evs[i], i
	}
	f := evs[0].Type.HiComp()
	sort.SliceStable(v, func(j, k int) bool {
		return f(m[v[j]], m[v[k]]) < 0
	})
	for i = 1; i < n && m[v[i-1]] != nil && m[v[i]] != nil && m[v[i-1]].HiRank == m[v[i]].HiRank; i++ {
	}
	return v, i
}

// LoOrder determines the order for v, low to high, using LoComp. Returns
// indices and pivot of winning vs losing. If there are no low evals the pivot
// will be 0.
func LoOrder(evs []*Eval) ([]int, int) {
	if len(evs) == 0 {
		return nil, 0
	}
	i, n := 0, len(evs)
	m, v := make(map[int]*Eval, n), make([]int, n)
	for ; i < n; i++ {
		m[i], v[i] = evs[i], i
	}
	f := evs[0].Type.LoComp()
	sort.SliceStable(v, func(j, k int) bool {
		return f(m[v[j]], m[v[k]]) < 0
	})
	if m[v[0]] == nil || m[v[0]].LoRank == Invalid {
		return nil, 0
	}
	for i = 1; i < n && m[v[i-1]] != nil && m[v[i]] != nil && m[v[i-1]].LoRank == m[v[i]].LoRank; i++ {
	}
	return v, i
}

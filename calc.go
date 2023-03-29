package cardrank

import (
	"context"
	"fmt"
	"sort"
)

// Calc calculates run odds.
type Calc struct {
	typ     Type
	runs    []*Run
	active  map[int]bool
	discard bool
	deep    bool
}

// NewCalc creates a new run odds calc.
func NewCalc(typ Type, opts ...CalcOption) *Calc {
	c := &Calc{
		typ: typ,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// u builds the set of unused cards.
func (c *Calc) u() []Card {
	var ex [][]Card
	for _, run := range c.runs {
		if c.discard {
			ex = append(ex, run.Discard)
		}
		ex = append(ex, run.Hi, run.Lo)
		if c.active == nil {
			ex = append(ex, run.Pockets...)
		} else {
			for i := 0; i < len(run.Pockets); i++ {
				if c.active[i] {
					ex = append(ex, run.Pockets[i])
				}
			}
		}
	}
	return c.typ.DeckType().Exclude(ex...)
}

// Calc calculates odds.
func (c *Calc) Calc(ctx context.Context) (*Odds, *Odds, bool) {
	// check runs and pocket count
	n := len(c.runs)
	if n == 0 {
		return nil, nil, false
	}
	// ensure at least 1 pocket pair has been dealt
	count := len(c.runs[n-1].Pockets)
	if count == 0 {
		return nil, nil, false
	}
	b, low, double := c.typ.Board(), c.typ.Low(), c.typ.Double()
	run := c.runs[n-1].Dupe()
	k, u := b-len(run.Hi), c.u()
	// if pocket == 2, board == 0, use lookup
	if !c.deep && b == k {
		hi, lo := run.CalcStart(100, low || double)
		return hi, lo, true
	}
	// expand hi + lo boards
	run.Hi = append(run.Hi, make([]Card, k)...)
	if double {
		run.Lo = append(run.Lo, make([]Card, k)...)
	}
	// setup odds
	i, g, v, hi := 0, NewBinGen(len(u), k), make([]int, k), NewOdds(count, u)
	var lo *Odds
	if low || double {
		lo = NewOdds(count, u)
	}
	hiSuits, loSuits := countRunSuits(run, double)
	// iterate combinations
	for g.Next(v) {
		// check context
		select {
		case <-ctx.Done():
			return hi, lo, false
		default:
		}
		// populate hi + lo boards
		for i = 0; i < k; i++ {
			run.Hi[b-k+i] = u[v[i]]
		}
		if double {
			copy(run.Lo[b-k:], run.Hi[b-k:])
		}
		// eval
		evs := run.Eval(c.typ, c.active, true)
		// add to odds
		hi.Add(evs, hiSuits, run.Hi[b-k:], false)
		switch {
		case low:
			lo.Add(evs, loSuits, run.Hi[b-k:], true)
		case double:
			lo.Add(evs, loSuits, run.Lo[b-k:], true)
		}
	}
	return hi, lo, true
}

// Odds are calculated run odds.
type Odds struct {
	N int
	V []int
	O []map[Card]bool
	S [][]Suit
	U map[Card]bool
	D bool
}

// NewOdds creates a new odds.
func NewOdds(count int, u []Card) *Odds {
	odds := &Odds{
		V: make([]int, count),
		O: make([]map[Card]bool, count),
		S: make([][]Suit, count),
		U: make(map[Card]bool),
	}
	for i := 0; i < count; i++ {
		odds.O[i] = make(map[Card]bool)
	}
	for _, c := range u {
		odds.U[c] = true
	}
	return odds
}

// Add adds the eval results to the odds.
func (odds *Odds) Add(evs []*Eval, suits [][4]int, v []Card, low bool) {
	indices, pivot := Order(evs, low)
	s := make([][4]int, len(suits))
	copy(s, suits)
	for i := 0; i < pivot; i++ {
		odds.V[indices[i]]++
		for j := 0; j < len(v); j++ {
			odds.O[indices[i]][v[j]] = true
		}
	}
	odds.N += pivot
}

// Float32 returns the odds as a slice of float32.
func (odds *Odds) Float32() []float32 {
	n := len(odds.V)
	v := make([]float32, len(odds.V))
	for i := 0; i < n; i++ {
		v[i] = float32(odds.V[i]) / float32(max(odds.N, 1))
	}
	return v
}

// Percent returns the odds for pos calculated as a percent.
func (odds *Odds) Percent(pos int) float32 {
	return float32(odds.V[pos]) / float32(max(odds.N, 1)) * 100
}

// Outs returns the out cards and suits for pos.
func (odds *Odds) Outs(pos int, distinct bool) ([]Card, []Suit) {
	v, s := odds.outs(pos, distinct)
	sort.Slice(v, func(i, j int) bool {
		m, n := v[i].Suit(), v[j].Suit()
		if m == n {
			return v[j].Rank() < v[i].Rank()
		}
		return m < n
	})
	sort.Slice(s, func(i, j int) bool {
		return s[i] > s[j]
	})
	return v, s
}

// outs returns the out cards and suits for pos.
func (odds *Odds) outs(pos int, distinct bool) ([]Card, []Suit) {
	v := make([]Card, len(odds.O[pos]))
	var j int
	for c := range odds.O[pos] {
		v[j] = c
		j++
	}
	if !distinct {
		return v, odds.S[pos]
	}
	return nil, nil
}

// Format satisfies the [fmt.Formatter] interface.
func (odds *Odds) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		if i, ok := f.Width(); ok {
			fmt.Fprintf(f, "%0.1f%% (%d/%d)", odds.Percent(i), odds.V[i], odds.N)
		}
	case 'o', 'O':
		odds.formatOuts(f, 's', verb == 'O')
	case 'b', 'B':
		odds.formatOuts(f, 'b', verb == 'B')
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, odds)", verb)
	}
}

// formatOuts formats outs to f.
func (odds *Odds) formatOuts(f fmt.State, verb rune, distinct bool) {
	if i, ok := f.Width(); ok {
		v, s := odds.Outs(i, distinct)
		switch n, m := len(v), len(s); {
		case n == 0 && m == 0:
			if odds.D {
				f.Write([]byte("drawing dead"))
			} else {
				f.Write([]byte("none"))
			}
		default:
			if n != 0 {
				CardFormatter(v).Format(f, verb)
				if m != 0 {
					f.Write([]byte(", "))
				}
			}
			if m != 0 {
				f.Write([]byte("any ["))
				for i := 0; i < m; i++ {
					if i != 0 {
						f.Write([]byte(", "))
					}
					switch verb {
					case 'b':
						f.Write([]byte(string(s[i].UnicodeBlack())))
					default:
						f.Write([]byte(s[i].Name()))
					}
				}
				f.Write([]byte{']'})
			}
		}
	}
}

// CalcOption is a run odds calc option.
type CalcOption func(*Calc)

// WithCalcRuns is a run odds calc option to set the runs.
func WithCalcRuns(runs []*Run) CalcOption {
	return func(c *Calc) {
		c.runs = runs
	}
}

// WithCalcPockets is a run odds calc option to run with the pockets, board.
func WithCalcPockets(pockets [][]Card, board []Card) CalcOption {
	return func(c *Calc) {
		run := NewRun(len(pockets))
		run.Pockets, run.Hi = pockets, board
		c.runs = append(c.runs, run)
	}
}

// WithCalcActive is a run odds calc option to run with the active map.
func WithCalcActive(active map[int]bool) CalcOption {
	return func(c *Calc) {
		c.active = active
	}
}

// WithCalcDiscard is a run odds calc option to set whether the run's discarded
// cards should be excluded.
func WithCalcDiscard(discard bool) CalcOption {
	return func(c *Calc) {
		c.discard = discard
	}
}

// WithCalcDeep is a run odds calc option to set whether the run should run
// deep calculations.
func WithCalcDeep(deep bool) CalcOption {
	return func(c *Calc) {
		c.deep = deep
	}
}

// BinGen is a binomial combination generator.
type BinGen struct {
	i int
	n int
	k int
	v []int
}

// NewBinGen creates a binomial combination generator.
func NewBinGen(n, k int) *BinGen {
	// calculate amount
	i, l := -1, k
	if n >= 0 && k >= 0 && n > k {
		if k > n/2 {
			l = n - k
		}
		i = 1
		for j := 1; j <= l; j++ {
			i = (n - l + j) * i / j
		}
	}
	return &BinGen{
		i: i,
		n: n,
		k: k,
	}
}

// Next generates the next binomial combination, storing the result in v.
func (g *BinGen) Next(v []int) bool {
	switch {
	case g.i <= 0:
		g.i = -1
		return false
	case g.v == nil:
		g.v = make([]int, g.k)
		for i := range g.v {
			g.v[i] = i
		}
	default:
		for i := g.k - 1; i >= 0; i-- {
			if g.v[i] == g.n+i-g.k {
				continue
			}
			g.v[i]++
			for j := i + 1; j < g.k; j++ {
				g.v[j] = g.v[i] + j - i
			}
			break
		}
	}
	g.i--
	copy(v, g.v)
	return true
}

// CalcStart returns the starting pocket value between 0 and 1.
func CalcStart(pocket []Card) (float32, bool) {
	if len(pocket) != 2 {
		return 0, false
	}
	r0, r1 := pocket[0].Rank(), pocket[1].Rank()
	if r0 < r1 {
		r0, r1 = r1, r0
	}
	i := 0
	if r0 != r1 && pocket[0].Suit() != pocket[1].Suit() {
		i = 1
	}
	return 1.0 - float32(starting[string([]byte{r0.Byte(), r1.Byte()})][i])/169.0, true
}

// countRunSuits returns the suit counts for each of the run's pockets.
func countRunSuits(run *Run, double bool) ([][4]int, [][4]int) {
	hi := countCardSuits(run.Pockets, run.Hi)
	var lo [][4]int
	if double {
		lo = countCardSuits(run.Pockets, run.Lo)
	}
	return hi, lo
}

// countCardSuits returns suit counts.
func countCardSuits(pockets [][]Card, board []Card) [][4]int {
	count := len(pockets)
	if count == 0 {
		return nil
	}
	base := make([]int, 4)
	countSuits(base, board)
	v := make([][4]int, count)
	for i := 0; i < count; i++ {
		copy(v[i][:], base)
	}
	return v
}

// countSuits counts the suits in v, adding to d.
func countSuits(d []int, v []Card) {
	for _, c := range v {
		d[c.SuitIndex()]++
	}
}

// starting are starting pockets.
//
// The first number is the pocket or suited value, the second is the non-suited
// value.
var starting = map[string][]uint8{
	"AA": {0},
	"AK": {3, 10},
	"AQ": {5, 17},
	"AJ": {7, 26},
	"AT": {11, 41},
	"A9": {18, 75},
	"A8": {23, 90},
	"A7": {29, 101},
	"A6": {33, 112},
	"A5": {27, 100},
	"A4": {31, 103},
	"A3": {32, 108},
	"A2": {38, 116},
	"J2": {88, 154},
	"J3": {86, 152},
	"J4": {85, 151},
	"J5": {81, 148},
	"J6": {78, 146},
	"J7": {63, 128},
	"J8": {40, 107},
	"J9": {25, 79},
	"JJ": {4},
	"JT": {15, 46},
	"KK": {1},
	"KQ": {6, 19},
	"KJ": {8, 30},
	"KT": {13, 44},
	"K9": {21, 80},
	"K8": {36, 111},
	"K7": {43, 121},
	"K6": {52, 124},
	"K5": {54, 127},
	"K4": {57, 131},
	"K3": {58, 132},
	"K2": {59, 134},
	"Q9": {24, 82},
	"Q8": {42, 114},
	"Q7": {60, 130},
	"Q6": {65, 136},
	"Q5": {68, 140},
	"Q4": {70, 142},
	"Q3": {71, 143},
	"Q2": {74, 145},
	"QQ": {2},
	"QJ": {12, 34},
	"QT": {14, 48},
	"TT": {9},
	"T9": {22, 72},
	"T8": {37, 99},
	"T7": {56, 123},
	"T6": {73, 139},
	"T5": {92, 156},
	"T4": {94, 157},
	"T3": {95, 159},
	"T2": {97, 161},
	"99": {16},
	"98": {39, 98},
	"97": {53, 118},
	"96": {67, 133},
	"95": {87, 149},
	"94": {105, 163},
	"93": {106, 164},
	"92": {110, 165},
	"88": {20},
	"87": {47, 113},
	"86": {61, 125},
	"85": {77, 138},
	"84": {93, 155},
	"83": {115, 166},
	"82": {117, 167},
	"77": {28},
	"76": {55, 120},
	"75": {66, 129},
	"74": {84, 144},
	"73": {102, 160},
	"72": {119, 168},
	"66": {35},
	"65": {62, 122},
	"64": {69, 135},
	"63": {89, 147},
	"62": {109, 162},
	"55": {45},
	"54": {64, 126},
	"53": {76, 137},
	"52": {91, 150},
	"44": {49},
	"43": {83, 141},
	"42": {96, 153},
	"33": {50},
	"32": {104, 158},
	"22": {51},
}

package cardrank

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
)

// OddsCalc calculates run odds.
type OddsCalc struct {
	typ     Type
	runs    []*Run
	active  map[int]bool
	folded  bool
	discard bool
	deep    bool
}

// NewOddsCalc creates a new run odds calc.
func NewOddsCalc(typ Type, opts ...CalcOption) *OddsCalc {
	c := &OddsCalc{
		typ: typ,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// u builds the set of unused cards.
func (c *OddsCalc) u() []Card {
	var ex [][]Card
	for _, run := range c.runs {
		if c.discard {
			ex = append(ex, run.Discard)
		}
		ex = append(ex, run.Hi, run.Lo)
		if c.active == nil || c.folded {
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
func (c *OddsCalc) Calc(ctx context.Context) (*Odds, *Odds, bool) {
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
		hi, lo := run.CalcStart(low || double)
		return hi, lo, true
	}
	// expand hi + lo boards
	run.Hi = append(run.Hi, make([]Card, k)...)
	if double {
		run.Lo = append(run.Lo, make([]Card, k)...)
	}
	// setup odds
	hi := NewOdds(count, u)
	var lo *Odds
	if low || double {
		lo = NewOdds(count, u)
	}
	hiSuits, loSuits := countRunSuits(run, double)
	// iterate combinations
	offset := b - k
	for g, v := NewCombinGen(u, k); g.Next(); {
		// check context
		select {
		case <-ctx.Done():
			return hi, lo, false
		default:
		}
		// populate hi + lo boards
		copy(run.Hi[offset:], v)
		if double {
			copy(run.Lo[offset:], v)
		}
		// eval
		evs := run.Eval(c.typ, c.active, true)
		// add to odds
		hi.Add(evs, hiSuits, run.Hi[offset:], false)
		switch {
		case low:
			lo.Add(evs, loSuits, run.Hi[offset:], true)
		case double:
			lo.Add(evs, loSuits, run.Lo[offset:], true)
		}
	}
	return hi, lo, true
}

// Odds are calculated run odds.
type Odds struct {
	// Total is the total number of outcomes.
	Total int
	// Counts is each position's outcome count for wins and splits.
	Counts []int
	// Outs are map of the available outs for a position.
	Outs []map[Card]bool
	// Suits [][]Suit
	// Dead  bool
}

// NewOdds creates a new odds.
func NewOdds(count int, u []Card) *Odds {
	odds := &Odds{
		Counts: make([]int, count),
		Outs:   make([]map[Card]bool, count),
		// Suits: make([][]Suit, count),
	}
	for i := 0; i < count; i++ {
		odds.Outs[i] = make(map[Card]bool)
	}
	return odds
}

// Add adds the eval results to the odds.
func (odds *Odds) Add(evs []*Eval, suits [][4]int, v []Card, low bool) {
	indices, pivot := Order(evs, low)
	s := make([][4]int, len(suits))
	copy(s, suits)
	for i := 0; i < pivot; i++ {
		odds.Counts[indices[i]]++
		for j := 0; j < len(v); j++ {
			odds.Outs[indices[i]][v[j]] = true
		}
	}
	odds.Total += pivot
}

// Float32 returns the odds as a slice of float32.
func (odds *Odds) Float32() []float32 {
	n := len(odds.Counts)
	v := make([]float32, len(odds.Counts))
	for i := 0; i < n; i++ {
		v[i] = float32(odds.Counts[i]) / float32(max(odds.Total, 1))
	}
	return v
}

// Percent returns the odds for pos calculated as a percent.
func (odds *Odds) Percent(pos int) float32 {
	return float32(odds.Counts[pos]) / float32(max(odds.Total, 1)) * 100
}

/*
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
		return s[j] < s[i]
	})
	return v, s
}

// outs returns the out cards and suits for pos.
func (odds *Odds) outs(pos int, distinct bool) ([]Card, []Suit) {
	v := make([]Card, len(odds.OutsMap[pos]))
	var j int
	for c := range odds.OutsMap[pos] {
		v[j] = c
		j++
	}
	if !distinct {
		return v, odds.Suits[pos]
	}
	return nil, nil
}
*/

// Format satisfies the [fmt.Formatter] interface.
func (odds *Odds) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		if i, ok := f.Width(); ok {
			fmt.Fprintf(f, "%0.1f%% (%d/%d)", odds.Percent(i), odds.Counts[i], odds.Total)
		}
	/*
		case 'o', 'O':
			odds.formatOuts(f, 's', verb == 'O')
		case 'b', 'B':
			odds.formatOuts(f, 'b', verb == 'B')
	*/
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, odds)", verb)
	}
}

/*
// formatOuts formats outs to f.
func (odds *Odds) formatOuts(f fmt.State, verb rune, distinct bool) {
	if _, ok := f.Width(); ok {
		v, s := odds.Outs(i, distinct)
		switch n, m := len(v), len(s); {
		case n == 0 && m == 0:
			if odds.Dead {
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
*/

// ExpValueCalc is a expected value calculator.
type ExpValueCalc struct {
	typ       Type
	pocket    []Card
	board     []Card
	opponents int
}

// NewExpValueCalc creates a new expected value calculator.
func NewExpValueCalc(typ Type, pocket []Card, opts ...CalcOption) *ExpValueCalc {
	c := &ExpValueCalc{
		typ:       typ,
		pocket:    pocket,
		opponents: 1,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// u builds the set of unused cards.
func (c *ExpValueCalc) u() []Card {
	return c.typ.DeckType().Exclude(c.pocket, c.board)
}

// Calc calculates the expected value.
func (c *ExpValueCalc) Calc(ctx context.Context) (*ExpValue, bool) {
	u, b, nb := c.u(), c.typ.Board(), len(c.board)
	if len(c.pocket) == 2 && nb == 0 {
		return StartingExpValue(c.pocket), true
	}
	v := make([]Card, b)
	copy(v, c.board)
	count, expv, g := int64(0), c.NewExpValue(), NewBinGenInit(u, b-nb, false, v[b-(b-nb):])
	for g.Next() {
		select {
		case <-ctx.Done():
			return expv, false
		default:
		}
		avail, board := Exclude(u, v[b-(b-nb):]), make([]Card, b)
		copy(board, v)
		atomic.AddInt64(&count, 1)
		go c.do(ctx, expv, board, avail, &count)
	}
	for {
		if atomic.LoadInt64(&count) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return expv, false
		case <-time.After(50 * time.Millisecond):
		}
	}
	return expv, true
}

func (c *ExpValueCalc) do(ctx context.Context, expv *ExpValue, board, avail []Card, wait *int64) {
	// setup evals
	evs := make([]*Eval, 2)
	for i := 0; i < len(evs); i++ {
		evs[i] = EvalOf(c.typ)
	}
	// eval pocket
	f := calcs[c.typ]
	f(evs[0], c.pocket, board)
	// set up variables for loop
	var i, pivot int
	var indices []int
	var win bool
	for g, v := NewCombinGen(avail, c.typ.Pocket()); g.Next(); {
		// eval and order
		evs[1].HiRank = Invalid
		f(evs[1], v, board)
		indices, pivot = Order(evs, false)
		// determine if win
		for i, win = 0, false; i < pivot; i++ {
			win = win || indices[i] == 0
		}
		// tally splits/wins/losses
		switch {
		case win && pivot != 1:
			atomic.AddUint64(&expv.Splits, 1)
		case win:
			atomic.AddUint64(&expv.Wins, 1)
		default:
			atomic.AddUint64(&expv.Losses, 1)
		}
		atomic.AddUint64(&expv.Total, 1)
	}
	atomic.AddInt64(wait, -1)
}

// NewExpValue creates a new expected value.
func (c *ExpValueCalc) NewExpValue() *ExpValue {
	return &ExpValue{
		Opponents: c.opponents,
	}
}

// ExpValue is the result of a expected value calculation.
type ExpValue struct {
	Opponents int
	Wins      uint64
	Splits    uint64
	Losses    uint64
	Total     uint64
}

// Float32 returns the expected value as a float32.
func (expv *ExpValue) Float64() float64 {
	if expv.Total != 0 {
		return (float64(expv.Wins) + float64(expv.Splits)/float64(expv.Opponents+1)) / float64(expv.Total)
	}
	return 0.0
}

// Percent returns the expected value calculated as a percent.
func (expv *ExpValue) Percent() float64 {
	return expv.Float64() * 100.0
}

// Format satisfies the [fmt.Formatter] interface.
func (expv *ExpValue) Format(f fmt.State, verb rune) {
	switch verb {
	case 'f':
		fmt.Fprintf(f, "%f", expv.Float64())
	case 's', 'v':
		fmt.Fprintf(f, "%0.1f%% (%d,%d/%d)", expv.Percent(), expv.Wins, expv.Splits, expv.Total)
	default:
		fmt.Fprintf(
			f,
			"%%!%c(ERROR=unknown verb, ExpValue<%d, %d, %d, %d>)",
			verb,
			expv.Wins, expv.Splits, expv.Losses, expv.Total,
		)
	}
}

// CalcOption is calculator option.
type CalcOption func(interface{})

// WithRuns is a run odds calc option to set the runs.
func WithRuns(runs []*Run) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*OddsCalc); ok {
			c.runs = runs
		}
	}
}

// WithPocketsBoard is a run odds calc option to run with the pockets, board.
func WithPocketsBoard(pockets [][]Card, board []Card) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*OddsCalc); ok {
			run := NewRun(len(pockets))
			run.Pockets, run.Hi = pockets, board
			c.runs = append(c.runs, run)
		}
	}
}

// WithActive is a run odds calc option to run with the active map and whether
// or not folded positions should be included.
func WithActive(active map[int]bool, folded bool) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*OddsCalc); ok {
			c.active, c.folded = active, folded
		}
	}
}

// WithDiscard is a run odds calc option to set whether the run's discarded
// cards should be excluded.
func WithDiscard(discard bool) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*OddsCalc); ok {
			c.discard = discard
		}
	}
}

// WithDeep is a run odds calc option to set whether the run should run deep
// calculations.
func WithDeep(deep bool) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*OddsCalc); ok {
			c.deep = deep
		}
	}
}

// WithBoard is an expected value calculator option to set the board.
func WithBoard(board []Card) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*ExpValueCalc); ok {
			c.board = board
		}
	}
}

// WithOpponents is an expected value calculator option to set the opponents.
func WithOpponents(opponents int) CalcOption {
	return func(v interface{}) {
		if c, ok := v.(*ExpValueCalc); ok {
			c.opponents = opponents
		}
	}
}

// BinGen is a binomial combination generator.
type BinGen[T any] struct {
	s []T
	i int
	n int
	k int
	v []int
	f func()
	d []T
}

// NewBinGen creates a uninitialized binomial combination generator. The
// generator must be manually initialized by calling [Init].
func NewBinGen[T any](s []T, k int) *BinGen[T] {
	// calculate iterations
	i, n, l := -1, len(s), k
	if 0 <= n && 0 <= k && k < n {
		if n/2 < k {
			l = n - k
		}
		i = 1
		for j := 1; j <= l; j++ {
			i = (n - l + j) * i / j
		}
	}
	return &BinGen[T]{
		s: s,
		i: i,
		n: n,
		k: k,
		f: func() {},
	}
}

// NewBinGenInit creates and initializes a binomial combination generator using
// f and d.
func NewBinGenInit[T any](s []T, k int, unused bool, d []T) *BinGen[T] {
	g := NewBinGen(s, k)
	if !unused {
		g.f = g.Copy
	} else {
		g.f = g.Unused
	}
	g.d = d
	return g
}

// NewCombinGen creates a binomial combination generator.
func NewCombinGen[T any](s []T, k int) (*BinGen[T], []T) {
	d := make([]T, k)
	g := NewBinGenInit(s, k, false, d)
	return g, d
}

// NewCombinUnusedGen creates a binomial combination generator that also copies
// the unused values.
func NewCombinUnusedGen[T any](s []T, k int) (*BinGen[T], []T) {
	d := make([]T, len(s))
	g := NewBinGenInit(s, k, true, d)
	return g, d
}

// Next generates the next binomial combination.
func (g *BinGen[T]) Next() bool {
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
		for i := g.k - 1; 0 <= i; i-- {
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
	g.f()
	return true
}

// Copy copies the next combination, storing in d.
func (g *BinGen[T]) Copy() {
	for i := 0; i < g.k; i++ {
		g.d[i] = g.s[g.v[i]]
	}
}

// Unused copies the next combination, storing in d along with the unused
// values.
func (g *BinGen[T]) Unused() {
	m := make(map[int]bool)
	for i := 0; i < g.k; i++ {
		m[g.v[i]], g.d[i] = true, g.s[g.v[i]]
	}
	pos := g.k
	for i := 0; i < g.n; i++ {
		if !m[i] {
			g.d[pos] = g.s[i]
			pos++
		}
	}
}

// startingExpValue is the preloaded map of starting expected value
// calculations.
var startingExpValue map[string]ExpValue

func init() {
	var err error
	if startingExpValue, err = holdemStarting(); err != nil {
		panic(err)
	}
}

// StartingExpValue returns the starting pocket value between 0 and 1.
func StartingExpValue(pocket []Card) *ExpValue {
	if len(pocket) != 2 {
		return nil
	}
	expv := startingExpValue[HashKey(pocket[0], pocket[1])]
	return &expv
}

// HashKey returns the hash key of the pocket cards.
func HashKey(c0, c1 Card) string {
	r0, r1 := c0.Rank(), c1.Rank()
	if r0 < r1 {
		r0, r1 = r1, r0
	}
	switch {
	case r0 == r1:
		return string([]byte{r0.Byte(), r1.Byte()})
	case c0.Suit() != c1.Suit():
		return string([]byte{r0.Byte(), r1.Byte(), 'o'})
	}
	return string([]byte{r0.Byte(), r1.Byte(), 's'})
}

// HoldemStarting returns the starting Holdem pockets.
func HoldemStarting() map[string]ExpValue {
	m, err := holdemStarting()
	if err != nil {
		panic(fmt.Sprintf("unable to load starting pockets: %v", err))
	}
	return m
}

// holdemStarting returns the starting Holdem pockets.
func holdemStarting() (map[string]ExpValue, error) {
	r := csv.NewReader(bytes.NewReader(starting))
	r.FieldsPerRecord = 5
	lines, err := r.ReadAll()
	switch {
	case err != nil:
		return nil, err
	case len(lines) != 170:
		return nil, fmt.Errorf("invalid starting pocket length %d", len(lines))
	}
	re := regexp.MustCompile(`^[2-9AKQJT]{2}[os]?$`)
	m := make(map[string]ExpValue)
	for i, line := range lines[1:] {
		if !re.MatchString(line[0]) {
			return nil, fmt.Errorf("line %d: invalid key %q", i+1, line[0])
		}
		w, _ := strconv.ParseUint(line[1], 10, 64)
		s, _ := strconv.ParseUint(line[2], 10, 64)
		l, _ := strconv.ParseUint(line[3], 10, 64)
		if w+s+l != startingTotal {
			return nil, fmt.Errorf("line %d: wins, splits, losses do not total %d: %d + %d + %d", i+1, startingTotal, w, s, l)
		}
		expv := ExpValue{
			Opponents: 1,
			Wins:      w,
			Splits:    s,
			Losses:    l,
			Total:     startingTotal,
		}
		if fmt.Sprintf("%f", expv.Float64()) != line[4] {
			return nil, fmt.Errorf("line %d: calculated %f does not equal %s!", i+1, expv.Float64(), line[4])
		}
		m[line[0]] = expv
	}
	return m, nil
}

// starting is the embedded starting pocket data.
//
//go:embed starting.csv
var starting []byte

// startingTotal is the total for each starting pocket pair.
const startingTotal = 2097572400

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

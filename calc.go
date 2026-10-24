package cardrank

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"iter"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"sync"
)

// RunOdds are a run's Hi and Lo odds.
type RunOdds struct {
	Hi *Odds
	// Lo is nil when the type has no Lo.
	Lo *Odds
}

// CalcOdds calculates the odds for the last of the runs, where the cards dealt
// in all the runs are unavailable. When ctx is done, returns the partial odds
// and the context's error.
func CalcOdds(ctx context.Context, typ Type, runs []*Run, opts ...OddsOption) (RunOdds, error) {
	c := &oddsCalc{
		typ:  typ,
		runs: runs,
	}
	for _, o := range opts {
		o.applyOdds(c)
	}
	hi, lo, ok := c.calc(ctx)
	switch {
	case ctx.Err() != nil:
		return RunOdds{Hi: hi, Lo: lo}, ctx.Err()
	case !ok:
		return RunOdds{}, ErrInvalidCalc
	}
	return RunOdds{Hi: hi, Lo: lo}, nil
}

// oddsCalc calculates run odds.
type oddsCalc struct {
	typ     Type
	deep    bool
	runs    []*Run
	active  []bool
	folded  bool
	discard bool
}

// u builds the set of unused cards.
func (c *oddsCalc) u() []Card {
	var ex []Cards
	for _, run := range c.runs {
		if c.discard {
			ex = append(ex, run.Discard)
		}
		ex = append(ex, run.Hi, run.Lo)
		for i, p := range run.Pockets {
			if c.active == nil || c.folded || c.active[i] {
				ex = append(ex, p)
			}
		}
	}
	return c.typ.DeckType().Exclude(ex...)
}

// calc calculates odds.
func (c *oddsCalc) calc(ctx context.Context) (*Odds, *Odds, bool) {
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
		hi, lo := run.calcStart(low || double)
		return hi, lo, true
	}
	// expand hi + lo boards
	run.Hi = append(run.Hi, make([]Card, k)...)
	if double {
		run.Lo = append(run.Lo, make([]Card, k)...)
	}
	// setup odds
	hi := newOdds(count)
	var lo *Odds
	if low || double {
		lo = newOdds(count)
	}
	// iterate combinations
	offset := b - k
	for g, v := newCombinGen(u, k); g.Next(); {
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
		evs := run.eval(c.typ, c.active, true)
		// add to odds
		hi.add(evs, false)
		if lo != nil {
			lo.add(evs, true)
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
}

// newOdds creates a new odds.
func newOdds(count int) *Odds {
	return &Odds{
		Counts: make([]int, count),
	}
}

// add adds the eval results to the odds.
func (odds *Odds) add(evs []*Eval, low bool) {
	r := Order(hands(evs, low))
	for _, i := range r.Winners() {
		odds.Counts[i]++
	}
	odds.Total += r.Pivot
}

// Float32 returns the odds as a slice of float32.
func (odds *Odds) Float32() []float32 {
	n := len(odds.Counts)
	v := make([]float32, len(odds.Counts))
	for i := range n {
		v[i] = float32(odds.Counts[i]) / float32(max(odds.Total, 1))
	}
	return v
}

// Percent returns the odds for pos calculated as a percent.
func (odds *Odds) Percent(pos int) float32 {
	return float32(odds.Counts[pos]) / float32(max(odds.Total, 1)) * 100
}

// Format satisfies the [fmt.Formatter] interface.
func (odds *Odds) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		if i, ok := f.Width(); ok {
			fmt.Fprintf(f, "%0.1f%% (%d/%d)", odds.Percent(i), odds.Counts[i], odds.Total)
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, odds)", verb)
	}
}

// expValueCalc is an expected value calculator.
type expValueCalc struct {
	typ       Type
	deep      bool
	pocket    []Card
	board     []Card
	opponents int
}

// CalcExpValue calculates the expected value of the pocket against all
// possible opponent pockets, for all possible remaining boards. When ctx is
// done, returns the partial expected value and the context's error.
func CalcExpValue(ctx context.Context, typ Type, pocket, board Cards, opts ...ExpValueOption) (ExpValue, error) {
	c := &expValueCalc{
		typ:       typ,
		pocket:    pocket,
		board:     board,
		opponents: 1,
	}
	for _, o := range opts {
		o.applyExpValue(c)
	}
	expv, ok := c.calc(ctx)
	switch {
	case ctx.Err() != nil:
		return *expv, ctx.Err()
	case !ok:
		return ExpValue{}, ErrInvalidCalc
	}
	return *expv, nil
}

// u builds the set of unused cards.
func (c *expValueCalc) u() []Card {
	return c.typ.DeckType().Exclude(c.pocket, c.board)
}

// Calc calculates the expected value. Boards are evaluated concurrently by
// [runtime.GOMAXPROCS] workers. When ctx is done, returns the partial expected
// value and false.
func (c *expValueCalc) calc(ctx context.Context) (*ExpValue, bool) {
	u, b, nb := c.u(), c.typ.Board(), len(c.board)
	switch np := len(c.pocket); {
	case !c.deep && 1 < np && np < 7 && nb == 0:
		return startingHandExpValue(c.pocket), true
	case nb == 0:
		return NewExpValue(1), false
	}
	// start workers
	boards := make(chan []Card)
	results := make([]*ExpValue, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for i := range results {
		results[i] = c.newExpValue()
		wg.Go(func() {
			for board := range boards {
				c.do(ctx, results[i], board, Cards(u).Exclude(board[nb:]))
			}
		})
	}
	// send boards
	v := make([]Card, b)
	copy(v, c.board)
send:
	for g := newBinGenInit(u, b-nb, false, v[nb:]); g.Next(); {
		select {
		case <-ctx.Done():
			break send
		case boards <- slices.Clone(v):
		}
	}
	close(boards)
	wg.Wait()
	expv := c.newExpValue()
	for _, res := range results {
		expv.Add(res)
	}
	return expv, ctx.Err() == nil
}

// do tallies the wins, splits, and losses for the pocket against all possible
// opponent pockets from avail, adding them to expv.
func (c *expValueCalc) do(ctx context.Context, expv *ExpValue, board, avail []Card) {
	if ctx.Err() != nil {
		return
	}
	e, ev, opp := evaluators[c.typ], EvalOf(c.typ), EvalOf(c.typ)
	e.Rank(ev, c.pocket, board)
	n := 0
	for g, v := newCombinGen(avail, c.typ.Pocket()); g.Next(); n++ {
		if n&0x3ff == 0x3ff && ctx.Err() != nil {
			return
		}
		opp.Hi.Rank = Invalid
		e.Rank(opp, v, board)
		switch ev.Hi.Compare(opp.Hi) {
		case 0:
			expv.Splits++
		case -1:
			expv.Wins++
		default:
			expv.Losses++
		}
		expv.Total++
	}
}

// newExpValue creates a new expected value.
func (c *expValueCalc) newExpValue() *ExpValue {
	return NewExpValue(c.opponents)
}

// ExpValue is the result of an expected value calculation.
type ExpValue struct {
	Opponents int
	Wins      uint64
	Splits    uint64
	Losses    uint64
	Total     uint64
}

// NewExpValue creates a new expected value.
func NewExpValue(opponents int) *ExpValue {
	return &ExpValue{
		Opponents: opponents,
	}
}

// Add adds b to the expected value.
func (expv *ExpValue) Add(v *ExpValue) {
	expv.Wins += v.Wins
	expv.Splits += v.Splits
	expv.Losses += v.Losses
	expv.Total += v.Total
}

// Float64 returns the expected value as a float64.
func (expv ExpValue) Float64() float64 {
	if expv.Total != 0 {
		return (float64(expv.Wins) + float64(expv.Splits)/float64(expv.Opponents+1)) / float64(expv.Total)
	}
	return 0.0
}

// Percent returns the expected value calculated as a percent.
func (expv ExpValue) Percent() float64 {
	return expv.Float64() * 100.0
}

// Format satisfies the [fmt.Formatter] interface.
func (expv ExpValue) Format(f fmt.State, verb rune) {
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

// OddsOption is an option for [CalcOdds].
type OddsOption interface {
	applyOdds(*oddsCalc)
}

// ExpValueOption is an option for [CalcExpValue].
type ExpValueOption interface {
	applyExpValue(*expValueCalc)
}

// CalcOption is an option for both [CalcOdds] and [CalcExpValue].
type CalcOption interface {
	OddsOption
	ExpValueOption
}

// RunOption is an option for both [CalcOdds] and [CalcOuts].
type RunOption interface {
	OddsOption
	OutsOption
}

// runOption is a [RunOption] func.
type runOption func(*oddsCalc)

func (o runOption) applyOdds(c *oddsCalc) { o(c) }
func (o runOption) applyOuts(c *outsCalc) { o(&c.oddsCalc) }

// expValueOption is an [ExpValueOption] func.
type expValueOption func(*expValueCalc)

func (o expValueOption) applyExpValue(c *expValueCalc) { o(c) }

// deepOption is the [WithDeep] option.
type deepOption bool

func (o deepOption) applyOdds(c *oddsCalc)         { c.deep = bool(o) }
func (o deepOption) applyExpValue(c *expValueCalc) { c.deep = bool(o) }

// WithDeep is a calc option to set whether to run deep calculations, instead
// of using the precalculated starting pocket odds when no board has been
// dealt.
func WithDeep(deep bool) CalcOption {
	return deepOption(deep)
}

// WithActive is a run option to set the active positions. Inactive positions
// are not included in the odds or outs.
func WithActive(active []bool) RunOption {
	return runOption(func(c *oddsCalc) {
		c.active = active
	})
}

// WithFolded is a run option to set whether the cards of inactive positions
// are unavailable (ie, they are known).
func WithFolded(folded bool) RunOption {
	return runOption(func(c *oddsCalc) {
		c.folded = folded
	})
}

// WithDiscard is a run option to set whether the runs' discarded cards are
// unavailable (ie, they are known).
func WithDiscard(discard bool) RunOption {
	return runOption(func(c *oddsCalc) {
		c.discard = discard
	})
}

// WithOpponents is an expected value option to set the number of opponents.
func WithOpponents(opponents int) ExpValueOption {
	return expValueOption(func(c *expValueCalc) {
		c.opponents = opponents
	})
}

// binGen is a binomial combination generator.
type binGen[T any] struct {
	s []T
	i int
	n int
	k int
	v []int
	f func()
	d []T
}

// newBinGen creates an uninitialized binomial combination generator. The
// generator must be manually initialized by setting f.
func newBinGen[T any](s []T, k int) *binGen[T] {
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
	return &binGen[T]{
		s: s,
		i: i,
		n: n,
		k: k,
		f: func() {},
	}
}

// newBinGenInit creates and initializes a binomial combination generator for k
// elements in s, copying the values to d.
func newBinGenInit[T any](s []T, k int, unused bool, d []T) *binGen[T] {
	g := newBinGen(s, k)
	if !unused {
		g.f = g.cpy
	} else {
		g.f = g.unused
	}
	g.d = d
	return g
}

// newCombinGen creates a binomial combination generator for k elements in s.
// Returns the generator and a slice where the values will be copied after each
// to [binGen.Next].
func newCombinGen[T any](s []T, k int) (*binGen[T], []T) {
	d := make([]T, k)
	g := newBinGenInit(s, k, false, d)
	return g, d
}

// newCombinUnusedGen creates a binomial combination generator for k elements
// in s, and len(s)-k unused elements. Returns the generator and a slice where
// the values will be copied after each to [binGen.Next].
func newCombinUnusedGen[T any](s []T, k int) (*binGen[T], []T) {
	d := make([]T, len(s))
	g := newBinGenInit(s, k, true, d)
	return g, d
}

// Combinations returns an iterator over each combination of k elements of s.
// The yielded slice is reused between iterations, so clone it (such as with
// [slices.Clone]) to retain it.
func Combinations[T any](s []T, k int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		for g, v := newCombinGen(s, k); g.Next(); {
			if !yield(v) {
				return
			}
		}
	}
}

// CombinationsUnused returns an iterator over each combination of k elements
// of s, and the remaining unused elements of s. The yielded slices are reused
// between iterations, so clone them (such as with [slices.Clone]) to retain
// them.
func CombinationsUnused[T any](s []T, k int) iter.Seq2[[]T, []T] {
	return func(yield func([]T, []T) bool) {
		for g, v := newCombinUnusedGen(s, k); g.Next(); {
			if !yield(v[:k:k], v[k:]) {
				return
			}
		}
	}
}

// Next generates the next binomial combination.
func (g *binGen[T]) Next() bool {
	switch {
	case g.i <= 0:
		g.i = -1
		return false
	case g.v == nil:
		g.v = make([]int, g.k)
		for i := range g.k {
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

// cpy copies the next combination to d.
func (g *binGen[T]) cpy() {
	for i := range g.k {
		g.d[i] = g.s[g.v[i]]
	}
}

// unused copies the next combination to d along with the unused values.
func (g *binGen[T]) unused() {
	m := make(map[int]bool)
	for i := range g.k {
		m[g.v[i]], g.d[i] = true, g.s[g.v[i]]
	}
	k := g.k
	for i := range g.n {
		if !m[i] {
			g.d[k] = g.s[i]
			k++
		}
	}
}

// startingExpValues is the preloaded map of starting expected value
// calculations.
var startingExpValues map[string]ExpValue

// startingCactus is the preloaded map of starting cactus values.
var startingCactus map[string]EvalRank

func init() {
	var err error
	if startingExpValues, startingCactus, err = holdemStarting(); err != nil {
		panic(err)
	}
}

// startingHandExpValue returns the starting pocket expected value.
func startingHandExpValue(pocket []Card) *ExpValue {
	var f func([]Card) ([][]Card, int)
	switch len(pocket) {
	case 2:
		f = take2c2
	case 3:
		f = take3c2
	case 4:
		f = take4c2
	case 5:
		f = take5c2
	case 6:
		f = take6c2
	default:
		return nil
	}
	pockets, n := f(pocket)
	expv := NewExpValue(1)
	for i := range n {
		v := startingExpValues[StartingKey(pockets[i][0], pockets[i][1])]
		expv.Add(&v)
	}
	return expv
}

// startingHandRank returns the worst (highest) possible resulting 5-card rank
// for the pocket.
//
// For example, a single [King] would be the highest non-[Straight] and
// non-[Flush] value between [Pair] and [HighCard].
func startingHandRank(pocket []Card) EvalRank {
	var f func([]Card) ([][]Card, int)
	switch n := len(pocket); n {
	case 0:
		return Nothing
	case 1:
		switch pocket[0].Rank() {
		case Ace:
			return AceHigh
		case King:
			return KingHigh
		case Queen:
			return QueenHigh
		case Jack:
			return JackHigh
		case Ten:
			return TenHigh
		case Nine:
			return NineHigh
		case Eight:
			return EightHigh
		case Seven:
			return SevenHigh
		case Six:
			return SixHigh
		case Five:
			return FiveHigh
		case Four:
			return FourHigh
		case Three:
			return ThreeHigh
		case Two:
			return TwoHigh
		}
	case 2:
		return startingCactus[StartingKey(pocket[0], pocket[1])]
	case 3:
		f = take3c2
	case 4:
		f = take4c2
	case 5:
		f = take5c2
	case 6:
		f = take6c2
	default:
		return Invalid
	}
	pockets, n := f(pocket)
	r := Invalid
	for i := range n {
		r = min(r, startingCactus[StartingKey(pockets[i][0], pockets[i][1])])
	}
	return r
}

// StartingKey returns the hash key of the pocket cards.
func StartingKey(c0, c1 Card) string {
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

// StartingHand returns the precalculated expected value (against 1 opponent)
// of a Holdem or Omaha starting pocket of 2 to 6 cards, and the worst possible
// resulting 5 card rank for the pocket. Returns a zero expected value and
// [Invalid] for other pocket sizes.
func StartingHand(pocket Cards) (ExpValue, EvalRank) {
	expv := startingHandExpValue(pocket)
	if expv == nil {
		return ExpValue{}, Invalid
	}
	return *expv, startingHandRank(pocket)
}

// holdemStarting returns the starting Holdem pockets.
func holdemStarting() (map[string]ExpValue, map[string]EvalRank, error) {
	r := csv.NewReader(bytes.NewReader(starting))
	r.FieldsPerRecord = 6
	lines, err := r.ReadAll()
	switch {
	case err != nil:
		return nil, nil, fmt.Errorf("unable to read starting pockets: %w", err)
	case len(lines) != 170:
		return nil, nil, fmt.Errorf("invalid starting pocket length %d", len(lines))
	}
	re := regexp.MustCompile(`^[2-9AKQJT]{2}[os]?$`)
	m, v := make(map[string]ExpValue), make(map[string]EvalRank)
	for i, line := range lines[1:] {
		if !re.MatchString(line[0]) {
			return nil, nil, fmt.Errorf("line %d: invalid key %q", i+1, line[0])
		}
		w, _ := strconv.ParseUint(line[1], 10, 64)
		s, _ := strconv.ParseUint(line[2], 10, 64)
		l, _ := strconv.ParseUint(line[3], 10, 64)
		if w+s+l != startingTotal {
			return nil, nil, fmt.Errorf("line %d: wins, splits, losses do not total %d: %d + %d + %d", i+1, startingTotal, w, s, l)
		}
		expv := ExpValue{
			Opponents: 1,
			Wins:      w,
			Splits:    s,
			Losses:    l,
			Total:     startingTotal,
		}
		if fmt.Sprintf("%f", expv.Float64()) != line[4] {
			return nil, nil, fmt.Errorf("line %d: calculated %f does not equal %s", i+1, expv.Float64(), line[4])
		}
		// parse cactus
		r, _ := strconv.ParseUint(line[5], 10, 64)
		if r == 0 || uint64(Nothing) < r {
			return nil, nil, fmt.Errorf("line %d: invalid cactus rank parsed from %q", i+1, line[5])
		}
		m[line[0]], v[line[0]] = expv, EvalRank(r)
	}
	return m, v, nil
}

// starting is the embedded starting pocket data.
//
//go:embed starting.csv
var starting []byte

// startingTotal is the total for each starting pocket pair.
const startingTotal = 2097572400

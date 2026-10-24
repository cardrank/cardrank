package cardrank

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
)

// Shuffler is an interface for a deck shuffler. Compatible with
// math/rand.Rand's Shuffle method.
type Shuffler interface {
	Shuffle(n int, swap func(int, int))
}

// DeckType is a deck type.
type DeckType uint8

// Deck types.
const (
	// DeckFrench is a standard deck of 52 playing cards.
	DeckFrench = DeckType(Two)
	// DeckShort is a deck of 36 playing cards of rank 6+ (see [Short]).
	DeckShort = DeckType(Six)
	// DeckManila is a deck of 32 playing cards of rank 7+ (see [Manila]).
	DeckManila = DeckType(Seven)
	// DeckSpanish is a deck of 28 playing cards of rank 8+ (see [Spanish]).
	DeckSpanish = DeckType(Eight)
	// DeckRoyal is a deck of 20 playing cards of rank 10+ (see [Royal]).
	DeckRoyal = DeckType(Ten)
	// DeckKuhn is a deck of 3 playing cards, a [King], [Queen], and a [Jack]
	// (see [Kuhn]).
	DeckKuhn = DeckType(^uint8(0) - 1)
	// DeckLeduc is a deck of 6 playing cards, a [King], [Queen], and a [Jack]
	// of the [Spade] and [Heart] suits (see [Leduc]).
	DeckLeduc = DeckType(^uint8(0) - 2)
	// DeckFrenchJoker is a standard [DeckFrench] with two additional "wild"
	// Joker cards (as represented by [InvalidCard]).
	DeckFrenchJoker = DeckType(^uint8(0) - 3)
)

// Name returns the deck name.
func (typ DeckType) Name() string {
	switch typ {
	case DeckFrench:
		return "French"
	case DeckShort:
		return "Short"
	case DeckManila:
		return "Manila"
	case DeckSpanish:
		return "Spanish"
	case DeckRoyal:
		return "Royal"
	case DeckKuhn:
		return "Kuhn"
	case DeckLeduc:
		return "Leduc"
	case DeckFrenchJoker:
		return "French Joker"
	}
	return ""
}

// Desc returns the deck description.
func (typ DeckType) Desc(short bool) string {
	switch french := typ == DeckFrench; {
	case french && short:
		return ""
	case french, typ == DeckKuhn, typ == DeckLeduc, typ == DeckFrenchJoker:
		return typ.Name()
	}
	return typ.Name() + " (" + strconv.Itoa(int(typ+2)) + "+)"
}

// Ordinal returns the deck ordinal.
func (typ DeckType) Ordinal() int {
	return int(typ + 2)
}

// Format satisfies the [fmt.Formatter] interface.
func (typ DeckType) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'd':
		buf = []byte(strconv.Itoa(int(typ)))
	case 'n':
		buf = []byte(typ.Name())
	case 'o':
		buf = []byte(strconv.Itoa(typ.Ordinal()))
	case 's', 'S':
		buf = []byte(typ.Desc(verb != 's'))
	case 'v':
		buf = []byte("DeckType(" + Rank(typ).Name() + ")")
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, deck: %d)", verb, int(typ)))
	}
	_, _ = f.Write(buf)
}

// Unshuffled returns a set of the deck's unshuffled cards.
func (typ DeckType) Unshuffled() []Card {
	switch typ {
	case DeckFrench, DeckShort, DeckManila, DeckSpanish, DeckRoyal:
		v := make([]Card, 4*(Ace-Rank(typ)+1))
		var i int
		for _, s := range []Suit{Spade, Heart, Diamond, Club} {
			for r := Rank(typ); r <= Ace; r++ {
				v[i] = New(r, s)
				i++
			}
		}
		return v
	case DeckFrenchJoker:
		return append(DeckFrench.Unshuffled(), InvalidCard, InvalidCard)
	case DeckKuhn:
		return []Card{
			New(King, Spade), New(Queen, Spade), New(Jack, Spade),
		}
	case DeckLeduc:
		return []Card{
			New(King, Spade), New(Queen, Spade), New(Jack, Spade),
			New(King, Heart), New(Queen, Heart), New(Jack, Heart),
		}
	}
	return nil
}

// deck cards.
var (
	deckFrench      []Card
	deckShort       []Card
	deckManila      []Card
	deckSpanish     []Card
	deckRoyal       []Card
	deckKuhn        []Card
	deckLeduc       []Card
	deckFrenchJoker []Card
)

func init() {
	deckFrench = DeckFrench.Unshuffled()
	deckShort = DeckShort.Unshuffled()
	deckManila = DeckManila.Unshuffled()
	deckSpanish = DeckSpanish.Unshuffled()
	deckRoyal = DeckRoyal.Unshuffled()
	deckKuhn = DeckKuhn.Unshuffled()
	deckLeduc = DeckLeduc.Unshuffled()
	deckFrenchJoker = DeckFrenchJoker.Unshuffled()
}

// v returns the cards for the type.
func (typ DeckType) v() []Card {
	switch typ {
	case DeckFrench:
		return deckFrench
	case DeckShort:
		return deckShort
	case DeckManila:
		return deckManila
	case DeckSpanish:
		return deckSpanish
	case DeckRoyal:
		return deckRoyal
	case DeckKuhn:
		return deckKuhn
	case DeckLeduc:
		return deckLeduc
	case DeckFrenchJoker:
		return deckFrenchJoker
	}
	return nil
}

// Shoe creates a card shoe composed of count number of decks of unshuffled
// cards.
func (typ DeckType) Shoe(count int) *Deck {
	v := typ.v()
	n := len(v)
	d := &Deck{
		v: make([]Card, n*count),
		l: count * n,
	}
	for i := range count {
		copy(d.v[i*n:], v)
	}
	return d
}

// New returns a new deck.
func (typ DeckType) New() *Deck {
	return typ.Shoe(1)
}

// Shuffle returns a new deck, shuffled by the shuffler.
func (typ DeckType) Shuffle(shuffler Shuffler) *Deck {
	d := typ.Shoe(1)
	d.Shuffle(shuffler)
	return d
}

// Exclude returns the deck's unshuffled cards, excluding any supplied cards.
func (typ DeckType) Exclude(ex ...Cards) Cards {
	return Cards(typ.v()).Exclude(ex...)
}

// Deck is a set of playing cards.
type Deck struct {
	i int
	l int
	v []Card
}

// DeckOf creates a deck for the provided cards.
func DeckOf(cards ...Card) *Deck {
	return &Deck{
		v: cards,
		l: len(cards),
	}
}

// Limit limits the cards for the deck, for use with card shoes composed of
// more than one deck of cards. The limit is clamped to the number of cards in
// the deck.
func (d *Deck) Limit(limit int) {
	d.l = max(0, min(limit, len(d.v)))
}

// Empty returns true when there are no cards remaining in the deck.
func (d *Deck) Empty() bool {
	return d.l <= d.i
}

// Remaining returns the number of remaining cards in the deck.
func (d *Deck) Remaining() int {
	if n := d.l - d.i; 0 <= n {
		return n
	}
	return 0
}

// All returns an iterator over all cards in the deck, including drawn cards,
// without drawing.
func (d *Deck) All() iter.Seq[Card] {
	return slices.Values(d.v[:d.l])
}

// Reset resets the deck.
func (d *Deck) Reset() {
	d.i = 0
}

// Draw draws count cards from the top (front) of the deck.
func (d *Deck) Draw(count int) Cards {
	if count < 0 {
		return nil
	}
	var cards Cards
	for l := min(d.i+count, d.l); d.i < l; d.i++ {
		cards = append(cards, d.v[d.i])
	}
	return cards
}

// Shuffle shuffles the deck's cards using the shuffler.
func (d *Deck) Shuffle(shuffler Shuffler) {
	shuffler.Shuffle(len(d.v), func(i, j int) {
		d.v[i], d.v[j] = d.v[j], d.v[i]
	})
}

// Dealer deals the streets of a [Type], handling multiple run outs, and
// determining the results of each run. See usage details in the [package
// example].
//
// [package example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package
type Dealer struct {
	desc    TypeDesc
	deck    *Deck
	count   int
	active  []bool
	runs    []*Run
	results []*Result
	n       int
	st      int
	s       int
	r       int
	e       int
}

// NewDealer creates a new dealer for the type, dealing from the deck to the
// number of players.
func NewDealer(typ Type, deck *Deck, players int) *Dealer {
	d := &Dealer{
		desc:  typ.Desc(),
		deck:  deck,
		count: players,
	}
	d.init()
	return d
}

// init inits the street position and active positions.
func (d *Dealer) init() {
	d.active = make([]bool, d.count)
	for i := range d.active {
		d.active[i] = true
	}
	d.runs = []*Run{NewRun(d.count)}
	d.results = nil
	d.n = 1
	d.st = -1
	d.s = -1
	d.r = -1
	d.e = -1
}

// Type returns the dealer's type.
func (d *Dealer) Type() Type {
	return d.desc.Type
}

// Deck returns the dealer's deck.
func (d *Dealer) Deck() *Deck {
	return d.deck
}

// Players returns the number of players.
func (d *Dealer) Players() int {
	return d.count
}

// Active returns true when the position has not folded.
func (d *Dealer) Active(position int) bool {
	return 0 <= position && position < len(d.active) && d.active[position]
}

// activeCount returns the number of active positions.
func (d *Dealer) activeCount() int {
	var n int
	for _, active := range d.active {
		if active {
			n++
		}
	}
	return n
}

// Fold folds the positions, which will not be dealt further cards and will not
// be included in results. Positions cannot be folded after the runs have been
// changed.
func (d *Dealer) Fold(positions ...int) error {
	if d.r != -1 && d.r != 0 {
		return ErrInvalidState
	}
	for _, position := range positions {
		if position < 0 || len(d.active) <= position {
			return ErrInvalidPosition
		}
	}
	for _, position := range positions {
		d.active[position] = false
	}
	return nil
}

// Street returns the current street, or a zero street description when no
// street has been dealt.
func (d *Dealer) Street() StreetDesc {
	if 0 <= d.s && d.s < len(d.desc.Streets) {
		return d.desc.Streets[d.s]
	}
	return StreetDesc{}
}

// StreetIndex returns the index of the current street (see [Dealer.Street]) in
// the type's streets (see [Type.Streets]), or -1 when no street is being
// dealt (before dealing, or after dealing is complete).
func (d *Dealer) StreetIndex() int {
	if 0 <= d.s && d.s < len(d.desc.Streets) {
		return d.s
	}
	return -1
}

// Format satisfies the [fmt.Formatter] interface.
func (d *Dealer) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'n': // name
		buf = []byte(d.Street().Name)
	case 's':
		if street := d.Street(); street.ID != 0 {
			buf = []byte(street.Desc())
		}
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, dealer)", verb))
	}
	_, _ = f.Write(buf)
}

// Deal returns an iterator that deals each street of each run, discarding
// cards prior to dealing pocket and board cards, yielding the run index and
// run. Positions can be folded, and the runs changed, from within the loop.
//
// Streets are dealt while there are at least 2 active positions for a [Type]
// having Max greater than 1.
func (d *Dealer) Deal() iter.Seq2[int, *Run] {
	return func(yield func(int, *Run) bool) {
		for d.next() {
			if !yield(d.r, d.runs[d.r]) {
				return
			}
		}
	}
}

// Results returns an iterator over each run's index and result. Call after
// dealing all streets (see [Dealer.Deal]).
func (d *Dealer) Results() iter.Seq2[int, *Result] {
	return func(yield func(int, *Result) bool) {
		for d.nextResult() {
			if !yield(d.e, d.results[d.e]) {
				return
			}
		}
	}
}

// hasActive returns true when there is more than 1 active positions.
func (d *Dealer) hasActive() bool {
	return 0 <= d.s && (d.desc.Max == 1 || 1 < d.activeCount())
}

// HasCalc returns true when odds are available for calculation.
func (d *Dealer) HasCalc() bool {
	if d.count != 0 && 0 <= d.r && d.r < d.n && d.desc.Type.Cactus() {
		p, b := d.desc.Type.Pocket(), d.desc.Type.Board()
		if p != 2 && d.s == 0 {
			return false
		}
		return b != 0 && len(d.runs[d.r].Pockets[0]) >= p
	}
	return false
}

// Discarded returns the cards discarded on the current street and run.
func (d *Dealer) Discarded() []Card {
	if 0 <= d.s && d.s <= len(d.desc.Streets) && 0 <= d.r && d.r < d.n {
		return d.runs[d.r].Discard
	}
	return nil
}

// Odds calculates the current run's odds.
func (d *Dealer) Odds(ctx context.Context, opts ...OddsOption) (RunOdds, error) {
	if d.r < 0 || d.n <= d.r {
		return RunOdds{}, ErrInvalidState
	}
	return CalcOdds(ctx, d.desc.Type, d.runs[:d.r+1], append([]OddsOption{WithActive(d.active)}, opts...)...)
}

// Reset resets the dealer and deck.
func (d *Dealer) Reset() {
	d.deck.Reset()
	d.init()
}

// ChangeRuns changes the number of runs. Can only be called once, while
// dealing the first run.
func (d *Dealer) ChangeRuns(runs int) error {
	switch {
	case d.r != 0,
		d.n != 1,
		len(d.runs) != 1,
		len(d.desc.Streets) <= d.s,
		!d.hasActive(),
		runs < 1:
		return ErrInvalidState
	}
	d.runs = append(d.runs, make([]*Run, runs-1)...)
	for run := 1; run < runs; run++ {
		d.runs[run] = d.runs[0].Dupe()
	}
	d.st, d.n = d.s, runs
	return nil
}

// next iterates the current street and run, dealing the street. Returns true
// when there are at least 2 active positions for a [Type] having Max greater
// than 1 and when there are additional streets or runs.
func (d *Dealer) next() bool {
	switch {
	case d.s == -1 && d.r == -1:
		d.s, d.r = 0, 0
	default:
		d.s++
	}
	switch n := len(d.desc.Streets); {
	case n <= d.s && d.r == d.n-1, !d.hasActive():
		return false
	case len(d.desc.Streets) <= d.s && d.r < d.n:
		d.s, d.r = d.st+1, d.r+1
	}
	d.deal(d.s, d.runs[d.r])
	return d.s < len(d.desc.Streets) || d.r < d.n-1
}

// nextResult iterates the next result.
func (d *Dealer) nextResult() bool {
	if d.results == nil {
		switch n := d.activeCount(); {
		case n == 1 && d.n == 1 && d.desc.Max != 1:
			// only one active position, which wins
			i := slices.Index(d.active, true)
			evs := make([]*Eval, d.count)
			evs[i] = EvalOf(d.desc.Type)
			r := Ranking{Order: []int{i}, Pivot: 1}
			res := &Result{Evals: evs}
			res.Hi = Win{Ranking: r, evals: evs}
			if d.desc.Low || d.desc.Double {
				res.Lo = Win{Ranking: r, evals: evs, low: true}
			}
			d.results = []*Result{res}
		case n > 1 || d.desc.Max == 1:
			d.results = make([]*Result, d.n)
			for i := range d.n {
				d.results[i] = newResult(d.desc.Type, d.runs[i], d.active, false)
			}
		}
	}
	if d.n <= d.e {
		return false
	}
	d.e++
	return d.e < d.n && d.e < len(d.results)
}

// deal deals pocket and board cards for the street and run, discarding cards
// accordingly.
func (d *Dealer) deal(street int, run *Run) {
	desc := d.desc.Streets[street]
	// pockets
	if p := desc.Pocket; 0 < p {
		if n := desc.PocketDiscard; 0 < n {
			run.Discard = append(run.Discard, d.deck.Draw(n)...)
		}
		for range p {
			for i := range d.count {
				run.Pockets[i] = append(run.Pockets[i], d.deck.Draw(1)...)
			}
		}
	}
	// board
	if b := desc.Board; 0 < b {
		// hi
		disc := desc.BoardDiscard
		if 0 < disc {
			run.Discard = append(run.Discard, d.deck.Draw(disc)...)
		}
		run.Hi = append(run.Hi, d.deck.Draw(b)...)
		// lo
		if d.desc.Double {
			if 0 < disc {
				run.Discard = append(run.Discard, d.deck.Draw(disc)...)
			}
			run.Lo = append(run.Lo, d.deck.Draw(b)...)
		}
	}
}

// Run holds pockets, and a Hi/Lo board for a deal.
type Run struct {
	Discard Cards
	Pockets []Cards
	Hi      Cards
	Lo      Cards
}

// NewRun creates a new run for the pocket count.
func NewRun(count int) *Run {
	return &Run{
		Pockets: make([]Cards, count),
	}
}

// Dupe creates a duplicate of run, with a copy of the pockets and Hi and Lo
// board.
func (run *Run) Dupe() *Run {
	r := &Run{
		Hi: slices.Clone(run.Hi),
		Lo: slices.Clone(run.Lo),
	}
	if run.Pockets != nil {
		r.Pockets = make([]Cards, len(run.Pockets))
		for i, pocket := range run.Pockets {
			r.Pockets[i] = slices.Clone(pocket)
		}
	}
	return r
}

// eval returns the evals for the run's active positions (or all positions
// when active is nil), ranking only when calc is true.
func (run *Run) eval(typ Type, active []bool, calc bool) []*Eval {
	n := len(run.Pockets)
	evs := make([]*Eval, n)
	e := evaluators[typ]
	f := e.Eval
	if calc {
		f = e.Rank
	}
	for i, double := 0, typ.Double(); i < n; i++ {
		if active == nil || active[i] {
			evs[i] = EvalOf(typ)
			f(evs[i], run.Pockets[i], run.Hi)
			if double {
				ev := EvalOf(typ)
				f(ev, run.Pockets[i], run.Lo)
				evs[i].Lo.Rank, evs[i].Lo.Best, evs[i].Lo.Unused = ev.Hi.Rank, ev.Hi.Best, ev.Hi.Unused
			}
		}
	}
	return evs
}

// calcStart returns the run's starting odds.
func (run *Run) calcStart(low bool) (*Odds, *Odds) {
	count := len(run.Pockets)
	hi := newOdds(count)
	hi.Total = startingTotal
	var lo *Odds
	if low {
		lo = newOdds(count)
		lo.Total = startingTotal
	}
	for i, pocket := range run.Pockets {
		expv := startingHandExpValue(pocket)
		if expv == nil {
			return nil, nil
		}
		hi.Counts[i] = int(expv.Wins + expv.Losses)
		if low {
			lo.Counts[i] = int(expv.Wins + expv.Losses)
		}
	}
	return hi, lo
}

// Result contains a run's evals and Hi and Lo wins.
type Result struct {
	// Evals are the evals for each position, nil for inactive positions.
	Evals []*Eval
	// Hi is the Hi win.
	Hi Win
	// Lo is the Lo win, which is not valid when there is no qualifying Lo or
	// the type has no Lo.
	Lo Win
}

// newResult creates a result for the run, storing the calculated or evaluated
// result.
func newResult(typ Type, run *Run, active []bool, calc bool) *Result {
	evs := run.eval(typ, active, calc)
	res := &Result{Evals: evs}
	if typ.Low() || typ.Double() {
		res.Lo = Win{Ranking: Order(hands(evs, true)), evals: evs, low: true}
	}
	res.Hi = Win{
		Ranking: Order(hands(evs, false)),
		Scoop:   typ.Low() && !res.Lo.Valid(),
		evals:   evs,
	}
	return res
}

// Win is the winning positions of a result's Hi or Lo.
type Win struct {
	Ranking
	// Scoop is true when the Hi winners also win the Lo.
	Scoop bool
	evals []*Eval
	low   bool
}

// Valid returns true when there are valid winners.
func (win Win) Valid() bool {
	if win.Pivot == 0 || len(win.evals) == 0 || len(win.Order) == 0 {
		return false
	}
	return win.hand(win.Order[0]).Valid()
}

// hand returns the Hi or Lo hand for the position.
func (win Win) hand(position int) Hand {
	switch ev := win.evals[position]; {
	case ev == nil:
		return Hand{Rank: Invalid}
	case win.low:
		return ev.Lo
	default:
		return ev.Hi
	}
}

// Hands returns the winning hands.
func (win Win) Hands() []Hand {
	var v []Hand
	for _, i := range win.Winners() {
		if h := win.hand(i); h.Valid() {
			v = append(v, h)
		}
	}
	return v
}

// Describe describes the win, using names for the positions (or the position
// numbers for positions without a name).
func (win Win) Describe(names ...string) string {
	if !win.Valid() {
		return "None"
	}
	var v []string
	for _, i := range win.Winners() {
		if i < len(names) {
			v = append(v, names[i])
		} else {
			v = append(v, strconv.Itoa(i))
		}
	}
	return fmt.Sprintf("%s %s with %s", strings.Join(v, ", "), win.Verb(), win.hand(win.Order[0]))
}

// Format satisfies the [fmt.Formatter] interface.
func (win Win) Format(f fmt.State, verb rune) {
	switch verb {
	case 'd':
		var v []string
		for _, i := range win.Winners() {
			v = append(v, strconv.Itoa(i))
		}
		fmt.Fprint(f, strings.Join(v, ", ")+" "+win.Verb())
	case 's':
		win.hand(win.Order[0]).Format(f, 's')
	case 'S':
		fmt.Fprint(f, win.Describe())
	case 'V':
		fmt.Fprint(f, win.Verb())
	case 'v':
		var v []string
		for _, i := range win.Winners() {
			v = append(v, fmt.Sprintf("%v", win.hand(i).Best))
		}
		fmt.Fprint(f, strings.Join(v, ", "))
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, win)", verb)
	}
}

// Verb returns the win verb.
func (win Win) Verb() string {
	switch {
	case win.Scoop:
		return "scoops"
	case win.Pivot > 2:
		return "push"
	case win.Pivot == 2:
		return "split"
	case win.Pivot == 0:
		return "none"
	}
	return "wins"
}

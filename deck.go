package cardrank

import (
	"fmt"
	"strconv"
	"strings"
)

// Shuffler is an interface for a deck shuffler. Compatible with
// math/rand.Rand's Shuffle method.
type Shuffler interface {
	Shuffle(int, func(int, int))
}

// DeckType is a deck type.
type DeckType uint8

// Deck types.
const (
	// DeckFrench is a standard deck of 52 cards (aka "French" deck).
	DeckFrench = DeckType(Two)
	// DeckShort is a deck of Short (6+) cards.
	DeckShort = DeckType(Six)
	// DeckManila is a deck of Manila (7+) cards.
	DeckManila = DeckType(Seven)
	// DeckSpanish is a deck of Spanish (8+) cards.
	DeckSpanish = DeckType(Eight)
	// DeckRoyal is a deck of Royal (10+) cards.
	DeckRoyal = DeckType(Ten)
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
	}
	return ""
}

// Desc returns the deck description.
func (typ DeckType) Desc(short bool) string {
	switch french := typ == DeckFrench; {
	case french && short:
		return ""
	case french:
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
	}
	return nil
}

// deck cards.
var (
	deckFrench  []Card
	deckShort   []Card
	deckManila  []Card
	deckSpanish []Card
	deckRoyal   []Card
)

func init() {
	deckFrench = DeckFrench.Unshuffled()
	deckShort = DeckShort.Unshuffled()
	deckManila = DeckManila.Unshuffled()
	deckSpanish = DeckSpanish.Unshuffled()
	deckRoyal = DeckRoyal.Unshuffled()
}

// Shoe creates a card shoe composed of count number of decks of unshuffled
// cards.
func (typ DeckType) Shoe(count int) *Deck {
	var v []Card
	switch typ {
	case DeckFrench:
		v = deckFrench
	case DeckShort:
		v = deckShort
	case DeckManila:
		v = deckManila
	case DeckSpanish:
		v = deckSpanish
	case DeckRoyal:
		v = deckRoyal
	default:
		return nil
	}
	n := len(v)
	d := &Deck{
		v: make([]Card, n*count),
		l: count * n,
	}
	for i := 0; i < count; i++ {
		copy(d.v[i*n:], v)
	}
	return d
}

// New returns a new deck for the deck type.
func (typ DeckType) New() *Deck {
	return typ.Shoe(1)
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

// NewDeck creates a French deck of 52 unshuffled cards.
func NewDeck() *Deck {
	return DeckFrench.New()
}

// NewShoe creates a card shoe with multiple sets of 52 unshuffled cards.
func NewShoe(count int) *Deck {
	return DeckFrench.Shoe(count)
}

// Limit limits the cards for the deck, for use with card shoes composed of
// more than one deck of cards.
func (d *Deck) Limit(limit int) {
	d.l = limit
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

// All returns a copy of all cards in the deck, without advancing.
func (d *Deck) All() []Card {
	v := make([]Card, d.l)
	copy(v, d.v)
	return v
}

// Reset resets the deck.
func (d *Deck) Reset() {
	d.i = 0
}

// Draw draws count cards from the top (front) of the deck.
func (d *Deck) Draw(count int) []Card {
	if count < 0 {
		return nil
	}
	var cards []Card
	for l := min(d.i+count, d.l); d.i < l; d.i++ {
		cards = append(cards, d.v[d.i])
	}
	return cards
}

// Shuffle shuffles the deck's cards using the shuffler times.
func (d *Deck) Shuffle(shuffler Shuffler, shuffles int) {
	for m := 0; m < shuffles; m++ {
		shuffler.Shuffle(len(d.v), func(i, j int) {
			d.v[i], d.v[j] = d.v[j], d.v[i]
		})
	}
}

// Dealer maintains deal state for a type, streets, deck, positions, runs,
// results, and wins.
type Dealer struct {
	TypeDesc
	Count   int
	Deck    *Deck
	Active  map[int]bool
	Discard []Card
	Runs    []*Run
	Results []*Result
	runs    int
	d       int
	i       int
	r       int
}

// NewDealer creates a new dealer for a provided deck and pocket count.
func NewDealer(desc TypeDesc, deck *Deck, count int) *Dealer {
	d := &Dealer{
		TypeDesc: desc,
		Count:    count,
		Deck:     deck,
	}
	d.init()
	return d
}

// NewShuffledDealer creates a new deck and dealer, shuffling the deck multiple
// times and returning the dealer with the created deck and pocket count.
func NewShuffledDealer(desc TypeDesc, shuffler Shuffler, shuffles, count int) *Dealer {
	d := desc.Type.Deck()
	d.Shuffle(shuffler, shuffles)
	return NewDealer(desc, d, count)
}

// init inits the street position and active positions.
func (d *Dealer) init() {
	d.Active = make(map[int]bool)
	d.Runs = []*Run{NewRun(d.Count)}
	d.Results = nil
	d.runs = 1
	d.d = 0
	d.i = -1
	d.r = -1
	for i := 0; i < d.Count; i++ {
		d.Active[i] = true
	}
}

// Format satisfies the [fmt.Formatter] interface.
func (d *Dealer) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'n': // name
		buf = []byte(d.Streets[d.i].Name)
	case 's':
		buf = []byte(d.Streets[d.i].Desc())
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, dealer)", verb))
	}
	_, _ = f.Write(buf)
}

// Id returns the street id.
func (d *Dealer) Id() byte {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].Id
	}
	return 0
}

// NextId returns the next street id.
func (d *Dealer) NextId() byte {
	if -1 <= d.i && d.i < len(d.Streets)-1 {
		return d.Streets[d.i+1].Id
	}
	return 0
}

// HasNext returns true when there is one or more remaining streets.
func (d *Dealer) HasNext() bool {
	n := len(d.Streets)
	return n != 0 && d.i < n-1
}

// HasPocket returns true when one or more pocket cards are dealt for the
// street.
func (d *Dealer) HasPocket() bool {
	return 0 <= d.i && d.i < len(d.Streets) && 0 < d.Streets[d.i].Pocket
}

// HasBoard returns true when one or more board cards are dealt for the
// street.
func (d *Dealer) HasBoard() bool {
	return 0 <= d.i && d.i < len(d.Streets) && 0 < d.Streets[d.i].Board
}

// HasActive returns true when there is more than 1 active positions.
func (d *Dealer) HasActive() bool {
	return 0 <= d.i && (d.Type.Max() == 1 || 1 < len(d.Active))
}

// Pocket returns the number of pocket cards to be dealt on the street.
func (d *Dealer) Pocket() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].Pocket
	}
	return 0
}

// PocketUp returns the number of pocket cards to be turned up on the current
// street.
func (d *Dealer) PocketUp() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].PocketUp
	}
	return 0
}

// PocketDiscard returns the number of cards to be discarded prior to dealing
// pockets on the current street.
func (d *Dealer) PocketDiscard() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].PocketDiscard
	}
	return 0
}

// PocketDraw returns the number of pocket cards that can be drawn on current
// the street.
func (d *Dealer) PocketDraw() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].PocketDraw
	}
	return 0
}

// Board returns the number of board cards to be dealt on the street.
func (d *Dealer) Board() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].Board
	}
	return 0
}

// BoardDiscard returns the number of board cards to be discarded prior to dealing
// a board on the current street.
func (d *Dealer) BoardDiscard() int {
	if 0 <= d.i && d.i < len(d.Streets) {
		return d.Streets[d.i].BoardDiscard
	}
	return 0
}

// Discarded returns the number of pocket ard board cards discarded on the
// current street.
func (d *Dealer) Discarded() []Card {
	if v := d.Discard[d.d:]; len(v) != 0 {
		return v
	}
	return nil
}

// Inactive returns the inactive positions.
func (d *Dealer) Inactive() []int {
	var v []int
	for i := 0; i < d.Count; i++ {
		if !d.Active[i] {
			v = append(v, i)
		}
	}
	return v
}

// Deactivate deactivates positions, which will not be dealt further cards and
// will not be included during eval.
func (d *Dealer) Deactivate(positions ...int) {
	for _, position := range positions {
		delete(d.Active, position)
	}
}

// Reset resets the iterator to i.
func (d *Dealer) Reset() {
	d.Deck.Reset()
	d.init()
}

// ChangeRuns changes the number of runs, returning true if successful.
func (d *Dealer) ChangeRuns(runs int) bool {
	if d.runs != 1 || len(d.Runs) != 1 || !d.HasActive() {
		return false
	}
	d.Runs = append(d.Runs, make([]*Run, runs-1)...)
	for run := 1; run < runs; run++ {
		d.Runs[run] = d.Runs[0].Dupe()
	}
	if d.i != -1 {
		d.i++
	}
	d.r, d.runs = -1, runs
	return true
}

// Next iterates the run and the street, discarding cards prior to dealing
// additional pocket and board cards for each run. Returns true when there are
// additional runs or streets, and having at least 2 active positions for a
// [Type] having more than 1 player.
func (d *Dealer) Next() bool {
	switch {
	case d.i == -1 && d.r == -1:
		d.i, d.r = 0, 0
	case d.i < len(d.Streets) && d.r == d.runs-1:
		d.i, d.r = d.i+1, 0
	default:
		d.r++
	}
	d.d = len(d.Discard)
	if len(d.Streets) <= d.i || !d.HasActive() {
		d.r = -1
		return false
	}
	d.Deal(d.r)
	return d.i < len(d.Streets)
}

// Run returns the current street and run.
func (d *Dealer) Run() (int, *Run) {
	if 0 <= d.r && d.r < d.runs {
		return d.r, d.Runs[d.r]
	}
	return -1, nil
}

// NextResult iterates the next result.
func (d *Dealer) NextResult() bool {
	if d.Results == nil {
		d.eval()
	}
	d.r++
	return d.r < d.runs
}

// Result returns the current result.
func (d *Dealer) Result() (int, *Result) {
	if 0 <= d.r && d.r < d.runs {
		return d.r, d.Results[d.r]
	}
	return -1, nil
}

// Deal deals pocket and board cards for the run.
func (d *Dealer) Deal(run int) {
	desc := d.Streets[d.i]
	// pockets
	if p := desc.Pocket; 0 < p {
		if n := desc.PocketDiscard; 0 < n {
			d.Discard = append(d.Discard, d.Deck.Draw(n)...)
		}
		for j := 0; j < p; j++ {
			for i := 0; i < d.Count; i++ {
				d.Runs[run].Pockets[i] = append(d.Runs[run].Pockets[i], d.Deck.Draw(1)...)
			}
		}
	}
	// board
	if b := desc.Board; 0 < b {
		// hi
		disc := desc.BoardDiscard
		if 0 < disc {
			d.Discard = append(d.Discard, d.Deck.Draw(disc)...)
		}
		d.Runs[run].Hi = append(d.Runs[run].Hi, d.Deck.Draw(b)...)
		// lo
		if d.Double {
			if 0 < disc {
				d.Discard = append(d.Discard, d.Deck.Draw(disc)...)
			}
			d.Runs[run].Lo = append(d.Runs[run].Lo, d.Deck.Draw(b)...)
		}
	}
}

// eval evaluates the results.
func (d *Dealer) eval() {
	switch n := len(d.Active); {
	case d.Results != nil:
	case n == 1 && d.runs == 1 && d.Max != 1:
		// only one active position
		var i int
		for ; i < d.Count && !d.Active[i]; i++ {
		}
		res := &Result{
			Evals:   []*Eval{EvalOf(d.Type)},
			HiOrder: []int{i},
			HiPivot: 1,
		}
		if d.Low || d.Double {
			res.LoOrder, res.LoPivot = res.HiOrder, res.HiPivot
		}
		d.Results = []*Result{res}
	case n > 1 || d.Max == 1:
		d.Results = make([]*Result, d.runs)
		for run := 0; run < d.runs; run++ {
			d.Results[run] = d.Eval(run)
		}
	}
}

// Eval evals the run, returning the result.
func (d *Dealer) Eval(run int) *Result {
	evs := d.Runs[run].Eval(d.Type, d.Active, d.Double)
	hiOrder, hiPivot := Order(evs, false)
	var loOrder []int
	var loPivot int
	if d.Low || d.Double {
		loOrder, loPivot = Order(evs, true)
	}
	return &Result{
		Evals:   evs,
		HiOrder: hiOrder,
		HiPivot: hiPivot,
		LoOrder: loOrder,
		LoPivot: loPivot,
	}
}

// Run holds pockets, and a hi/lo board for a deal.
type Run struct {
	Pockets [][]Card
	Hi      []Card
	Lo      []Card
}

// Eval returns the evals for the run.
func (run *Run) Eval(typ Type, active map[int]bool, double bool) []*Eval {
	n := len(run.Pockets)
	evs := make([]*Eval, n)
	for i := 0; i < n; i++ {
		if active[i] {
			evs[i] = typ.New()
			evs[i].Eval(run.Pockets[i], run.Hi)
			if double {
				ev := EvalOf(typ)
				ev.Eval(run.Pockets[i], run.Lo)
				evs[i].LoRank, evs[i].LoBest, evs[i].LoUnused = ev.HiRank, ev.HiBest, ev.HiUnused
			}
		}
	}
	return evs
}

// NewRun creates a new run for the pocket count.
func NewRun(count int) *Run {
	return &Run{
		Pockets: make([][]Card, count),
	}
}

// Dupe creates a duplicate of run, with a copy of the pockets and Hi and Lo
// board.
func (run *Run) Dupe() *Run {
	r := new(Run)
	if run.Pockets != nil {
		r.Pockets = make([][]Card, len(run.Pockets))
		for i := 0; i < len(run.Pockets); i++ {
			r.Pockets[i] = make([]Card, len(run.Pockets[i]))
			copy(r.Pockets[i], run.Pockets[i])
		}
	}
	if run.Hi != nil {
		r.Hi = make([]Card, len(run.Hi))
		copy(r.Hi, run.Hi)
	}
	if run.Lo != nil {
		r.Lo = make([]Card, len(run.Lo))
		copy(r.Lo, run.Lo)
	}
	return r
}

// Result contains dealer eval results.
type Result struct {
	Evals   []*Eval
	HiOrder []int
	HiPivot int
	LoOrder []int
	LoPivot int
}

// Win returns the hi and lo win.
func (res *Result) Win() (*Win, *Win) {
	low := res.Evals[res.HiOrder[0]].Type.Low()
	var lo *Win
	if res.LoOrder != nil && res.LoPivot != 0 {
		lo = NewWin(res.Evals, res.LoOrder, res.LoPivot, true, false)
	}
	hi := NewWin(res.Evals, res.HiOrder, res.HiPivot, false, low && lo == nil)
	return hi, lo
}

// Win formats win information.
type Win struct {
	Evals []*Eval
	Order []int
	Pivot int
	Low   bool
	Scoop bool
}

// NewWin creates a new win.
func NewWin(evs []*Eval, order []int, pivot int, low, scoop bool) *Win {
	return &Win{
		Evals: evs,
		Order: order,
		Pivot: pivot,
		Low:   low,
		Scoop: scoop,
	}
}

// Desc returns the eval descriptions.
func (win *Win) Desc() []*EvalDesc {
	var v []*EvalDesc
	for i := 0; i < win.Pivot; i++ {
		if d := win.Evals[win.Order[i]].Desc(win.Low); d != nil && d.Rank != 0 && d.Rank != Invalid {
			v = append(v, d)
		}
	}
	return v
}

// Invalid returns true when there are no valid winners.
func (win *Win) Invalid() bool {
	switch {
	case win == nil, win.Pivot == 0,
		len(win.Evals) == 0, len(win.Order) == 0:
		return false
	}
	d := win.Evals[win.Order[0]].Desc(win.Low)
	return d == nil || d.Rank == 0 || d.Rank == Invalid
}

// Format satisfies the [fmt.Formatter] interface.
func (win *Win) Format(f fmt.State, verb rune) {
	switch verb {
	case 'd':
		var v []string
		for i := 0; i < win.Pivot; i++ {
			v = append(v, strconv.Itoa(win.Order[i]))
		}
		fmt.Fprint(f, strings.Join(v, ", ")+" "+win.Verb())
	case 's':
		win.Evals[win.Order[0]].Desc(win.Low).Format(f, 's')
	case 'S':
		if !win.Invalid() {
			win.Format(f, 'd')
			fmt.Fprint(f, " with ")
			win.Format(f, 's')
		} else {
			fmt.Fprint(f, "None")
		}
	case 'v':
		var v []string
		for i := 0; i < win.Pivot; i++ {
			desc := win.Evals[win.Order[i]].Desc(win.Low)
			v = append(v, fmt.Sprintf("%v", desc.Best))
		}
		fmt.Fprint(f, strings.Join(v, ", "))
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, win)", verb)
	}
}

// Verb returns the win verb.
func (win *Win) Verb() string {
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

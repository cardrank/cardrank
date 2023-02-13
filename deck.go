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

// Ordinal returns the ordinal for the deck.
func (typ DeckType) Ordinal() int {
	return int(typ + 2)
}

// Format satisfies the fmt.Formatter interface.
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
	}
	_, _ = f.Write(buf)
}

// Unshuffled returns a set of unshuffled cards for the deck type.
func (typ DeckType) Unshuffled() []Card {
	switch typ {
	case DeckFrench, DeckShort, DeckManila, DeckRoyal:
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
	deckFrench []Card
	deckShort  []Card
	deckManila []Card
	deckRoyal  []Card
)

func init() {
	deckFrench = DeckFrench.Unshuffled()
	deckShort = DeckShort.Unshuffled()
	deckManila = DeckManila.Unshuffled()
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

// NewDeck creates a deck of 52 unshuffled cards.
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

// Draw draws the next n cards from the top (front) of the deck.
func (d *Deck) Draw(n int) []Card {
	if n < 0 {
		return nil
	}
	var cards []Card
	for l := min(d.i+n, d.l); d.i < l; d.i++ {
		cards = append(cards, d.v[d.i])
	}
	return cards
}

// Shuffle shuffles the deck's cards using the shuffler multiple times.
func (d *Deck) Shuffle(shuffler Shuffler, shuffles int) {
	for m := 0; m < shuffles; m++ {
		shuffler.Shuffle(len(d.v), func(i, j int) {
			d.v[i], d.v[j] = d.v[j], d.v[i]
		})
	}
}

// Dealer maintains deal state for a type, deck, streets, positions, runs, and
// wins.
type Dealer struct {
	TypeDesc
	Count   int
	Deck    *Deck
	Active  map[int]bool
	Discard []Card
	Pockets [][]Card
	Boards  []Board
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
	d.Pockets = make([][]Card, d.Count)
	d.Active = make(map[int]bool)
	d.Boards = make([]Board, 1)
	d.Results = nil
	d.runs = 1
	d.d = 0
	d.i = -1
	d.r = -1
	for i := 0; i < d.Count; i++ {
		d.Active[i] = true
	}
}

// Format satisfies the fmt.Formatter interface.
func (d *Dealer) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'n': // name
		buf = []byte(d.Streets[d.i].Name)
	case 's':
		buf = []byte(d.Streets[d.i].Desc())
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
	return 0 <= d.i && 1 < len(d.Active)
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

// Runs changes the number of runs, returns true if successful.
func (d *Dealer) Runs(runs int) bool {
	if d.runs != 1 || runs <= 1 || len(d.Boards) != 1 || !d.HasActive() {
		return false
	}
	d.Boards = append(d.Boards, make([]Board, runs-1)...)
	for run := 1; run < runs; run++ {
		d.Boards[run] = d.Boards[0].Dupe()
	}
	d.runs = runs
	return true
}

// Next iterates the street, discarding cards prior to dealing additional
// pocket and board cards for each run. Returns true when there are additional
// streets, and when at least 2 active positions.
func (d *Dealer) Next() bool {
	d.i++
	d.d = len(d.Discard)
	if len(d.Streets) <= d.i || !d.HasActive() {
		d.eval()
		return false
	}
	d.DealPocket()
	for run := 0; run < d.runs; run++ {
		d.DealBoard(run)
	}
	return true
}

// NextResult iterates the next result.
func (d *Dealer) NextResult() bool {
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

// DealPocket deals pocket cards for the street.
func (d *Dealer) DealPocket() {
	// pockets
	desc := d.Streets[d.i]
	if p := desc.Pocket; 0 < p {
		if n := desc.PocketDiscard; 0 < n {
			d.Discard = append(d.Discard, d.Deck.Draw(n)...)
		}
		for j := 0; j < p; j++ {
			for i := 0; i < d.Count; i++ {
				d.Pockets[i] = append(d.Pockets[i], d.Deck.Draw(1)...)
			}
		}
	}
}

// DealBoard deals board cards for the street and run.
func (d *Dealer) DealBoard(run int) {
	desc := d.Streets[d.i]
	if b := desc.Board; 0 < b {
		// hi
		disc := desc.BoardDiscard
		if 0 < disc {
			d.Discard = append(d.Discard, d.Deck.Draw(disc)...)
		}
		d.Boards[run].Hi = append(d.Boards[run].Hi, d.Deck.Draw(b)...)
		// lo
		if d.Double {
			if 0 < disc {
				d.Discard = append(d.Discard, d.Deck.Draw(disc)...)
			}
			d.Boards[run].Lo = append(d.Boards[run].Lo, d.Deck.Draw(b)...)
		}
	}
}

// eval evaluates the results.
func (d *Dealer) eval() {
	switch n := len(d.Active); {
	case d.Results != nil:
	case n == 1 && d.runs == 1:
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
	case n > 1:
		d.Results = make([]*Result, d.runs)
		for run := 0; run < d.runs; run++ {
			d.Results[run] = d.EvalRun(run)
		}
	}
}

// EvalRun evals the run.
func (d *Dealer) EvalRun(run int) *Result {
	evs := d.EvalBoard(d.Boards[run])
	hiOrder, hiPivot := HiOrder(evs)
	var loOrder []int
	var loPivot int
	if d.Low || d.Double {
		loOrder, loPivot = LoOrder(evs)
	}
	return &Result{
		Evals:   evs,
		HiOrder: hiOrder,
		HiPivot: hiPivot,
		LoOrder: loOrder,
		LoPivot: loPivot,
	}
}

// EvalBoard evals the board for the pockets.
func (d *Dealer) EvalBoard(board Board) []*Eval {
	evs := make([]*Eval, d.Count)
	for i := 0; i < d.Count; i++ {
		if d.Active[i] {
			evs[i] = d.Type.New(d.Pockets[i], board.Hi)
			if d.Double {
				evs[i].Double(d.Pockets[i], board.Lo)
			}
		}
	}
	return evs
}

// Board holds boards.
type Board struct {
	Hi []Card
	Lo []Card
}

// Dupe creates a duplicate of the hi and lo portions of the board, excluding
// any eval info collected.
func (board Board) Dupe() Board {
	b := Board{}
	if board.Hi != nil {
		b.Hi = make([]Card, len(board.Hi))
		copy(b.Hi, board.Hi)
	}
	if board.Lo != nil {
		b.Lo = make([]Card, len(board.Lo))
		copy(b.Lo, board.Lo)
	}
	return b
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

// Format satisfies the fmt.Formatter interface.
func (win *Win) Format(f fmt.State, verb rune) {
	switch verb {
	case 'S':
		var v []string
		for i := 0; i < win.Pivot; i++ {
			var desc *Desc
			if !win.Low {
				desc = win.Evals[win.Order[i]].HiDesc()
			} else {
				desc = win.Evals[win.Order[i]].LoDesc()
			}
			v = append(v, fmt.Sprintf("%v", desc.Best))
		}
		fmt.Fprint(f, strings.Join(v, ", "))
	case 's':
		var v []string
		for i := 0; i < win.Pivot; i++ {
			v = append(v, strconv.Itoa(win.Order[i]))
		}
		fmt.Fprint(f, strings.Join(v, ", ")+" "+win.Verb())
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

package cardrank

import (
	"fmt"
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
	// DeckFrench is deck of French (52) cards.
	DeckFrench = DeckType(Two)
	// DeckShort is a deck of Short (6+) cards.
	DeckShort = DeckType(Six)
	// DeckManila is a deck of Manila (7+) cards.
	DeckManila = DeckType(Seven)
	// DeckRoyal is a deck of Royal (10+) cards.
	DeckRoyal = DeckType(Ten)
)

// String satisfies the fmt.Stringer interface.
func (typ DeckType) String() string {
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

// Unshuffled returns a set of unshuffled cards for
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

// New returns a new deck for the deck type.
func (typ DeckType) New() *Deck {
	var v []Card
	switch typ {
	case DeckFrench:
		v = unshuffledFrench
	case DeckShort:
		v = unshuffledShort
	case DeckManila:
		v = unshuffledManila
	case DeckRoyal:
		v = unshuffledRoyal
	default:
		return nil
	}
	n := len(v)
	d := &Deck{
		v: make([]Card, n),
		l: n,
	}
	copy(d.v, v)
	return d
}

// unshuffled cards.
var (
	unshuffledFrench []Card
	unshuffledShort  []Card
	unshuffledManila []Card
	unshuffledRoyal  []Card
)

func init() {
	unshuffledFrench = DeckFrench.Unshuffled()
	unshuffledShort = DeckShort.Unshuffled()
	unshuffledManila = DeckManila.Unshuffled()
	unshuffledRoyal = DeckRoyal.Unshuffled()
}

// Deck is a set of playing cards.
type Deck struct {
	i int
	l int
	v []Card
}

// NewDeck returns a new deck of 52 unshuffled cards.
func NewDeck() *Deck {
	return DeckFrench.New()
}

// NewShoeDeck creates a new unshuffled deck "shoe" composed of n decks of
// unshuffled cards.
func NewShoeDeck(n int) *Deck {
	cards := make([]Card, len(unshuffledFrench)*n)
	for i := 0; i < n; i++ {
		copy(cards[i*len(unshuffledFrench):], unshuffledFrench)
	}
	return &Deck{
		l: len(cards),
		v: cards,
	}
}

// SetLimit sets a limit for the deck.
//
// Useful when using a card deck "shoe" composed of more than one deck of
// cards.
func (d *Deck) SetLimit(limit int) {
	d.l = limit
}

// Shuffle shuffles the deck's cards using the provided shuffler.
func (d *Deck) Shuffle(shuffler Shuffler) {
	shuffler.Shuffle(len(d.v), func(i, j int) {
		d.v[i], d.v[j] = d.v[j], d.v[i]
	})
}

// ShuffleN shuffles the deck's cards, n times, using the provided shuffler.
func (d *Deck) ShuffleN(shuffler Shuffler, n int) {
	for m := 0; m < n; m++ {
		shuffler.Shuffle(len(d.v), func(i, j int) {
			d.v[i], d.v[j] = d.v[j], d.v[i]
		})
	}
}

// Draw draws the next n cards from the top (front) of the deck.
func (d *Deck) Draw(n int) []Card {
	if n < 0 {
		return nil
	}
	var hand []Card
	for l := min(d.i+n, d.l); d.i < l; d.i++ {
		hand = append(hand, d.v[d.i])
	}
	return hand
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

// Deal draws one card successively for each hand until each hand has n cards.
func (d *Deck) Deal(hands, n int) [][]Card {
	// deal pockets
	pockets := make([][]Card, hands)
	for i := 0; i < hands; i++ {
		pockets[i] = make([]Card, n)
	}
	for j := 0; j < n; j++ {
		for i := 0; i < hands; i++ {
			pockets[i][j] = d.Draw(1)[0]
		}
	}
	return pockets
}

// Board draws board cards by discarding discard cards, and drawing count cards each
// for each count in counts.
func (d *Deck) Board(discard int, counts ...int) []Card {
	var board []Card
	for _, count := range counts {
		_ = d.Draw(discard)
		board = append(board, d.Draw(count)...)
	}
	return board
}

// MultiBoard draws n boards of cards, discarding cards, and drawing count
// cards for each count in counts.
func (d *Deck) MultiBoard(n int, discard int, counts ...int) [][]Card {
	boards := make([][]Card, n)
	for _, count := range counts {
		for i := 0; i < n; i++ {
			_ = d.Draw(discard)
			boards[i] = append(boards[i], d.Draw(count)...)
		}
	}
	return boards
}

// DealFor deals hands for the type.
func (d *Deck) DealFor(typ Type, hands int) ([][]Card, []Card) {
	return NewShuffledDealer(typ.Desc(), d).DealAll(hands)
}

// Holdem draws hands for Texas Holdem, returning the set of pockets (one per
// hand) and board cards. Deals 1 card per player until each player has 2
// pocket cards, then discards a card, deals 3 board cards, discards another,
// deals another board card, discards another, and deals a final card to the
// board.
func (d *Deck) Holdem(hands int) ([][]Card, []Card) {
	return d.DealFor(Holdem, hands)
}

// Omaha draws hands for Omaha, returning the set of pockets (one per hand) and
// board cards.
func (d *Deck) Omaha(hands int) ([][]Card, []Card) {
	return d.DealFor(Omaha, hands)
}

// Stud draws hands for Stud, returning the sets of pockets (one per hand).
// Deals no board cards.
func (d *Deck) Stud(hands int) ([][]Card, []Card) {
	return d.DealFor(Stud, hands)
}

// Badugi draws hands for Badugi, returning the sets of pockets (one per hand).
// Deals no board cards.
func (d *Deck) Badugi(hands int) ([][]Card, []Card) {
	return d.DealFor(Badugi, hands)
}

// Dealer is a deck and street iterator.
type Dealer struct {
	TypeDesc
	d *Deck
	i int
}

// NewDealer creates a new dealer.
func NewDealer(desc TypeDesc, shuffler Shuffler, n int) *Dealer {
	d := desc.Type.Deck()
	d.ShuffleN(shuffler, n)
	return &Dealer{
		TypeDesc: desc,
		d:        d,
		i:        -1,
	}
}

// NewShuffledDealer creates a new dealer for an already shuffled deck.
func NewShuffledDealer(desc TypeDesc, d *Deck) *Dealer {
	return &Dealer{
		TypeDesc: desc,
		d:        d,
		i:        -1,
	}
}

// Format satisfies the fmt.Formatter interface.
func (d *Dealer) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		desc := d.Street()
		var v []string
		if 0 < desc.Pocket {
			if 0 < desc.PocketDiscard {
				v = append(v, fmt.Sprintf("D: %d", desc.PocketDiscard))
			}
			v = append(v, fmt.Sprintf("p: %d", desc.Pocket))
			if 0 < desc.PocketUp {
				v = append(v, fmt.Sprintf("u: %d", desc.PocketUp))
			}
		}
		if 0 < desc.Board {
			if 0 < desc.BoardDiscard {
				v = append(v, fmt.Sprintf("d: %d", desc.BoardDiscard))
			}
			v = append(v, fmt.Sprintf("b: %d", desc.Board))
		}
		var s string
		if len(v) != 0 {
			s = " (" + strings.Join(v, ", ") + ")"
		}
		fmt.Fprintf(f, "%d:%q %s%s", d.i, desc.Id, desc.Name, s)
	}
}

// All returns a copy of all cards in the deck, without advancing.
func (d *Dealer) All() []Card {
	return d.d.All()
}

// Next returns true when there are more betting streets defined.
func (d *Dealer) Next() bool {
	d.i++
	return d.i < len(d.Streets)
}

// Street returns the current street.
func (d *Dealer) Street() StreetDesc {
	return d.Streets[d.i]
}

// Pocket returns the current street pocket.
func (d *Dealer) Pocket() int {
	return d.Streets[d.i].Pocket
}

// Board returns the current street board.
func (d *Dealer) Board() int {
	return d.Streets[d.i].Board
}

// Deal deals cards for the street.
func (d *Dealer) Deal(pockets [][]Card, board []Card, hands int) ([][]Card, []Card) {
	return d.DealPockets(pockets, hands, true), d.DealBoard(board, true)
}

// DealPockets deals and appends pockets, returning the appended slice.
func (d *Dealer) DealPockets(pockets [][]Card, hands int, discard bool) [][]Card {
	if p := d.Streets[d.i].Pocket; 0 < p {
		if n := d.Streets[d.i].PocketDiscard; discard && 0 < n {
			_ = d.d.Draw(n)
		}
		if pockets == nil {
			pockets = make([][]Card, hands)
		}
		for j := 0; j < p; j++ {
			for i := 0; i < hands; i++ {
				pockets[i] = append(pockets[i], d.d.Draw(1)[0])
			}
		}
	}
	return pockets
}

// DealBoard deals and appends the board, returning the appended slice.
func (d *Dealer) DealBoard(board []Card, discard bool) []Card {
	if p := d.Streets[d.i].Board; 0 < p {
		if n := d.Streets[d.i].BoardDiscard; discard && 0 < n {
			_ = d.d.Draw(n)
		}
		board = append(board, d.d.Draw(p)...)
	}
	return board
}

// Reset resets the iterator to i.
func (d *Dealer) Reset() {
	d.d.Reset()
	d.i = -1
}

// DealAll deals all pockets, board for the hands. Resets the dealer and the
// deck.
func (d *Dealer) DealAll(hands int) ([][]Card, []Card) {
	d.Reset()
	var pockets [][]Card
	var board []Card
	for d.Next() {
		pockets, board = d.Deal(pockets, board, hands)
	}
	return pockets, board
}

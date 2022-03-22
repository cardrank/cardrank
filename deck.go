package cardrank

const (
	// UnshuffledSize is the unshuffled deck size.
	UnshuffledSize = 52
	// UnshuffledShortSize is the unshuffled short deck size.
	UnshuffledShortSize = 36
)

// unshuffled is an unshuffled set of cards.
var unshuffled = Unshuffled()

// unshuffledShortDeck is an unshuffled set of shortdeck (6+) cards.
var unshuffledShortDeck = UnshuffledShortDeck()

// Unshuffled generates an unshuffled set of standard playing cards.
func Unshuffled() []Card {
	v := make([]Card, UnshuffledSize)
	var i int
	for _, s := range []Suit{Spade, Heart, Diamond, Club} {
		for r := Two; r <= Ace; r++ {
			v[i] = New(r, s)
			i++
		}
	}
	return v
}

// UnshuffledShortDeck generates an unshuffled set of short deck cards (ie,
// excluding card ranks 2 through 5).
func UnshuffledShortDeck() []Card {
	v := make([]Card, UnshuffledShortSize)
	var i int
	for _, s := range []Suit{Spade, Heart, Diamond, Club} {
		for r := Six; r <= Ace; r++ {
			v[i] = New(r, s)
			i++
		}
	}
	return v
}

// UnshuffledExclude generates an unshuffled set of cards, with excluded
// cards removed.
func UnshuffledExclude(exclude []Card) []Card {
	m := make(map[uint32]bool)
	for _, c := range exclude {
		m[uint32(c)] = true
	}
	var v []Card
	for _, s := range []Suit{Spade, Heart, Diamond, Club} {
		for r := Two; r <= Ace; r++ {
			if c := New(r, s); !m[uint32(c)] {
				v = append(v, c)
			}
		}
	}
	return v
}

// Deck is a set of playing cards.
type Deck struct {
	i uint16
	l uint16
	v []Card
}

// NewDeck creates a new deck of cards. If no cards are provided, then a deck
// will be created using the standard unshuffled cards.
func NewDeck(cards ...Card) *Deck {
	if cards == nil {
		cards = unshuffled
	}
	d := &Deck{
		v: make([]Card, len(cards)),
		l: uint16(len(cards)),
	}
	copy(d.v, cards)
	return d
}

// NewShortDeck creates a new deck of short cards.
func NewShortDeck() *Deck {
	return NewDeck(unshuffledShortDeck...)
}

// NewDeckShoe creates a card deck "shoe" composed of n decks of
// unshuffled cards.
func NewDeckShoe(n int) *Deck {
	cards := make([]Card, len(unshuffled)*n)
	for i := 0; i < n; i++ {
		copy(cards[i*len(unshuffled):], unshuffled)
	}
	return &Deck{
		l: uint16(len(cards)),
		v: cards,
	}
}

// SetLimit sets a limit for the deck.
//
// Useful when using a card deck "shoe" composed of more than one deck of
// cards.
func (d *Deck) SetLimit(limit int) {
	d.l = uint16(limit)
}

// Shuffle shuffles the deck's cards using f (same interface as
// math/rand.Shuffle).
func (d *Deck) Shuffle(f func(int, func(i, j int))) {
	f(len(d.v), func(i, j int) {
		d.v[i], d.v[j] = d.v[j], d.v[i]
	})
}

// ShuffleN shuffles the deck's cards, n times, using f (same interface as
// math/rand.Shuffle).
func (d *Deck) ShuffleN(f func(n int, swap func(i, j int)), n int) {
	for i := 0; i < n; i++ {
		d.Shuffle(f)
	}
}

// Draw draws the next n cards from the top (front) of the deck.
func (d *Deck) Draw(n int) []Card {
	if n < 0 {
		panic("n cannot be negative")
	}
	var hand []Card
	for l := min(d.i+uint16(n), d.l); d.i < l; d.i++ {
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
	if n := int(d.l) - int(d.i); 0 <= n {
		return n
	}
	return 0
}

// Deal draws one card successively for each hand until each hand has n cards.
func (d *Deck) Deal(hands, n int) [][]Card {
	// deal pockets
	pockets := make([][]Card, hands)
	for i := 0; i < n*hands; i++ {
		if i%n == 0 {
			pockets[i/n] = make([]Card, n)
		}
		pockets[i/n][i%n] = d.Draw(1)[0]
	}
	return pockets
}

// Board draws board cards by discarding a card and drawing n cards for each n
// in counts.
func (d *Deck) Board(counts ...int) []Card {
	var board []Card
	for _, n := range counts {
		board = append(board, d.Draw(n)[1:]...)
	}
	return board
}

// Simple draws board cards and hands of n cards. Useful for examples.
func (d *Deck) Simple(board, hands, n int) ([][]Card, []Card) {
	b := d.Draw(board)
	pockets := make([][]Card, hands)
	for i := 0; i < hands; i++ {
		pockets[i] = d.Draw(n)
	}
	return pockets, b
}

// Holdem draws hands for Texas Holdem, returning the set of pockets (one per
// hand) and board cards. Deals 1 card per player until each player has 2
// pocket cards, then discards a card, deals 3 board cards, discards another,
// deals another board card, discards another, and deals a final card to the
// board.
func (d *Deck) Holdem(hands int) ([][]Card, []Card) {
	return d.Deal(hands, 2), d.Board(4, 2, 2)
}

// HoldemSimple draws hands for Texas Holdem, returning the set of pockets (one
// per hand) and board cards. Useful for examples. Deals 5 board cards prior to
// dealing pocket cards for each hand.
func (d *Deck) HoldemSimple(hands int) ([][]Card, []Card) {
	return d.Simple(5, hands, 2)
}

// Omaha draws hands for Omaha, returning the set of pockets (one per hand) and
// board cards.
func (d *Deck) Omaha(hands int) ([][]Card, []Card) {
	return d.Deal(hands, 4), d.Board(4, 2, 2)
}

// OmahaSimple draws hands for Omaha, returning the set of pockets (one per
// hand) and board cards. Useful for examples. Deals 5 board cards prior to
// dealing pocket cards for each hand.
func (d *Deck) OmahaSimple(hands int) ([][]Card, []Card) {
	return d.Simple(5, hands, 4)
}

package cardrank

import (
	"fmt"
	"sort"
)

// Hand is a poker hand.
type Hand struct {
	typ    Type
	pocket []Card
	board  []Card
	rank   HandRank
	hand   []Card
	best   []Card
	unused []Card
}

// NewHandOf creates a new poker hand of the specified type.
func NewHandOf(typ Type, pocket, board []Card, f func([]Card) HandRank) *Hand {
	h := &Hand{
		typ:    typ,
		pocket: make([]Card, len(pocket)),
		board:  make([]Card, len(board)),
	}
	copy(h.pocket, pocket)
	copy(h.board, board)
	unused := h.eval(f)
	high := Five
	if typ == ShortDeck {
		high = Nine
	}
	switch h.rank.Fixed() {
	case StraightFlush:
		h.best, h.unused = bestStraightFlush(h.hand, high)
	case FourOfAKind:
		h.best, h.unused = bestSet(h.hand)
	case FullHouse:
		h.best, h.unused = bestSet(h.hand)
	case Flush:
		h.best, h.unused = bestFlush(h.hand)
	case Straight:
		h.best, h.unused = bestStraight(h.hand, high)
	case ThreeOfAKind:
		h.best, h.unused = bestSet(h.hand)
	case TwoPair:
		h.best, h.unused = bestSet(h.hand)
	case Pair:
		h.best, h.unused = bestSet(h.hand)
	case Nothing:
		h.best, h.unused = h.hand[:5], h.hand[5:]
	default:
		panic("invalid card rank")
	}
	if Omaha <= typ && typ < Stud {
		h.unused = unused
	}
	return h
}

// NewHand creates a new poker hand.
func NewHand(pocket, board []Card, f func([]Card) HandRank) *Hand {
	typ := Holdem
	if len(pocket) == 4 {
		typ = Omaha
	}
	return NewHandOf(typ, pocket, board, f)
}

// eval evaluates the poker hand rank for the hands best ranking cards chosen
// from the hand's board/pocket using f. See Type for evaluation rules.
func (h *Hand) eval(f func([]Card) HandRank) []Card {
	var unused []Card
	h.rank, h.hand, unused = h.typ.Best(h.pocket, h.board, f)
	// order hand high to low
	sort.Slice(h.hand, func(i, j int) bool {
		m, n := h.hand[i].Rank(), h.hand[j].Rank()
		if m == n {
			return h.hand[i].Suit() > h.hand[j].Suit()
		}
		return m > n
	})
	return unused
}

// Type returns the poker hand's type.
func (h *Hand) Type() Type {
	return h.typ
}

// Pocket returns the poker hand's pocket.
func (h *Hand) Pocket() []Card {
	return h.pocket
}

// Board returns the poker hand's board.
func (h *Hand) Board() []Card {
	return h.board
}

// Rank returns the poker hand's rank.
func (h *Hand) Rank() HandRank {
	return h.rank
}

// Fixed returns the poker hand's fixed rank.
func (h *Hand) Fixed() HandRank {
	return h.rank.Fixed()
}

// Hand returns the poker hand's hand.
func (h *Hand) Hand() []Card {
	return h.hand
}

// Best returns the poker hand's best-five cards.
func (h *Hand) Best() []Card {
	return h.best
}

// Unused returns the hand's unused cards.
func (h *Hand) Unused() []Card {
	return h.unused
}

// Format satisfies the fmt.Formatter interface.
func (h *Hand) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		fmt.Fprintf(f, "%s %s", h.Description(), h.best)
	case 'q':
		fmt.Fprintf(f, "\"%s %s\"", h.Description(), h.best)
	case 'S':
		fmt.Fprintf(f, "%s %S", h.Description(), HandFormatter(h.best))
	case 'b':
		fmt.Fprintf(f, "%s %b", h.Description(), h.best)
	case 'h':
		fmt.Fprintf(f, "%s %h", h.Description(), HandFormatter(h.best))
	case 'c':
		fmt.Fprintf(f, "%s %c", h.Description(), h.best)
	case 'C':
		fmt.Fprintf(f, "%s %C", h.Description(), HandFormatter(h.best))
	case 'f':
		for _, c := range h.best {
			c.Format(f, 's')
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, hand: %s/%s)", verb, h.best, h.unused)
	}
}

// Description describes the hand's best-five cards.
//
// Examples:
//
//	Straight Flush, Ace-high, Royal
//	Straight Flush, King-high
//	Straight Flush, Five-high, Steel Wheel
//	Four of a Kind, Nines, kicker Jack
//	Full House, Sixes full of Fours
//	Flush, Ten-high
//	Straight, Eight-high
//	Three of a Kind, Fours, kickers Ace, King
//	Two Pair, Nines over Sixes, kicker Jack
//	Pair, Aces, kickers King, Queen, Nine
//	Nothing, Seven-high, kickers Six, Five, Three, Two
//
func (h *Hand) Description() string {
	switch h.rank.Fixed() {
	case StraightFlush:
		switch r := h.best[0].Rank(); {
		case r == Ace:
			return fmt.Sprintf("Straight Flush, %N-high, Royal", h.best[0])
		case r == Nine && h.typ == ShortDeck:
			return fmt.Sprintf("Straight Flush, %N-high, Iron Maiden", h.best[0])
		case r == Five:
			return fmt.Sprintf("Straight Flush, %N-high, Steel Wheel", h.best[0])
		}
		return fmt.Sprintf("Straight Flush, %N-high", h.best[0])
	case FourOfAKind:
		return fmt.Sprintf("Four of a Kind, %P, kicker %N", h.best[0], h.best[4])
	case FullHouse:
		return fmt.Sprintf("Full House, %P full of %P", h.best[0], h.best[3])
	case Flush:
		return fmt.Sprintf("Flush, %N-high", h.best[0])
	case Straight:
		return fmt.Sprintf("Straight, %N-high", h.best[0])
	case ThreeOfAKind:
		return fmt.Sprintf("Three of a Kind, %P, kickers %N, %N", h.best[0], h.best[3], h.best[4])
	case TwoPair:
		return fmt.Sprintf("Two Pair, %P over %P, kicker %N", h.best[0], h.best[2], h.best[4])
	case Pair:
		return fmt.Sprintf("Pair, %P, kickers %N, %N, %N", h.best[0], h.best[2], h.best[3], h.best[4])
	}
	return fmt.Sprintf("Nothing, %N-high, kickers %N, %N, %N, %N", h.best[0], h.best[1], h.best[2], h.best[3], h.best[4])
}

// Compare compares the hand ranks.
func (h *Hand) Compare(b *Hand) int {
	switch hf, bf := h.rank.Fixed(), b.rank.Fixed(); {
	case h.typ == ShortDeck && hf == Flush && bf == FullHouse:
		return -1
	case h.typ == ShortDeck && hf == FullHouse && bf == Flush:
		return +1
	case h.rank < b.rank:
		return -1
	case b.rank < h.rank:
		return +1
	}
	return 0
}

// HandFormatter wraps formatting a set of cards. Allows `go test` to function
// without disabling vet.
type HandFormatter []Card

// Format satisfies the fmt.Formatter interface.
func (hand HandFormatter) Format(f fmt.State, verb rune) {
	_, _ = f.Write([]byte{'['})
	for i, c := range hand {
		if i != 0 {
			_, _ = f.Write([]byte{' '})
		}
		c.Format(f, verb)
	}
	_, _ = f.Write([]byte{']'})
}

// bestStraightFlush returns the best-five straight flush in the hand.
func bestStraightFlush(hand []Card, high Rank) ([]Card, []Card) {
	v := orderSuits(hand)
	var b, d []Card
	for _, c := range hand {
		switch c.Suit() {
		case v[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	e, f := bestStraight(b, high)
	e = append(e, append(d, f...)...)
	return e[:5], e[5:]
}

// bestFlush returns the best-five flush in the hand.
func bestFlush(hand []Card) ([]Card, []Card) {
	v := orderSuits(hand)
	var b, d []Card
	for _, c := range hand {
		switch c.Suit() {
		case v[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	b = append(b, d...)
	return b[:5], b[5:]
}

// bestStraight returns the best-five straight in the hand.
func bestStraight(hand []Card, high Rank) ([]Card, []Card) {
	m := make(map[Rank][]Card)
	for _, c := range hand {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	var b []Card
	for i := Ace; i >= high; i-- {
		// last card index
		j := i - Six
		// check ace
		if i == high {
			j = Ace
		}
		if m[i] != nil && m[i-1] != nil && m[i-2] != nil && m[i-3] != nil && m[j] != nil {
			// collect b, removing from m
			b = []Card{m[i][0], m[i-1][0], m[i-2][0], m[i-3][0], m[j][0]}
			m[i] = m[i][1:]
			m[i-1] = m[i-1][1:]
			m[i-2] = m[i-2][1:]
			m[i-3] = m[i-3][1:]
			m[j] = m[j][1:]
			break
		}
	}
	// collect remaining
	var d []Card
	for i := int(Ace); i >= 0; i-- {
		if _, ok := m[Rank(i)]; ok && m[Rank(i)] != nil {
			d = append(d, m[Rank(i)]...)
		}
	}
	b = append(b, d...)
	return b[:5], b[5:]
}

// bestSet returns the best matching sets in the hand.
func bestSet(hand []Card) ([]Card, []Card) {
	v := orderRanks(hand)
	var a, b, d []Card
	for _, c := range hand {
		switch c.Rank() {
		case v[0]:
			a = append(a, c)
		case v[1]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	b = append(a, append(b, d...)...)
	return b[:5], b[5:]
}

// orderSuits order's a hand's card suits by count.
func orderSuits(hand []Card) []Suit {
	m := make(map[Suit]int)
	var v []Suit
	for _, c := range hand {
		s := c.Suit()
		if _, ok := m[s]; !ok {
			v = append(v, s)
		}
		m[s]++
	}
	sort.Slice(v, func(i, j int) bool {
		if m[v[i]] == m[v[j]] {
			return v[i] > v[j]
		}
		return m[v[i]] > m[v[j]]
	})
	return v
}

// orderRanks orders a hand's card ranks by count.
func orderRanks(hand []Card) []Rank {
	m := make(map[Rank]int)
	var v []Rank
	for _, c := range hand {
		r := c.Rank()
		if _, ok := m[r]; !ok {
			v = append(v, r)
		}
		m[r]++
	}
	sort.Slice(v, func(i, j int) bool {
		if m[v[i]] == m[v[j]] {
			return v[i] > v[j]
		}
		return m[v[i]] > m[v[j]]
	})
	return v
}

// OrderHands orders hands by rank, low to high, returning 'pivot' of winning
// vs losing hands.
func OrderHands(hands []*Hand) ([]int, int) {
	i, n := 0, len(hands)
	m, h := make(map[int]*Hand, n), make([]int, n)
	for ; i < n; i++ {
		m[i], h[i] = hands[i], i
	}
	sort.SliceStable(h, func(j, k int) bool {
		return m[h[j]].Compare(m[h[k]]) < 0
	})
	for i = 1; i < n && m[h[i-1]].Rank() == m[h[i]].Rank(); i++ {
	}
	return h, i
}

package cardrank

import (
	"fmt"
	"sort"
)

// Type is a hand type.
type Type uint32

// Hand types.
const (
	Holdem Type = iota
	Short
	Royal
	Omaha
	OmahaHiLo
	Stud
	StudHiLo
	Razz
	Badugi // FIXME: not yet available
)

// String satisfies the fmt.Stringer interface.
func (typ Type) String() string {
	switch typ {
	case Holdem:
		return "Holdem"
	case Short:
		return "Short"
	case Royal:
		return "Royal"
	case Omaha:
		return "Omaha"
	case OmahaHiLo:
		return "OmahaHiLo"
	case Stud:
		return "Stud"
	case StudHiLo:
		return "StudHiLo"
	case Razz:
		return "Razz"
	case Badugi:
		return "Badugi"
	}
	return fmt.Sprintf("Type(%d)", typ)
}

// NewDeck returns a new deck for the type.
func (typ Type) Deck() *Deck {
	switch typ {
	case Short:
		return NewShortDeck()
	case Royal:
		return NewRoyalDeck()
	}
	return NewDeck()
}

// DealShuffle creates a new deck for the type and shuffles the deck count
// number of times, returning the pockets and board for the number of hands
// specified.
func (typ Type) DealShuffle(shuffle func(int, func(int, int)), count, hands int) ([][]Card, []Card) {
	d := typ.Deck()
	for i := 0; i < count; i++ {
		d.Shuffle(shuffle)
	}
	switch typ {
	case Holdem, Short, Royal:
		return d.Holdem(hands)
	case Omaha, OmahaHiLo:
		return d.Omaha(hands)
	case Stud, StudHiLo, Razz:
		return d.Stud(hands)
	case Badugi:
		return d.Badugi(hands)
	}
	return nil, nil
}

// Deal creates a new deck for the type, shuffling it once, returning the
// pockets and board for the number of hands specified.
//
// Use DealShuffle when needing to shuffle the deck more than once.
func (typ Type) Deal(shuffle func(int, func(int, int)), hands int) ([][]Card, []Card) {
	return typ.DealShuffle(shuffle, 1, hands)
}

// RankHand ranks the hand.
func (typ Type) RankHand(pocket, board []Card) *Hand {
	return NewHand(typ, pocket, board)
}

// RankHands ranks the hands.
func (typ Type) RankHands(pockets [][]Card, board []Card) []*Hand {
	hands := make([]*Hand, len(pockets))
	for i := 0; i < len(pockets); i++ {
		hands[i] = typ.RankHand(pockets[i], board)
	}
	return hands
}

// Best returns the best hand and rank for the provided pocket, board. Returns
// the rank of the best available hand, its best cards, and its unused cards.
func (typ Type) Best(pocket, board []Card) (HandRank, []Card, []Card, HandRank, []Card, []Card) {
	switch typ {
	case Holdem, Short, Royal, Stud, StudHiLo:
		rank, best, unused := best(typ, pocket, board)
		if typ == StudHiLo {
			lowRank, lowBest, lowUnused := bestLow(pocket, board, EightOrBetterRanker, eightOrBetterMaxRank)
			if lowRank < eightOrBetterMaxRank {
				return rank, best, unused, lowRank, lowBest, lowUnused
			}
		}
		return rank, best, unused, 0, nil, nil
	case Omaha, OmahaHiLo:
		return bestOmaha(typ, pocket, board)
	case Razz:
		rank, best, unused := bestLow(pocket, board, RazzRanker, ^uint16(0))
		if rank >= lowMaxRank {
			r := Invalid - rank
			switch r.Fixed() {
			case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
				best, _ = bestSet(best)
			default:
				panic("invalid hand rank")
			}
		}
		return rank, best, unused, 0, nil, nil
	}
	panic("invalid type")
}

// MaxPlayers returns the max players for the type.
func (typ Type) MaxPlayers() int {
	switch typ {
	case Holdem, Short, Omaha, OmahaHiLo:
		return 10
	case Royal:
		return 5
	case Stud, StudHiLo, Razz:
		return 7
	case Badugi:
		return 8
	}
	return 0
}

// Hand is a poker hand.
type Hand struct {
	typ       Type
	pocket    []Card
	board     []Card
	rank      HandRank
	best      []Card
	unused    []Card
	lowRank   HandRank
	lowBest   []Card
	lowUnused []Card
}

// NewHand creates a new hand of the specified type.
func NewHand(typ Type, pocket, board []Card) *Hand {
	h := &Hand{
		typ:    typ,
		pocket: make([]Card, len(pocket)),
		board:  make([]Card, len(board)),
	}
	copy(h.pocket, pocket)
	copy(h.board, board)
	h.rank, h.best, h.unused, h.lowRank, h.lowBest, h.lowUnused = typ.Best(h.pocket, h.board)
	return h
}

// LowValid returns true if is a valid low hand.
func (h *Hand) LowValid() bool {
	return len(h.lowBest) != 0
}

// Pocket returns the hand's pocket.
func (h *Hand) Pocket() []Card {
	return h.pocket
}

// Board returns the hand's board.
func (h *Hand) Board() []Card {
	return h.board
}

// Rank returns the hand's rank.
func (h *Hand) Rank() HandRank {
	return h.rank
}

// Fixed returns the hand's fixed rank.
func (h *Hand) Fixed() HandRank {
	return h.rank.Fixed()
}

// Best returns the hand's best-five cards.
func (h *Hand) Best() []Card {
	return h.best
}

// Unused returns the hand's unused cards.
func (h *Hand) Unused() []Card {
	return h.unused
}

// LowRank returns the hand's low rank.
func (h Hand) LowRank() HandRank {
	return h.lowRank
}

// LowBest returns the hand's best-five low cards.
func (h Hand) LowBest() []Card {
	return h.lowBest
}

// LowUnused returns the poker hand's unused-five low cards.
func (h Hand) LowUnused() []Card {
	return h.lowUnused
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
	r := h.rank
	switch {
	case h.typ == Razz && h.rank < lowMaxRank:
		return h.best[0].Rank().Name() + "-low"
	case h.typ == Razz:
		r = Invalid - r
	}
	switch r.Fixed() {
	case StraightFlush:
		switch r := h.best[0].Rank(); {
		case r == Ace:
			return fmt.Sprintf("Straight Flush, %N-high, Royal", h.best[0])
		case r == Nine && h.typ == Short:
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

// LowDescription describes the hands best-five low cards.
func (h *Hand) LowDescription() string {
	if len(h.lowBest) == 0 {
		return "None"
	}
	return h.lowBest[0].Rank().Name() + "-low"
}

// Compare compares the hand ranks.
func (h *Hand) Compare(b *Hand) int {
	switch hf, bf := h.rank.Fixed(), b.rank.Fixed(); {
	case h.typ == Short && hf == Flush && bf == FullHouse:
		return -1
	case h.typ == Short && hf == FullHouse && bf == Flush:
		return +1
	case h.rank < b.rank:
		return -1
	case b.rank < h.rank:
		return +1
	}
	return 0
}

// LowCompare compares the low hand ranks.
func (h *Hand) LowCompare(b *Hand) int {
	switch {
	case h.lowRank == 0 && b.lowRank != 0:
		return +1
	case b.lowRank == 0 && h.lowRank != 0:
		return -1
	case h.lowRank < b.lowRank:
		return -1
	case b.lowRank > h.lowRank:
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

// bestOmaha returns the best omaha hand hi/lo hands for pocket, board.
func bestOmaha(typ Type, pocket, board []Card) (HandRank, []Card, []Card, HandRank, []Card, []Card) {
	hand := make([]Card, 5)
	var r HandRank
	rank, best, unused := Invalid, make([]Card, 5), make([]Card, 4)
	lowRank, lowBest, lowUnused := Invalid, make([]Card, 5), make([]Card, 4)
	for i := 0; i < 6; i++ {
		for j := 0; j < 10; j++ {
			hand[0], hand[1], hand[2], hand[3], hand[4] = pocket[t4c2[i][0]], pocket[t4c2[i][1]], board[t5c3[j][0]], board[t5c3[j][1]], board[t5c3[j][2]]
			if r = DefaultRanker(hand); r < rank {
				rank = r
				copy(best, hand)
				unused[0], unused[1], unused[2], unused[3] = pocket[t4c2[i][2]], pocket[t4c2[i][3]], board[t5c3[j][3]], board[t5c3[j][4]]
			}
			if typ == OmahaHiLo {
				if r = HandRank(EightOrBetterRanker(hand[0], hand[1], hand[2], hand[3], hand[4])); r < lowRank && r < eightOrBetterMaxRank {
					lowRank = r
					copy(lowBest, hand)
					lowUnused[0], lowUnused[1], lowUnused[2], lowUnused[3] = pocket[t4c2[i][2]], pocket[t4c2[i][3]], board[t5c3[j][3]], board[t5c3[j][4]]
				}
			}
		}
	}
	// order best
	sort.Slice(best, func(i, j int) bool {
		m, n := best[i].Rank(), best[j].Rank()
		if m == n {
			return best[i].Suit() > best[j].Suit()
		}
		return m > n
	})
	switch rank.Fixed() {
	case StraightFlush:
		best, _ = bestStraightFlush(best, Five)
	case Flush:
		best, _ = bestFlush(best)
	case Straight:
		best, _ = bestStraight(best, Five)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		best, _ = bestSet(best)
	case Nothing:
	default:
		panic("invalid hand rank")
	}
	if typ == OmahaHiLo && lowRank < eightOrBetterMaxRank {
		sort.Slice(lowBest, func(i, j int) bool {
			return (lowBest[i].Rank()+1)%13 > (lowBest[j].Rank()+1)%13
		})
		return rank, best, unused, lowRank, lowBest, lowUnused
	}
	return rank, best, unused, 0, nil, nil
}

// best returns the best hand for the pocket, board.
func best(typ Type, pocket, board []Card) (HandRank, []Card, []Card) {
	f := DefaultRanker
	if typ == Short {
		f = DefaultSixPlusRanker
	}
	// copy
	hand := make([]Card, len(pocket)+len(board))
	copy(hand, pocket)
	copy(hand[len(pocket):], board)
	rank := f(hand)
	// order hand high to low
	sort.Slice(hand, func(i, j int) bool {
		m, n := hand[i].Rank(), hand[j].Rank()
		if m == n {
			return hand[i].Suit() > hand[j].Suit()
		}
		return m > n
	})
	// determine high for straights
	high := Five
	if typ == Short {
		high = Nine
	}
	var best, unused []Card
	switch rank.Fixed() {
	case StraightFlush:
		best, unused = bestStraightFlush(hand, high)
	case Flush:
		best, unused = bestFlush(hand)
	case Straight:
		best, unused = bestStraight(hand, high)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		best, unused = bestSet(hand)
	case Nothing:
		best, unused = hand[:5], hand[5:]
	default:
		panic("invalid hand rank")
	}
	return rank, best, unused
}

// bestLow uses f to determine the best low hand of a 7 card hand.
func bestLow(pocket, board []Card, f RankFiveFunc, maxRank uint16) (HandRank, []Card, []Card) {
	hand := make([]Card, len(pocket)+len(board))
	copy(hand, pocket)
	copy(hand[len(pocket):], board)
	if len(hand) != 7 {
		panic("bad pocket or board")
	}
	rank, r := uint16(Invalid), uint16(0)
	best, unused := make([]Card, 5), make([]Card, 2)
	for i := 0; i < 21; i++ {
		if r = f(
			hand[t7c5[i][0]],
			hand[t7c5[i][1]],
			hand[t7c5[i][2]],
			hand[t7c5[i][3]],
			hand[t7c5[i][4]],
		); r < rank && r < maxRank {
			rank = r
			best[0], best[1], best[2], best[3], best[4] = hand[t7c5[i][0]], hand[t7c5[i][1]], hand[t7c5[i][2]], hand[t7c5[i][3]], hand[t7c5[i][4]]
			unused[0], unused[1] = hand[t7c5[i][5]], hand[t7c5[i][6]]
		}
	}
	// order
	sort.Slice(best, func(i, j int) bool {
		return (best[i].Rank()+1)%13 > (best[j].Rank()+1)%13
	})
	if rank < maxRank {
		return HandRank(rank), best, unused
	}
	return 0, nil, nil
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

// Order orders hands by rank, low to high, returning 'pivot' of winning vs
// losing hands. Pivot will always be 1 or higher.
func Order(hands []*Hand) ([]int, int) {
	i, n := 0, len(hands)
	m, h := make(map[int]*Hand, n), make([]int, n)
	for ; i < n; i++ {
		m[i], h[i] = hands[i], i
	}
	sort.SliceStable(h, func(j, k int) bool {
		return m[h[j]].Compare(m[h[k]]) < 0
	})
	for i = 1; i < n && m[h[i-1]].rank == m[h[i]].rank; i++ {
	}
	return h, i
}

// LowOrder orders hands by rank, low to high, returning 'pivot' of winning vs
// losing hands. If there are no low hands the pivot will be 0.
func LowOrder(hands []*Hand) ([]int, int) {
	i, n := 0, len(hands)
	m, h := make(map[int]*Hand, n), make([]int, n)
	for ; i < n; i++ {
		m[i], h[i] = hands[i], i
	}
	sort.SliceStable(h, func(j, k int) bool {
		return m[h[j]].LowCompare(m[h[k]]) < 0
	})
	if m[h[0]].lowRank == 0 {
		return nil, 0
	}
	for i = 1; i < n && m[h[i-1]].lowRank == m[h[i]].lowRank; i++ {
	}
	return h, i
}

package cardrank

import (
	"fmt"
)

// ranker is the hand ranker.
var ranker func([]Card) HandRank

func init() {
	if Hybrid.Available() {
		ranker = Hybrid.Rank
	} else {
		ranker = CactusFast.Rank
	}
}

// Type is a type of game.
type Type uint32

// Type values.
const (
	Holdem    Type = 0x0010
	ShortDeck Type = 0x0011
	Omaha     Type = 0x0100
	OmahaHiLo Type = 0x0101
	Stud      Type = 0x1000
	StudHiLo  Type = 0x1001
	Razz      Type = 0x1010
)

// String satisfies the fmt.Stringer interface.
func (typ Type) String() string {
	switch typ {
	case Holdem:
		return "Holdem"
	case ShortDeck:
		return "ShortDeck"
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
	}
	return fmt.Sprintf("Type(%d)", typ)
}

// NewDeck returns a new deck for the type.
func (typ Type) Deck() *Deck {
	if typ == ShortDeck {
		return NewShortDeck()
	}
	return NewDeck()
}

// MultiShuffleDeal shuffles a deck multiple times and deals cards for the type.
func (typ Type) MultiShuffleDeal(shuffle func(int, func(int, int)), count, hands int) ([][]Card, []Card) {
	d := typ.Deck()
	for i := 0; i < count; i++ {
		d.Shuffle(shuffle)
	}
	switch {
	case typ&Holdem != 0:
		return d.Holdem(hands)
	case typ&Omaha != 0:
		return d.Omaha(hands)
	}
	return nil, nil
}

// Deal deals cards for the type.
func (typ Type) Deal(shuffle func(int, func(int, int)), hands int) ([][]Card, []Card) {
	return typ.MultiShuffleDeal(shuffle, 1, hands)
}

// RankHand ranks the hand.
func (typ Type) RankHand(pocket, board []Card) *Hand {
	f := ranker
	if typ == ShortDeck {
		f = CactusFastSixPlus.Rank
	}
	return NewHandOf(typ, pocket, board, f)
}

// RankHands ranks the hands.
func (typ Type) RankHands(pockets [][]Card, board []Card) []*Hand {
	hands := make([]*Hand, len(pockets))
	for i := 0; i < len(pockets); i++ {
		hands[i] = typ.RankHand(pockets[i], board)
	}
	return hands
}

// LowRankHand ranks the low hand.
func (typ Type) LowRankHand(pocket, board []Card) *Hand {
	return NewHandOf(typ, pocket, board, EightOrBetter.Rank)
}

// LowRankHands ranks the hands.
func (typ Type) LowRankHands(pockets [][]Card, board []Card) []*Hand {
	hands := make([]*Hand, len(pockets))
	for i := 0; i < len(pockets); i++ {
		hands[i] = typ.LowRankHand(pockets[i], board)
	}
	return hands
}

// Best returns the best hand and rank for the provided pocket, board using f
// to evaluate ranks of possible hands.
func (typ Type) Best(pocket, board []Card, f func([]Card) HandRank) (HandRank, []Card, []Card) {
	switch {
	case Holdem <= typ && typ < Omaha:
		hand := make([]Card, len(pocket)+len(board))
		copy(hand, pocket)
		copy(hand[len(pocket):], board)
		return f(hand), hand, nil
	case Omaha <= typ && typ < Stud:
		rank, h, best, unused := HandRank(9999), make([]Card, 5), make([]Card, 5), make([]Card, 4)
		for i := 0; i < 6; i++ {
			for j := 0; j < 10; j++ {
				h[0], h[1], h[2], h[3], h[4] = pocket[p4c2[i][0]], pocket[p4c2[i][1]], board[b5c3[j][0]], board[b5c3[j][1]], board[b5c3[j][2]]
				if r := f(h); r < rank {
					rank = r
					copy(best, h)
					unused[0], unused[1], unused[2], unused[3] = pocket[p4c2[i][2]], pocket[p4c2[i][3]], board[b5c3[j][3]], board[b5c3[j][4]]
				}
			}
		}
		return rank, best, unused
	}
	return Invalid, nil, nil
}

// Max returns the max players for the type.
func (typ Type) Max() int {
	return 10
}

// p4c2 is used for choosing 2 of 4 pocket cards.
var p4c2 = [6][4]int{
	{0, 1, 2, 3},
	{0, 2, 1, 3},
	{0, 3, 1, 2},
	{1, 2, 0, 3},
	{1, 3, 0, 2},
	{2, 3, 0, 1},
}

// b5c3 is used for choosing 3 of 5 board cards.
var b5c3 = [10][5]int{
	{0, 1, 2, 3, 4},
	{0, 1, 3, 2, 4},
	{0, 1, 4, 2, 3},
	{0, 2, 3, 1, 4},
	{0, 2, 4, 1, 3},
	{0, 3, 4, 1, 2},
	{1, 2, 3, 0, 4},
	{1, 2, 4, 0, 3},
	{1, 3, 4, 0, 2},
	{2, 3, 4, 0, 1},
}

//go:build !portable && !embedded

package cardrank

import (
	"bytes"
	_ "embed"
	"encoding/binary"
)

func init() {
	twoPlus = NewTwoPlusRanker()
}

// TwoPlusRanker is a two plus hand ranker.
type TwoPlusRanker struct {
	ranks []uint32
	cards map[Card]uint32
	types [10]uint32
}

// TwoPlusRanker creates a new two plus hand ranker.
func NewTwoPlusRanker() RankerFunc {
	var buf []byte
	for _, v := range [][]byte{
		handranks00,
		handranks01,
		handranks02,
		handranks03,
		handranks04,
		handranks05,
		handranks06,
		handranks07,
		handranks08,
		handranks09,
		handranks10,
		handranks11,
		handranks12,
	} {
		buf = append(buf, v...)
	}
	if len(buf)%4 != 0 || len(buf)/4 != 32487834 {
		panic("invalid file")
	}
	ranks := make([]uint32, len(buf)/4)
	if err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, ranks); err != nil {
		panic(err)
	}
	// build cards
	cards := make(map[Card]uint32, 52)
	for i, r := uint32(0), Two; r <= Ace; r++ {
		for _, s := range []Suit{Spade, Heart, Club, Diamond} {
			cards[New(r, s)] = i + 1
			i++
		}
	}
	p := &TwoPlusRanker{
		ranks: ranks,
		cards: cards,
		types: [10]uint32{
			uint32(Invalid),
			uint32(HighCard),
			uint32(Pair),
			uint32(TwoPair),
			uint32(ThreeOfAKind),
			uint32(Straight),
			uint32(Flush),
			uint32(FullHouse),
			uint32(FourOfAKind),
			uint32(StraightFlush),
		},
	}
	return p.rank
}

// rank satisfies the Ranker interface.
func (p *TwoPlusRanker) rank(hand []Card) HandRank {
	i := uint32(53)
	for _, c := range hand {
		i = p.ranks[i+p.cards[c]]
	}
	if len(hand) < 7 {
		i = p.ranks[i]
	}
	return HandRank(p.types[i>>12] - i&0xfff + 1)
}

//go:embed handranks00.dat
var handranks00 []byte

//go:embed handranks01.dat
var handranks01 []byte

//go:embed handranks02.dat
var handranks02 []byte

//go:embed handranks03.dat
var handranks03 []byte

//go:embed handranks04.dat
var handranks04 []byte

//go:embed handranks05.dat
var handranks05 []byte

//go:embed handranks06.dat
var handranks06 []byte

//go:embed handranks07.dat
var handranks07 []byte

//go:embed handranks08.dat
var handranks08 []byte

//go:embed handranks09.dat
var handranks09 []byte

//go:embed handranks10.dat
var handranks10 []byte

//go:embed handranks11.dat
var handranks11 []byte

//go:embed handranks12.dat
var handranks12 []byte

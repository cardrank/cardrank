//go:build !portable && !embedded

package cardrank

import (
	"bytes"
	_ "embed"
	"encoding/binary"
)

func init() {
	twoPlusTwo = NewTwoPlusTwoRanker()
}

// TwoPlusTwoRanker is a implementation of the 2+2 poker forum hand ranker.
// Uses the embedded twoplustwo*.dat files to provide extremely fast 7 card
// hand lookup. Uses Cactus Kev values.
//
// The lookup table is contained in the 'twoplustwo*.dat' files, and were
// broken up from a single file to get around GitHub's size limitations. Files
// were generated with 'gen.go', which is a pure-Go implementation of the code
// generator available at: https://github.com/tangentforks/TwoPlusTwoHandEvaluator
//
// When recombined, the lookup table has the same hash as the original table
// generated using the C code.
type TwoPlusTwoRanker struct {
	ranks []uint32
	cards map[Card]uint32
	types [10]uint32
}

// TwoPlusRanker creates a new two plus hand ranker.
func NewTwoPlusTwoRanker() RankerFunc {
	var buf []byte
	for _, v := range [][]byte{
		twoplustwo00Dat,
		twoplustwo01Dat,
		twoplustwo02Dat,
		twoplustwo03Dat,
		twoplustwo04Dat,
		twoplustwo05Dat,
		twoplustwo06Dat,
		twoplustwo07Dat,
		twoplustwo08Dat,
		twoplustwo09Dat,
		twoplustwo10Dat,
		twoplustwo11Dat,
		twoplustwo12Dat,
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
	p := &TwoPlusTwoRanker{
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
func (p *TwoPlusTwoRanker) rank(hand []Card) HandRank {
	i := uint32(53)
	for _, c := range hand {
		i = p.ranks[i+p.cards[c]]
	}
	if len(hand) < 7 {
		i = p.ranks[i]
	}
	return HandRank(p.types[i>>12] - i&0xfff + 1)
}

// Embedded lookup table.
var (
	//go:embed twoplustwo00.dat
	twoplustwo00Dat []byte
	//go:embed twoplustwo01.dat
	twoplustwo01Dat []byte
	//go:embed twoplustwo02.dat
	twoplustwo02Dat []byte
	//go:embed twoplustwo03.dat
	twoplustwo03Dat []byte
	//go:embed twoplustwo04.dat
	twoplustwo04Dat []byte
	//go:embed twoplustwo05.dat
	twoplustwo05Dat []byte
	//go:embed twoplustwo06.dat
	twoplustwo06Dat []byte
	//go:embed twoplustwo07.dat
	twoplustwo07Dat []byte
	//go:embed twoplustwo08.dat
	twoplustwo08Dat []byte
	//go:embed twoplustwo09.dat
	twoplustwo09Dat []byte
	//go:embed twoplustwo10.dat
	twoplustwo10Dat []byte
	//go:embed twoplustwo11.dat
	twoplustwo11Dat []byte
	//go:embed twoplustwo12.dat
	twoplustwo12Dat []byte
)

package cardrank

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Rank is a card rank.
type Rank uint8

// Card ranks.
const (
	Ace Rank = 12 - iota
	King
	Queen
	Jack
	Ten
	Nine
	Eight
	Seven
	Six
	Five
	Four
	Three
	Two
)

// InvalidRank is an invalid rank.
const InvalidRank = Rank(^uint8(0))

// RankFromRune returns the card rank for the rune.
func RankFromRune(r rune) Rank {
	switch r {
	case 'A', 'a':
		return Ace
	case 'K', 'k':
		return King
	case 'Q', 'q':
		return Queen
	case 'J', 'j':
		return Jack
	case 'T', 't':
		return Ten
	case '9':
		return Nine
	case '8':
		return Eight
	case '7':
		return Seven
	case '6':
		return Six
	case '5':
		return Five
	case '4':
		return Four
	case '3':
		return Three
	case '2':
		return Two
	}
	return InvalidRank
}

// String satisfies the fmt.Stringer interface.
func (rank Rank) String() string {
	return string(rank.Byte())
}

// Byte returns the byte representation for the card rank.
func (rank Rank) Byte() byte {
	switch rank {
	case Ace:
		return 'A'
	case King:
		return 'K'
	case Queen:
		return 'Q'
	case Jack:
		return 'J'
	case Ten:
		return 'T'
	case Nine:
		return '9'
	case Eight:
		return '8'
	case Seven:
		return '7'
	case Six:
		return '6'
	case Five:
		return '5'
	case Four:
		return '4'
	case Three:
		return '3'
	case Two:
		return '2'
	}
	return 0
}

// Index the int index for the card rank (0-12 for Two-Ace).
func (rank Rank) Index() int {
	return int(rank)
}

// Name returns the name of the card rank.
func (rank Rank) Name() string {
	switch rank {
	case Ace:
		return "Ace"
	case King:
		return "King"
	case Queen:
		return "Queen"
	case Jack:
		return "Jack"
	case Ten:
		return "Ten"
	case Nine:
		return "Nine"
	case Eight:
		return "Eight"
	case Seven:
		return "Seven"
	case Six:
		return "Six"
	case Five:
		return "Five"
	case Four:
		return "Four"
	case Three:
		return "Three"
	case Two:
		return "Two"
	}
	return ""
}

// PluralName returns the plural name of the card rank.
func (rank Rank) PluralName() string {
	if rank == Six {
		return "Sixes"
	}
	return rank.Name() + "s"
}

// Suit is a card suit.
type Suit uint8

// Card suits.
const (
	Spade Suit = 1 << iota
	Heart
	Diamond
	Club
)

// InvalidSuit
const InvalidSuit = Suit(^uint8(0))

// SuitFromRune returns the card suit for the rune.
func SuitFromRune(r rune) Suit {
	switch r {
	case 'S', 's', UnicodeSpadeBlack, UnicodeSpadeWhite:
		return Spade
	case 'H', 'h', UnicodeHeartBlack, UnicodeHeartWhite:
		return Heart
	case 'D', 'd', UnicodeDiamondBlack, UnicodeDiamondWhite:
		return Diamond
	case 'C', 'c', UnicodeClubBlack, UnicodeClubWhite:
		return Club
	}
	return InvalidSuit
}

// String satisfies the fmt.Stringer interface.
func (suit Suit) String() string {
	return string(suit.Byte())
}

// Byte returns the byte representation for the card suit.
func (suit Suit) Byte() byte {
	switch suit {
	case Spade:
		return 's'
	case Heart:
		return 'h'
	case Diamond:
		return 'd'
	case Club:
		return 'c'
	}
	return 0
}

// Index returns the int index for the card suit (0-3 for Spade, Heart,
// Diamond, Club).
func (suit Suit) Index() int {
	switch suit {
	case Spade:
		return 0
	case Heart:
		return 1
	case Diamond:
		return 2
	case Club:
		return 3
	}
	return 0
}

// Name returns the name of the card suit.
func (suit Suit) Name() string {
	switch suit {
	case Spade:
		return "Spade"
	case Heart:
		return "Heart"
	case Diamond:
		return "Diamond"
	case Club:
		return "Club"
	}
	return ""
}

// PluralName returns the plural name of the card suit.
func (suit Suit) PluralName() string {
	return suit.Name() + "s"
}

// UnicodeBlack returns the black unicode pip rune for the card suit.
func (suit Suit) UnicodeBlack() rune {
	switch suit {
	case Spade:
		return UnicodeSpadeBlack
	case Heart:
		return UnicodeHeartBlack
	case Diamond:
		return UnicodeDiamondBlack
	case Club:
		return UnicodeClubBlack
	}
	return 0
}

// UnicodeWhite returns the white unicode pip rune for the card suit.
func (suit Suit) UnicodeWhite() rune {
	switch suit {
	case Spade:
		return UnicodeSpadeWhite
	case Heart:
		return UnicodeHeartWhite
	case Diamond:
		return UnicodeDiamondWhite
	case Club:
		return UnicodeClubWhite
	}
	return 0
}

// PlayingCardRune returns the unicode playing card rune for the card rank and
// suit.
func PlayingCardRune(rank Rank, suit Suit) rune {
	var v rune
	switch suit {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch rank {
	case Ace:
	case King:
		v += 13
	case Queen:
		v += 12
	default:
		v += rune(rank + 1)
	}
	return v
}

// PlayingCardKnightRune returns the unicode playing card rune for the card
// rank and suit, substituting knights for jacks.
func PlayingCardKnightRune(rank Rank, suit Suit) rune {
	var v rune
	switch suit {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch rank {
	case Ace:
	case King:
		v += 13
	case Queen:
		v += 12
	case Jack:
		v += 11
	default:
		v += rune(rank + 1)
	}
	return v
}

// Card is a card consisting of a rank (23456789TJQKA) and suit (shdc).
type Card uint32

// InvalidCard is an invalid card.
const InvalidCard = Card(^uint32(0))

// New creates a card for the specified rank and suit.
func New(rank Rank, suit Suit) Card {
	if Ace < rank || (suit != Spade && suit != Heart && suit != Diamond && suit != Club) {
		return InvalidCard
	}
	return Card(1<<uint32(rank)<<16 | uint32(suit)<<12 | uint32(rank)<<8 | uint32(primes[rank]))
}

// FromRune creates a card from a unicode playing card rune.
func FromRune(r rune) Card {
	switch {
	case unicode.Is(rangeS, r):
		return New(runeCardRank(r, UnicodeSpadeAce), Spade)
	case unicode.Is(rangeH, r):
		return New(runeCardRank(r, UnicodeHeartAce), Heart)
	case unicode.Is(rangeD, r):
		return New(runeCardRank(r, UnicodeDiamondAce), Diamond)
	case unicode.Is(rangeC, r):
		return New(runeCardRank(r, UnicodeClubAce), Club)
	}
	return InvalidCard
}

// FromString creates a card from a string.
func FromString(s string) Card {
	if strings.HasPrefix(s, "10") {
		s = "T" + s[2:]
	}
	switch v := []rune(s); len(v) {
	case 1:
		return FromRune(v[0])
	case 2:
		return New(RankFromRune(v[0]), SuitFromRune(v[1]))
	}
	return InvalidCard
}

// FromIndex creates a card from a numerical index (0-51).
func FromIndex(i int) Card {
	if i < 52 {
		return New(Rank(i%13), Suit(1<<(i/13)))
	}
	return InvalidCard
}

// Parse parses strings of Card representations from v. Ignores whitespace
// between cards and case. Combines all parsed representations into a single
// Card slice.
//
// Cards can described using common text strings (such as "Ah", "ah", "aH", or
// "AH"), or having a white or black unicode pip for the suit (such as "J♤" or
// "K♠"), or single unicode playing card runes (such as "🃆" or "🂣").
func Parse(v ...string) ([]Card, error) {
	var hand []Card
	for n, s := range v {
		for i, r := 0, []rune(s); i < len(r); i++ {
			switch {
			case unicode.IsSpace(r[i]):
				continue
			case unicode.Is(rangeA, r[i]):
				c := FromRune(r[i])
				if c == InvalidCard {
					return nil, &ParseError{
						S:   s,
						N:   n,
						I:   i,
						Err: ErrInvalidCard,
					}
				}
				hand = append(hand, c)
				continue
			case len(r)-i < 2:
				return nil, &ParseError{
					S:   s,
					N:   n,
					I:   i,
					Err: ErrInvalidCard,
				}
			}
			c := r[i]
			// handle '10'
			if len(r)-i > 2 && c == '1' && r[i+1] == '0' {
				c, i = 'T', i+1
			}
			card := New(RankFromRune(c), SuitFromRune(r[i+1]))
			if card == InvalidCard {
				return nil, &ParseError{
					S:   s,
					N:   n,
					I:   i,
					Err: ErrInvalidCard,
				}
			}
			hand = append(hand, card)
			i++
		}
	}
	return hand, nil
}

// Must creates a Card slice from v.
//
// See Parse for overview of accepted string representations of cards.
func Must(v ...string) []Card {
	hand, err := Parse(v...)
	if err == nil {
		return hand
	}
	panic(err)
}

// Rank returns the rank of the card.
func (c Card) Rank() Rank {
	return Rank(c >> 8 & 0xf)
}

// RankByte returns the rank byte of the card.
func (c Card) RankByte() byte {
	return c.Rank().Byte()
}

// RankIndex returns the rank index of the card.
func (c Card) RankIndex() int {
	return c.Rank().Index()
}

// Suit returns the suit of the card.
func (c Card) Suit() Suit {
	return Suit(c >> 12 & 0xf)
}

// SuitByte returns the suit byte of the card.
func (c Card) SuitByte() byte {
	return c.Suit().Byte()
}

// SuitIndex returns the suit index of the card.
func (c Card) SuitIndex() int {
	return c.Suit().Index()
}

// Index returns the index of the card.
func (c Card) Index() int {
	return c.SuitIndex()*13 + c.RankIndex()
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (c *Card) UnmarshalText(buf []byte) error {
	var err error
	*c = FromString(string(buf))
	return err
}

// MarshalText satisfies the encoding.TextMarshaler interface.
func (c Card) MarshalText() ([]byte, error) {
	return []byte{c.RankByte(), c.SuitByte()}, nil
}

// String satisfies the fmt.Stringer interface.
func (c Card) String() string {
	return string(c.RankByte()) + string(c.SuitByte())
}

// Format satisfies the fmt.Formatter interface.
//
// Supported verbs:
//
//	s - rank (23456789TJQKA) and suit (shdc) (ex: Ks Ah)
//	S - same as s, uppercased (ex: KS AH)
//	q - same as s, quoted (ex: "Ks" "Ah")
//	v - same as s
//	r - rank (as in s) without suit (ex: K A)
//	u - suit (as in s) without rank (shdc)
//	b - rank (as in s) and the black unicode pip rune (♠♥♦♣) (ex: K♠ A♥)
//	B - black unicode pip rune (as in b) without rank (♠♥♦♣)
//	h - rank (as in s) and the white unicode pip rune (♤♡♢♧) (ex: K♤ A♡)
//	H - white unicode pip rune (as in h) without rank (♤♡♢♧)
//	c - playing card rune (ex: 🂡  🂱  🃁  🃑)
//	C - playing card rune (as in c), substituting knights for jacks (ex: 🂬  🂼  🃌  🃜)
//	n - rank name, lower cased (ex: one two jack queen king ace)
//	N - rank name, title cased (ex: One Two Jack Queen King Ace)
//	p - plural rank name, lower cased (ex: ones twos sixes)
//	P - plural rank name, title cased (ex: Ones Twos Sixes)
//	t - suit name, lower cased (spade heart diamond club)
//	T - suit name, title cased (Spade Heart Diamond Club)
//	l - plural suit name, lower cased (spades hearts diamonds clubs)
//	L - plural suit name, title cased (Spades Hearts Diamonds Clubs)
//	d - base 10 integer value
func (c Card) Format(f fmt.State, verb rune) {
	r, s := c.Rank(), c.Suit()
	var buf []byte
	switch verb {
	case 's', 'S', 'v':
		buf = append(buf, r.Byte(), s.Byte())
		if verb == 'S' {
			buf = bytes.ToUpper(buf)
		}
	case 'q':
		buf = append(buf, '"', r.Byte(), s.Byte(), '"')
	case 'r':
		buf = append(buf, r.Byte())
	case 'u':
		buf = append(buf, s.Byte())
	case 'b':
		buf = append(buf, (string(r.Byte()) + string(s.UnicodeBlack()))...)
	case 'B':
		buf = append(buf, string(s.UnicodeBlack())...)
	case 'h':
		buf = append(buf, (string(r.Byte()) + string(s.UnicodeWhite()))...)
	case 'H':
		buf = append(buf, string(s.UnicodeWhite())...)
	case 'c':
		buf = append(buf, string(PlayingCardRune(r, s))...)
	case 'C':
		buf = append(buf, string(PlayingCardKnightRune(r, s))...)
	case 'n', 'N':
		buf = append(buf, r.Name()...)
		if verb == 'n' {
			buf = bytes.ToLower(buf)
		}
	case 'p', 'P':
		buf = append(buf, r.PluralName()...)
		if verb == 'p' {
			buf = bytes.ToLower(buf)
		}
	case 't', 'T':
		buf = append(buf, s.Name()...)
		if verb == 't' {
			buf = bytes.ToLower(buf)
		}
	case 'l', 'L':
		buf = append(buf, s.PluralName()...)
		if verb == 'l' {
			buf = bytes.ToLower(buf)
		}
	case 'd':
		buf = append(buf, strconv.Itoa(int(c))...)
	default:
		buf = append(buf, fmt.Sprintf("%%!%c(ERROR=unknown verb, card: %s)", verb, string(r.Byte())+string(s.Byte()))...)
	}
	_, _ = f.Write(buf)
}

// CardFormatter wraps formatting a set of cards. Allows `go test` to function
// without disabling vet.
type CardFormatter []Card

// Format satisfies the fmt.Formatter interface.
func (v CardFormatter) Format(f fmt.State, verb rune) {
	_, _ = f.Write([]byte{'['})
	for i, c := range v {
		if i != 0 {
			_, _ = f.Write([]byte{' '})
		}
		c.Format(f, verb)
	}
	_, _ = f.Write([]byte{']'})
}

// ParseError is a parse error.
type ParseError struct {
	S   string
	N   int
	I   int
	Err error
}

// Error satisfies the error interface.
func (err *ParseError) Error() string {
	return fmt.Sprintf("parse %q %d, %d: %v", err.S, err.N, err.I, err.Err)
}

// Unwrap satisfies the errors.Unwrap interface.
func (err *ParseError) Unwrap() error {
	return err.Err
}

// Unicode card runes.
const (
	UnicodeSpadeAce     rune = '🂡'
	UnicodeHeartAce     rune = '🂱'
	UnicodeDiamondAce   rune = '🃁'
	UnicodeClubAce      rune = '🃑'
	UnicodeSpadeBlack   rune = '♠'
	UnicodeSpadeWhite   rune = '♤'
	UnicodeHeartBlack   rune = '♥'
	UnicodeHeartWhite   rune = '♡'
	UnicodeDiamondBlack rune = '♦'
	UnicodeDiamondWhite rune = '♢'
	UnicodeClubBlack    rune = '♣'
	UnicodeClubWhite    rune = '♧'
)

// runeCardRank converts the unicode rune offset to a card rank.
func runeCardRank(rank, ace rune) Rank {
	r := Rank(rank - ace)
	switch {
	case r == 0:
		return Ace
	case r >= 11:
		return r - 2
	}
	return r - 1
}

func init() {
	s, h, d, c := make([]rune, 14), make([]rune, 14), make([]rune, 14), make([]rune, 14)
	for i := 0; i < 14; i++ {
		s[i] = UnicodeSpadeAce + rune(i)
		h[i] = UnicodeHeartAce + rune(i)
		d[i] = UnicodeDiamondAce + rune(i)
		c[i] = UnicodeClubAce + rune(i)
	}
	rangeS = newRangeTable(s...)
	rangeH = newRangeTable(h...)
	rangeD = newRangeTable(d...)
	rangeC = newRangeTable(c...)
	a := make([]rune, 14*4)
	copy(a[0:14], s[:])
	copy(a[14:28], h[:])
	copy(a[28:42], d[:])
	copy(a[42:56], c[:])
	rangeA = newRangeTable(a...)
}

// range tables for unicode playing card runes.
var (
	rangeS *unicode.RangeTable // spadees
	rangeH *unicode.RangeTable // hearts
	rangeD *unicode.RangeTable // diamonds
	rangeC *unicode.RangeTable // clubs
	rangeA *unicode.RangeTable // all
)

func newRangeTable(r ...rune) *unicode.RangeTable {
	if len(r) == 0 {
		return &unicode.RangeTable{}
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})
	// Remove duplicates.
	k := 1
	for i := 1; i < len(r); i++ {
		if r[k-1] != r[i] {
			r[k] = r[i]
			k++
		}
	}
	rt := new(unicode.RangeTable)
	for _, r := range r[:k] {
		if r <= 0xFFFF {
			rt.R16 = append(rt.R16, unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1})
		} else {
			rt.R32 = append(rt.R32, unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1})
		}
	}
	return rt
}

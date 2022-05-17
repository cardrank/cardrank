package cardrank

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
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

// RankFromRune returns the card rank for the rune.
func RankFromRune(r rune) (Rank, error) {
	switch r {
	case 'A', 'a':
		return Ace, nil
	case 'K', 'k':
		return King, nil
	case 'Q', 'q':
		return Queen, nil
	case 'J', 'j':
		return Jack, nil
	case 'T', 't':
		return Ten, nil
	case '9':
		return Nine, nil
	case '8':
		return Eight, nil
	case '7':
		return Seven, nil
	case '6':
		return Six, nil
	case '5':
		return Five, nil
	case '4':
		return Four, nil
	case '3':
		return Three, nil
	case '2':
		return Two, nil
	}
	return 0, ErrInvalidCardRank
}

// String satisfies the fmt.Stringer interface.
func (r Rank) String() string {
	return string(r.Byte())
}

// Byte returns the byte representation for the card rank.
func (r Rank) Byte() byte {
	switch r {
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

// Index the int index for the card rank (0-13 for Two-Ace).
func (r Rank) Index() int {
	return int(r)
}

// Name returns the name of the card rank.
func (r Rank) Name() string {
	switch r {
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
func (r Rank) PluralName() string {
	if r == Six {
		return "Sixes"
	}
	return r.Name() + "s"
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

// SuitFromRune returns the card suit for the rune.
func SuitFromRune(r rune) (Suit, error) {
	switch r {
	case 'S', 's', UnicodeSpadeBlack, UnicodeSpadeWhite:
		return Spade, nil
	case 'H', 'h', UnicodeHeartBlack, UnicodeHeartWhite:
		return Heart, nil
	case 'D', 'd', UnicodeDiamondBlack, UnicodeDiamondWhite:
		return Diamond, nil
	case 'C', 'c', UnicodeClubBlack, UnicodeClubWhite:
		return Club, nil
	}
	return 0, ErrInvalidCardSuit
}

// String satisfies the fmt.Stringer interface.
func (s Suit) String() string {
	return string(s.Byte())
}

// Byte returns the byte representation for the card suit.
func (s Suit) Byte() byte {
	switch s {
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
func (s Suit) Index() int {
	switch s {
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
func (s Suit) Name() string {
	switch s {
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
func (s Suit) PluralName() string {
	return s.Name() + "s"
}

// UnicodeBlack returns the black unicode pip rune for the card suit.
func (s Suit) UnicodeBlack() rune {
	switch s {
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
func (s Suit) UnicodeWhite() rune {
	switch s {
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
func PlayingCardRune(r Rank, s Suit) rune {
	var v rune
	switch s {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch r {
	case Ace:
	case King:
		v += 13
	case Queen:
		v += 12
	default:
		v += rune(r + 1)
	}
	return v
}

// PlayingCardKnightRune returns the unicode playing card rune for the card
// rank and suit, substituting knights for jacks.
func PlayingCardKnightRune(r Rank, s Suit) rune {
	var v rune
	switch s {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch r {
	case Ace:
	case King:
		v += 13
	case Queen:
		v += 12
	case Jack:
		v += 11
	default:
		v += rune(r + 1)
	}
	return v
}

// Card is a card consisting of a rank (23456789TJQKA) and suit (shdc).
type Card uint32

// New creates a card for the specified rank and suit.
func New(r Rank, s Suit) Card {
	return Card(1<<uint32(r)<<16 | uint32(s)<<12 | uint32(r)<<8 | uint32(primes[r]))
}

// FromRune creates a card from a unicode playing card rune.
func FromRune(r rune) (Card, error) {
	switch {
	case unicode.Is(rangeS, r):
		return New(runeCardRank(r, UnicodeSpadeAce), Spade), nil
	case unicode.Is(rangeH, r):
		return New(runeCardRank(r, UnicodeHeartAce), Heart), nil
	case unicode.Is(rangeD, r):
		return New(runeCardRank(r, UnicodeDiamondAce), Diamond), nil
	case unicode.Is(rangeC, r):
		return New(runeCardRank(r, UnicodeClubAce), Club), nil
	}
	return 0, ErrInvalidCard
}

// FromString creates a card from a string.
func FromString(str string) (Card, error) {
	if strings.HasPrefix(str, "10") {
		str = "T" + str[2:]
	}
	switch v := []rune(str); len(v) {
	case 1:
		return FromRune(v[0])
	case 2:
		r, err := RankFromRune(v[0])
		if err != nil {
			return 0, err
		}
		s, err := SuitFromRune(v[1])
		if err != nil {
			return 0, err
		}
		return New(r, s), nil
	}
	return 0, ErrInvalidCard
}

// Parse parses card representations in v.
func Parse(v ...string) ([]Card, error) {
	var hand []Card
	for _, s := range v {
		for i, r := 0, []rune(s); i < len(r); i++ {
			switch {
			case unicode.IsSpace(r[i]):
				continue
			case unicode.Is(rangeA, r[i]):
				c, err := FromRune(r[i])
				if err != nil {
					return nil, err
				}
				hand = append(hand, c)
				continue
			case len(r)-i < 2:
				return nil, ErrInvalidCard
			}
			c := r[i]
			// handle '10'
			if len(r)-i > 2 && c == '1' && r[i+1] == '0' {
				c, i = 'T', i+1
			}
			rank, err := RankFromRune(c)
			if err != nil {
				return nil, err
			}
			suit, err := SuitFromRune(r[i+1])
			if err != nil {
				return nil, err
			}
			hand = append(hand, New(rank, suit))
			i++
		}
	}
	return hand, nil
}

// MustCard creates card from s.
func MustCard(s string) Card {
	c, err := FromString(s)
	if err == nil {
		return c
	}
	panic(err)
}

// Must creates a hand from v.
func Must(v ...string) []Card {
	hand, err := Parse(v...)
	if err == nil {
		return hand
	}
	panic(err)
}

// Rank returns the rank of the card.
func (c Card) Rank() Rank {
	return Rank((uint32(c) >> 8) & 0xf)
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
	return Suit((uint32(c) >> 12) & 0xf)
}

// SuitByte returns the suit byte of the card.
func (c Card) SuitByte() byte {
	return c.Suit().Byte()
}

// SuitIndex returns the suit index of the card.
func (c Card) SuitIndex() int {
	return c.Suit().Index()
}

// BitRank returns the bit rank of the card.
func (c Card) BitRank() uint32 {
	return (uint32(c) >> 16) & 0x1fff
}

// Prime returns the prime value of the card.
func (c Card) Prime() uint32 {
	return uint32(c) & 0x3f
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (c *Card) UnmarshalText(buf []byte) error {
	var err error
	*c, err = FromString(string(buf))
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
func runeCardRank(r, ace rune) Rank {
	i := Rank(r - ace)
	switch {
	case i == 0:
		return Ace
	case i >= 11:
		return i - 2
	}
	return i - 1
}

func init() {
	s, h, d, c := make([]rune, 14), make([]rune, 14), make([]rune, 14), make([]rune, 14)
	for i := 0; i < 14; i++ {
		s[i] = UnicodeSpadeAce + rune(i)
		h[i] = UnicodeHeartAce + rune(i)
		d[i] = UnicodeDiamondAce + rune(i)
		c[i] = UnicodeClubAce + rune(i)
	}
	rangeS = rangetable.New(s...)
	rangeH = rangetable.New(h...)
	rangeD = rangetable.New(d...)
	rangeC = rangetable.New(c...)
	rangeA = rangetable.Merge(rangeS, rangeH, rangeD, rangeC)
}

// range tables for unicode playing card runes.
var (
	rangeS *unicode.RangeTable // spadees
	rangeH *unicode.RangeTable // hearts
	rangeD *unicode.RangeTable // diamonds
	rangeC *unicode.RangeTable // clubs
	rangeA *unicode.RangeTable // all
)

// Error is a error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidCard is the invalid card error.
	ErrInvalidCard Error = "invalid card"
	// ErrInvalidCardRank is the invalid card rank error.
	ErrInvalidCardRank Error = "invalid card rank"
	// ErrInvalidCardSuit is the invalid card suit error.
	ErrInvalidCardSuit Error = "invalid card suit"
)

// primes are the first 13 prime numbers (one per card rank).
var primes = [...]uint8{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}

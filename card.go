package cardrank

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// AlternateEmoji are alternate emoji pips.
var AlternateEmoji = [4]string{
	"♠️",
	"♥️",
	"🔷",
	"☘️",
}

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

// InvalidRank is an invalid card rank.
const InvalidRank = ^Rank(0)

// RankFromRune returns a rune's card rank.
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

// String satisfies the [fmt.Stringer] interface.
func (rank Rank) String() string {
	return string(rank.Byte())
}

// Byte returns the card rank byte.
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
	return '0'
}

// Index the card rank int index (0-12 for [Two]-[Ace]).
func (rank Rank) Index() int {
	return int(rank)
}

// Name returns the card rank name.
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

// PluralName returns the card rank plural name.
func (rank Rank) PluralName() string {
	if rank == Six {
		return "Sixes"
	}
	return rank.Name() + "s"
}

// StraightFlushName returns the card rank [StraightFlush] name.
func (rank Rank) StraightFlushName() string {
	switch rank {
	case Ace:
		return "Royal"
	case King:
		return "Platinum Oxide"
	case Queen:
		return "Silver Tongue"
	case Jack:
		return "Bronze Fist"
	case Ten:
		return "Golden Ratio"
	case Nine:
		return "Iron Maiden"
	case Eight:
		return "Tin Cup"
	case Seven:
		return "Brass Axe"
	case Six:
		return "Aluminum Window"
	case Five:
		return "Steel Wheel"
	}
	return ""
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

// InvalidSuit is an invalid card suit.
const InvalidSuit = ^Suit(0)

// SuitFromRune returns a rune's card suit.
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

// String satisfies the [fmt.Stringer] interface.
func (suit Suit) String() string {
	return string(suit.Byte())
}

// Byte returns the card suit byte.
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
	return '0'
}

// Index returns the card suit int index (0-3 for Spade, Heart, Diamond, Club).
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

// Name returns the card suit name.
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

// PluralName returns the card suit plural name.
func (suit Suit) PluralName() string {
	return suit.Name() + "s"
}

// UnicodeBlack returns the card suit black unicode pip rune.
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

// UnicodeWhite returns the card suit white unicode pip rune.
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

// Emoji returns the card suit pip emoji.
func (suit Suit) Emoji() string {
	switch suit {
	case Spade, Heart, Diamond, Club:
		return string([]rune{suit.UnicodeBlack(), '\ufe0f'})
	}
	return ""
}

// AlternateEmoji returns the card suit alternate pip emoji.
func (suit Suit) AlternateEmoji() string {
	switch suit {
	case Spade:
		return AlternateEmoji[0]
	case Heart:
		return AlternateEmoji[1]
	case Diamond:
		return AlternateEmoji[2]
	case Club:
		return AlternateEmoji[3]
	}
	return ""
}

// Card is a card consisting of a [Rank] (23456789TJQKA) and [Suit] (shdc).
type Card uint32

// InvalidCard is an invalid card.
const InvalidCard = ^Card(0)

// New creates a card for the rank and suit.
func New(rank Rank, suit Suit) Card {
	if Ace < rank || (suit != Spade && suit != Heart && suit != Diamond && suit != Club) {
		return InvalidCard
	}
	return 1<<Card(rank)<<16 | Card(suit)<<12 | Card(rank)<<8 | Card(primes[rank])
}

// FromRune creates a card from a unicode playing card rune.
func FromRune(r rune) Card {
	for _, v := range [...]struct {
		ace  rune
		suit Suit
	}{
		{UnicodeSpadeAce, Spade},
		{UnicodeHeartAce, Heart},
		{UnicodeDiamondAce, Diamond},
		{UnicodeClubAce, Club},
	} {
		// ace through king, including the knight
		if v.ace <= r && r < v.ace+14 {
			return New(runeCardRank(r, v.ace), v.suit)
		}
	}
	return InvalidCard
}

// FromString creates a card from a string.
func FromString(s string) Card {
	if rest, ok := strings.CutPrefix(s, "10"); ok {
		s = "T" + rest
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
	if 0 <= i && i < 52 {
		return New(Rank(i%13), Suit(1<<(i/13)))
	}
	return InvalidCard
}

// Parse parses common string representations of [Card] values contained in v,
// ignoring case and whitespace.
//
// Accepts the following:
//   - a rank followed by a suit (ex: "Ah", "ks", "10s", "Tc", "8d", "6c")
//   - a rank followed by a white or black unicode suit pip (ex: "J♤", "K♠")
//   - unicode playing card runes (ex: "🃆", "🂣").
//
// Returns a single slice of all cards from all strings in v.
func Parse(v ...string) ([]Card, error) {
	var cards []Card
	for n, s := range v {
		for i, r := 0, []rune(s); i < len(r); i++ {
			switch {
			case unicode.IsSpace(r[i]):
				continue
			case isCardRune(r[i]):
				c := FromRune(r[i])
				if c == InvalidCard {
					return nil, &ParseError{
						S:   s,
						N:   n,
						I:   i,
						Err: ErrInvalidCard,
					}
				}
				cards = append(cards, c)
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
			// parse '10'
			if 2 < len(r)-i && c == '1' && r[i+1] == '0' {
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
			cards = append(cards, card)
			i++
		}
	}
	return cards, nil
}

// Must parses common string representations of [Card] values contained in v,
// ignoring case and whitespace and panicing on any error.
//
// Returns a single slice of all cards from all strings in v.
//
// See [Parse] for overview of accepted string representations.
func Must(v ...string) []Card {
	cards, err := Parse(v...)
	if err == nil {
		return cards
	}
	panic(err)
}

// Rank returns the card rank.
func (c Card) Rank() Rank {
	return Rank(c >> 8 & 0xf)
}

// RankByte returns the card rank byte.
func (c Card) RankByte() byte {
	return c.Rank().Byte()
}

// RankIndex returns the card rank index.
func (c Card) RankIndex() int {
	return c.Rank().Index()
}

// Suit returns the card suit.
func (c Card) Suit() Suit {
	return Suit(c >> 12 & 0xf)
}

// SuitByte returns the card suit byte.
func (c Card) SuitByte() byte {
	return c.Suit().Byte()
}

// SuitIndex returns the card suit index.
func (c Card) SuitIndex() int {
	return c.Suit().Index()
}

// Index returns the card index.
func (c Card) Index() int {
	return c.SuitIndex()*13 + c.RankIndex()
}

// AceRank returns the card [Ace]-low index.
func (c Card) AceRank() int {
	return int(c>>8&0xf+1) % 13
}

// Rune returns the card's unicode playing card rune.
func (c Card) Rune() rune {
	if c == InvalidCard {
		return '0'
	}
	var v rune
	switch c.Suit() {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch rank := c.Rank(); rank {
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

// KnightRune returns the card's unicode playing card rune, substituting
// knights for [Jack] cards.
func (c Card) KnightRune() rune {
	if c == InvalidCard {
		return '0'
	}
	var v rune
	switch c.Suit() {
	case Spade:
		v = UnicodeSpadeAce
	case Heart:
		v = UnicodeHeartAce
	case Diamond:
		v = UnicodeDiamondAce
	case Club:
		v = UnicodeClubAce
	}
	switch rank := c.Rank(); rank {
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

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface.
func (c *Card) UnmarshalText(buf []byte) error {
	if *c = FromString(string(buf)); *c == InvalidCard {
		return ErrInvalidCard
	}
	return nil
}

// MarshalText satisfies the [encoding.TextMarshaler] interface.
func (c Card) MarshalText() ([]byte, error) {
	if c != InvalidCard {
		return []byte{c.RankByte(), c.SuitByte()}, nil
	}
	return nil, ErrInvalidCard
}

// String satisfies the [fmt.Stringer] interface.
func (c Card) String() string {
	return string(c.RankByte()) + string(c.SuitByte())
}

// Format satisfies the [fmt.Formatter] interface.
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
//	e - rank (as in s) and the emoji pip (♠️ ❤️ ♦️ ♣️ ) (ex: K♠️ A❤️ )
//	E - emoji pip (as in e) without rank (♠️ ❤️ ♦️ ♣️ )
//	a - rank (as in s) and the alternate emoji pip (see [AlternateEmoji])
//	A - alternate emoji pip (see [AlternateEmoji])
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
//	F - straight flush rank name
func (c Card) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 's', 'S', 'v':
		buf = append(buf, c.RankByte(), c.SuitByte())
		if verb == 'S' {
			buf = bytes.ToUpper(buf)
		}
	case 'q':
		buf = append(buf, '"', c.RankByte(), c.SuitByte(), '"')
	case 'r':
		buf = append(buf, c.RankByte())
	case 'u':
		buf = append(buf, c.SuitByte())
	case 'b':
		buf = append(buf, (string(c.RankByte()) + string(c.Suit().UnicodeBlack()))...)
	case 'B':
		buf = append(buf, string(c.Suit().UnicodeBlack())...)
	case 'h':
		buf = append(buf, (string(c.RankByte()) + string(c.Suit().UnicodeWhite()))...)
	case 'H':
		buf = append(buf, string(c.Suit().UnicodeWhite())...)
	case 'e':
		buf = append(buf, (string(c.RankByte()) + c.Suit().Emoji())...)
	case 'E':
		buf = append(buf, c.Suit().Emoji()...)
	case 'a':
		buf = append(buf, (string(c.RankByte()) + c.Suit().AlternateEmoji())...)
	case 'A':
		buf = append(buf, c.Suit().AlternateEmoji()...)
	case 'c':
		buf = append(buf, string(c.Rune())...)
	case 'C':
		buf = append(buf, string(c.KnightRune())...)
	case 'n', 'N':
		buf = append(buf, c.Rank().Name()...)
		if verb == 'n' {
			buf = bytes.ToLower(buf)
		}
	case 'p', 'P':
		buf = append(buf, c.Rank().PluralName()...)
		if verb == 'p' {
			buf = bytes.ToLower(buf)
		}
	case 't', 'T':
		buf = append(buf, c.Suit().Name()...)
		if verb == 't' {
			buf = bytes.ToLower(buf)
		}
	case 'l', 'L':
		buf = append(buf, c.Suit().PluralName()...)
		if verb == 'l' {
			buf = bytes.ToLower(buf)
		}
	case 'F':
		buf = append(buf, c.Rank().StraightFlushName()...)
	case 'd':
		buf = append(buf, strconv.Itoa(int(c))...)
	default:
		buf = append(buf, fmt.Sprintf(
			"%%!%c(ERROR=unknown verb, card: %s)",
			verb,
			string(c.RankByte())+string(c.SuitByte()))...,
		)
	}
	_, _ = f.Write(buf)
}

// Formatter wraps formatting a set of cards. Allows `go test` to function
// without disabling vet.
type Formatter []Card

// Format satisfies the [fmt.Formatter] interface.
func (v Formatter) Format(f fmt.State, verb rune) {
	_, _ = f.Write([]byte{'['})
	for i, c := range v {
		if i != 0 {
			_, _ = f.Write([]byte{' '})
		}
		c.Format(f, verb)
	}
	_, _ = f.Write([]byte{']'})
}

// Combine combines the cards in v.
func (v Formatter) Combine() Card {
	n := len(v)
	if n == 0 {
		return InvalidCard
	}
	return v.combine(n)
}

// combine combines the cards in v.
func (v Formatter) combine(n int) Card {
	c := v[0]
	for i := 1; i < n; i++ {
		c |= v[i]
	}
	return c
}

// Suits returns unused suits in v.
func (v Formatter) Suits() []Suit {
	n := len(v)
	if n == 0 {
		return []Suit{Spade, Heart, Diamond, Club}
	}
	s := v.combine(n).Suit()
	var suits []Suit
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		if s&suit == 0 {
			suits = append(suits, suit)
		}
	}
	return suits
}

// Ranks returns unused ranks in v.
func (v Formatter) Ranks() []Rank {
	n := len(v)
	if n == 0 {
		return []Rank{Ace, King, Queen, Jack, Ten, Nine, Eight, Seven, Six, Five, Four, Three, Two}
	}
	m := make(map[Rank]bool, len(v))
	for i := range n {
		m[v[i].Rank()] = true
	}
	var ranks []Rank
	for rank := Ace; rank >= Two && rank != InvalidRank; rank-- {
		if !m[rank] {
			ranks = append(ranks, rank)
		}
	}
	return ranks
}

// ParseError is a parse error.
type ParseError struct {
	S   string
	N   int
	I   int
	Err error
}

// Error satisfies the [error] interface.
func (err *ParseError) Error() string {
	return fmt.Sprintf("parse %q %d, %d: %v", err.S, err.N, err.I, err.Err)
}

// Unwrap satisfies the [errors.Unwrap] interface.
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

// Exclude returns a copy of v excluding any specified cards.
func Exclude(v []Card, ex ...[]Card) []Card {
	if len(ex) == 0 {
		return slices.Clone(v)
	}
	m := make(map[Card]bool)
	for _, u := range ex {
		for _, c := range u {
			m[c] = true
		}
	}
	var u []Card
	for _, c := range v {
		if !m[c] {
			u = append(u, c)
		}
	}
	return u
}

// runeCardRank converts the unicode rune offset to a card rank.
func runeCardRank(rank, ace rune) Rank {
	r := Rank(rank - ace)
	switch {
	case r == 0:
		return Ace
	case 11 <= r:
		return r - 2
	}
	return r - 1
}

// isCardRune returns true when r is within the unicode playing card block.
func isCardRune(r rune) bool {
	return UnicodeSpadeAce <= r && r < UnicodeClubAce+14
}

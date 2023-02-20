package cardrank

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Type is a package level eval type, providing a package-level interface for
// using DefaultTypes], and wraps the [TypeDesc].
//
// [DefaultTypes] will be registered during init, unless using the
// [noinit][#noinit] tag.
//
// See [TypeDesc].
type Type uint16

// Types.
const (
	Holdem         Type = 'H'<<8 | 'h' // Hh
	Split          Type = 'H'<<8 | 'l' // Hl
	Short          Type = 'H'<<8 | 's' // Hs
	Manila         Type = 'H'<<8 | 'm' // Hm
	Spanish        Type = 'H'<<8 | 'p' // Hp
	Royal          Type = 'H'<<8 | 'r' // Hr
	Double         Type = 'H'<<8 | 'd' // Hd
	Showtime       Type = 'H'<<8 | 't' // Ht
	Swap           Type = 'H'<<8 | 'w' // Hw
	River          Type = 'H'<<8 | 'v' // Hv
	Dallas         Type = 'H'<<8 | 'a' // Ha
	Houston        Type = 'H'<<8 | 'u' // Hu
	Draw           Type = 'D'<<8 | 'h' // Dh
	DrawHiLo       Type = 'D'<<8 | 'l' // Dl
	Stud           Type = 'S'<<8 | 'h' // Sh
	StudHiLo       Type = 'S'<<8 | 'l' // Sl
	StudFive       Type = 'S'<<8 | '5' // S5
	Video          Type = 'J'<<8 | 'h' // Jh
	Omaha          Type = 'O'<<8 | '4' // O4
	OmahaHiLo      Type = 'O'<<8 | 'l' // Ol
	OmahaDouble    Type = 'O'<<8 | 'd' // Od
	OmahaFive      Type = 'O'<<8 | '5' // O5
	OmahaSix       Type = 'O'<<8 | '6' // O6
	Courchevel     Type = 'O'<<8 | 'c' // Oc
	CourchevelHiLo Type = 'O'<<8 | 'e' // Oe
	Fusion         Type = 'O'<<8 | 'f' // Of
	FusionHiLo     Type = 'O'<<8 | 'F' // OF
	Soko           Type = 'K'<<8 | 'h' // Kh
	SokoHiLo       Type = 'K'<<8 | 'l' // Kl
	Lowball        Type = 'L'<<8 | '1' // L1
	LowballTriple  Type = 'L'<<8 | '3' // L3
	Razz           Type = 'R'<<8 | 'a' // Ra
	Badugi         Type = 'B'<<8 | 'a' // Ba
)

// DefaultTypes returns the default type descriptions.
func DefaultTypes() []TypeDesc {
	var v []TypeDesc
	for _, d := range []struct {
		id   string
		typ  Type
		name string
		opt  TypeOption
	}{
		{"Hh", Holdem, "Holdem", WithHoldem(false)},
		{"Hl", Split, "Split", WithHoldem(true)},
		{"Hs", Short, "Short", WithShort()},
		{"Hm", Manila, "Manila", WithManila()},
		{"Hp", Spanish, "Spanish", WithSpanish()},
		{"Hr", Royal, "Royal", WithRoyal()},
		{"Hd", Double, "Double", WithDouble()},
		{"Ht", Showtime, "Showtime", WithShowtime(false)},
		{"Hw", Swap, "Swap", WithSwap(false)},
		{"Hv", River, "River", WithRiver(false)},
		{"Ha", Dallas, "Dallas", WithDallas(false)},
		{"Hu", Houston, "Houston", WithHouston(false)},
		{"Dh", Draw, "Draw", WithDraw(false)},
		{"Dl", DrawHiLo, "DrawHiLo", WithDraw(true)},
		{"Sh", Stud, "Stud", WithStud(false)},
		{"Sl", StudHiLo, "StudHiLo", WithStud(true)},
		{"S5", StudFive, "StudFive", WithStudFive(false)},
		{"Jh", Video, "Video", WithVideo(false)},
		{"O4", Omaha, "Omaha", WithOmaha(false)},
		{"Ol", OmahaHiLo, "OmahaHiLo", WithOmaha(true)},
		{"Od", OmahaDouble, "OmahaDouble", WithOmahaDouble()},
		{"O5", OmahaFive, "OmahaFive", WithOmahaFive(false)},
		{"O6", OmahaSix, "OmahaSix", WithOmahaSix(false)},
		{"Oc", Courchevel, "Courchevel", WithCourchevel(false)},
		{"Oe", CourchevelHiLo, "CourchevelHiLo", WithCourchevel(true)},
		{"Of", Fusion, "Fusion", WithFusion(false)},
		{"OF", FusionHiLo, "FusionHiLo", WithFusion(true)},
		{"Kh", Soko, "Soko", WithSoko(false)},
		{"Kl", SokoHiLo, "SokoHiLo", WithSoko(true)},
		{"L1", Lowball, "Lowball", WithLowball(false)},
		{"L3", LowballTriple, "LowballTriple", WithLowball(true)},
		{"Ra", Razz, "Razz", WithRazz()},
		{"Ba", Badugi, "Badugi", WithBadugi()},
	} {
		desc, err := NewType(d.id, d.typ, d.name, d.opt)
		if err != nil {
			panic(err)
		}
		v = append(v, *desc)
	}
	return v
}

// IdToType converts id to its type.
func IdToType(id string) (Type, error) {
	switch {
	case len(id) != 2,
		!unicode.IsLetter(rune(id[0])) && !unicode.IsNumber(rune(id[0])),
		!unicode.IsLetter(rune(id[1])) && !unicode.IsNumber(rune(id[1])):
		return 0, ErrInvalidId
	}
	return Type(id[0])<<8 | Type(id[1]), nil
}

// MarshalText satisfies the encoding.TextMarshaler interface.
func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.Id()), nil
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (typ *Type) UnmarshalText(buf []byte) error {
	name := strings.ToLower(string(buf))
	for t, desc := range descs {
		if strings.ToLower(desc.Name) == name {
			*typ = t
			return nil
		}
	}
	if len(name) == 2 {
		if id, err := IdToType(string(buf)); err == nil {
			*typ = id
			return nil
		}
	}
	return ErrInvalidType
}

// Id returns the type's id.
func (typ Type) Id() string {
	return string([]byte{byte(typ >> 8), byte(typ)})
}

// Format satisfies the fmt.Formatter interface.
func (typ Type) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'c':
		buf = []byte(typ.Id())
	case 's', 'v':
		if desc, ok := descs[typ]; ok {
			buf = []byte(desc.Name)
		} else {
			buf = []byte("Type(" + strconv.Itoa(int(typ)) + ")")
		}
	case 'l':
		if desc, ok := descs[typ]; ok {
			buf = []byte(desc.Eval.Name())
			if desc.Low {
				buf = append(buf, " Hi/Lo"...)
			}
		} else {
			buf = []byte("Type(" + strconv.Itoa(int(typ)) + ")")
		}
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, type: %d)", verb, int(typ)))
	}
	_, _ = f.Write(buf)
}

// Desc returns the type's description.
func (typ Type) Desc() TypeDesc {
	return descs[typ]
}

// Name returns the type's name.
func (typ Type) Name() string {
	return descs[typ].Name
}

// Max returns the type's max players.
func (typ Type) Max() int {
	return descs[typ].Max
}

// Low returns true when the type supports 8-or-better lo eval.
func (typ Type) Low() bool {
	return descs[typ].Low
}

// Double returns true when the type has double boards.
func (typ Type) Double() bool {
	return descs[typ].Double
}

// Show returns true when the type shows folded cards.
func (typ Type) Show() bool {
	return descs[typ].Show
}

// Once returns true when draws are limited to one time.
func (typ Type) Once() bool {
	return descs[typ].Once
}

// Blinds returns the type's blind names.
func (typ Type) Blinds() []string {
	if desc, ok := descs[typ]; ok {
		v := make([]string, len(desc.Blinds))
		copy(v, desc.Blinds)
		return v
	}
	return nil
}

// Streets returns the type's street descriptions.
func (typ Type) Streets() []StreetDesc {
	if desc, ok := descs[typ]; ok {
		v := make([]StreetDesc, len(desc.Streets))
		copy(v, desc.Streets)
		return v
	}
	return nil
}

// Pocket returns the type's total dealt pocket cards.
func (typ Type) Pocket() int {
	if desc, ok := descs[typ]; ok {
		var count int
		for i := 0; i < len(desc.Streets); i++ {
			count += desc.Streets[i].Pocket
		}
		return count
	}
	return 0
}

// PocketDiscard returns the type's total pocket discard.
func (typ Type) PocketDiscard() int {
	if desc, ok := descs[typ]; ok {
		var count int
		for i := 0; i < len(desc.Streets); i++ {
			count += desc.Streets[i].PocketDiscard
		}
		return count
	}
	return 0
}

// Board returns the type's total dealt board cards.
func (typ Type) Board() int {
	if desc, ok := descs[typ]; ok {
		var count int
		for i := 0; i < len(desc.Streets); i++ {
			count += desc.Streets[i].Board
		}
		return count
	}
	return 0
}

// BoardDiscard returns the type's total board discard.
func (typ Type) BoardDiscard() int {
	if desc, ok := descs[typ]; ok {
		var count int
		for i := 0; i < len(desc.Streets); i++ {
			count += desc.Streets[i].BoardDiscard
		}
		return count
	}
	return 0
}

// DeckType returns the type's deck type.
func (typ Type) DeckType() DeckType {
	return descs[typ].Deck
}

// Deck returns a new deck for the type.
func (typ Type) Deck() *Deck {
	return descs[typ].Deck.New()
}

// Dealer creates a new dealer with a deck shuffled by shuffles, for the pocket
// count.
func (typ Type) Dealer(shuffler Shuffler, shuffles, count int) *Dealer {
	if desc, ok := descs[typ]; ok {
		return NewShuffledDealer(desc, shuffler, shuffles, count)
	}
	return nil
}

// Deal creates a new dealer for the type, shuffling the deck by shuffles and
// returning the count dealt pockets and hi board.
func (typ Type) Deal(shuffler Shuffler, shuffles, count int) ([][]Card, []Card) {
	if d := typ.Dealer(shuffler, shuffles, count); d != nil {
		for d.Next() {
		}
		return d.Runs[0].Pockets, d.Runs[0].Hi
	}
	return nil, nil
}

// New creates a new eval for the type and pocket.
func (typ Type) New() *Eval {
	return EvalOf(typ)
}

// Eval creates a new eval, and evaluates the type and pocket.
func (typ Type) Eval(pocket, board []Card) *Eval {
	ev := EvalOf(typ)
	ev.Eval(pocket, board)
	return ev
}

// EvalPockets creates new evals and evaluates each of the pockets and board.
func (typ Type) EvalPockets(pockets [][]Card, board []Card) []*Eval {
	evs := make([]*Eval, len(pockets))
	for i := 0; i < len(pockets); i++ {
		evs[i] = typ.Eval(pockets[i], board)
	}
	return evs
}

// TypeDesc is a type description.
type TypeDesc struct {
	// Num is the registered number.
	Num int
	// Type is the type.
	Type Type
	// Name is the type name.
	Name string
	// Max is the max number of players.
	Max int
	// Low is true when the type's Cactus or Omaha eval is the Hi/Lo variant,
	// enabling 8-or-better on the lo eval.
	Low bool
	// Double is true when there are double the number boards.
	Double bool
	// Show is true when folded cards are shown.
	Show bool
	// Once is true when a draw can only occur once.
	Once bool
	// Blinds are the blind names.
	Blinds []string
	// Streets are the betting streets.
	Streets []StreetDesc
	// Deck is the deck type.
	Deck DeckType
	// Eval is the eval type.
	Eval EvalType
	// HiDesc is the hi desc type.
	HiDesc DescType
	// LoDesc is the lo desc type.
	LoDesc DescType
}

// NewType creates a new type description. Any created type must be registered
// by calling [RegisterType] before it can be used for eval.
func NewType(id string, typ Type, name string, opts ...TypeOption) (*TypeDesc, error) {
	switch id, err := IdToType(id); {
	case err != nil:
		return nil, err
	case id != typ:
		return nil, ErrInvalidId
	}
	desc := &TypeDesc{
		Type:   typ,
		Name:   name,
		Deck:   DeckFrench,
		Eval:   EvalCactus,
		HiDesc: DescCactus,
		LoDesc: DescLow,
	}
	for _, o := range opts {
		o(desc)
	}
	return desc, nil
}

// Apply applies street options.
//
//nolint:gosec
func (desc *TypeDesc) Apply(opts ...StreetOption) {
	for _, o := range opts {
		for i, street := range desc.Streets {
			o(i, &street)
			desc.Streets[i] = street
		}
	}
}

// TypeOption is a type description option.
type TypeOption func(*TypeDesc)

// WithHoldem is a type description option to set Holdem definitions.
//
// Holdem is a best-5 card game played with a standard (French) deck,
// comprising a pocket of 2 cards, 5 community board cards, and a Pre-Flop,
// Flop, Turn, and River streets. 2 pocket cards are dealt on the Pre-Flop,
// with 3 board cards on the Flop, 1 board card on the Turn, and one on the
// River. 1 card is discarded on the Flop, Turn, and River, prior to the board
// cards being dealt.
func WithHoldem(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Apply(opts...)
	}
}

// WithShort is a type description option to set Short definitions.
//
// Short is a Holdem variant played with a Short (6+) deck. Flushes rank over
// Full Houses.
func WithShort(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 6
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Deck = DeckShort
		desc.Eval = EvalShort
		desc.HiDesc = DescFlushOver
		desc.Apply(opts...)
	}
}

// WithManila is a type description option to set Manila definitions.
//
// Manila is a Holdem variant played with a Manila (7+) deck that forces the
// use of 2 pocket cards, adding a Drop street before the Flop, with the 5
// (instead of 4) streets receiving 1 board card each. Flushes rank over Full
// Houses.
func WithManila(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 6
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 0, 1, 1, 1)
		desc.Deck = DeckManila
		desc.Eval = EvalManila
		desc.HiDesc = DescFlushOver
		desc.Streets[0].Board = 1
		desc.Streets = insert(desc.Streets, 1, StreetDesc{
			Id:    'd',
			Name:  "Drop",
			Board: 1,
		})
		desc.Apply(opts...)
	}
}

// WithSpanish is a type description option to set Spanish definitions.
//
// Spanish is a Holdem variant played with a Spanish deck (8+) that forces the
// use of 2 pocket cards, adding a Drop street before the Flop, with the 5
// (instead of 4) streets receiving 1 board card each. Flushes rank over Full
// Houses.
func WithSpanish(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 6
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 0, 1, 1, 1)
		desc.Deck = DeckSpanish
		desc.Eval = EvalSpanish
		desc.HiDesc = DescFlushOver
		desc.Streets[0].Board = 1
		desc.Streets = insert(desc.Streets, 1, StreetDesc{
			Id:    'd',
			Name:  "Drop",
			Board: 1,
		})
		desc.Apply(opts...)
	}
}

// WithRoyal is a type description option to set Royal definitions.
//
// Royal is a Holdem variant played with a Royal deck.
func WithRoyal(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 5
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Deck = DeckRoyal
		desc.Apply(opts...)
	}
}

// WithDouble is a type description option to set Double definitions.
//
// Double is a Holdem variant where there are two community boards dealt in
// tandem, comprising the Hi and Lo.
func WithDouble(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Double = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.LoDesc = DescCactus
		desc.Apply(opts...)
	}
}

// WithShowtime is a type description option to set Showtime definitions.
//
// Showtime is a Holdem variant where folded cards are shown.
func WithShowtime(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Show = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Apply(opts...)
	}
}

// WithSwap is a type description option to set Swap definitions.
//
// Swap is a Holdem variant that allows the swapping of cards.
func WithSwap(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Once = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		for i := 1; i < len(desc.Streets); i++ {
			desc.Streets[i].PocketDraw = 2
		}
		desc.Apply(opts...)
	}
}

// WithRiver is a type description option to set River definitions.
//
// River is a Holdem variant that deals 1 pocket card on the river, and no
// board cards on the river.
func WithRiver(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Once = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 0)
		desc.Streets[3].Pocket = 1
		desc.Apply(opts...)
	}
}

// WithDallas is a type description option to set Dallas definitions.
//
// Dallas is Holdem variant that forces the use of the 2 pocket cards and any 3
// board cards. Essentially Omaha, but with 2 pocket cards instead of 4.
func WithDallas(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Eval = EvalDallas
		desc.Apply(opts...)
	}
}

// WithHouston is a type description option to set Houston definitions.
//
// Houston is a Holdem variant with 3 pocket cards, 2 board cards dealt on the
// Flop, and requires using 2 of the 3 pocket cards, and any 3 board cards to
// make the best-5 cards. Essentially Omaha, but with 3 pocket cards instead of
// 4.
func WithHouston(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(3, 1, 2, 1, 1)
		desc.Eval = EvalHouston
		desc.Apply(opts...)
	}
}

// WithDraw is a type description option to set Draw definitions.
//
// Draw is a best-5 card game played with a standard (French) deck, comprising
// a pocket of 5 cards, no community cards, with a Ante, 6th, and River streets.
// 5 cards are dealt on the Ante, and up to 5 cards can be drawn (exchanged) on
// the 6th street.
func WithDraw(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Blinds = StudBlinds()
		desc.Streets = NumberedStreets(5, 0, 0)
		desc.Streets[1].PocketDraw = 5
		desc.Apply(opts...)
	}
}

// WithStud is a type description option to set Stud definitions.
//
// Stud is a best-5 card game, played with a standard (French) deck, comprising
// a pocket of 7 cards, no community cards, with Ante, 4th, 5th, 6th and River
// streets. 3 pocket cards are dealt on the Ante, with 1 up, and 1 pocket card
// is dealt up on the 4th, 5th, and 6th streets, with a final pocket card dealt
// down on the 7th street.
func WithStud(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Low = low
		desc.Blinds = StudBlinds()
		desc.Streets = StudStreets()
		desc.Apply(opts...)
	}
}

// WithStudFive is a type description option to set StudFive definitions.
//
// StudFive is a best-5 card game palyed with a standard (French) deck,
// comprising a pocket of 5 cards, no community cards, with Ante, 3rd, 4th, and
// River streets. 2 pocket cards are dealt on the Ante, with 1 up, and 1 pocket
// card dealt up on the 3rd, 4th, and 5th streets.
func WithStudFive(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Low = low
		desc.Blinds = StudBlinds()
		desc.Streets = StudFiveStreets()
		desc.Apply(opts...)
	}
}

// WithVideo is a type description option to set Video definitions.
//
// Video is a best-5 card game, played with a standard (French) deck,
// comprising a pocket of 5 cards, no community cards, with a Ante and River. 5
// pocket cards are dealt on the Ante, all up. Up to 5 cards can be drawn
// (exchanged) on the River.
func WithVideo(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 1
		desc.Low = low
		desc.Blinds = StudBlinds()
		desc.Streets = NumberedStreets(5, 0)
		desc.Streets[0].PocketUp = 5
		desc.Streets[1].PocketDraw = 5
		desc.Eval = EvalJacksOrBetter
		desc.Apply(opts...)
	}
}

// WithOmaha is a type description option to set Omaha definitions.
//
// Omaha is a Holdem variant with 4 pocket cards, requiring the use of 2 of the
// pocket and 3 of the board to make the best-5 cards.
func WithOmaha(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(4, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.Apply(opts...)
	}
}

// WithOmahaDouble is a type description option to set OmahaDouble definitions.
//
// OmahaDouble is a Omaha variant where there are two community boards dealt in
// tandem, comprising the Hi and Lo.
func WithOmahaDouble(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Double = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(4, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.LoDesc = DescCactus
		desc.Apply(opts...)
	}
}

// WithOmahaFive is a type description option to set OmahaFive definitions.
//
// Omaha is a Holdem variant with 5 pocket cards, requiring the use of 2 of the
// pocket and 3 of the board to make the best-5 cards.
func WithOmahaFive(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(5, 0, 3, 1, 1)
		desc.Eval = EvalOmahaFive
		desc.Apply(opts...)
	}
}

// WithOmahaSix is a type description option to set OmahaSix definitions.
//
// Omaha is a Holdem variant with 6 pocket cards, requiring the use of 2 of the
// pocket and 3 of the board to make the best-5 cards.
func WithOmahaSix(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(6, 0, 3, 1, 1)
		desc.Eval = EvalOmahaSix
		desc.Apply(opts...)
	}
}

// WithCourchevel is a type description option to set Courchevel definitions.
//
// Courchevel is a OmahaFive variant, where 1 board card is dealt on the
// Pre-Flop, and 2 on the Flop.
func WithCourchevel(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(5, 0, 3, 1, 1)
		desc.Eval = EvalOmahaFive
		// pre-flop
		desc.Streets[0].Pocket = 5
		desc.Streets[0].Board = 1
		desc.Streets[0].BoardDiscard = 1
		// flop
		desc.Streets[1].Board = 2
		desc.Streets[1].BoardDiscard = 0
		desc.Apply(opts...)
	}
}

// WithFusion is a type description option to set Fusion definitions.
//
// Fusion is a Omaha variant where only 2 pocket cards are dealt on the
// Pre-Flop, with 1 additional pocket card dealt on the Flop and Turn.
func WithFusion(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		// flop and turn get additional pocket
		desc.Streets[1].Pocket = 1
		desc.Streets[2].Pocket = 1
		desc.Apply(opts...)
	}
}

// WithSoko is a type description option to set Soko definitions.
//
// Soko is a Holdem variant with 2 additional hand ranks, a Four Flush
// comprising 4 cards of the same suit, and a Four Straight, comprising 4 cards
// in sequential rank (no wrapping straights), that beat Pair and Nothing.
func WithSoko(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Low = low
		desc.Streets = NumberedStreets(2, 3)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalSoko
		desc.HiDesc = DescSoko
		desc.Apply(opts...)
	}
}

// WithLowball is a type description option to set Lowball definitions.
//
// Lowball is a best-5 low game played with a standard (French) deck,
// comprising 5 pocket cards, no community cards, and a Ante, 6th, 7th, and
// River streets using a 2-to-7 low ("Lowball") inverted ranking system, where
// Aces always play high, and non-Flush, and non-Straight lows are best. For
// multi variants, up to 5 cards can be drawn (exchanged) on the 6th, 7th, and
// River streets. Non-multi variants may only exchange up to 5 cards only once
// on either the 6th, 7th, or River streets.
func WithLowball(multi bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Once = !multi
		desc.Streets = NumberedStreets(5, 0, 0, 0)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalLowball
		desc.HiDesc = DescLowball
		for i := 1; i < 4; i++ {
			desc.Streets[i].PocketDraw = 5
		}
		desc.Apply(opts...)
	}
}

// WithRazz is a type description option to set Razz definitions.
//
// Razz is a Stud variant, using a A-to-5 low ranking, where Aces play low, and
// Flushes and Straights do not affect ranking.
func WithRazz(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Blinds = HoldemBlinds()
		desc.Streets = StudStreets()
		desc.Eval = EvalRazz
		desc.HiDesc = DescRazz
		desc.Apply(opts...)
	}
}

// WithBadugi is a type description option to set Badugi definitions.
//
// Badugi is a best-4, lowest non-matching-suit card game played with a
// standard (French) deck, comprising a pocket of 4 cards, no community cards,
// and Ante, 5th, 6th, and River streets.
//
//	4 cards, low evaluation of separate suits
//	All 4 face down pre-flop
//	3 rounds of player discards (up to 4)
func WithBadugi(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Streets = NumberedStreets(4, 0, 0, 0)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalBadugi
		desc.HiDesc = DescLow
		for i := 1; i < 4; i++ {
			desc.Streets[i].PocketDraw = 4
		}
		desc.Apply(opts...)
	}
}

// StreetDesc is a type's street description.
type StreetDesc struct {
	// Id is the id of the street.
	Id byte
	// Name is the name of the street.
	Name string
	// Pocket is the count of cards to deal.
	Pocket int
	// PocketUp is the count of cards to reveal.
	PocketUp int
	// PocketDiscard is the count of cards to discard before pockets dealt.
	PocketDiscard int
	// PocketDraw is the count of cards to draw.
	PocketDraw int
	// Board is the count of board cards to deal.
	Board int
	// BoardDiscard is the count of cards to discard before board dealt.
	BoardDiscard int
}

// Desc returns a description of the street.
func (desc StreetDesc) Desc() string {
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
	if 0 < desc.PocketDraw {
		v = append(v, fmt.Sprintf("w: %d", desc.PocketDraw))
	}
	var s string
	if len(v) != 0 {
		s = " (" + strings.Join(v, ", ") + ")"
	}
	return fmt.Sprintf("%c: %s%s", desc.Id, desc.Name, s)
}

// HoldemBlinds returns the Holdem blind names.
func HoldemBlinds() []string {
	return []string{
		"Small Blind",
		"Big Blind",
		"Straddle",
	}
}

// StudBlinds returns the Stud blind names.
func StudBlinds() []string {
	return []string{
		"Ante",
		"Bring In",
	}
}

// HoldemStreets creates Holdem streets for the pre-flop, flop, turn, and
// river.
func HoldemStreets(pocket, discard, flop, turn, river int) []StreetDesc {
	d := func(id byte, name string, pocket int, board int) StreetDesc {
		n := discard
		if id == 'p' {
			n = 0
		}
		return StreetDesc{
			Id:           id,
			Name:         name,
			Pocket:       pocket,
			Board:        board,
			BoardDiscard: n,
		}
	}
	return []StreetDesc{
		d('p', "Pre-Flop", pocket, 0),
		d('f', "Flop", 0, flop),
		d('t', "Turn", 0, turn),
		d('r', "River", 0, river),
	}
}

// StudStreets creates Stud streets for the ante, third street, fourth street,
// fifth street, sixth street and river.
func StudStreets() []StreetDesc {
	v := NumberedStreets(3, 1, 1, 1, 1)
	for i := 0; i < 4; i++ {
		v[i].PocketUp = 1
	}
	return v
}

// StudFiveStreets creates Stud streets for the ante, third street, fourth
// street, and river.
func StudFiveStreets() []StreetDesc {
	v := NumberedStreets(2, 1, 1, 1)
	for i := 0; i < 4; i++ {
		v[i].PocketUp = 1
	}
	return v
}

// NumberedStreets returns numbered streets (ante, first, second, ...).
func NumberedStreets(pockets ...int) []StreetDesc {
	var v []StreetDesc
	var count, total int
	for i := 0; i < len(pockets); i++ {
		total += pockets[i]
	}
	for i := 0; i < len(pockets); i++ {
		count += pockets[i]
		name, id := ordinal(count), '0'+byte(count)
		switch {
		case i == 0:
			name = "Ante"
		case i == len(pockets)-1:
			name = "River"
			if pockets[i] == 0 && i != 0 {
				id = v[i-1].Id + 1
			}
		case i != 0 && pockets[i] == 0:
			n := int(v[i-1].Id-'0') + 1
			name, id = ordinal(n), '0'+byte(n)
		}
		v = append(v, StreetDesc{
			Id:     id,
			Name:   name,
			Pocket: pockets[i],
		})
	}
	return v
}

// EvalType is a eval type.
type EvalType uint8

// Eval types.
const (
	EvalCactus        EvalType = 0
	EvalJacksOrBetter EvalType = 'j'
	EvalShort         EvalType = 't'
	EvalManila        EvalType = 'm'
	EvalSpanish       EvalType = 'p'
	EvalDallas        EvalType = 'a'
	EvalHouston       EvalType = 'u'
	EvalOmaha         EvalType = 'o'
	EvalOmahaFive     EvalType = 'v'
	EvalOmahaSix      EvalType = 'i'
	EvalSoko          EvalType = 'k'
	EvalLowball       EvalType = 'l'
	EvalRazz          EvalType = 'r'
	EvalBadugi        EvalType = 'b'
)

// New creates the eval type.
func (typ EvalType) New(low bool) EvalFunc {
	switch typ {
	case EvalCactus:
		return NewCactusEval(low)
	case EvalJacksOrBetter:
		return NewJacksOrBetterEval(low)
	case EvalShort:
		return NewShortEval()
	case EvalManila:
		return NewManilaEval()
	case EvalSpanish:
		return NewSpanishEval()
	case EvalDallas:
		return NewDallasEval(RankCactus, Five, low, nil)
	case EvalHouston:
		return NewHoustonEval(RankCactus, Five, nil)
	case EvalOmaha:
		return NewOmahaEval(low)
	case EvalOmahaFive:
		return NewOmahaFiveEval(low)
	case EvalOmahaSix:
		return NewOmahaSixEval(low)
	case EvalSoko:
		return NewSokoEval(low)
	case EvalLowball:
		return NewLowballEval()
	case EvalRazz:
		return NewRazzEval()
	case EvalBadugi:
		return NewBadugiEval()
	}
	return nil
}

// Format satisfies the fmt.Formatter interface.
func (typ EvalType) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'd':
		buf = []byte(strconv.Itoa(int(typ)))
	case 'c':
		buf = []byte{typ.Byte()}
	case 's', 'v':
		buf = []byte(typ.Name())
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, eval: %d)", verb, int(typ)))
	}
	_, _ = f.Write(buf)
}

// Byte returns the eval type as a byte.
func (typ EvalType) Byte() byte {
	switch typ {
	case EvalCactus:
		return 'c'
	case EvalJacksOrBetter,
		EvalShort,
		EvalManila,
		EvalSpanish,
		EvalDallas,
		EvalHouston,
		EvalOmaha,
		EvalOmahaFive,
		EvalOmahaSix,
		EvalSoko,
		EvalLowball,
		EvalRazz,
		EvalBadugi:
		return byte(typ)
	}
	return ' '
}

// Name returns the eval type's name.
func (typ EvalType) Name() string {
	switch typ {
	case EvalCactus:
		return "Cactus"
	case EvalJacksOrBetter:
		return "JacksOrBetter"
	case EvalShort:
		return "Short"
	case EvalManila:
		return "Manila"
	case EvalSpanish:
		return "Spanish"
	case EvalDallas:
		return "Dallas"
	case EvalHouston:
		return "Houston"
	case EvalOmaha:
		return "Omaha"
	case EvalOmahaFive:
		return "OmahaFive"
	case EvalOmahaSix:
		return "OmahaSix"
	case EvalSoko:
		return "Soko"
	case EvalLowball:
		return "Lowball"
	case EvalRazz:
		return "Razz"
	case EvalBadugi:
		return "Badugi"
	}
	return ""
}

// DescType is a description type.
type DescType uint8

// Desc types.
const (
	DescCactus    DescType = 0
	DescFlushOver DescType = 'f'
	DescSoko      DescType = 'k'
	DescLow       DescType = 'l'
	DescLowball   DescType = 'b'
	DescRazz      DescType = 'r'
)

// Format satisfies the fmt.Formatter interface.
func (typ DescType) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'd':
		buf = []byte(strconv.Itoa(int(typ)))
	case 'c':
		buf = []byte{typ.Byte()}
	case 's', 'v':
		buf = []byte(typ.Name())
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, desc: %d)", verb, int(typ)))
	}
	_, _ = f.Write(buf)
}

// Byte returns the desc type as a byte.
func (typ DescType) Byte() byte {
	switch typ {
	case DescCactus:
		return 'c'
	case DescFlushOver,
		DescSoko,
		DescLow,
		DescLowball,
		DescRazz:
		return byte(typ)
	}
	return ' '
}

// Name returns the desc type's name.
func (typ DescType) Name() string {
	switch typ {
	case DescCactus:
		return "Cactus"
	case DescFlushOver:
		return "FlushOver"
	case DescSoko:
		return "Soko"
	case DescLow:
		return "Low"
	case DescLowball:
		return "Lowball"
	case DescRazz:
		return "Razz"
	}
	return ""
}

// Desc writes a description for the verb to f.
func (typ DescType) Desc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch verb {
	case 'd':
		fmt.Fprintf(f, "%d", int(rank))
	case 'u':
		CardFormatter(unused).Format(f, 's')
	case 'v', 's':
		switch typ {
		case DescCactus:
			CactusDesc(f, verb, rank, best, unused)
		case DescLow:
			LowDesc(f, verb, rank, best, unused)
		case DescFlushOver:
			FlushOverDesc(f, verb, rank, best, unused)
		case DescRazz:
			RazzDesc(f, verb, rank, best, unused)
		case DescLowball:
			LowballDesc(f, verb, rank, best, unused)
		case DescSoko:
			SokoDesc(f, verb, rank, best, unused)
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, desc: %d)", verb, int(typ))
	}
}

// StreetOption is a street option.
type StreetOption func(int, *StreetDesc)

// WithStreetPocket is a street option to set the pocket for a street.
func WithStreetPocket(street, pocket int) StreetOption {
	return func(n int, desc *StreetDesc) {
		if n == street {
			desc.Pocket = pocket
		}
	}
}

// CactusDesc writes a Cactus description to f for the rank, best, and unused
// cards.
//
// Examples:
//
//	Straight Flush, Ace-high, Royal
//	Straight Flush, King-high, Platinum Oxide
//	Straight Flush, Five-high, Steel Wheel
//	Four of a Kind, Nines, kicker Jack
//	Full House, Sixes full of Fours
//	Flush, Ten-high
//	Straight, Eight-high
//	Three of a Kind, Fours, kickers Ace, King
//	Two Pair, Nines over Sixes, kicker Jack
//	Pair, Aces, kickers King, Queen, Nine
//	Seven-high, kickers Six, Five, Three, Two
func CactusDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch rank.Fixed() {
	case StraightFlush:
		fmt.Fprintf(f, "Straight Flush, %N-high, %F", best[0], best[0])
	case FourOfAKind:
		fmt.Fprintf(f, "Four of a Kind, %P, kicker %N", best[0], best[4])
	case FullHouse:
		fmt.Fprintf(f, "Full House, %P full of %P", best[0], best[3])
	case Flush:
		fmt.Fprintf(f, "Flush, %N-high, kickers %N, %N, %N, %N", best[0], best[1], best[2], best[3], best[4])
	case Straight:
		fmt.Fprintf(f, "Straight, %N-high", best[0])
	case ThreeOfAKind:
		fmt.Fprintf(f, "Three of a Kind, %P, kickers %N, %N", best[0], best[3], best[4])
	case TwoPair:
		fmt.Fprintf(f, "Two Pair, %P over %P, kicker %N", best[0], best[2], best[4])
	case Pair:
		fmt.Fprintf(f, "Pair, %P, kickers %N, %N, %N", best[0], best[2], best[3], best[4])
	case Nothing:
		fmt.Fprintf(f, "%N-high, kickers %N, %N, %N, %N", best[0], best[1], best[2], best[3], best[4])
	default:
		fmt.Fprint(f, "None")
	}
}

// FlushOverDesc writes a FlushOver description to f for the rank, best, and
// unused cards.
func FlushOverDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	CactusDesc(f, verb, rank.FromFlushOver(), best, unused)
}

// SokoDesc writes a Soko description to f for the rank, best, and unused cards.
func SokoDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch {
	case rank <= TwoPair:
		CactusDesc(f, verb, rank, best, unused)
	case rank <= sokoFlush:
		fmt.Fprintf(f, "Four Flush, %N-high, kickers %N, %N, %N, %N", best[0], best[1], best[2], best[3], best[4])
	case rank <= sokoStraight:
		fmt.Fprintf(f, "Four Straight, %N-high, kicker %N", best[0], best[4])
	default:
		CactusDesc(f, verb, rank-sokoStraight+TwoPair, best, unused)
	}
}

// LowDesc writes a Low description to f for the rank, best, and unused cards.
func LowDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch {
	case rank == 0, rank == Invalid:
		_, _ = f.Write([]byte("None"))
	default:
		for i := 0; i < len(best); i++ {
			if i != 0 {
				_, _ = f.Write([]byte(", "))
			}
			best[i].Format(f, 'N')
		}
		_, _ = f.Write([]byte("-low"))
	}
}

// LowballDesc writes a Lowball description to f for the rank, best, and unused
// cards.
func LowballDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch r := rank.FromLowball(); {
	case rank <= StraightFlush:
		LowDesc(f, verb, r, best, unused)
		fmt.Fprintf(f, ", No. %d", rank)
	case Pair < r && r <= Nothing || r == Straight:
		LowDesc(f, verb, r, best, unused)
	case r == StraightFlush:
		CactusDesc(f, verb, Flush, best, unused)
	default:
		CactusDesc(f, verb, r, best, unused)
	}
}

// RazzDesc writes a Razz description to f for the rank, best, and unused
// cards.
func RazzDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card) {
	switch {
	case rank < aceFiveMax:
		LowDesc(f, verb, rank, best, unused)
	default:
		CactusDesc(f, verb, Invalid-rank, best, unused)
	}
}

// ordinal returns the ordinal string for n (1st, 2nd, ...).
func ordinal(n int) string {
	switch p, q := n%10, n%100; {
	case p == 1 && q != 11:
		return fmt.Sprintf("%dst", n)
	case p == 2 && q != 12:
		return fmt.Sprintf("%dnd", n)
	case p == 3 && q != 13:
		return fmt.Sprintf("%drd", n)
	}
	return fmt.Sprintf("%dth", n)
}

package cardrank

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Type is a hand eval type.
type Type uint16

// Hand eval types.
const (
	Holdem         Type = 'H'<<8 | 'h' // Hh
	Short          Type = 'H'<<8 | 's' // Hs
	Manila         Type = 'H'<<8 | 'm' // Hm
	Royal          Type = 'H'<<8 | 'r' // Hr
	Double         Type = 'H'<<8 | 'd' // Hd
	Showtime       Type = 'H'<<8 | 't' // Ht
	Swap           Type = 'H'<<8 | 'w' // Hw
	Omaha          Type = 'O'<<8 | '4' // O4
	OmahaHiLo      Type = 'O'<<8 | 'l' // Ol
	OmahaDouble    Type = 'O'<<8 | 'd' // Od
	OmahaFive      Type = 'O'<<8 | '5' // O5
	OmahaSix       Type = 'O'<<8 | '6' // O6
	Courchevel     Type = 'O'<<8 | 'c' // Oc
	CourchevelHiLo Type = 'O'<<8 | 'e' // Oe
	Fusion         Type = 'O'<<8 | 'f' // Of
	FusionHiLo     Type = 'O'<<8 | 'F' // OF
	Stud           Type = 'S'<<8 | 'h' // Sh
	StudHiLo       Type = 'S'<<8 | 'l' // Sl
	Razz           Type = 'R'<<8 | 'a' // Ra
	Badugi         Type = 'B'<<8 | 'a' // Ba
	Lowball        Type = 'L'<<8 | '1' // L1
	LowballTriple  Type = 'L'<<8 | '3' // L3
	Soko           Type = 'K'<<8 | 'o' // Ko
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
		{"Hh", Holdem, "Holdem", WithHoldem()},
		{"Hs", Short, "Short", WithShort()},
		{"Hm", Manila, "Manila", WithManila()},
		{"Hr", Royal, "Royal", WithRoyal()},
		{"Hd", Double, "Double", WithDouble()},
		{"Ht", Showtime, "Showtime", WithShowtime()},
		{"Hw", Swap, "Swap", WithSwap()},
		{"O4", Omaha, "Omaha", WithOmaha(false)},
		{"Ol", OmahaHiLo, "OmahaHiLo", WithOmaha(true)},
		{"Od", OmahaDouble, "OmahaDouble", WithOmahaDouble()},
		{"O5", OmahaFive, "OmahaFive", WithOmahaFive(false)},
		{"O6", OmahaSix, "OmahaSix", WithOmahaSix(false)},
		{"Oc", Courchevel, "Courchevel", WithCourchevel(false)},
		{"Oe", CourchevelHiLo, "CourchevelHiLo", WithCourchevel(true)},
		{"Of", Fusion, "Fusion", WithFusion(false)},
		{"OF", FusionHiLo, "FusionHiLo", WithFusion(true)},
		{"Sh", Stud, "Stud", WithStud(false)},
		{"Sl", StudHiLo, "StudHiLo", WithStud(true)},
		{"Ra", Razz, "Razz", WithRazz()},
		{"Ba", Badugi, "Badugi", WithBadugi()},
		{"L1", Lowball, "Lowball", WithLowball(false)},
		{"L3", LowballTriple, "LowballTriple", WithLowball(true)},
		{"Ko", Soko, "Soko", WithSoko()},
	} {
		desc, err := NewTypeDesc(d.id, d.typ, d.name, d.opt)
		if err != nil {
			panic(err)
		}
		v = append(v, *desc)
	}
	return v
}

// Types returns the registered types.
func Types() []Type {
	var v []TypeDesc
	for _, desc := range descs {
		v = append(v, desc)
	}
	sort.Slice(v, func(i, j int) bool {
		return v[i].Num < v[j].Num
	})
	types := make([]Type, len(v))
	for i := 0; i < len(types); i++ {
		types[i] = v[i].Type
	}
	return types
}

// MarshalText satisfies the encoding.TextMarshaler interface.
func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
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

// String satisfies the fmt.Stringer interface.
func (typ Type) String() string {
	return string([]byte{byte(typ >> 8 & 0xf), byte(typ & 0xf)})
}

// Format satisfies the fmt.Formatter interface.
func (typ Type) Format(f fmt.State, verb rune) {
	switch verb {
	case 'c':
		fmt.Fprint(f, typ.String())
		return
	}
	if desc, ok := descs[typ]; ok {
		fmt.Fprint(f, desc.Name)
	} else {
		fmt.Fprintf(f, "Type(%d)", typ)
	}
}

// Desc returns the type description.
func (typ Type) Desc() TypeDesc {
	return descs[typ]
}

// Name returns the type name.
func (typ Type) Name() string {
	return descs[typ].Name
}

// Max returns the type max players.
func (typ Type) Max() int {
	return descs[typ].Max
}

// Low returns true when the type has a low board.
func (typ Type) Low() bool {
	return descs[typ].Low
}

// Double returns true when the type has a double board.
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

// Streets returns the type's street names.
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

// HiComp returns a hi compare func.
func (typ Type) HiComp() func(*Eval, *Eval) int {
	f, low := descs[typ].HiComp.Comp, descs[typ].Low
	loMax := Invalid
	if low {
		loMax = rankEightOrBetterMax
	}
	return func(a, b *Eval) int {
		switch {
		case a == nil && b == nil:
			return -1
		case a == nil:
			return +1
		case b == nil:
			return -1
		}
		return f(a, b, loMax)
	}
}

// LoComp returns a lo compare func.
func (typ Type) LoComp() func(*Eval, *Eval) int {
	f, low := descs[typ].LoComp.Comp, descs[typ].Low
	loMax := Invalid
	if low {
		loMax = rankEightOrBetterMax
	}
	return func(a, b *Eval) int {
		switch {
		case a == nil && b == nil:
			return -1
		case a == nil:
			return +1
		case b == nil:
			return -1
		}
		return f(a, b, loMax)
	}
}

// HiDesc returns the type's hi desc type.
func (typ Type) HiDesc() DescType {
	return descs[typ].HiDesc
}

// LoDesc returns the type's lo desc type.
func (typ Type) LoDesc() DescType {
	return descs[typ].LoDesc
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
		return d.Pockets, d.Boards[0].Hi
	}
	return nil, nil
}

// New creates a new eval for the type, and evaluates the pocket and board.
func (typ Type) New(pocket, board []Card) *Eval {
	return EvalOf(typ).Eval(pocket, board)
}

// Eval creates a new eval for the type for each of the pockets and board.
func (typ Type) Eval(pockets [][]Card, board []Card) []*Eval {
	evs := make([]*Eval, len(pockets))
	for i := 0; i < len(pockets); i++ {
		evs[i] = typ.New(pockets[i], board)
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
	// Low is true when the type is a Hi/Lo variant.
	Low bool
	// Double is true when there are two boards.
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
	// HiComp is the hi compare type.
	HiComp CompType
	// LoComp is the lo compare type.
	LoComp CompType
	// HiDesc is the hi desc type.
	HiDesc DescType
	// LoDesc is the lo desc type.
	LoDesc DescType
}

// NewTypeDesc creates a new type description.
func NewTypeDesc(id string, typ Type, name string, opts ...TypeOption) (*TypeDesc, error) {
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
		HiComp: CompHi,
		LoComp: CompLo,
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
func WithHoldem(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Apply(opts...)
	}
}

// WithShort is a type description option to set Short definitions.
func WithShort(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 6
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Deck = DeckShort
		desc.Eval = EvalShort
		desc.HiComp = CompShort
		desc.HiDesc = DescShort
		desc.LoDesc = DescShort
		desc.Apply(opts...)
	}
}

// WithManila is a type description option to set Manila definitions.
func WithManila(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 6
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 0, 1, 1, 1)
		desc.Deck = DeckManila
		desc.Eval = EvalManila
		desc.HiComp = CompShort
		desc.HiDesc = DescManila
		desc.LoDesc = DescManila
		desc.Streets[0].Board = 1
		desc.Streets = append(desc.Streets[:2], append([]StreetDesc{
			{
				Id:    'l',
				Name:  "Flop",
				Board: 1,
			},
		}, desc.Streets[2:]...)...)
		desc.Apply(opts...)
	}
}

// WithRoyal is a type description option to set Royal definitions.
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
func WithDouble(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Double = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.LoComp = CompLo
		desc.LoDesc = DescCactus
		desc.Apply(opts...)
	}
}

// WithShowtime is a type description option to set Showtime definitions.
func WithShowtime(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Show = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Apply(opts...)
	}
}

// WithSwap is a type description option to set Swap definitions.
func WithSwap(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 10
		desc.Once = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		for i := 1; i < len(desc.Streets); i++ {
			desc.Streets[i].PocketDraw = 2
		}
		desc.Apply(opts...)
	}
}

// WithOmaha is a type description option to set Omaha definitions.
func WithOmaha(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(4, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.LoComp = CompLo
		desc.Apply(opts...)
	}
}

// WithOmahaDouble is a type description option to set OmahaDouble definitions.
func WithOmahaDouble(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Double = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(4, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.LoComp = CompHi
		desc.LoDesc = DescCactus
		desc.Apply(opts...)
	}
}

// WithOmahaFive is a type description option to set OmahaFive definitions.
func WithOmahaFive(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(5, 0, 3, 1, 1)
		desc.Eval = EvalOmahaFive
		desc.LoComp = CompLo
		desc.Apply(opts...)
	}
}

// WithOmahaSix is a type description option to set OmahaSix definitions.
func WithOmahaSix(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(6, 0, 3, 1, 1)
		desc.Eval = EvalOmahaSix
		desc.LoComp = CompLo
		desc.Apply(opts...)
	}
}

// WithCourchevel is a type description option to set Courchevel definitions.
func WithCourchevel(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(5, 0, 3, 1, 1)
		desc.Eval = EvalOmahaFive
		desc.LoComp = CompLo
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
func WithFusion(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Low = low
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(2, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.LoComp = CompLo
		// flop and turn get additional pocket
		desc.Streets[1].Pocket = 1
		desc.Streets[2].Pocket = 1
		desc.Apply(opts...)
	}
}

// WithStud is a type description option to set Stud definitions.
func WithStud(low bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Low = low
		desc.Blinds = StudBlinds()
		desc.Streets = StudStreets()
		desc.Eval = EvalStud
		desc.LoComp = CompLo
		desc.Apply(opts...)
	}
}

// WithRazz is a type description option to set Razz definitions.
//
// Same as Stud, but with a Ace-to-Five low card ranking.
func WithRazz(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Blinds = HoldemBlinds()
		desc.Streets = StudStreets()
		desc.Eval = EvalRazz
		desc.HiDesc = DescRazz
		desc.LoDesc = DescRazz
		desc.Apply(opts...)
	}
}

// WithBadugi is a type description option to set Badugi definitions.
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
		desc.LoDesc = DescLow
		for i := 1; i < 4; i++ {
			desc.Streets[i].PocketDraw = 4
		}
		desc.Apply(opts...)
	}
}

// WithLowball is a type description option to set Lowball definitions.
func WithLowball(multi bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Once = !multi
		desc.Streets = NumberedStreets(5, 0, 0, 0)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalLowball
		desc.HiDesc = DescLowball
		desc.LoDesc = DescLowball
		desc.HiComp = CompLowball
		for i := 1; i < 4; i++ {
			desc.Streets[1].PocketDraw = 5
		}
		desc.Apply(opts...)
	}
}

// WithSoko is a type description option to set Soko definitions.
func WithSoko(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Streets = NumberedStreets(2, 3)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalSoko
		desc.HiDesc = DescSoko
		desc.LoDesc = DescSoko
		desc.HiComp = CompSoko
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
		v[0].PocketUp = 1
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
	EvalCactus    EvalType = 0
	EvalShort     EvalType = 't'
	EvalManila    EvalType = 'm'
	EvalOmaha     EvalType = 'o'
	EvalOmahaFive EvalType = 'v'
	EvalOmahaSix  EvalType = 'i'
	EvalStud      EvalType = 's'
	EvalRazz      EvalType = 'r'
	EvalBadugi    EvalType = 'b'
	EvalLowball   EvalType = '2'
	EvalSoko      EvalType = 'k'
)

// New creates the eval type.
func (typ EvalType) New(low bool) EvalFunc {
	switch typ {
	case EvalCactus:
		return NewCactusEval(DefaultEval, Five)
	case EvalShort:
		return NewShortEval()
	case EvalManila:
		return NewManilaEval()
	case EvalOmaha, EvalOmahaFive, EvalOmahaSix, EvalStud:
		loMax := Invalid
		if low {
			loMax = rankEightOrBetterMax
		}
		switch typ {
		case EvalOmaha:
			return NewOmahaEval(loMax)
		case EvalOmahaFive:
			return NewOmahaFiveEval(loMax)
		case EvalOmahaSix:
			return NewOmahaSixEval(loMax)
		case EvalStud:
			return NewStudEval(loMax)
		}
	case EvalRazz:
		return NewRazzEval()
	case EvalBadugi:
		return NewBadugiEval()
	case EvalLowball:
		return NewLowballEval()
	case EvalSoko:
		return NewSokoEval()
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
		return 'h'
	case EvalShort,
		EvalManila,
		EvalOmaha,
		EvalOmahaFive,
		EvalOmahaSix,
		EvalStud,
		EvalRazz,
		EvalBadugi,
		EvalLowball,
		EvalSoko:
		return byte(typ)
	}
	return ' '
}

// Name returns the eval type's name.
func (typ EvalType) Name() string {
	switch typ {
	case EvalCactus:
		return "Cactus"
	case EvalShort:
		return "Short"
	case EvalManila:
		return "Manila"
	case EvalOmaha:
		return "Omaha"
	case EvalOmahaFive:
		return "OmahaFive"
	case EvalOmahaSix:
		return "OmahaSix"
	case EvalStud:
		return "Stud"
	case EvalRazz:
		return "Razz"
	case EvalBadugi:
		return "Badugi"
	case EvalLowball:
		return "Lowball"
	case EvalSoko:
		return "Soko"
	}
	return ""
}

// CompType is a compare type.
type CompType uint8

// Comp types.
const (
	CompHi      CompType = 0
	CompLo      CompType = 'l'
	CompShort   CompType = 's'
	CompManila  CompType = 'm'
	CompLowball CompType = '2'
	CompSoko    CompType = 'k'
)

// Comp compares a, b.
func (typ CompType) Comp(a, b *Eval, loMax EvalRank) int {
	switch typ {
	case CompHi:
		return HiComp(a, b, loMax)
	case CompLo:
		return LoComp(a, b, loMax)
	case CompShort:
		return ShortComp(a, b, loMax)
	case CompManila:
		return ManilaComp(a, b, loMax)
	case CompLowball:
		return LowballComp(a, b, loMax)
	case CompSoko:
		return SokoComp(a, b, loMax)
	}
	return 0
}

// Format satisfies the fmt.Formatter interface.
func (typ CompType) Format(f fmt.State, verb rune) {
	var buf []byte
	switch verb {
	case 'd':
		buf = []byte(strconv.Itoa(int(typ)))
	case 'c':
		buf = []byte{typ.Byte()}
	case 's', 'v':
		buf = []byte(typ.Name())
	default:
		buf = []byte(fmt.Sprintf("%%!%c(ERROR=unknown verb, comp: %d)", verb, int(typ)))
	}
	_, _ = f.Write(buf)
}

// Byte returns the comp type as a byte.
func (typ CompType) Byte() byte {
	switch typ {
	case CompHi:
		return 'h'
	case CompLo,
		CompShort,
		CompManila,
		CompLowball,
		CompSoko:
		return byte(typ)
	}
	return ' '
}

// Name returns the comp type's name.
func (typ CompType) Name() string {
	switch typ {
	case CompHi:
		return "Hi"
	case CompLo:
		return "Lo"
	case CompShort:
		return "Short"
	case CompManila:
		return "Manila"
	case CompLowball:
		return "Lowball"
	case CompSoko:
		return "Soko"
	}
	return ""
}

// DescType is a description type.
type DescType uint8

// Desc types.
const (
	DescCactus  DescType = 0
	DescLow     DescType = 'l'
	DescShort   DescType = 's'
	DescManila  DescType = 'm'
	DescRazz    DescType = 'r'
	DescLowball DescType = '2'
	DescSoko    DescType = 'k'
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
		return 'h'
	case DescLow,
		DescShort,
		DescManila,
		DescRazz,
		DescLowball,
		DescSoko:
		return byte(typ)
	}
	return ' '
}

// Name returns the desc type's name.
func (typ DescType) Name() string {
	switch typ {
	case DescCactus:
		return "Cactus"
	case DescLow:
		return "Low"
	case DescShort:
		return "Short"
	case DescManila:
		return "Manila"
	case DescRazz:
		return "Razz"
	case DescLowball:
		return "Lowball"
	case DescSoko:
		return "Soko"
	}
	return ""
}

// Desc writes a description for the verb to f.
func (typ DescType) Desc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	switch verb {
	case 'd':
		fmt.Fprintf(f, "%d", int(rank))
	case 'u':
		CardFormatter(unused).Format(f, 's')
	case 'v', 's':
		switch typ {
		case DescCactus:
			CactusDesc(f, verb, rank, best, unused, low)
		case DescLow:
			LowDesc(f, verb, rank, best, unused, low)
		case DescShort:
			ShortDesc(f, verb, rank, best, unused, low)
		case DescManila:
			ManilaDesc(f, verb, rank, best, unused, low)
		case DescRazz:
			RazzDesc(f, verb, rank, best, unused, low)
		case DescLowball:
			LowballDesc(f, verb, rank, best, unused, low)
		case DescSoko:
			SokoDesc(f, verb, rank, best, unused, low)
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

// HiComp is a hi eval compare func.
func HiComp(a, b *Eval, _ EvalRank) int {
	switch {
	case a.HiRank < b.HiRank:
		return -1
	case b.HiRank < a.HiRank:
		return +1
	}
	return 0
}

// LoComp is a lo eval compare func.
func LoComp(a, b *Eval, loMax EvalRank) int {
	switch low := loMax != Invalid; {
	case low && a.LoRank == Invalid && b.LoRank != Invalid:
		return +1
	case low && b.LoRank == Invalid && a.LoRank != Invalid:
		return -1
	case a.LoRank < b.LoRank:
		return -1
	case b.LoRank < a.LoRank:
		return +1
	}
	return 0
}

// ShortComp is the Short compare func.
func ShortComp(a, b *Eval, _ EvalRank) int {
	switch af, bf := a.HiRank.Fixed(), b.HiRank.Fixed(); {
	case af == Flush && bf == FullHouse:
		return -1
	case af == FullHouse && bf == Flush:
		return +1
	case a.HiRank < b.HiRank:
		return -1
	case b.HiRank < a.HiRank:
		return +1
	}
	return 0
}

// ManilaComp is the Manila compare func.
func ManilaComp(a, b *Eval, _ EvalRank) int {
	switch af, bf := a.HiRank.Fixed(), b.HiRank.Fixed(); {
	case af == Flush && bf == FullHouse:
		return -1
	case af == FullHouse && bf == Flush:
		return +1
	case a.HiRank < b.HiRank:
		return -1
	case b.HiRank < a.HiRank:
		return +1
	}
	return 0
}

// LowballComp is the Lowball compare func.
func LowballComp(a, b *Eval, _ EvalRank) int {
	switch af, bf := rankMax-a.HiRank, rankMax-b.HiRank; {
	case af == bf:
	}
	// log.Printf("a: %v // b: %v", a.HiRank, b.HiRank)
	return 0
}

// SokoComp is the Soko compare func.
func SokoComp(a, b *Eval, _ EvalRank) int {
	return 0
}

// CactusDesc returns a Cactus description for the rank, best, and unused
// cards.
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
func CactusDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	// add additional straight flush names (at some point)
	// A: Royal
	// K: Platinum Oxide
	// Q: Silver Tongue
	// J: Bronze Fist
	// T: Golden Ratio
	// 9: Iron Maiden
	// 8: Tin Cup
	// 7: Brass Axe
	// 6: Aluminum Window
	// 5: Steel Wheel
	switch rank.Fixed() {
	case StraightFlush:
		switch best[0].Rank() {
		case Ace:
			fmt.Fprintf(f, "Straight Flush, %N-high, Royal", best[0])
		case Five:
			fmt.Fprintf(f, "Straight Flush, %N-high, Steel Wheel", best[0])
		default:
			fmt.Fprintf(f, "Straight Flush, %N-high", best[0])
		}
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
	default:
		fmt.Fprintf(f, "Nothing, %N-high, kickers %N, %N, %N, %N", best[0], best[1], best[2], best[3], best[4])
	}
}

// LowDesc returns a Low description for the rank, best, and unused cards.
func LowDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	switch {
	case rank == 0, rank == Invalid, !low:
		fmt.Fprint(f, "None")
	default:
		v := make([]string, len(best))
		for i := 0; i < len(best); i++ {
			v[i] = best[i].Rank().Name()
		}
		fmt.Fprint(f, strings.Join(v, ", ")+"-low")
	}
}

// ShortDesc returns a Short description for the rank, best, and unused cards.
func ShortDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	switch {
	case rank.Fixed() == StraightFlush && best[0].Rank() == Nine:
		fmt.Fprintf(f, "Straight Flush, %N-high, Iron Maiden", best[0])
	default:
		CactusDesc(f, verb, rank, best, unused, low)
	}
}

// ManilaDesc returns a Manila description for the rank, best, and unused cards.
func ManilaDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	switch {
	case rank.Fixed() == StraightFlush && best[0].Rank() == Ten:
		fmt.Fprintf(f, "Straight Flush, %N-high, Golden Ratio", best[0])
	default:
		CactusDesc(f, verb, rank, best, unused, low)
	}
}

// RazzDesc returns a Razz description for the rank, best, and unused cards.
func RazzDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	switch {
	case rank < rankAceFiveMax:
		LowDesc(f, verb, rank, best, unused, true)
	default:
		CactusDesc(f, verb, Invalid-rank, best, unused, false)
	}
}

// LowballDesc returns a Lowball description for the rank, best, and unused cards.
func LowballDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	fmt.Fprintf(f, "(lowball desc incomplete: %d)", rank)
}

// SokoDesc returns a Soko description for the rank, best, and unused cards.
func SokoDesc(f fmt.State, verb rune, rank EvalRank, best, unused []Card, low bool) {
	fmt.Fprintf(f, "(soko desc incomplete: %d)", rank)
}

// IdToType converts an id to a type.
func IdToType(id string) (Type, error) {
	switch {
	case len(id) != 2,
		!unicode.IsLetter(rune(id[0])) && !unicode.IsNumber(rune(id[0])),
		!unicode.IsLetter(rune(id[1])) && !unicode.IsNumber(rune(id[1])):
		return 0, ErrInvalidId
	}
	return Type(id[0])<<8 | Type(id[1]), nil
}

// bestHoldem sets the best holdem.
func bestHoldem(ev *Eval, v []Card, straightHigh Rank) {
	// order high to low
	sort.Slice(v, func(i, j int) bool {
		m, n := v[i].Rank(), v[j].Rank()
		if m == n {
			return v[j].Suit() < v[i].Suit()
		}
		return n < m
	})
	// set best, unused
	switch ev.HiRank.Fixed() {
	case StraightFlush:
		v = bestStraightFlush(v, straightHigh, true)
	case Flush:
		v = bestFlush(v)
	case Straight:
		v = bestStraight(v, straightHigh, true)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		v = bestSet(v)
	case Nothing:
	default:
		panic("bad rank")
	}
	ev.HiBest, ev.HiUnused = v[:5], v[5:]
}

// bestOmaha sets the best omaha on the eval.
func bestOmaha(ev *Eval, loMax EvalRank) {
	// order best
	sort.Slice(ev.HiBest, func(i, j int) bool {
		m, n := ev.HiBest[i].Rank(), ev.HiBest[j].Rank()
		if m == n {
			return ev.HiBest[j].Suit() < ev.HiBest[i].Suit()
		}
		return n < m
	})
	switch ev.HiRank.Fixed() {
	case StraightFlush:
		ev.HiBest = bestStraightFlush(ev.HiBest, Five, true)
	case Flush:
		ev.HiBest = bestFlush(ev.HiBest)
	case Straight:
		ev.HiBest = bestStraight(ev.HiBest, Five, true)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		ev.HiBest = bestSet(ev.HiBest)
	case Nothing:
	default:
		panic("bad rank")
	}
	if loMax != Invalid && ev.LoRank < loMax {
		sort.Slice(ev.LoBest, func(i, j int) bool {
			return ev.LoBest[j].AceIndex() < ev.LoBest[i].AceIndex()
		})
	} else {
		ev.LoBest, ev.LoUnused = nil, nil
	}
}

// bestLowball returns the best lowball.
func bestLowball(ev *Eval, v []Card) {
	// order high to low
	sort.Slice(v, func(i, j int) bool {
		m, n := v[i].Rank(), v[j].Rank()
		if m == n {
			return v[j].Suit() < v[i].Suit()
		}
		return n < m
	})
	// set best, unused
	switch ev.HiRank.Fixed() {
	case StraightFlush:
		ev.HiBest = bestStraightFlush(v, Seven, false)
	case Flush:
		ev.HiBest = bestFlush(v)
	case Straight:
		ev.HiBest = bestStraight(v, Seven, false)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		ev.HiBest = bestSet(v)
	case Nothing:
		ev.HiBest = v
	default:
		panic("bad rank")
	}
	bestHoldem(ev, v, Six)
}

// bestStraightFlush returns the best-five straight flush in v.
func bestStraightFlush(v []Card, high Rank, wrap bool) []Card {
	s := orderSuits(v)
	var b, d []Card
	for _, c := range v {
		switch c.Suit() {
		case s[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	return append(bestStraight(b, high, wrap), d...)
}

// bestFlush returns the best-five flush in v.
func bestFlush(v []Card) []Card {
	suits := orderSuits(v)
	var b, d []Card
	for _, c := range v {
		switch c.Suit() {
		case suits[0]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	return append(b, d...)
}

// bestStraight returns the best-five straight in v.
func bestStraight(v []Card, high Rank, wrap bool) []Card {
	m := make(map[Rank][]Card)
	for _, c := range v {
		r := c.Rank()
		m[r] = append(m[r], c)
	}
	var b []Card
	for i := Ace; high <= i; i-- {
		// last card index
		j := i - Six
		// check ace
		if i == high && wrap {
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
	for i := int(Ace); 0 <= i; i-- {
		if _, ok := m[Rank(i)]; ok && m[Rank(i)] != nil {
			d = append(d, m[Rank(i)]...)
		}
	}
	return append(b, d...)
}

// bestSet returns the best matching sets in v.
func bestSet(v []Card) []Card {
	ranks := orderRanks(v)
	var a, b, d []Card
	for _, c := range v {
		switch c.Rank() {
		case ranks[0]:
			a = append(a, c)
		case ranks[1]:
			b = append(b, c)
		default:
			d = append(d, c)
		}
	}
	return append(a, append(b, d...)...)
}

// orderSuits orders v's card suits by count.
func orderSuits(v []Card) []Suit {
	m := make(map[Suit]int)
	var suits []Suit
	for _, c := range v {
		s := c.Suit()
		if _, ok := m[s]; !ok {
			suits = append(suits, s)
		}
		m[s]++
	}
	sort.Slice(suits, func(i, j int) bool {
		if m[suits[i]] == m[suits[j]] {
			return suits[j] < suits[i]
		}
		return m[suits[j]] < m[suits[i]]
	})
	return suits
}

// orderRanks orders v's card ranks by count.
func orderRanks(v []Card) []Rank {
	m := make(map[Rank]int)
	var ranks []Rank
	for _, c := range v {
		r := c.Rank()
		if _, ok := m[r]; !ok {
			ranks = append(ranks, r)
		}
		m[r]++
	}
	sort.Slice(ranks, func(i, j int) bool {
		if m[ranks[i]] == m[ranks[j]] {
			return ranks[j] < ranks[i]
		}
		return m[ranks[j]] < m[ranks[i]]
	})
	return ranks
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

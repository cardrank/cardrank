package cardrank

import (
	"fmt"
	"sort"
	"strings"
)

// Type is a hand type.
type Type uint16

// Hand types.
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
	var types []TypeDesc
	for _, v := range []struct {
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
		{"Sh", Stud, "Stud", WithStud(false)},
		{"Sl", StudHiLo, "StudHiLo", WithStud(true)},
		{"Ra", Razz, "Razz", WithRazz()},
		{"Ba", Badugi, "Badugi", WithBadugi()},
		{"L1", Lowball, "Lowball", WithLowball(false)},
		{"L3", LowballTriple, "LowballTriple", WithLowball(true)},
		{"Ko", Soko, "Soko", WithSoko()},
	} {
		desc, err := NewTypeDesc(v.id, v.typ, v.name, v.opt)
		if err != nil {
			panic(err)
		}
		types = append(types, *desc)
	}
	return types
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

// Desc returns the type description.
func (typ Type) Desc() TypeDesc {
	return descs[typ]
}

// Name returns the type name.
func (typ Type) Name() string {
	if desc, ok := descs[typ]; ok {
		return desc.Name
	}
	return typ.String()
}

// Max returns the type max players.
func (typ Type) Max() int {
	if desc, ok := descs[typ]; ok {
		return desc.Max
	}
	return 0
}

// Double returns true when the type has a double board.
func (typ Type) Double() bool {
	return descs[typ].Double
}

// Show returns true when the type shows folded cards.
func (typ Type) Show() bool {
	return descs[typ].Show
}

// Low returns true when the type has a low board.
func (typ Type) Low() bool {
	return descs[typ].Low
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

// DeckType returns the type's deck type.
func (typ Type) DeckType() DeckType {
	return descs[typ].Deck
}

// Deck returns a new deck for the type.
func (typ Type) Deck() *Deck {
	if desc, ok := descs[typ]; ok {
		return desc.Deck.New()
	}
	return nil
}

// Eval evals the hand for the type.
func (typ Type) Eval(h *Hand) {
	evals[typ](h)
}

// HiCompare returns a hi compare func.
func (typ Type) HiCompare() func(*Hand, *Hand) int {
	f, low := descs[typ].HiComp.Compare, descs[typ].Low
	loMax := Invalid
	if low {
		loMax = rankEightOrBetterMax
	}
	return func(a, b *Hand) int {
		return f(a, b, loMax)
	}
}

// LoCompare returns a lo compare func.
func (typ Type) LoCompare() func(*Hand, *Hand) int {
	f, low := descs[typ].LoComp.Compare, descs[typ].Low
	loMax := Invalid
	if low {
		loMax = rankEightOrBetterMax
	}
	return func(a, b *Hand) int {
		return f(a, b, loMax)
	}
}

// Dealer creates a new dealer for the type.
func (typ Type) Dealer(shuffler Shuffler, n int) *Dealer {
	if desc, ok := descs[typ]; ok {
		return NewDealer(desc, shuffler, n)
	}
	return nil
}

// DealShuffle creates a new deck for the type and shuffles the deck count
// number of times, returning the pockets and board for the number of hands
// specified.
func (typ Type) DealShuffle(shuffler Shuffler, count, hands int) ([][]Card, []Card) {
	if d := typ.Dealer(shuffler, count); d != nil {
		return d.DealAll(hands)
	}
	return nil, nil
}

// Deal creates a new deck for the type, shuffling it once, returning the
// pockets and board for the number of hands specified.
//
// Use DealShuffle when needing to shuffle the deck more than once.
func (typ Type) Deal(shuffler Shuffler, hands int) ([][]Card, []Card) {
	return typ.DealShuffle(shuffler, 1, hands)
}

// RankHand creates a new hand for the pocket, board.
func (typ Type) RankHand(pocket, board []Card) *Hand {
	return NewHand(typ, pocket, board)
}

// RankHands creates a new hand for the provided pockets and board.
func (typ Type) RankHands(pockets [][]Card, board []Card) []*Hand {
	hands := make([]*Hand, len(pockets))
	for i := 0; i < len(pockets); i++ {
		hands[i] = typ.RankHand(pockets[i], board)
	}
	return hands
}

// String satisfies the fmt.Stringer interface.
func (typ Type) String() string {
	return string([]byte{byte(typ >> 8 & 0xff), byte(typ & 0xff)})
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
		if id, err := IdToType(name); err == nil {
			*typ = id
			return nil
		}
	}
	return ErrInvalidType
}

// descs are the registered type descriptions.
var descs map[Type]TypeDesc = make(map[Type]TypeDesc)

// evals are eval funcs.
var evals map[Type]EvalFunc = make(map[Type]EvalFunc)

// RegisterType registers a type.
func RegisterType(desc TypeDesc) error {
	if _, ok := descs[desc.Type]; ok {
		return ErrInvalidId
	}
	// check street ids
	m := make(map[byte]bool)
	for i, street := range desc.Streets {
		if m[street.Id] {
			return fmt.Errorf("%s street %d id %c must be unique", desc.Type, i, street.Id)
		}
	}
	desc.Num = len(descs)
	descs[desc.Type] = desc
	evals[desc.Type] = desc.Eval.New(desc.Low)
	return nil
}

// RegisterDefaultTypes registers default types.
func RegisterDefaultTypes() error {
	for _, desc := range DefaultTypes() {
		if err := RegisterType(desc); err != nil {
			return err
		}
	}
	return nil
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
		Type: typ,
		Name: name,
	}
	for _, o := range opts {
		o(desc)
	}
	return desc, nil
}

// Apply applies street options.
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

// WithOmaha is a type description option to set standard Omaha definitions.
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

// WithOmahaDouble is a type description option to set standard OmahaDouble
// definitions.
func WithOmahaDouble(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 9
		desc.Double = true
		desc.Blinds = HoldemBlinds()
		desc.Streets = HoldemStreets(4, 1, 3, 1, 1)
		desc.Eval = EvalOmaha
		desc.Apply(opts...)
	}
}

// WithOmahaFive is a type description option to set standard OmahaFive
// definitions.
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

// WithOmahaSix is a type description option to set standard OmahaSix
// definitions.
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

// WithCourchevel is a type description option to set standard Courchevel
// definitions.
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

// WithFusion is a type description option to set standard Fusion definitions.
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

// WithStud is a type description option to set standard Stud definitions.
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

// WithRazz is a type description option to set standard Razz definitions.
//
// Same as Stud, but with a Ace-to-Five low card ranking.
func WithRazz(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 7
		desc.Blinds = HoldemBlinds()
		desc.Streets = StudStreets()
		desc.Eval = EvalRazz
		desc.Apply(opts...)
	}
}

// WithBadugi is a type description option to set standard Badugi definitions.
//
// 4 cards, low evaluation of separate suits
// All 4 face down pre-flop
// 3 rounds of player discards (up to 4)
func WithBadugi(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Streets = NumberedStreets(4, 0, 0, 0)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalBadugi
		for i := 1; i < 4; i++ {
			desc.Streets[i].PocketDraw = 4
		}
		desc.Apply(opts...)
	}
}

// WithLowball is a type description option to set standard Lowball
// definitions.
func WithLowball(multi bool, opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Once = !multi
		desc.Streets = NumberedStreets(5, 0, 0, 0)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalLowball
		desc.HiComp = CompLowball
		for i := 1; i < 4; i++ {
			desc.Streets[1].PocketDraw = 5
		}
		desc.Apply(opts...)
	}
}

// WithSoko is a type description option to set standard Soko
// definitions.
func WithSoko(opts ...StreetOption) TypeOption {
	return func(desc *TypeDesc) {
		desc.Max = 8
		desc.Streets = NumberedStreets(2, 3)
		desc.Blinds = HoldemBlinds()
		desc.Eval = EvalSoko
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
	sd := func(id byte, name string, pocket int, board int) StreetDesc {
		return StreetDesc{
			Id:           id,
			Name:         name,
			Pocket:       pocket,
			Board:        board,
			BoardDiscard: discard,
		}
	}
	return []StreetDesc{
		sd('p', "Pre-Flop", pocket, 0),
		sd('f', "Flop", 0, flop),
		sd('t', "Turn", 0, turn),
		sd('r', "River", 0, river),
	}
}

// StudStreets creates Stud streets for the ante, third street, fourth street,
// fifth street, sixth street and river.
func StudStreets() []StreetDesc {
	streets := NumberedStreets(3, 1, 1, 1, 1)
	for i := 0; i < 4; i++ {
		streets[0].PocketUp = 1
	}
	return streets
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

// EvalFunc is a hand rank eval func.
type EvalFunc func(*Hand)

// EvalType is a hand rank eval type.
type EvalType uint8

// Eval types.
const (
	EvalHoldem EvalType = iota
	EvalShort
	EvalManila
	EvalOmaha
	EvalOmahaFive
	EvalOmahaSix
	EvalStud
	EvalRazz
	EvalBadugi
	EvalLowball
	EvalSoko
)

// New creates the eval type.
func (typ EvalType) New(low bool) EvalFunc {
	switch typ {
	case EvalHoldem:
		return NewHoldemEval(DefaultRank, Five)
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

// CompType is a compare type.
type CompType uint8

// Compare types.
const (
	CompHi CompType = iota
	CompLo
	CompShort
	CompManila
	CompLowball
	CompSoko
)

// Compare compares a, b.
func (typ CompType) Compare(a, b *Hand, loMax HandRank) int {
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

// NewHoldemEval creates a Holdem hand rank eval func.
func NewHoldemEval(f HandRankFunc, straightHigh Rank) EvalFunc {
	return func(h *Hand) {
		hand := h.Hand()
		h.HiRank = f(hand)
		bestHoldem(h, hand, straightHigh)
	}
}

// NewShortEval creates a Short hand rank eval func.
func NewShortEval() EvalFunc {
	return NewHoldemEval(NewRankFunc(func(c0, c1, c2, c3, c4 Card) HandRank {
		r := DefaultCactus(c0, c1, c2, c3, c4)
		switch r {
		case 747: // Straight Flush, 9, 8, 7, 6, Ace
			return 6
		case 6610: // Straight, 9, 8, 7, 6, Ace
			return 1605
		}
		return r
	}), Nine)
}

// NewManilaEval creates a Manila hand rank eval func.
func NewManilaEval() EvalFunc {
	return NewHoldemEval(NewRankFunc(func(c0, c1, c2, c3, c4 Card) HandRank {
		r := DefaultCactus(c0, c1, c2, c3, c4)
		switch r {
		case 747: // Straight Flush, 10, 9, 8, 7, Ace
			return 6
		case 6610: // Straight, 10, 9, 8, 7, Ace
			return 1605
		}
		return r
	}), Ten)
}

// NewOmahaEval creates a Omaha hand rank eval func.
func NewOmahaEval(loMax HandRank) EvalFunc {
	return func(h *Hand) {
		h.Init(5, 4, loMax)
		v, r := make([]Card, 5), HandRank(0)
		for i := 0; i < 6; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = h.Pocket[t4c2[i][0]], h.Pocket[t4c2[i][1]] // pocket
				v[2], v[3] = h.Board[t5c3[j][0]], h.Board[t5c3[j][1]]   // board
				v[4] = h.Board[t5c3[j][2]]                              // board
				if r = DefaultRank(v); r < h.HiRank {
					copy(h.HiBest, v)
					h.HiRank = r
					h.HiUnused[0], h.HiUnused[1] = h.Pocket[t4c2[i][2]], h.Pocket[t4c2[i][3]] // pocket
					h.HiUnused[2], h.HiUnused[3] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = HandRank(RankEightOrBetter(v[0], v[1], v[2], v[3], v[4])); r < h.LoRank && r < loMax {
						copy(h.LoBest, v)
						h.LoRank = r
						h.LoUnused[0], h.LoUnused[1] = h.Pocket[t4c2[i][2]], h.Pocket[t4c2[i][3]] // pocket
						h.LoUnused[2], h.LoUnused[3] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(h, loMax)
	}
}

// NewOmahaFiveEval creates a new Omaha5 hand rank eval func.
func NewOmahaFiveEval(loMax HandRank) EvalFunc {
	return func(h *Hand) {
		h.Init(5, 5, loMax)
		v, r := make([]Card, 5), HandRank(0)
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = h.Pocket[t5c2[i][0]], h.Pocket[t5c2[i][1]] // pocket
				v[2], v[3] = h.Board[t5c3[j][0]], h.Board[t5c3[j][1]]   // board
				v[4] = h.Board[t5c3[j][2]]                              // board
				if r = DefaultRank(v); r < h.HiRank {
					copy(h.HiBest, v)
					h.HiRank = r
					h.HiUnused[0], h.HiUnused[1] = h.Pocket[t5c2[i][2]], h.Pocket[t5c2[i][3]] // pocket
					h.HiUnused[2] = h.Pocket[t5c2[i][4]]                                      // pocket
					h.HiUnused[3], h.HiUnused[4] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = HandRank(RankEightOrBetter(v[0], v[1], v[2], v[3], v[4])); r < h.LoRank && r < loMax {
						copy(h.LoBest, v)
						h.LoRank = r
						h.LoUnused[0], h.LoUnused[1] = h.Pocket[t5c2[i][2]], h.Pocket[t5c2[i][3]] // pocket
						h.LoUnused[2] = h.Pocket[t5c2[i][4]]                                      // pocket
						h.LoUnused[3], h.LoUnused[4] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(h, loMax)
	}
}

// NewOmahaSixEval creates a new Omaha6 hand rank eval func.
func NewOmahaSixEval(loMax HandRank) EvalFunc {
	return func(h *Hand) {
		h.Init(5, 6, loMax)
		v, r := make([]Card, 5), HandRank(0)
		for i := 0; i < 15; i++ {
			for j := 0; j < 10; j++ {
				v[0], v[1] = h.Pocket[t6c2[i][0]], h.Pocket[t6c2[i][1]] // pocket
				v[2], v[3] = h.Board[t5c3[j][0]], h.Board[t5c3[j][1]]   // board
				v[4] = h.Board[t5c3[j][2]]                              // board
				if r = DefaultRank(v); r < h.HiRank {
					copy(h.HiBest, v)
					h.HiRank = r
					h.HiUnused[0], h.HiUnused[1] = h.Pocket[t6c2[i][2]], h.Pocket[t6c2[i][3]] // pocket
					h.HiUnused[2], h.HiUnused[3] = h.Pocket[t6c2[i][4]], h.Pocket[t6c2[i][5]] // pocket
					h.HiUnused[4], h.HiUnused[5] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
				}
				if loMax != Invalid {
					if r = HandRank(RankEightOrBetter(v[0], v[1], v[2], v[3], v[4])); r < h.LoRank && r < loMax {
						copy(h.LoBest, v)
						h.LoRank = r
						h.LoUnused[0], h.LoUnused[1] = h.Pocket[t6c2[i][2]], h.Pocket[t6c2[i][3]] // pocket
						h.LoUnused[2], h.LoUnused[3] = h.Pocket[t6c2[i][4]], h.Pocket[t6c2[i][5]] // pocket
						h.LoUnused[4], h.LoUnused[5] = h.Board[t5c3[j][3]], h.Board[t5c3[j][4]]   // board
					}
				}
			}
		}
		bestOmaha(h, loMax)
	}
}

// NewStudEval creates a Stud hand rank eval func.
func NewStudEval(loMax HandRank) EvalFunc {
	hi := NewHoldemEval(DefaultRank, Five)
	lo := NewLowEval(RankEightOrBetter, loMax)
	return func(h *Hand) {
		hi(h)
		if loMax != Invalid {
			v := NewUnevaluatedHand(StudHiLo, h.Pocket, h.Board)
			lo(v)
			if v.HiRank < loMax {
				h.LoRank, h.LoBest, h.LoUnused = v.HiRank, v.HiBest, v.HiUnused
			}
		}
	}
}

// NewRazzEval creates a Razz hand rank eval func.
func NewRazzEval() EvalFunc {
	f := NewLowEval(RankRazz, Invalid)
	return func(h *Hand) {
		f(h)
		if rankLowMax <= h.HiRank {
			switch r := Invalid - h.HiRank; r.Fixed() {
			case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
				h.HiBest, _ = bestSet(h.HiBest)
			default:
				panic("bad rank")
			}
		}
	}
}

// NewBadugiEval creates a Badugi hand rank eval func.
func NewBadugiEval() EvalFunc {
	return func(h *Hand) {
		s := make([][]Card, 4)
		for i := 0; i < len(h.Pocket) && i < 4; i++ {
			idx := h.Pocket[i].SuitIndex()
			s[idx] = append(s[idx], h.Pocket[i])
		}
		sort.SliceStable(s, func(i, j int) bool {
			a, b := len(s[i]), len(s[j])
			switch {
			case a != b:
				return a < b
			case a == 0:
				return true
			case b == 0:
				return false
			}
			return s[i][0].AceIndex() < s[j][0].AceIndex()
		})
		count, rank := 4, 0
		for i := 0; i < 4; i++ {
			sort.Slice(s[i], func(j, k int) bool {
				return s[i][j].AceIndex() < s[i][k].AceIndex()
			})
			captured, r := false, 0
			for j := 0; j < len(s[i]); j++ {
				if r = 1 << s[i][j].AceIndex(); rank&r == 0 && !captured {
					captured, h.HiBest = true, append(h.HiBest, s[i][j])
					rank |= r
					count--
				} else {
					h.HiUnused = append(h.HiUnused, s[i][j])
				}
			}
		}
		sort.Slice(h.HiBest, func(i, j int) bool {
			return h.HiBest[i].AceIndex() > h.HiBest[j].AceIndex()
		})
		sort.Slice(h.HiUnused, func(i, j int) bool {
			if a, b := h.HiUnused[i].AceIndex(), h.HiUnused[j].AceIndex(); a != b {
				return a > b
			}
			return h.HiUnused[i].Suit() < h.HiUnused[j].Suit()
		})
		h.HiRank = HandRank(count<<13 | rank)
	}
}

// NewLowballEval creates a Lowball hand rank eval func.
func NewLowballEval() EvalFunc {
	f := NewRankFunc(RankLowball)
	return func(h *Hand) {
		if len(h.Pocket) != 5 {
			panic("bad pocket")
		}
		h.HiRank = f(h.Pocket)
		h.Init(5, 0, Invalid)
		copy(h.HiBest, h.Pocket)
	}
}

// NewSokoEval creates a Soko hand rank eval func.
func NewSokoEval() EvalFunc {
	f := NewHoldemEval(DefaultRank, Five)
	return func(h *Hand) {
		f(h)
	}
}

// NewLowEval creates a low hand rank eval func, using f to determine the best
// low hand of a 7 card hand.
func NewLowEval(f RankFunc, loMax HandRank) EvalFunc {
	return func(h *Hand) {
		hand := h.Hand()
		if len(hand) != 7 {
			panic("bad hand")
		}
		best, unused := make([]Card, 5), make([]Card, 2)
		rank, r := Invalid, HandRank(0)
		for i := 0; i < 21; i++ {
			if r = HandRank(f(
				hand[t7c5[i][0]],
				hand[t7c5[i][1]],
				hand[t7c5[i][2]],
				hand[t7c5[i][3]],
				hand[t7c5[i][4]],
			)); r < rank && r < loMax {
				rank = r
				best[0], best[1] = hand[t7c5[i][0]], hand[t7c5[i][1]]
				best[2], best[3] = hand[t7c5[i][2]], hand[t7c5[i][3]]
				best[4] = hand[t7c5[i][4]]
				unused[0], unused[1] = hand[t7c5[i][5]], hand[t7c5[i][6]]
			}
		}
		if loMax <= rank {
			return
		}
		// order
		sort.Slice(best, func(i, j int) bool {
			return (best[i].Rank()+1)%13 > (best[j].Rank()+1)%13
		})
		h.HiRank, h.HiBest, h.HiUnused = rank, best, unused
	}
}

// HiComp is a hi eval compare func.
func HiComp(a, b *Hand, _ HandRank) int {
	switch {
	case a.HiRank < b.HiRank:
		return -1
	case a.HiRank > b.HiRank:
		return +1
	}
	return 0
}

// LoComp is a lo eval compare func.
func LoComp(a, b *Hand, loMax HandRank) int {
	switch low := loMax != Invalid; {
	case low && a.LoRank == Invalid && b.LoRank != Invalid:
		return +1
	case low && b.LoRank == Invalid && a.LoRank != Invalid:
		return -1
	case a.LoRank < b.LoRank:
		return -1
	case a.LoRank > b.LoRank:
		return +1
	}
	return 0
}

// ShortComp is the Short compare func.
func ShortComp(a, b *Hand, _ HandRank) int {
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
func ManilaComp(a, b *Hand, _ HandRank) int {
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
func LowballComp(a, b *Hand, _ HandRank) int {
	switch af, bf := rankMax-a.HiRank, rankMax-b.HiRank; {
	case af == bf:
	}
	// log.Printf(">>> a: %v // b: %v", a.HiRank, b.HiRank)
	return 0
}

// SokoComp is the Soko compare func.
func SokoComp(a, b *Hand, _ HandRank) int {
	return 0
}

// IdToType converts an id to a type.
func IdToType(id string) (Type, error) {
	if len(id) != 2 {
		return 0, ErrInvalidId
	}
	return Type(id[0])<<8 | Type(id[1]), nil
}

// bestHoldem sets the best holdem.
func bestHoldem(h *Hand, hand []Card, straightHigh Rank) {
	// order hand high to low
	sort.Slice(hand, func(i, j int) bool {
		m, n := hand[i].Rank(), hand[j].Rank()
		if m == n {
			return hand[i].Suit() > hand[j].Suit()
		}
		return m > n
	})
	// set best, unused
	switch h.HiRank.Fixed() {
	case StraightFlush:
		h.HiBest, h.HiUnused = bestStraightFlush(hand, straightHigh)
	case Flush:
		h.HiBest, h.HiUnused = bestFlush(hand)
	case Straight:
		h.HiBest, h.HiUnused = bestStraight(hand, straightHigh)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		h.HiBest, h.HiUnused = bestSet(hand)
	case Nothing:
		h.HiBest, h.HiUnused = hand[:5], hand[5:]
	default:
		panic("bad rank")
	}
}

// bestOmaha sets the best omaha on the eval.
func bestOmaha(h *Hand, loMax HandRank) {
	// order best
	sort.Slice(h.HiBest, func(i, j int) bool {
		m, n := h.HiBest[i].Rank(), h.HiBest[j].Rank()
		if m == n {
			return h.HiBest[i].Suit() > h.HiBest[j].Suit()
		}
		return m > n
	})
	switch h.HiRank.Fixed() {
	case StraightFlush:
		h.HiBest, _ = bestStraightFlush(h.HiBest, Five)
	case Flush:
		h.HiBest, _ = bestFlush(h.HiBest)
	case Straight:
		h.HiBest, _ = bestStraight(h.HiBest, Five)
	case FourOfAKind, FullHouse, ThreeOfAKind, TwoPair, Pair:
		h.HiBest, _ = bestSet(h.HiBest)
	case Nothing:
	default:
		panic("bad rank")
	}
	if loMax != Invalid && h.LoRank < loMax {
		sort.Slice(h.LoBest, func(i, j int) bool {
			return (h.LoBest[i].Rank()+1)%13 > (h.LoBest[j].Rank()+1)%13
		})
	} else {
		h.LoBest, h.LoUnused = nil, nil
	}
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

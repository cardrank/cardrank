package cardrank

import (
	"math/bits"
	"slices"
)

// profile is an evaluator profile: how an evaluator ranks hands.
type profile struct {
	id       string
	name     string
	deck     DeckType
	levels   []level
	classify func(EvalRank) (string, string, bool)
}

// level is a profile's hand level.
type level struct {
	id       string
	name     string
	example  Cards
	variants []variant
}

// variant is a variant of a hand level.
type variant struct {
	id      string
	name    string
	example Cards
}

// info returns the profile's catalog info.
func (p *profile) info() ProfileInfo {
	info := ProfileInfo{
		ID:   p.id,
		Name: label("profile", p.id, p.name),
		Deck: deckID(p.deck),
	}
	for _, l := range p.levels {
		li := LevelInfo{
			ID:      l.id,
			Name:    label("level", l.id, l.name),
			Example: l.example,
		}
		for _, v := range l.variants {
			li.Variants = append(li.Variants, VariantInfo{
				ID:      v.id,
				Name:    label("variant", v.id, v.name),
				Example: v.example,
			})
		}
		info.Levels = append(info.Levels, li)
	}
	return info
}

// profileOrder are the profiles, in catalog order.
var profileOrder = []*profile{
	highProfile("high", "High", DeckFrench, false),
	highProfile("high_royal", "High (Royal Deck)", DeckRoyal, false),
	highProfile("high_short", "High (Short Deck)", DeckShort, true),
	highProfile("high_manila", "High (Manila Deck)", DeckManila, true),
	highProfile("high_spanish", "High (Spanish Deck)", DeckSpanish, true),
	sokoProfile(),
	lowballProfile(),
	razzProfile(),
	badugiProfile(),
	lowProfile(),
}

// profiles are the profiles, by ID.
var profiles = func() map[string]*profile {
	m := make(map[string]*profile, len(profileOrder))
	for _, p := range profileOrder {
		m[p.id] = p
	}
	return m
}()

// straightFlushIDs are the straight flush variant IDs, by top card.
var straightFlushIDs = map[Rank]string{
	Ace:   "royal_flush",
	King:  "platinum_crown",
	Queen: "gold_scepter",
	Jack:  "silver_sword",
	Ten:   "copper_ring",
	Nine:  "iron_orb",
	Eight: "tin_cup",
	Seven: "brass_axe",
	Six:   "lead_pipe",
	Five:  "steel_wheel",
}

// lowIDs are the no pair variant IDs, by top card.
var lowIDs = map[Rank]string{
	Ace:   "ace_low",
	King:  "king_low",
	Queen: "queen_low",
	Jack:  "jack_low",
	Ten:   "ten_low",
	Nine:  "nine_low",
	Eight: "eight_low",
	Seven: "seven_low",
	Six:   "six_low",
	Five:  "five_low",
}

// cactusLevels are the Cactus hand levels, best first.
var cactusLevels = []struct {
	id   string
	rank EvalRank
}{
	{"straight_flush", StraightFlush},
	{"four_of_a_kind", FourOfAKind},
	{"full_house", FullHouse},
	{"flush", Flush},
	{"straight", Straight},
	{"three_of_a_kind", ThreeOfAKind},
	{"two_pair", TwoPair},
	{"pair", Pair},
	{"high_card", Nothing},
}

// cactusClass returns the level and variant of a Cactus rank.
func cactusClass(r EvalRank) (string, string, bool) {
	if r == 0 || Nothing < r {
		return "", "", false
	}
	f := r.Fixed()
	for _, l := range cactusLevels {
		if l.rank == f {
			if f == StraightFlush {
				return l.id, straightFlushIDs[Ace-Rank(r-1)], true
			}
			return l.id, "", true
		}
	}
	return "", "", false
}

// highProfile creates a high profile for the deck, ranking a [Flush] over a
// [FullHouse] when flushOver is true. Levels and variants that the deck can't
// make are omitted.
func highProfile(id, name string, deck DeckType, flushOver bool) *profile {
	examples := highExamples[deck]
	p := &profile{
		id:   id,
		name: name,
		deck: deck,
		classify: func(r EvalRank) (string, string, bool) {
			if flushOver {
				r = r.FromFlushOver()
			}
			return cactusClass(r)
		},
	}
	order := cactusLevels
	if flushOver {
		// a flush ranks over a full house
		order = append(append(append([]struct {
			id   string
			rank EvalRank
		}{}, cactusLevels[:2]...), cactusLevels[3], cactusLevels[2]), cactusLevels[4:]...)
	}
	for _, l := range order {
		var lv level
		if l.rank == StraightFlush {
			lv = straightFlushLevel(deck, true)
		} else {
			example, ok := examples[l.id]
			if !ok {
				continue
			}
			lv = level{id: l.id, name: l.rank.Title(), example: Must(example)}
		}
		p.levels = append(p.levels, lv)
	}
	return p
}

// highExamples are the examples of each deck's high levels (other than
// straight flushes). Levels that a deck can't make have no example.
var highExamples = map[DeckType]map[string]string{
	DeckFrench: {
		"four_of_a_kind":  "9s 9h 9d 9c Kd",
		"full_house":      "Ks Kh Kd 7c 7s",
		"flush":           "Ah Jh 8h 5h 2h",
		"straight":        "Ts 9h 8d 7c 6s",
		"three_of_a_kind": "Qs Qh Qd 9c 4s",
		"two_pair":        "Js Jh 6d 6c As",
		"pair":            "Ts Th Ad 8c 3s",
		"high_card":       "As Qh 9d 6c 3s",
	},
	// a royal deck can't make a flush (only a royal flush) or a high card
	// (only a straight)
	DeckRoyal: {
		"four_of_a_kind":  "As Ah Ad Ac Ks",
		"full_house":      "Ks Kh Kd Qs Qh",
		"straight":        "As Kh Qd Jc Ts",
		"three_of_a_kind": "Qs Qh Qd As Kh",
		"two_pair":        "Js Jh Ts Th As",
		"pair":            "As Ah Ks Qh Jd",
	},
	DeckShort: {
		"four_of_a_kind":  "9s 9h 9d 9c Ad",
		"full_house":      "Ks Kh Kd 7c 7s",
		"flush":           "Ah Jh 9h 7h 6h",
		"straight":        "Ts 9h 8d 7c 6s",
		"three_of_a_kind": "Qs Qh Qd 9c 7s",
		"two_pair":        "Js Jh 7d 7c As",
		"pair":            "Ts Th Ad 8c 6s",
		"high_card":       "As Qh 9d 7c 6s",
	},
	DeckManila: {
		"four_of_a_kind":  "9s 9h 9d 9c Ad",
		"full_house":      "Ks Kh Kd 7c 7s",
		"flush":           "Ah Jh 9h 8h 7h",
		"straight":        "Js Th 9d 8c 7s",
		"three_of_a_kind": "Qs Qh Qd 9c 7s",
		"two_pair":        "Js Jh 8d 8c As",
		"pair":            "Ts Th Ad 8c 7s",
		"high_card":       "As Qh 9d 8c 7s",
	},
	DeckSpanish: {
		"four_of_a_kind":  "9s 9h 9d 9c Ad",
		"full_house":      "Ks Kh Kd 8c 8s",
		"flush":           "Ah Qh Th 9h 8h",
		"straight":        "Qs Jh Td 9c 8s",
		"three_of_a_kind": "Qs Qh Qd 9c 8s",
		"two_pair":        "Js Jh 8d 8c As",
		"pair":            "Ts Th Ad 9c 8s",
		"high_card":       "As Kh Td 9c 8s",
	},
}

// straightFlushLevel returns the straight flush level for the deck, with a
// variant for each straight flush the deck can make, including the deck's
// wheel (such as 5-4-3-2-A), when wheel is true.
func straightFlushLevel(deck DeckType, wheel bool) level {
	l := level{id: "straight_flush", name: StraightFlush.Title()}
	base := Rank(deck)
	for top := Ace; base+3 <= top && top != InvalidRank; top-- {
		var v Cards
		switch {
		case base+4 <= top:
			for r := top; top-4 <= r && r != InvalidRank; r-- {
				v = append(v, New(r, Heart))
			}
		case wheel && top+1 < Ace:
			// the deck's wheel, with the ace low
			for r := top; base <= r && r != InvalidRank; r-- {
				v = append(v, New(r, Heart))
			}
			v = append(v, New(Ace, Heart))
		default:
			continue
		}
		l.variants = append(l.variants, variant{
			id:      straightFlushIDs[top],
			name:    top.StraightFlushName(),
			example: v,
		})
	}
	// use the second best (or only) straight flush as the level's example
	l.example = l.variants[min(1, len(l.variants)-1)].example
	return l
}

// noPairLevel returns a no pair level, with a variant for each top card, where
// examples are made from the top card and the lowest other cards.
func noPairLevel(tops []Rank, aceLow bool) level {
	l := level{id: "no_pair", name: "No Pair"}
	for _, top := range tops {
		low := []Rank{Five, Four, Three, Two}
		if aceLow {
			low = []Rank{Four, Three, Two, Ace}
		}
		suits := []Suit{Spade, Heart, Diamond, Club, Spade}
		v := Cards{New(top, suits[0])}
		for _, r := range low {
			if r != top && len(v) < 5 {
				v = append(v, New(r, suits[len(v)]))
			}
		}
		if len(v) < 5 {
			// the top card is one of the lowest cards (such as a wheel)
			for _, r := range []Rank{Five, Four, Three, Two, Ace} {
				if !containsRank(v, r) && len(v) < 5 {
					v = append(v, New(r, suits[len(v)]))
				}
			}
		}
		l.variants = append(l.variants, variant{
			id:      lowIDs[top],
			name:    top.Name() + "-low",
			example: v,
		})
	}
	l.example = l.variants[min(1, len(l.variants)-1)].example
	return l
}

// containsRank returns true when v contains a card of the rank.
func containsRank(v Cards, rank Rank) bool {
	for _, c := range v {
		if c.Rank() == rank {
			return true
		}
	}
	return false
}

// sokoProfile creates the [Soko] profile, which adds four card flushes and
// straights between two pair and a pair.
func sokoProfile() *profile {
	p := highProfile("soko", "Soko", DeckFrench, false)
	p.classify = func(r EvalRank) (string, string, bool) {
		switch {
		case r == 0 || SokoNothing < r:
			return "", "", false
		case r <= TwoPair:
			return cactusClass(r)
		case r <= SokoFlush:
			return "four_flush", "", true
		case r <= SokoStraight:
			return "four_straight", "", true
		}
		return cactusClass(r - SokoStraight + TwoPair)
	}
	var levels []level
	for _, l := range p.levels {
		if l.id == "pair" {
			levels = append(levels,
				level{id: "four_flush", name: "Four Flush", example: Must("Ah Jh 8h 3h Kc")},
				level{id: "four_straight", name: "Four Straight", example: Must("9s 8h 7d 6c Ks")},
			)
		}
		levels = append(levels, l)
	}
	p.levels = levels
	return p
}

// lowballProfile creates the 2-7 [Lowball] profile, where the best hand is
// 7-5-4-3-2, aces are high, and straights and flushes count against.
func lowballProfile() *profile {
	// in 2-7 lowball, a royal flush is the worst hand
	sf := straightFlushLevel(DeckFrench, false)
	slices.Reverse(sf.variants)
	sf.example = Must("9h 8h 7h 6h 5h")
	return &profile{
		id:   "lowball_27",
		name: "2-7 Lowball",
		deck: DeckFrench,
		levels: []level{
			noPairLevel([]Rank{Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}, false),
			{id: "pair", name: Pair.Title(), example: Must("8s 8h 5d 3c 2s")},
			{id: "two_pair", name: TwoPair.Title(), example: Must("8s 8h 5d 5c 2s")},
			{id: "three_of_a_kind", name: ThreeOfAKind.Title(), example: Must("8s 8h 8d 5c 2s")},
			{id: "straight", name: Straight.Title(), example: Must("6s 5h 4d 3c 2s")},
			{id: "flush", name: Flush.Title(), example: Must("9h 7h 5h 3h 2h")},
			{id: "full_house", name: FullHouse.Title(), example: Must("8s 8h 8d 5c 5s")},
			{id: "four_of_a_kind", name: FourOfAKind.Title(), example: Must("8s 8h 8d 8c 2s")},
			sf,
		},
		classify: func(rank EvalRank) (string, string, bool) {
			if rank == 0 || Nothing < rank {
				return "", "", false
			}
			switch r := rank.FromLowball(); {
			case r == StraightFlush:
				// 5-4-3-2-A suited is an ace high flush
				return "flush", "", true
			case r == Straight:
				// 5-4-3-2-A is an ace high no pair
				return "no_pair", lowIDs[Ace], true
			case Pair < r:
				return "no_pair", lowIDs[highTop(r)], true
			default:
				return cactusClass(r)
			}
		},
	}
}

// highTop returns the top card of a no pair (high card) Cactus rank.
func highTop(r EvalRank) Rank {
	switch {
	case r <= AceHigh:
		return Ace
	case r <= KingHigh:
		return King
	case r <= QueenHigh:
		return Queen
	case r <= JackHigh:
		return Jack
	case r <= TenHigh:
		return Ten
	case r <= NineHigh:
		return Nine
	case r <= EightHigh:
		return Eight
	}
	return Seven
}

// razzProfile creates the A-5 low [Razz] profile, where the best hand is
// 5-4-3-2-A, aces are low, and straights and flushes don't count.
func razzProfile() *profile {
	return &profile{
		id:   "razz",
		name: "A-5 Low (Razz)",
		deck: DeckFrench,
		levels: []level{
			noPairLevel([]Rank{Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}, true),
			{id: "pair", name: Pair.Title(), example: Must("Ks Kh 3d 2c As")},
			{id: "two_pair", name: TwoPair.Title(), example: Must("Ks Kh 3d 3c As")},
			{id: "three_of_a_kind", name: ThreeOfAKind.Title(), example: Must("7s 7h 7d 2c As")},
			{id: "full_house", name: FullHouse.Title(), example: Must("7s 7h 7d 2c 2s")},
			{id: "four_of_a_kind", name: FourOfAKind.Title(), example: Must("7s 7h 7d 7c 2s")},
		},
		classify: func(rank EvalRank) (string, string, bool) {
			switch {
			case rank == 0 || rank == Invalid:
				return "", "", false
			case rank < AceFiveLow:
				return "no_pair", lowIDs[aceLowTop(rank)], true
			}
			return cactusClass(Invalid - rank)
		},
	}
}

// lowProfile creates the 8-or-better low profile, for the Lo of Hi/Lo types.
func lowProfile() *profile {
	return &profile{
		id:     "low_8",
		name:   "8-or-Better Low",
		deck:   DeckFrench,
		levels: []level{noPairLevel([]Rank{Five, Six, Seven, Eight}, true)},
		classify: func(rank EvalRank) (string, string, bool) {
			if rank == 0 || EightOrBetter <= rank {
				return "", "", false
			}
			return "no_pair", lowIDs[aceLowTop(rank)], true
		},
	}
}

// aceLowTop returns the top card of an A-5 low rank (see [RankAceFiveLow]).
func aceLowTop(rank EvalRank) Rank {
	n := bits.Len16(uint16(rank&0x1fff)) - 1
	if n <= 0 {
		return Ace
	}
	return Rank(n - 1)
}

// badugiProfile creates the [Badugi] profile, where a hand's level is the
// number of cards having different suits and ranks.
func badugiProfile() *profile {
	return &profile{
		id:   "badugi",
		name: "Badugi",
		deck: DeckFrench,
		levels: []level{
			{id: "badugi", name: "Badugi", example: Must("Ks 9h 6d 2c")},
			{id: "three_card", name: "Three-Card Hand", example: Must("As 2s 3h 4d")},
			{id: "two_card", name: "Two-Card Hand", example: Must("As 2s 3h 4h")},
			{id: "one_card", name: "One-Card Hand", example: Must("As 2s 3s 4s")},
		},
		classify: func(rank EvalRank) (string, string, bool) {
			if rank&0x1fff == 0 || rank == Invalid {
				return "", "", false
			}
			switch rank >> 13 {
			case 0:
				return "badugi", "", true
			case 1:
				return "three_card", "", true
			case 2:
				return "two_card", "", true
			case 3:
				return "one_card", "", true
			}
			return "", "", false
		},
	}
}

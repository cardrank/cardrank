package cardrank

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
)

// Label is a name to show: a stable key (for translation), and the English
// text.
type Label struct {
	// Key is a stable key, such as "cardrank.level.straight_flush".
	Key string `json:"key"`
	// Text is the English text, such as "Straight Flush".
	Text string `json:"text"`
}

// Catalog is a catalog of the registered types, their decks, and the hands
// they evaluate, for clients showing hands (see [HandCatalog]).
//
// IDs are stable: a renamed hand changes its label's text, never an ID.
type Catalog struct {
	// Revision is "sha256:" and the hex SHA-256 of the catalog's JSON
	// encoding (without the revision).
	Revision string `json:"revision"`
	// Decks are the decks of the games.
	Decks []DeckInfo `json:"decks"`
	// Games are the registered types.
	Games []GameInfo `json:"games"`
	// Profiles are the evaluator profiles, shared between games where the
	// hands are ranked the same.
	Profiles []ProfileInfo `json:"profiles"`
	// Statuses are the statuses of a hand that has no level.
	Statuses []StatusInfo `json:"statuses"`
}

// DeckInfo describes a deck.
type DeckInfo struct {
	// ID is the deck's ID, such as "french" or "short".
	ID string `json:"id"`
	// Name is the deck's name.
	Name Label `json:"name"`
	// Cards are the deck's unshuffled cards.
	Cards Cards `json:"cards"`
}

// GameInfo describes a registered type.
type GameInfo struct {
	// ID is the game's ID, such as "holdem" or "omaha_hi_lo".
	ID string `json:"id"`
	// Type is the type.
	Type Type `json:"type"`
	// Name is the game's name.
	Name Label `json:"name"`
	// Deck is the game's deck ([DeckInfo.ID]).
	Deck string `json:"deck"`
	// Max is the maximum number of players.
	Max int `json:"max"`
	// Pocket is the total number of pocket cards dealt.
	Pocket int `json:"pocket"`
	// Board is the total number of board cards dealt (on each board).
	Board int `json:"board"`
	// Draw is true when pocket cards can be drawn (discarded and replaced).
	Draw bool `json:"draw"`
	// Once is true when cards can only be drawn once.
	Once bool `json:"once"`
	// Blinds are the blinds (forced bets).
	Blinds []Label `json:"blinds"`
	// Streets are the game's streets.
	Streets []StreetInfo `json:"streets"`
	// Tracks are the hands evaluated for each player: the Hi, and either the
	// Lo or the second board's Hi.
	Tracks []TrackInfo `json:"tracks"`
}

// StreetInfo describes a street.
type StreetInfo struct {
	// ID is the street's ID within the game, such as "f" for the flop.
	ID string `json:"id"`
	// Name is the street's name.
	Name Label `json:"name"`
	// Pocket is the number of pocket cards dealt.
	Pocket int `json:"pocket"`
	// PocketUp is the number of pocket cards dealt face up.
	PocketUp int `json:"pocket_up"`
	// PocketDiscard is the number of cards discarded before the pocket cards.
	PocketDiscard int `json:"pocket_discard"`
	// PocketDraw is the number of pocket cards that can be drawn.
	PocketDraw int `json:"pocket_draw"`
	// Board is the number of board cards dealt.
	Board int `json:"board"`
	// BoardDiscard is the number of cards discarded before the board cards.
	BoardDiscard int `json:"board_discard"`
}

// TrackInfo describes a hand evaluated for each player.
type TrackInfo struct {
	// ID is "hi" or "lo".
	ID string `json:"id"`
	// Board is 0, or 1 for the second board of a double board game.
	Board int `json:"board"`
	// Profile is the evaluator profile ([ProfileInfo.ID]).
	Profile string `json:"profile"`
	// Qualifier is the qualifier for a hand to count, such as
	// "eight_or_better", or "" for none.
	Qualifier string `json:"qualifier"`
}

// ProfileInfo describes how an evaluator ranks hands.
type ProfileInfo struct {
	// ID is the profile's ID, such as "high" or "razz".
	ID string `json:"id"`
	// Name is the profile's name.
	Name Label `json:"name"`
	// Deck is the profile's deck ([DeckInfo.ID]).
	Deck string `json:"deck"`
	// Levels are the hand levels, best first.
	Levels []LevelInfo `json:"levels"`
}

// LevelInfo describes a hand level, such as a straight flush.
type LevelInfo struct {
	// ID is the level's ID within its profile, such as "straight_flush".
	ID string `json:"id"`
	// Name is the level's name.
	Name Label `json:"name"`
	// Example is a hand making the level.
	Example Cards `json:"example"`
	// Variants are the level's variants, best first.
	Variants []VariantInfo `json:"variants"`
}

// VariantInfo describes a variant of a hand level, such as a royal flush.
type VariantInfo struct {
	// ID is the variant's ID within its level, such as "royal_flush".
	ID string `json:"id"`
	// Name is the variant's name.
	Name Label `json:"name"`
	// Example is a hand making the variant.
	Example Cards `json:"example"`
}

// StatusInfo describes the status of a hand that has no level.
type StatusInfo struct {
	// ID is the status's ID, such as "no_qualifying_low".
	ID string `json:"id"`
	// Name is the status's name.
	Name Label `json:"name"`
}

// HandClass is the classification of a hand in the [Catalog]: its level and
// variant, or its status when it has no level.
type HandClass struct {
	// Track is the hand's track ([TrackInfo.ID]).
	Track string `json:"track"`
	// Board is the hand's board ([TrackInfo.Board]).
	Board int `json:"board"`
	// Profile is the track's profile ([ProfileInfo.ID]).
	Profile string `json:"profile"`
	// Level is the hand's level ([LevelInfo.ID]), or "" when the hand has a
	// status.
	Level string `json:"level"`
	// Variant is the hand's variant ([VariantInfo.ID]), or "" for none.
	Variant string `json:"variant"`
	// Status is the hand's status ([StatusInfo.ID]), or "" when the hand has
	// a level.
	Status string `json:"status"`
}

// Statuses.
const (
	// StatusNoQualifyingLow is the status of a Lo with no qualifying hand.
	StatusNoQualifyingLow = "no_qualifying_low"
	// StatusNoQualifyingHand is the status of a Hi with no qualifying hand
	// (such as below a pair of jacks for [Video]).
	StatusNoQualifyingHand = "no_qualifying_hand"
	// StatusNoHand is the status of a hand that couldn't be evaluated (such as
	// too few cards).
	StatusNoHand = "no_hand"
)

// HandCatalog returns the catalog of the registered types, built on first use.
// The catalog is shared, and must not be modified.
var HandCatalog = sync.OnceValue(func() *Catalog {
	c := buildCatalog()
	c.Revision = catalogRevision(c)
	return c
})

// Classes classifies the eval's hands in the [Catalog], one for each of its
// type's tracks: the Hi, and either the Lo or the second board's Hi.
func (ev *Eval) Classes() []HandClass {
	tracks := gameTracks(descs[ev.Type])
	v := make([]HandClass, len(tracks))
	for i, track := range tracks {
		h := ev.Hi
		if i != 0 {
			h = ev.Lo
		}
		v[i] = classify(track, h.Rank)
	}
	return v
}

// HiClass classifies the eval's Hi hand (see [Eval.Classes]).
func (ev *Eval) HiClass() HandClass {
	return ev.Classes()[0]
}

// LoClass classifies the eval's Lo hand (see [Eval.Classes]), or the second
// board's Hi for a double board type. Returns a zero class when the type has
// neither.
func (ev *Eval) LoClass() HandClass {
	if v := ev.Classes(); 1 < len(v) {
		return v[1]
	}
	return HandClass{}
}

// classify classifies the rank on the track.
func classify(track TrackInfo, rank EvalRank) HandClass {
	class := HandClass{
		Track:   track.ID,
		Board:   track.Board,
		Profile: track.Profile,
	}
	p, ok := profiles[track.Profile]
	if !ok {
		class.Status = StatusNoHand
		return class
	}
	if level, variant, ok := p.classify(rank); ok {
		class.Level, class.Variant = level, variant
		return class
	}
	switch track.Qualifier {
	case qualifierEightOrBetter:
		class.Status = StatusNoQualifyingLow
	case qualifierJacksOrBetter:
		class.Status = StatusNoQualifyingHand
	default:
		class.Status = StatusNoHand
	}
	return class
}

// Qualifiers.
const (
	qualifierEightOrBetter = "eight_or_better"
	qualifierJacksOrBetter = "jacks_or_better"
)

// buildCatalog builds the catalog of the registered types.
func buildCatalog() *Catalog {
	c := &Catalog{
		Statuses: []StatusInfo{
			{StatusNoQualifyingLow, label("status", StatusNoQualifyingLow, "No Qualifying Low")},
			{StatusNoQualifyingHand, label("status", StatusNoQualifyingHand, "No Qualifying Hand")},
			{StatusNoHand, label("status", StatusNoHand, "No Hand")},
		},
	}
	decks, used := make(map[DeckType]bool), make(map[string]bool)
	for _, typ := range Types() {
		desc := descs[typ]
		decks[desc.Deck] = true
		g := GameInfo{
			ID:     gameID(typ),
			Type:   typ,
			Name:   label("game", gameID(typ), desc.Name),
			Deck:   deckID(desc.Deck),
			Max:    desc.Max,
			Pocket: typ.Pocket(),
			Board:  typ.Board(),
			Draw:   typ.Draw(),
			Once:   desc.Once,
			Tracks: gameTracks(desc),
		}
		for _, name := range desc.Blinds {
			g.Blinds = append(g.Blinds, label("blind", blindIDs[name], name))
		}
		for _, s := range desc.Streets {
			g.Streets = append(g.Streets, StreetInfo{
				ID:            string(rune(s.ID)),
				Name:          label("street", streetIDs[s.Name], s.Name),
				Pocket:        s.Pocket,
				PocketUp:      s.PocketUp,
				PocketDiscard: s.PocketDiscard,
				PocketDraw:    s.PocketDraw,
				Board:         s.Board,
				BoardDiscard:  s.BoardDiscard,
			})
		}
		for _, track := range g.Tracks {
			used[track.Profile] = true
		}
		c.Games = append(c.Games, g)
	}
	for _, typ := range []DeckType{DeckFrench, DeckShort, DeckManila, DeckSpanish, DeckRoyal, DeckKuhn, DeckLeduc, DeckFrenchJoker} {
		if decks[typ] {
			c.Decks = append(c.Decks, DeckInfo{
				ID:    deckID(typ),
				Name:  label("deck", deckID(typ), typ.Name()),
				Cards: Cards(typ.Unshuffled()),
			})
		}
	}
	for _, p := range profileOrder {
		if used[p.id] {
			c.Profiles = append(c.Profiles, p.info())
		}
	}
	return c
}

// catalogRevision returns the revision of the catalog: the SHA-256 of its JSON
// encoding, without its revision.
func catalogRevision(c *Catalog) string {
	v := *c
	v.Revision = ""
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(buf)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// label returns a label with a "cardrank.<kind>.<id>" key, or an empty key
// when id is empty.
func label(kind, id, text string) Label {
	if id == "" {
		return Label{Text: text}
	}
	return Label{Key: "cardrank." + kind + "." + id, Text: text}
}

// gameTracks returns the type's tracks.
func gameTracks(desc TypeDesc) []TrackInfo {
	hi := TrackInfo{ID: "hi", Profile: hiProfile(desc)}
	if desc.Eval == EvalJacksOrBetter {
		hi.Qualifier = qualifierJacksOrBetter
	}
	switch {
	case desc.Double:
		second := hi
		second.Board = 1
		return []TrackInfo{hi, second}
	case desc.Low:
		return []TrackInfo{hi, {ID: "lo", Profile: "low_8", Qualifier: qualifierEightOrBetter}}
	}
	return []TrackInfo{hi}
}

// hiProfile returns the profile of the type's Hi, or "" for a type with its
// own evaluator.
func hiProfile(desc TypeDesc) string {
	if desc.Evaluator != nil {
		return ""
	}
	switch desc.Eval {
	case EvalCactus, EvalOmaha, EvalJacksOrBetter:
		switch desc.Deck {
		case DeckFrench:
			return "high"
		case DeckRoyal:
			return "high_royal"
		}
	case EvalShort:
		return "high_short"
	case EvalManila:
		return "high_manila"
	case EvalSpanish:
		return "high_spanish"
	case EvalSoko:
		return "soko"
	case EvalLowball:
		return "lowball_27"
	case EvalRazz:
		return "razz"
	case EvalBadugi:
		return "badugi"
	}
	return ""
}

// gameID returns the type's game ID.
func gameID(typ Type) string {
	if id, ok := gameIDs[typ]; ok {
		return id
	}
	return "type_" + typ.ID()
}

// gameIDs are the game IDs of the default types.
var gameIDs = map[Type]string{
	Holdem:         "holdem",
	Split:          "split",
	Short:          "short",
	Manila:         "manila",
	Spanish:        "spanish",
	Royal:          "royal",
	Double:         "double",
	Showtime:       "showtime",
	Swap:           "swap",
	River:          "river",
	Dallas:         "dallas",
	DallasDouble:   "dallas_double",
	Houston:        "houston",
	Draw:           "draw",
	DrawHiLo:       "draw_hi_lo",
	Stud:           "stud",
	StudHiLo:       "stud_hi_lo",
	StudFive:       "stud_five",
	Video:          "video",
	Omaha:          "omaha",
	OmahaHiLo:      "omaha_hi_lo",
	OmahaDouble:    "omaha_double",
	OmahaFive:      "omaha_five",
	OmahaSix:       "omaha_six",
	Jakarta:        "jakarta",
	Courchevel:     "courchevel",
	CourchevelHiLo: "courchevel_hi_lo",
	Fusion:         "fusion",
	FusionHiLo:     "fusion_hi_lo",
	Soko:           "soko",
	SokoHiLo:       "soko_hi_lo",
	Lowball:        "lowball",
	LowballTriple:  "lowball_triple",
	Razz:           "razz",
	Badugi:         "badugi",
}

// deckID returns the deck's ID.
func deckID(typ DeckType) string {
	switch typ {
	case DeckFrench:
		return "french"
	case DeckShort:
		return "short"
	case DeckManila:
		return "manila"
	case DeckSpanish:
		return "spanish"
	case DeckRoyal:
		return "royal"
	case DeckKuhn:
		return "kuhn"
	case DeckLeduc:
		return "leduc"
	case DeckFrenchJoker:
		return "french_joker"
	}
	return ""
}

// streetIDs are the IDs of the street names.
var streetIDs = map[string]string{
	"Ante":     "ante",
	"Pre-Flop": "pre_flop",
	"Flop":     "flop",
	"Turn":     "turn",
	"River":    "river",
	"Drop":     "drop",
	"3rd":      "third",
	"4th":      "fourth",
	"5th":      "fifth",
	"6th":      "sixth",
	"7th":      "seventh",
}

// blindIDs are the IDs of the blind names.
var blindIDs = map[string]string{
	"Small Blind": "small_blind",
	"Big Blind":   "big_blind",
	"Straddle":    "straddle",
	"Ante":        "ante",
	"Bring In":    "bring_in",
}

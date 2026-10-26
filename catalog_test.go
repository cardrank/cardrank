package cardrank

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func TestHandCatalog(t *testing.T) {
	c := HandCatalog()
	t.Logf("revision: %s", c.Revision)
	// games are the registered types
	types := Types()
	if n, exp := len(c.Games), len(types); n != exp {
		t.Fatalf("expected %d games, got: %d", exp, n)
	}
	for i, g := range c.Games {
		if g.Type != types[i] {
			t.Errorf("game %d expected type %s, got: %s", i, types[i], g.Type)
		}
		if _, ok := gameIDs[g.Type]; !ok {
			t.Errorf("game %s has no ID", g.Type)
		}
	}
	// decks
	var decks []string
	for _, d := range c.Decks {
		decks = append(decks, d.ID)
	}
	if exp := []string{"french", "short", "manila", "spanish", "royal"}; !slices.Equal(decks, exp) {
		t.Errorf("expected decks %v, got: %v", exp, decks)
	}
	unique(t, "deck", decks)
	unique(t, "game", ids(c.Games, func(g GameInfo) string { return g.ID }))
	unique(t, "profile", ids(c.Profiles, func(p ProfileInfo) string { return p.ID }))
	unique(t, "status", ids(c.Statuses, func(s StatusInfo) string { return s.ID }))
	// references resolve
	profiles := make(map[string]ProfileInfo)
	for _, p := range c.Profiles {
		profiles[p.ID] = p
		if !slices.Contains(decks, p.Deck) {
			t.Errorf("profile %s has unknown deck %q", p.ID, p.Deck)
		}
		checkLabel(t, p.Name)
		unique(t, "level", ids(p.Levels, func(l LevelInfo) string { return l.ID }))
		for _, l := range p.Levels {
			checkLabel(t, l.Name)
			unique(t, "variant", ids(l.Variants, func(v VariantInfo) string { return v.ID }))
			for _, v := range l.Variants {
				checkLabel(t, v.Name)
			}
		}
	}
	used := make(map[string]bool)
	for _, g := range c.Games {
		checkLabel(t, g.Name)
		if !slices.Contains(decks, g.Deck) {
			t.Errorf("game %s has unknown deck %q", g.ID, g.Deck)
		}
		for _, b := range g.Blinds {
			checkLabel(t, b)
		}
		for _, s := range g.Streets {
			checkLabel(t, s.Name)
		}
		for _, track := range g.Tracks {
			p, ok := profiles[track.Profile]
			switch {
			case !ok:
				t.Errorf("game %s track %s has unknown profile %q", g.ID, track.ID, track.Profile)
			case p.Deck != g.Deck:
				t.Errorf("game %s deck %s doesn't match profile %s deck %s", g.ID, g.Deck, p.ID, p.Deck)
			}
			used[track.Profile] = true
		}
	}
	for _, p := range c.Profiles {
		if !used[p.ID] {
			t.Errorf("profile %s is not used by a game", p.ID)
		}
	}
	for _, s := range c.Statuses {
		checkLabel(t, s.Name)
	}
}

func unique(t *testing.T, kind string, v []string) {
	t.Helper()
	m := make(map[string]bool)
	for _, id := range v {
		if id == "" || m[id] {
			t.Errorf("%s ID %q is empty or not unique", kind, id)
		}
		m[id] = true
	}
}

func ids[T any](v []T, f func(T) string) []string {
	s := make([]string, len(v))
	for i, x := range v {
		s[i] = f(x)
	}
	return s
}

func checkLabel(t *testing.T, l Label) {
	t.Helper()
	if !strings.HasPrefix(l.Key, "cardrank.") || l.Text == "" {
		t.Errorf("invalid label %+v", l)
	}
}

// profileTypes are a type evaluating each profile.
var profileTypes = map[string]Type{
	"high":         Holdem,
	"high_royal":   Royal,
	"high_short":   Short,
	"high_manila":  Manila,
	"high_spanish": Spanish,
	"soko":         Soko,
	"lowball_27":   Lowball,
	"razz":         Razz,
	"badugi":       Badugi,
	"low_8":        Split,
}

// profileHand evaluates the cards with the profile's type.
func profileHand(id string, v Cards) Hand {
	typ := profileTypes[id]
	var ev *Eval
	if typ.Board() == 0 {
		ev = typ.Eval(v, nil)
	} else {
		ev = typ.Eval(v[:2], v[2:])
	}
	if id == "low_8" {
		return ev.Lo
	}
	return ev.Hi
}

func TestHandCatalogExamples(t *testing.T) {
	for _, p := range HandCatalog().Profiles {
		for _, l := range p.Levels {
			h := profileHand(p.ID, l.Example)
			level, _, ok := profiles[p.ID].classify(h.Rank)
			t.Logf("%-12s %-15s %b %s", p.ID, l.ID, l.Example, h)
			if !ok || level != l.ID {
				t.Errorf("%s %s example %b classified as %q", p.ID, l.ID, l.Example, level)
			}
			for _, v := range l.Variants {
				h := profileHand(p.ID, v.Example)
				level, variant, ok := profiles[p.ID].classify(h.Rank)
				t.Logf("%-12s   %-15s %b %s", p.ID, v.ID, v.Example, h)
				if !ok || level != l.ID || variant != v.ID {
					t.Errorf("%s %s %s example %b classified as %q %q", p.ID, l.ID, v.ID, v.Example, level, variant)
				}
			}
		}
	}
}

// TestHandCatalogClassify classifies every rank each profile's deck can make,
// checking the classification against the hand's description, and that the
// levels and variants are ordered as the ranks are.
func TestHandCatalogClassify(t *testing.T) {
	for _, info := range HandCatalog().Profiles {
		t.Run(info.ID, func(t *testing.T) {
			p := profiles[info.ID]
			// a hand for each distinct rank
			n := 5
			if info.ID == "badugi" {
				n = 4
			}
			hands := make(map[EvalRank]Cards)
			for v := range Combinations(p.deck.Unshuffled(), n) {
				r := profileRank(info.ID, v)
				if _, ok := hands[r]; !ok {
					hands[r] = slices.Clone(v)
				}
			}
			t.Logf("%d ranks", len(hands))
			// the rank range of each level and variant
			type span struct{ min, max EvalRank }
			spans := make(map[string]*span)
			add := func(key string, r EvalRank) {
				if s, ok := spans[key]; ok {
					s.min, s.max = min(s.min, r), max(s.max, r)
				} else {
					spans[key] = &span{r, r}
				}
			}
			for r, v := range hands {
				level, variant, ok := p.classify(r)
				if !ok {
					if info.ID == "low_8" && r == Invalid {
						continue
					}
					t.Fatalf("rank %d %b not classified", r, v)
				}
				li := slices.IndexFunc(info.Levels, func(l LevelInfo) bool { return l.ID == level })
				if li == -1 {
					t.Fatalf("rank %d %b classified as unknown level %q", r, v, level)
				}
				if l := info.Levels[li]; (len(l.Variants) == 0) != (variant == "") ||
					variant != "" && !slices.ContainsFunc(l.Variants, func(x VariantInfo) bool { return x.ID == variant }) {
					t.Fatalf("rank %d %b classified as unknown variant %q of %q", r, v, variant, level)
				}
				add(level, r)
				if variant != "" {
					add(level+"/"+variant, r)
				}
				// the classification matches the description
				h := profileHand(info.ID, v)
				if h.Rank != r {
					t.Fatalf("%b expected rank %d, got: %d", v, r, h.Rank)
				}
				if expLevel, expVariant := descClass(info.ID, h); level != expLevel || variant != expVariant {
					t.Errorf("rank %d %b (%s) classified as %q %q, expected %q %q", r, v, h, level, variant, expLevel, expVariant)
				}
			}
			// levels and variants are best first, and every level and variant
			// can be made
			var prev *span
			for _, l := range info.Levels {
				s, ok := spans[l.ID]
				if !ok {
					t.Errorf("level %s is never made", l.ID)
					continue
				}
				t.Logf("%-15s %5d-%-5d", l.ID, s.min, s.max)
				if prev != nil && s.min <= prev.max {
					t.Errorf("level %s ranks %d-%d not after the previous level's %d-%d", l.ID, s.min, s.max, prev.min, prev.max)
				}
				prev = s
				var pv *span
				for _, v := range l.Variants {
					s, ok := spans[l.ID+"/"+v.ID]
					if !ok {
						t.Errorf("variant %s %s is never made", l.ID, v.ID)
						continue
					}
					if pv != nil && s.min <= pv.max {
						t.Errorf("variant %s %s not after the previous variant", l.ID, v.ID)
					}
					pv = s
				}
			}
		})
	}
}

// profileRank returns the profile's rank of the cards.
func profileRank(id string, v Cards) EvalRank {
	switch id {
	case "high", "high_royal":
		return RankCactus(v[0], v[1], v[2], v[3], v[4])
	case "high_short":
		return RankShort(v[0], v[1], v[2], v[3], v[4])
	case "high_manila":
		return RankManila(v[0], v[1], v[2], v[3], v[4])
	case "high_spanish":
		return RankSpanish(v[0], v[1], v[2], v[3], v[4])
	case "soko":
		return RankSoko(v[0], v[1], v[2], v[3], v[4])
	case "lowball_27":
		return RankLowball(v[0], v[1], v[2], v[3], v[4])
	case "razz":
		return RankRazz(v[0], v[1], v[2], v[3], v[4])
	case "low_8":
		if r := RankEightOrBetter(v[0], v[1], v[2], v[3], v[4]); r < EightOrBetter {
			return r
		}
		return Invalid
	}
	return profileHand(id, v).Rank
}

// descLevels are the level IDs of the short descriptions.
var descLevels = func() map[string]string {
	m := map[string]string{
		"Four Flush":    "four_flush",
		"Four Straight": "four_straight",
	}
	for _, l := range cactusLevels {
		m[l.rank.Title()] = l.id
	}
	return m
}()

// descClass returns the level and variant of a hand, from its description.
func descClass(profile string, h Hand) (string, string) {
	if profile == "badugi" {
		return [...]string{"", "one_card", "two_card", "three_card", "badugi"}[len(h.Best)], ""
	}
	s := fmt.Sprintf("%e", h)
	if id, ok := descLevels[s]; ok {
		if id == "straight_flush" {
			return id, straightFlushIDs[h.Best[0].Rank()]
		}
		return id, ""
	}
	switch {
	case strings.HasSuffix(s, "-high"):
		return "high_card", ""
	case strings.HasSuffix(s, "-low"):
		for r := Ace; r != InvalidRank; r-- {
			if r.Name()+"-low" == s {
				return "no_pair", lowIDs[r]
			}
		}
	}
	return "", ""
}

func TestHandCatalogRevision(t *testing.T) {
	c := HandCatalog()
	if !strings.HasPrefix(c.Revision, "sha256:") || len(c.Revision) != 7+64 {
		t.Fatalf("invalid revision %q", c.Revision)
	}
	// reproducible
	if r := catalogRevision(buildCatalog()); r != c.Revision {
		t.Errorf("expected revision %s, got: %s", c.Revision, r)
	}
	// changes when a name, example, or order changes
	for name, change := range map[string]func(*Catalog){
		"name": func(c *Catalog) {
			c.Profiles[0].Levels[0].Variants[0].Name.Text = "Royal"
		},
		"example": func(c *Catalog) {
			c.Profiles[0].Levels[1].Example[4] = New(Queen, Diamond)
		},
		"order": func(c *Catalog) {
			l := c.Profiles[0].Levels
			l[2], l[3] = l[3], l[2]
		},
		"game": func(c *Catalog) {
			c.Games[0].Max = 9
		},
	} {
		buf, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		var v Catalog
		if err := json.Unmarshal(buf, &v); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if r := catalogRevision(&v); r != c.Revision {
			t.Fatalf("expected round trip revision %s, got: %s", c.Revision, r)
		}
		change(&v)
		if r := catalogRevision(&v); r == c.Revision {
			t.Errorf("expected revision to change when the %s changes", name)
		}
	}
}

func TestEvalClasses(t *testing.T) {
	tests := []struct {
		typ    Type
		pocket string
		board  string
		exp    []HandClass
	}{
		{Holdem, "Ah Kh", "Qh Jh Th 2s 3s", []HandClass{{"hi", 0, "high", "straight_flush", "royal_flush", ""}}},
		{Holdem, "9h 8d", "2s 3c 4d", []HandClass{{"hi", 0, "high", "high_card", "", ""}}},
		{Short, "9h Ah", "8h 7h 6h Ks Kc", []HandClass{{"hi", 0, "high_short", "straight_flush", "iron_orb", ""}}},
		{Short, "Kh Kd", "Ks 7h 7s 9h Ah", []HandClass{{"hi", 0, "high_short", "full_house", "", ""}}},
		{Split, "Ah 2c", "3d 4s 8h Kc Kd", []HandClass{
			{"hi", 0, "high", "pair", "", ""},
			{"lo", 0, "low_8", "no_pair", "eight_low", ""},
		}},
		{Split, "Kh Qc", "Jd 9s 8h 2c 2d", []HandClass{
			{"hi", 0, "high", "pair", "", ""},
			{"lo", 0, "low_8", "", "", StatusNoQualifyingLow},
		}},
		{OmahaHiLo, "Ah 2h Kc Qd", "3c 4d 8s Kh Ts", []HandClass{
			{"hi", 0, "high", "pair", "", ""},
			{"lo", 0, "low_8", "no_pair", "eight_low", ""},
		}},
		{Video, "Th Td 4c 7s 2d", "", []HandClass{{"hi", 0, "high", "", "", StatusNoQualifyingHand}}},
		{Video, "Jh Jd 4c 7s 2d", "", []HandClass{{"hi", 0, "high", "pair", "", ""}}},
		{Razz, "As 2h 3d 4c 5s Kh Kd", "", []HandClass{{"hi", 0, "razz", "no_pair", "five_low", ""}}},
		{Lowball, "7s 5h 4d 3c 2s", "", []HandClass{{"hi", 0, "lowball_27", "no_pair", "seven_low", ""}}},
		{Soko, "Ah Jh 8h 3h Kc", "", []HandClass{{"hi", 0, "soko", "four_flush", "", ""}}},
		{Badugi, "As 2s 3h 4d", "", []HandClass{{"hi", 0, "badugi", "three_card", "", ""}}},
		{Badugi, "", "", []HandClass{{"hi", 0, "badugi", "", "", StatusNoHand}}},
	}
	for i, test := range tests {
		ev := test.typ.Eval(Must(test.pocket), Must(test.board))
		v := ev.Classes()
		t.Logf("%d %s %s: %s %+v", i, test.typ, ev, ev.Lo, v)
		if !slices.Equal(v, test.exp) {
			t.Errorf("test %d expected %+v, got: %+v", i, test.exp, v)
		}
		if ev.HiClass() != test.exp[0] {
			t.Errorf("test %d expected Hi class %+v, got: %+v", i, test.exp[0], ev.HiClass())
		}
	}
	// double board
	d := Double.Dealer(rand.New(rand.NewSource(1)), 2)
	for range d.Deal() {
	}
	for _, res := range d.Results() {
		v := res.Evals[0].Classes()
		if len(v) != 2 || v[0].Board != 0 || v[1].Board != 1 || v[1].Track != "hi" || v[1].Level == "" {
			t.Errorf("expected Hi classes for both boards, got: %+v", v)
		}
		if res.Evals[0].LoClass() != v[1] {
			t.Errorf("expected Lo class to be the second board's class")
		}
	}
}

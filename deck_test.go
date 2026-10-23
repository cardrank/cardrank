package cardrank

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDeck(t *testing.T) {
	tests := []struct {
		exp int
		typ DeckType
		r   string
	}{
		{52, DeckFrench, "23456789TJQKA"},
		{36, DeckShort, "6789TJQKA"},
		{32, DeckManila, "789TJQKA"},
		{28, DeckSpanish, "89TJQKA"},
		{20, DeckRoyal, "TJQKA"},
	}
	for _, test := range tests {
		t.Run(test.typ.Name(), func(t *testing.T) {
			testDeckNew(t, test.exp, test.typ, test.r)
			testDeckDraw(t, test.exp, test.typ)
			testDeckShoe(t, test.exp, test.typ)
		})
	}
}

func testDeckNew(t *testing.T, exp int, typ DeckType, r string) {
	t.Helper()
	v := typ.Unshuffled()
	switch {
	case len(v) != exp:
		t.Fatalf("expected length %d", exp)
	}
	// check cards
	d1, d2 := typ.New(), typ.New()
	for _, d := range []*Deck{d1, d2} {
		for _, r := range r {
			for _, s := range "shdc" {
				c := FromString(string(r) + string(s))
				if c == InvalidCard {
					t.Fatalf("expected valid card for %c%c", r, s)
				}
				if !slices.Contains(v, c) {
					t.Errorf("v does not contain %s", c)
				}
				if !slices.Contains(d.v, c) {
					t.Errorf("d.v does not contain %s", c)
				}
			}
		}
	}
	// check shuffle
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	d1.Shuffle(rnd)
	d2.Shuffle(rnd)
	if slices.Equal(d1.v, v) {
		t.Fatalf("expected d1.v != v")
	}
	if slices.Equal(d2.v, v) {
		t.Fatalf("expected d2.v != v")
	}
	if n, exp := len(d1.v), exp; n != exp {
		t.Fatalf("expected len(d1.v) == %d, got: %d", exp, n)
	}
	if n, exp := len(d2.v), exp; n != exp {
		t.Fatalf("expected len(d2.v) == %d, got: %d", exp, n)
	}
	for i := range exp {
		if !slices.Contains(d1.v, v[i]) {
			t.Errorf("d1.v does not contain %s", v[i])
		}
		if !slices.Contains(d2.v, v[i]) {
			t.Errorf("d2.v does not contain %s", v[i])
		}
	}
}

func testDeckDraw(t *testing.T, exp int, typ DeckType) {
	t.Helper()
	for n := 1; n < exp; n++ {
		d := typ.New()
		v := d.Draw(n)
		if len(v) != n {
			t.Fatalf("expected len(v) == %d, got: %d", n, len(v))
		}
		if d.Empty() {
			t.Fatalf("expected d to not be empty for %d", n)
		}
		d.Draw(exp - n)
		if !d.Empty() {
			t.Errorf("expected d to be empty for %d", n)
		}
	}
	d := typ.New()
	v := d.Draw(exp + 1)
	if n, exp := len(v), exp; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	for range 100 {
		_ = d.Draw(1)
		if n, exp := len(v), exp; n != exp {
			t.Errorf("expected len(v) == %d, got: %d", exp, n)
		}
	}
	if !d.Empty() {
		t.Errorf("expected d to be empty")
	}
	if n, exp := d.Remaining(), 0; n != exp {
		t.Errorf("expected d.Remaining() == %d, got: %d", exp, n)
	}
	d.Reset()
	if d.Remaining() != exp {
		t.Errorf("expected d.Remaining() == %d", exp)
	}
}

func testDeckShoe(t *testing.T, exp int, typ DeckType) {
	t.Helper()
	const count = 7
	d := typ.Shoe(count)
	d.Shuffle(rand.New(rand.NewSource(time.Now().UnixNano())))
	if n, exp := len(d.v), count*exp; n != exp {
		t.Fatalf("expected len(d.v) == %d, got: %d", exp, n)
	}
	m := make(map[Card]int, exp)
	for _, c := range d.v {
		m[c]++
	}
	if n, exp := len(m), exp; n != exp {
		t.Errorf("expected %d, got: %d", exp, n)
	}
	for _, c := range typ.Unshuffled() {
		switch i, ok := m[c]; {
		case !ok:
			t.Fatalf("expected m to contain %s", c)
		case i != count:
			t.Errorf("expected %d == %d", count, i)
		}
	}
	limit := (count - 2) * exp
	d.Limit(limit)
	v := d.Draw(count * exp)
	if n, exp := len(v), limit; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	if n, exp := d.Remaining(), 0; n != exp {
		t.Errorf("expected d.Remaining() == %d, got: %d", exp, n)
	}
}

func TestDealer(t *testing.T) {
	// seed := time.Now().UnixNano()
	// seed := int64(1676122011905868217)
	seed := int64(1677109206437341728)
	t.Logf("seed: %d", seed)
	r := rand.New(rand.NewSource(seed))
	for _, typ := range Types() {
		if maximum := typ.Max(); maximum != 1 {
			for i := 2; i <= maximum; i++ {
				count, s := i, r.Int63()
				t.Run(fmt.Sprintf("%s/%d", typ, count), func(t *testing.T) {
					testDealer(t, typ, count, s, nil)
				})
			}
		} else {
			for i := 1; i <= 8; i++ {
				s := r.Int63()
				t.Run(fmt.Sprintf("%s/%d", typ, i), func(t *testing.T) {
					testDealer(t, typ, 1, s, nil)
				})
			}
		}
	}
}

func TestDealerRuns(t *testing.T) {
	tests := []struct {
		typ   Type
		count int
		seed  int64
	}{
		{Holdem, 5, 17},
		{Double, 6, 22},
		{Omaha, 4, 100},
		{OmahaDouble, 4, 182},
		{OmahaHiLo, 4, 72},
		{FusionHiLo, 5, 256},
		{Manila, 3, 768},
		{Jakarta, 2, 101},
	}
	for _, test := range tests {
		t.Run(test.typ.Name(), func(t *testing.T) {
			testDealer(t, test.typ, test.count, test.seed, func(_ *rand.Rand, d *Dealer, run int) {
				if d.Street().ID == 'f' && run == 0 {
					if err := d.ChangeRuns(3); err != nil {
						t.Fatalf("expected no error, got: %v", err)
					}
					// fold positions 3 and 4, when present
					var fold []int
					for _, i := range []int{3, 4} {
						if i < d.Players() {
							fold = append(fold, i)
						}
					}
					if err := d.Fold(fold...); err != nil {
						t.Fatalf("expected no error, got: %v", err)
					}
				}
			})
		})
	}
}

type dealFunc func(r *rand.Rand, d *Dealer, run int)

func testDealer(t *testing.T, typ Type, count int, seed int64, f dealFunc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r := rand.New(rand.NewSource(seed))
	d := typ.Dealer(r, count)
	desc := typ.Desc()
	t.Logf("Eval: %l", typ)
	t.Logf("Desc: %s/%s", desc.HiDesc, desc.LoDesc)
	deck := slices.Collect(d.Deck().All())
	t.Logf("Deck: %s [%d]", desc.Deck, len(deck))
	for i := 0; i < len(deck); i += 8 {
		t.Logf("  %v", deck[i:min(i+8, len(deck))])
	}
	last := -1
	for i, run := range d.Deal() {
		if last != i {
			t.Logf("Run %d:", i)
		}
		last = i
		t.Logf("  %s", d)
		street := d.Street()
		if street.Pocket != 0 {
			for i := range count {
				t.Logf("    %d: %v", i, run.Pockets[i])
			}
		}
		if v := d.Discarded(); len(v) != 0 {
			t.Logf("    Discard: %v", v)
		}
		if street.Board != 0 {
			t.Logf("    Board: %v", run.Hi)
			if typ.Double() {
				t.Logf("           %v", run.Lo)
			}
		}
		if d.HasCalc() {
			if odds, err := d.Odds(ctx); err == nil && odds.Hi != nil {
				t.Log("    Calc:")
				for i := range len(odds.Hi.Counts) {
					t.Logf("      %d: %*s", i, i, odds.Hi)
					if odds.Lo != nil {
						t.Logf("         %*s", i, odds.Lo)
					}
				}
			}
		}
		if f != nil {
			f(r, d, i)
		}
	}
	t.Logf("Showdown:")
	for i, res := range d.Results() {
		run := d.runs[i]
		t.Logf("  Run %d:", i)
		if typ.Board() != 0 {
			t.Logf("    Board: %v", run.Hi)
			if typ.Double() {
				t.Logf("           %v", run.Lo)
			}
		}
		if typ.Pocket() != 0 {
			t.Log("    Pockets:")
			for i := range count {
				t.Logf("      %d: %v", i, run.Pockets[i])
			}
		}
		t.Log("    Evals:")
		for i := range count {
			if d.Active(i) {
				hi := res.Evals[i].Hi
				t.Logf("      %d: %v %v %s", i, hi.Best, hi.Unused, hi)
				if typ.Low() || typ.Double() {
					lo := res.Evals[i].Lo
					t.Logf("         %v %v %s", lo.Best, lo.Unused, lo)
				}
			} else {
				t.Logf("      %d: inactive", i)
			}
		}
		t.Log("    Result:")
		t.Logf("      %S", res.Hi)
		if res.Lo.Valid() {
			t.Logf("      %S", res.Lo)
		}
	}
}

func TestDealStreets(t *testing.T) {
	for _, typ := range Types() {
		if maximum := typ.Max(); maximum != 1 {
			for i := 2; i <= maximum; i++ {
				t.Run(fmt.Sprintf("%s/%d", typ, i), func(t *testing.T) {
					testDealStreets(t, typ, len(typ.Streets()), i)
				})
			}
		} else {
			for i := 1; i <= 8; i++ {
				t.Run(fmt.Sprintf("%s/%d", typ, i), func(t *testing.T) {
					testDealStreets(t, typ, len(typ.Streets()), 1)
				})
			}
		}
	}
}

func testDealStreets(t *testing.T, typ Type, exp, count int) {
	t.Helper()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := typ.Dealer(r, count)
	var streets int
	for range d.Deal() {
		streets++
	}
	if streets != exp {
		t.Errorf("expected %d, got: %d", exp, streets)
	}
	// dealing is complete
	for range d.Deal() {
		t.Fatal("expected no further streets")
	}
}

func TestRunOut(t *testing.T) {
	// seed := time.Now().UnixNano()
	const seed = 1679273183508957122
	t.Logf("seed: %d", seed)
	for _, typ := range Types() {
		if typ.Board() == 0 || typ.Draw() {
			continue
		}
		for n := 2; n <= typ.Max()-6; n++ {
			t.Run(fmt.Sprintf("%s/%d", typ, n), func(t *testing.T) {
				testRunOut(t, seed, typ, n)
			})
		}
	}
}

func testRunOut(t *testing.T, seed int64, typ Type, count int) {
	t.Helper()
	const runOuts = 4
	var deck Cards
	runs, results := make([][]*Run, runOuts), make([][]*Result, runOuts)
	r := rand.New(rand.NewSource(seed))
	d := typ.Dealer(r, count)
	deck = slices.Collect(d.Deck().All())
	for i := 0; i < len(deck); i += 8 {
		t.Logf("%v", deck[i:min(i+8, len(deck))])
	}
	v := make([]string, runOuts)
	for i := range runOuts {
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "-- %d --\n", i)
		d.Reset()
		for run := range d.Deal() {
			// change run on the flop
			fmt.Fprintf(buf, "%s\n", d)
			if i != 0 && run == 0 && d.Street().ID == 'f' {
				if err := d.ChangeRuns(i + 1); err != nil {
					t.Fatalf("unable to change runs to %d: %v", i+1, err)
				} else {
					fmt.Fprintf(buf, "runs changed to %d\n", i+1)
				}
			} else if i == 0 && run == 0 && d.Street().ID == 'f' {
				fmt.Fprintln(buf)
			}
		}
		for range d.Results() {
		}
		v[i], runs[i], results[i] = buf.String(), d.runs, d.results
	}
	t.Log("")
	sidebyside(t, "", "  ", v...)
	for i := range runOuts - 1 {
		t.Log("")
		for j := i + 1; j < runOuts; j++ {
			t.Logf("--- %d :: %d ---", i, j)
			/*
				if n := len(discard[j-1]); reflect.DeepEqual(discard[j-1], discard[j][:n]) {
					t.Logf("  discard: %v / %v", discard[j-1], discard[j][n:])
				} else {
					t.Errorf("  expected discard prefix %v, got: %v / %v", discard[j-1], discard[j][:n], discard[j][n:])
				}
			*/
			for k := range len(runs[i]) {
				t.Logf("%d:", k)
				dumpRuns(t, runs[i][k], runs[j][k])
				if !reflect.DeepEqual(runs[i][k], runs[j][k]) {
					t.Errorf("  run out %d and %d are incongruent for run %d", i, j, k)
				} else {
					t.Logf("  run out %d and %d are congruent for run %d", i, j, k)
				}
				if !reflect.DeepEqual(results[i][k], results[j][k]) {
					t.Errorf("  run out %d and %d are incongruent for result %d", i, j, k)
				} else {
					t.Logf("  run out %d and %d are congruent for result %d", i, j, k)
				}
			}
		}
	}
}

func dumpRuns(t *testing.T, runs ...*Run) {
	t.Helper()
	v := make([]string, len(runs))
	for i, run := range runs {
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "d: %v\n", run.Discard)
		fmt.Fprintf(buf, "b: %v\n", run.Hi)
		if len(run.Lo) != 0 {
			fmt.Fprintf(buf, "%v\n", run.Lo)
		}
		for i, pocket := range run.Pockets {
			fmt.Fprintf(buf, "%d: %v\n", i, pocket)
		}
		v[i] = buf.String()
	}
	sidebyside(t, "  ", "  ", v...)
}

func sidebyside(t *testing.T, pad, gap string, v ...string) {
	t.Helper()
	if len(v) == 0 {
		return
	}
	n, lines, widths := 0, make([][]string, len(v)), make([]int, len(v))
	for i, s := range v {
		lines[i] = strings.Split(strings.TrimSuffix(s, "\n"), "\n")
		for _, line := range lines[i] {
			widths[i] = max(widths[i], len(line))
		}
		n = max(n, len(lines[i]))
	}
	var x string
	for i := range n {
		s := pad
		for j := range v {
			if j != 0 {
				s += gap
			}
			if x = ""; i < len(lines[j]) {
				x = lines[j][i]
			}
			s += fmt.Sprintf("%-*s", widths[j], x)
		}
		t.Log(s)
	}
}

func TestDeckFrenchJoker(t *testing.T) {
	typ := DeckFrenchJoker
	for _, other := range []DeckType{DeckFrench, DeckShort, DeckManila, DeckSpanish, DeckRoyal, DeckKuhn, DeckLeduc} {
		if typ == other {
			t.Fatalf("expected DeckFrenchJoker to differ from %s", other.Name())
		}
	}
	if s, exp := typ.Name(), "French Joker"; s != exp {
		t.Errorf("expected Name() == %q, got: %q", exp, s)
	}
	if s, exp := typ.Desc(false), "French Joker"; s != exp {
		t.Errorf("expected Desc(false) == %q, got: %q", exp, s)
	}
	if n, exp := len(typ.Unshuffled()), 54; n != exp {
		t.Errorf("expected %d cards, got: %d", exp, n)
	}
}

func TestDeckLimit(t *testing.T) {
	tests := []struct {
		limit int
		exp   int
	}{
		{-1, 0},
		{0, 0},
		{10, 10},
		{52, 52},
		{60, 52},
	}
	for _, test := range tests {
		t.Run(strconv.Itoa(test.limit), func(t *testing.T) {
			d := DeckFrench.New()
			d.Limit(test.limit)
			if n := d.Remaining(); n != test.exp {
				t.Errorf("expected Remaining() == %d, got: %d", test.exp, n)
			}
			if n := len(d.Draw(100)); n != test.exp {
				t.Errorf("expected to draw %d cards, got: %d", test.exp, n)
			}
			if n := len(slices.Collect(d.All())); n != test.exp {
				t.Errorf("expected All() to return %d cards, got: %d", test.exp, n)
			}
		})
	}
}

func TestDealerFormatBeforeDeal(t *testing.T) {
	d := NewDealer(Holdem, DeckFrench.New(), 2)
	for _, verb := range []string{"%n", "%s"} {
		if s := fmt.Sprintf(verb, d); s != "" {
			t.Errorf("expected %s to be empty before dealing, got: %q", verb, s)
		}
	}
	for range d.Deal() {
		break
	}
	if s, exp := fmt.Sprintf("%n", d), "Pre-Flop"; s != exp {
		t.Errorf("expected %%n == %q, got: %q", exp, s)
	}
}

func TestWinValid(t *testing.T) {
	evs := []*Eval{EvalOf(Holdem)}
	for _, win := range []Win{
		{},
		{evals: evs, Ranking: Ranking{Order: []int{0}}},
		{evals: evs, Ranking: Ranking{Pivot: 1}},
		{Ranking: Ranking{Order: []int{0}, Pivot: 1}},
		// unevaluated
		{evals: evs, Ranking: Ranking{Order: []int{0}, Pivot: 1}},
	} {
		if win.Valid() {
			t.Errorf("expected %+v to be invalid", win)
		}
		if s, exp := fmt.Sprintf("%S", win), "None"; s != exp {
			t.Errorf("expected %%S == %q, got: %q", exp, s)
		}
	}
	ev := Holdem.Eval(Must("Ah Kh"), Must("Qh Jh Th 2s 3s"))
	win := Win{evals: []*Eval{ev}, Ranking: Ranking{Order: []int{0}, Pivot: 1}}
	if !win.Valid() {
		t.Errorf("expected win to be valid")
	}
	if s, exp := win.Describe("alice"), "alice wins with Straight Flush, Ace-high, Royal"; s != exp {
		t.Errorf("expected %q, got: %q", exp, s)
	}
}

func TestDealerStreetIndex(t *testing.T) {
	d := Holdem.Dealer(rand.New(rand.NewSource(1)), 2)
	if i := d.StreetIndex(); i != -1 {
		t.Fatalf("expected -1 before dealing, got: %d", i)
	}
	streets := Holdem.Streets()
	type dealt struct {
		run, street int
	}
	var v []dealt
	for run := range d.Deal() {
		i := d.StreetIndex()
		if i < 0 || len(streets) <= i {
			t.Fatalf("expected a valid street index, got: %d", i)
		}
		if exp := streets[i]; d.Street() != exp {
			t.Errorf("expected street %v, got: %v", exp, d.Street())
		}
		v = append(v, dealt{run, i})
		// run it twice, after the flop
		if d.Street().ID == 'f' && run == 0 {
			if err := d.ChangeRuns(2); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		}
	}
	// the second run deals the turn and river
	exp := []dealt{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}}
	if !slices.Equal(v, exp) {
		t.Errorf("expected %v, got: %v", exp, v)
	}
	if i := d.StreetIndex(); i != -1 {
		t.Errorf("expected -1 after dealing, got: %d", i)
	}
}

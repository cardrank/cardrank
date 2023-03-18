package cardrank

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
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
	for _, tt := range tests {
		test := tt
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
				if !contains(c, v...) {
					t.Errorf("v does not contain %s", c)
				}
				if !contains(c, d.v...) {
					t.Errorf("d.v does not contain %s", c)
				}
			}
		}
	}
	// check shuffle
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	d1.Shuffle(rnd, 1)
	d2.Shuffle(rnd, 1)
	if equals(d1.v, v) {
		t.Fatalf("expected d1.v != v")
	}
	if equals(d2.v, v) {
		t.Fatalf("expected d2.v != v")
	}
	if n, exp := len(d1.v), exp; n != exp {
		t.Fatalf("expected len(d1.v) == %d, got: %d", exp, n)
	}
	if n, exp := len(d2.v), exp; n != exp {
		t.Fatalf("expected len(d2.v) == %d, got: %d", exp, n)
	}
	for i := 0; i < exp; i++ {
		if !contains(v[i], d1.v...) {
			t.Errorf("d1.v does not contain %s", v[i])
		}
		if !contains(v[i], d2.v...) {
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
	for i := 0; i < 100; i++ {
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
	d.Shuffle(rand.New(rand.NewSource(time.Now().UnixNano())), 1)
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
	for _, tt := range Types() {
		if max := tt.Max(); max != 1 {
			for i := 2; i <= max; i++ {
				typ, count, s := tt, i, r.Int63()
				t.Run(fmt.Sprintf("%s/%d", typ, count), func(t *testing.T) {
					testDealer(t, typ, count, s, nil)
				})
			}
		} else {
			for i := 1; i <= 8; i++ {
				typ, s := tt, r.Int63()
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
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.typ.Name(), func(t *testing.T) {
			testDealer(t, test.typ, test.count, test.seed, func(r *rand.Rand, d *Dealer) {
				switch run, _ := d.Run(); {
				case d.Id() == 'f' && run == 0:
					if b, exp := d.ChangeRuns(3), true; b != exp {
						t.Fatalf("expected %t, got: %t", exp, b)
					}
					if b, exp := d.Deactivate(3, 4), true; b != exp {
						t.Fatalf("expected %t, got: %t", exp, b)
					}
				}
			})
		})
	}
}

type dealFunc func(r *rand.Rand, d *Dealer)

func testDealer(t *testing.T, typ Type, count int, seed int64, f dealFunc) {
	t.Helper()
	r := rand.New(rand.NewSource(seed))
	d := typ.Dealer(r, 1, count)
	desc := typ.Desc()
	t.Logf("Eval: %l", typ)
	t.Logf("Desc: %s/%s", desc.HiDesc, desc.LoDesc)
	deck := d.Deck.All()
	t.Logf("Deck: %s [%d]", desc.Deck, len(deck))
	for i := 0; i < len(deck); i += 8 {
		t.Logf("  %v", deck[i:min(i+8, len(deck))])
	}
	last := -1
	for d.Next() {
		i, run := d.Run()
		if last != i {
			t.Logf("Run %d:", i)
		}
		last = i
		t.Logf("  %s", d)
		if d.HasPocket() {
			for i := 0; i < count; i++ {
				t.Logf("    %d: %v", i, run.Pockets[i])
			}
		}
		if v := d.Discarded(); len(v) != 0 {
			t.Logf("    Discard: %v", v)
		}
		if d.HasBoard() {
			t.Logf("    Board: %v", run.Hi)
			if d.Double {
				t.Logf("           %v", run.Lo)
			}
		}
		if f != nil {
			f(r, d)
		}
	}
	t.Logf("Showdown:")
	for d.NextResult() {
		run, res := d.Result()
		t.Logf("  Run %d:", run)
		if d.Type.Board() != 0 {
			t.Logf("    Board: %v", d.Runs[run].Hi)
			if d.Double {
				t.Logf("           %v", d.Runs[run].Lo)
			}
		}
		if d.Type.Pocket() != 0 {
			t.Log("    Pockets:")
			for i := 0; i < count; i++ {
				t.Logf("      %d: %v", i, d.Runs[run].Pockets[i])
			}
		}
		t.Log("    Evals:")
		for i := 0; i < count; i++ {
			if d.Active[i] {
				hi := res.Evals[i].Desc(false)
				t.Logf("      %d: %v %v %s", i, hi.Best, hi.Unused, hi)
				if d.Low || d.Double {
					lo := res.Evals[i].Desc(true)
					t.Logf("         %v %v %s", lo.Best, lo.Unused, lo)
				}
			} else {
				t.Logf("      %d: inactive", i)
			}
		}
		hi, lo := res.Win()
		t.Log("    Result:")
		t.Logf("      %S", hi)
		if lo != nil {
			t.Logf("      %S", lo)
		}
	}
}

func TestHasNext(t *testing.T) {
	for _, tt := range Types() {
		typ := tt
		if max := typ.Max(); max != 1 {
			for i := 2; i <= max; i++ {
				t.Run(fmt.Sprintf("%s/%d", typ, i), func(t *testing.T) {
					testHasNext(t, typ, len(typ.Streets()), i)
				})
			}
		} else {
			for i := 1; i <= 8; i++ {
				t.Run(fmt.Sprintf("%s/%d", typ, i), func(t *testing.T) {
					testHasNext(t, typ, len(typ.Streets()), 1)
				})
			}
		}
	}
}

func testHasNext(t *testing.T, typ Type, exp, count int) {
	t.Helper()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := typ.Dealer(r, 1, count)
	if !d.HasNext() {
		t.Fatal("expected to have next")
	}
	var next int
	if d.HasNext() {
		next++
	}
	var streets int
	for d.Next() {
		if d.HasNext() {
			next++
		}
		streets++
	}
	switch {
	case streets != exp:
		t.Errorf("expected %d, got: %d", exp, streets)
	case next != exp:
		t.Errorf("expected %d, got: %d", exp, next)
	case d.HasNext():
		t.Error("expected to not have next")
	}
}

func TestRunOut(t *testing.T) {
	// seed := time.Now().UnixNano()
	const seed = 1679273183508957122
	t.Logf("seed: %d", seed)
	for _, tt := range Types() {
		if tt.Board() == 0 || tt.Draw() {
			continue
		}
		for n := 2; n <= tt.Max()-6; n++ {
			typ, count := tt, n
			t.Run(fmt.Sprintf("%s/%d", typ, count), func(t *testing.T) {
				testRunOut(t, seed, typ, count)
			})
		}
	}
}

func testRunOut(t *testing.T, seed int64, typ Type, count int) {
	t.Helper()
	const runOuts = 4
	var deck []Card
	runs, results := make([][]*Run, runOuts), make([][]*Result, runOuts)
	r := rand.New(rand.NewSource(seed))
	d := typ.Dealer(r, 3, count)
	deck = d.Deck.All()
	for i := 0; i < len(deck); i += 8 {
		t.Logf("%v", deck[i:min(i+8, len(deck))])
	}
	v := make([]string, runOuts)
	for i := 0; i < runOuts; i++ {
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "-- %d --\n", i)
		d.Reset()
		for d.Next() {
			// change run on the flop
			run, _ := d.Run()
			fmt.Fprintf(buf, "%s\n", d)
			if i != 0 && run == 0 && d.Id() == 'f' {
				if !d.ChangeRuns(i + 1) {
					t.Fatalf("unable to change runs to %d", i+1)
				} else {
					fmt.Fprintf(buf, "runs changed to %d\n", i+1)
				}
			} else if i == 0 && run == 0 && d.Id() == 'f' {
				fmt.Fprintln(buf)
			}
		}
		for d.NextResult() {
		}
		v[i], runs[i], results[i] = buf.String(), d.Runs, d.Results
	}
	t.Log("")
	sidebyside(t, "", "  ", v...)
	for i := 0; i < runOuts-1; i++ {
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
			for k := 0; k < len(runs[i]); k++ {
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
	for i := 0; i < len(v); i++ {
		lines[i] = strings.Split(strings.TrimSuffix(v[i], "\n"), "\n")
		for j := 0; j < len(lines[i]); j++ {
			widths[i] = max(widths[i], len(lines[i][j]))
		}
		n = max(n, len(lines[i]))
	}
	var x string
	for i := 0; i < n; i++ {
		s := pad
		for j := 0; j < len(v); j++ {
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

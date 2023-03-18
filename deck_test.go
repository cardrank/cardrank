package cardrank

import (
	"fmt"
	"math/rand"
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
				if !contains(v, c) {
					t.Errorf("v does not contain %s", c)
				}
				if !contains(d.v, c) {
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
		if !contains(d1.v, v[i]) {
			t.Errorf("d1.v does not contain %s", v[i])
		}
		if !contains(d2.v, v[i]) {
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
	for d.Next() {
		t.Logf("%s", d)
		rn, run := d.Run()
		t.Logf("  Run %d:", rn)
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
			if d.Low || d.Double {
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
	t.Parallel()
	for _, typ := range Types() {
		exp := len(typ.Streets())
		if max := typ.Max(); max != 1 {
			for i := 2; i <= max; i++ {
				testHasNext(t, typ, exp, i)
			}
		} else {
			for i := 1; i <= 8; i++ {
				testHasNext(t, typ, exp, 1)
			}
		}
	}
}

func testHasNext(t *testing.T, typ Type, exp, count int) {
	t.Helper()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := typ.Dealer(r, 1, count)
	if !d.HasNext() {
		t.Fatalf("%s expected to have next", typ)
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
		t.Errorf("%s expected %d, got: %d", typ, exp, streets)
	case next != exp:
		t.Errorf("%s expected %d, got: %d", typ, exp, next)
	case d.HasNext():
		t.Errorf("%s expected to not have next", typ)
	}
}

func TestRunOut(t *testing.T) {
}

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
		{20, DeckRoyal, "TJQKA"},
	}
	for _, v := range tests {
		test := v
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
	seed := int64(1676122011905868217)
	t.Logf("seed: %d", seed)
	r := rand.New(rand.NewSource(seed))
	for _, tt := range Types() {
		for i := 2; i <= tt.Max(); i++ {
			typ, count, s := tt, i, r.Int63()
			t.Run(fmt.Sprintf("%s/%d", typ, count), func(t *testing.T) {
				testDealer(t, typ, count, s, nil)
			})
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
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.typ.Name(), func(t *testing.T) {
			testDealer(t, test.typ, test.count, test.seed, func(r *rand.Rand, d *Dealer) bool {
				switch d.Id() {
				case 'f':
					d.Deactivate(3, 4)
					if b, exp := d.Runs(3), true; b != exp {
						t.Fatalf("expected %t, got: %t", exp, b)
					}
				}
				return false
			})
		})
	}
}

type dealFunc func(r *rand.Rand, d *Dealer) bool

func testDealer(t *testing.T, typ Type, count int, seed int64, f dealFunc) {
	t.Helper()
	r := rand.New(rand.NewSource(seed))
	d := typ.Dealer(r, 1, count)
	desc := typ.Desc()
	t.Logf("Eval: %s", desc.Eval)
	t.Logf("HiComp: %s LoComp: %s", desc.HiComp, desc.LoComp)
	t.Logf("HiDesc: %s LoDesc: %s", desc.HiDesc, desc.LoDesc)
	t.Logf("Deck: %s [%d]", desc.Deck, len(d.Deck.v))
	deck := d.Deck.All()
	for i := 0; i < len(deck); i += 8 {
		t.Logf("  %v", d.Deck.v[i:min(i+8, len(deck))])
	}
	for d.Next() {
		t.Logf("%s", d)
		if d.HasPocket() {
			for i := 0; i < count; i++ {
				t.Logf("  %d: %v", i, d.Pockets[i])
			}
		}
		if v := d.Discarded(); len(v) != 0 {
			t.Logf("  Discard: %v", v)
		}
		if d.HasBoard() {
			for i := 0; i < len(d.Boards); i++ {
				t.Logf("  Run %d: %v", i, d.Boards[i].Hi)
				if d.Double {
					t.Logf("         %v", d.Boards[i].Lo)
				}
			}
		}
		end := false
		if f != nil {
			end = f(r, d)
		}
		if end {
			break
		}
	}
	t.Logf("Showdown:")
	for d.NextResult() {
		run, res := d.Result()
		t.Logf("  Run %d:", run)
		for i := 0; i < count; i++ {
			if d.Active[i] {
				hi := res.Evals[i].HiDesc()
				t.Logf("    %d: %v %v %s", i, hi.Best, hi.Unused, hi)
				if d.Low || d.Double {
					lo := res.Evals[i].LoDesc()
					t.Logf("       %v %v %s", lo.Best, lo.Unused, lo)
				}
			} else {
				t.Logf("    %d: inactive", i)
			}
		}
		hi, lo := res.Win()
		t.Logf("    Result: %s with %S", hi, hi)
		if lo != nil {
			t.Logf("            %s with %S", lo, lo)
		}
	}
}

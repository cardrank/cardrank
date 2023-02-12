package cardrank

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestDeck(t *testing.T) {
	tests := []struct {
		typ DeckType
		r   string
		v   []Card
		exp int
	}{
		{DeckFrench, "23456789TJQKA", unshuffledFrench, 52},
		{DeckShort, "6789TJQKA", unshuffledShort, 36},
		{DeckManila, "789TJQKA", unshuffledManila, 32},
		{DeckRoyal, "TJQKA", unshuffledRoyal, 20},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.typ.String(), func(t *testing.T) {
			v := test.typ.Unshuffled()
			d := test.typ.New()
			switch {
			case len(v) != test.exp,
				len(d.v) != test.exp,
				len(test.v) != test.exp:
				t.Fatalf("expected length %d", test.exp)
			}
			// check cards
			for _, r := range test.r {
				for _, s := range "shdc" {
					c := FromString(string(r) + string(s))
					if c == InvalidCard {
						t.Fatalf("expected valid card for %c%c", r, s)
					}
					if !contains(v, c) {
						t.Errorf("does not contain %s", c)
					}
					if !contains(d.v, c) {
						t.Errorf("does not contain %s", c)
					}
					if !contains(test.v, c) {
						t.Errorf("does not contain %s", c)
					}
				}
			}
			// check deal
			d1, d2 := test.typ.New(), test.typ.New()
			if !reflect.DeepEqual(d1.v, d2.v) || !reflect.DeepEqual(d1.v, test.v) {
				t.Fatalf("expected d1.v == d2.v == test.v")
			}
			// shuffle
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			d1.Shuffle(r)
			d2.Shuffle(r)
			if reflect.DeepEqual(d1.v, test.v) {
				t.Fatalf("expected d1.v != test.v")
			}
			if reflect.DeepEqual(d2.v, test.v) {
				t.Fatalf("expected d2.v != test.v")
			}
			if reflect.DeepEqual(d1.v, d2.v) {
				t.Fatalf("expected d1.v != d2.v")
			}
			if n, exp := len(d1.v), test.exp; n != exp {
				t.Fatalf("expected len(d1.v) == %d, got: %d", exp, n)
			}
			if n, exp := len(d2.v), test.exp; n != exp {
				t.Fatalf("expected len(d2.v) == %d, got: %d", exp, n)
			}
			for i := 0; i < test.exp; i++ {
				if !contains(d1.v, test.v[i]) {
					t.Errorf("d1.v does not contain %s", test.v[i])
				}
				if !contains(d1.v, d2.v[i]) {
					t.Errorf("d1.v does not contain %s", d2.v[i])
				}
			}
		})
	}
}

func TestNewDeckShoe(t *testing.T) {
	const decks = 7
	d := NewShoeDeck(decks)
	d.Shuffle(rand.New(rand.NewSource(time.Now().UnixNano())))
	if n, exp := len(d.v), unshuffledSize*decks; n != exp {
		t.Fatalf("expected len(d.v) == %d, got: %d", exp, n)
	}
	m := make(map[uint]int, len(d.v))
	for _, c := range d.v {
		m[uint(c)]++
	}
	for _, c := range unshuffledFrench {
		i, ok := m[uint(c)]
		if !ok {
			t.Fatalf("expected m to contain %s", c)
		}
		if i != decks {
			t.Errorf("expected %d == %d", decks, i)
		}
	}
	limit := 5 * unshuffledSize
	d.SetLimit(limit)
	v := d.Draw(decks * unshuffledSize)
	if n, exp := len(v), limit; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	if n, exp := d.Remaining(), 0; n != exp {
		t.Errorf("expected d.Remaining() == %d, got: %d", exp, n)
	}
}

func TestDeckDraw(t *testing.T) {
	for exp := 1; exp < unshuffledSize; exp++ {
		d := NewDeck()
		v := d.Draw(exp)
		if len(v) != exp {
			t.Fatalf("expected len(v) == %d, got: %d", exp, len(v))
		}
		if d.Empty() {
			t.Fatalf("expected d to not be empty")
		}
		d.Draw(unshuffledSize - exp)
		if !d.Empty() {
			t.Errorf("expected d to be empty")
		}
	}
}

func TestDeckDrawAll(t *testing.T) {
	d := NewDeck()
	v := d.Draw(100)
	if n, exp := len(v), unshuffledSize; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	if !d.Empty() {
		t.Errorf("expected d to be empty")
	}
	if n, exp := d.Remaining(), 0; n != exp {
		t.Errorf("expeceted d.Remaining() == %d, got: %d", exp, n)
	}
}

func TestDeckDrawEmpty(t *testing.T) {
	d := NewDeck()
	if d.Empty() {
		t.Fatalf("expected d to not be empty")
	}
	v := d.Draw(unshuffledSize - 1)
	if d.Empty() {
		t.Fatalf("expected d to not be empty")
	}
	if n, exp := len(v), unshuffledSize-1; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	v = append(v, d.Draw(1)...)
	if !d.Empty() {
		t.Errorf("expected d to be empty, remaining: %d", d.Remaining())
	}
	if n, exp := len(v), unshuffledSize; n != exp {
		t.Fatalf("expected len(v) == %d, got: %d", exp, n)
	}
	for i := 0; i < unshuffledSize; i++ {
		if !contains(v, unshuffledFrench[i]) {
			t.Errorf("v does not contain %s", unshuffledFrench[i])
		}
	}
}

func TestDealer(t *testing.T) {
	// seed := time.Now().UnixNano()
	seed := int64(1676122011905868217)
	t.Logf("seed: %d", seed)
	r := rand.New(rand.NewSource(seed))
	for _, tt := range Types() {
		for i := 2; i <= tt.Max(); i++ {
			typ, n, s := tt, i, r.Int63()
			t.Run(fmt.Sprintf("%s/%d", typ, n), func(t *testing.T) {
				testDealer(t, n, typ, s)
			})
		}
	}
}

func testDealer(t *testing.T, hands int, typ Type, seed int64) {
	d := typ.Dealer(rand.New(rand.NewSource(seed)), 3)
	t.Logf("Deck (%s, %d):", typ.DeckType(), len(d.d.v))
	for i := 0; i < len(d.d.v); i += 8 {
		t.Logf("  %v", d.d.v[i:min(i+8, len(d.d.v))])
	}
	double, low := typ.Double(), typ.Low()
	var pockets [][]Card
	var b1, b2 []Card
	for d.Next() {
		t.Logf("%s:", d)
		pockets, b1 = d.Deal(pockets, b1, hands)
		if 0 < d.Pocket() {
			for i := 0; i < hands; i++ {
				t.Logf("  % 2d: %v", i, pockets[i])
			}
		}
		if 0 < d.Board() {
			t.Logf("  Board: %v", b1)
			if double {
				b2 = d.DealBoard(b2, false)
				t.Logf("         %v", b2)
			}
		}
	}
	h1 := typ.RankHands(pockets, b1)
	var h2 []*Hand
	if double {
		h2 = typ.RankHands(pockets, b2)
	}
	t.Logf("Showdown:")
	for i := 0; i < len(h1); i++ {
		t.Logf(" % 2d: %04d %v %v %s", i, h1[i].HiRank, h1[i].HiBest, h1[i].HiUnused, h1[i].Description())
		switch {
		case double:
			t.Logf("     %04d %v %v %s", h2[i].HiRank, h2[i].HiBest, h2[i].HiUnused, h2[i].Description())
		case low:
			t.Logf("     %04d %v %v %s", h1[i].LoRank, h1[i].LoBest, h1[i].LoUnused, h1[i].LowDescription())
		}
	}
	t.Logf("Result:")
	win := NewWin(h1, h2, low)
	t.Logf("  %s", win.HiDesc(func(_, i int) string {
		return strconv.Itoa(i)
	}))
	if !win.Scoop() && (double || low) {
		t.Logf("  %s", win.LoDesc(func(_, i int) string {
			return strconv.Itoa(i)
		}))
	}
}

const unshuffledSize = 52

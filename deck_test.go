package cardrank

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestNewDeck(t *testing.T) {
	d1, d2 := NewDeck(), NewDeck()
	if !reflect.DeepEqual(d1.v, d2.v) || !reflect.DeepEqual(d1.v, unshuffled) {
		t.Fatalf("expected d1.v == d2.v == unshuffled")
	}
	// shuffle
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d1.Shuffle(r)
	d2.Shuffle(r)
	if reflect.DeepEqual(d1.v, unshuffled) {
		t.Fatalf("expected d1.v != unshuffled")
	}
	if reflect.DeepEqual(d2.v, unshuffled) {
		t.Fatalf("expected d2.v != unshuffled")
	}
	if reflect.DeepEqual(d1.v, d2.v) {
		t.Fatalf("expected d1.v != d2.v")
	}
	if n, exp := len(d1.v), unshuffledSize; n != exp {
		t.Fatalf("expected len(d1.v) == %d, got: %d", exp, n)
	}
	if n, exp := len(d2.v), unshuffledSize; n != exp {
		t.Fatalf("expected len(d2.v) == %d, got: %d", exp, n)
	}
	for i := 0; i < unshuffledSize; i++ {
		if !contains(d1.v, unshuffled[i]) {
			t.Errorf("d1.v does not contain %s", unshuffled[i])
		}
		if !contains(d1.v, d2.v[i]) {
			t.Errorf("d1.v does not contain %s", d2.v[i])
		}
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
	for _, c := range unshuffled {
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
		hand := d.Draw(exp)
		if len(hand) != exp {
			t.Fatalf("expected len(hand) == %d, got: %d", exp, len(hand))
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
	hand := d.Draw(100)
	if n, exp := len(hand), unshuffledSize; n != exp {
		t.Errorf("expected len(hand) == %d, got: %d", exp, n)
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
	hand := d.Draw(unshuffledSize - 1)
	if d.Empty() {
		t.Fatalf("expected d to not be empty")
	}
	if n, exp := len(hand), unshuffledSize-1; n != exp {
		t.Errorf("expected len(hand) == %d, got: %d", exp, n)
	}
	hand = append(hand, d.Draw(1)...)
	if !d.Empty() {
		t.Errorf("expected d to be empty, remaining: %d", d.Remaining())
	}
	if n, exp := len(hand), unshuffledSize; n != exp {
		t.Fatalf("expected len(hand) == %d, got: %d", exp, n)
	}
	for i := 0; i < unshuffledSize; i++ {
		if !contains(hand, unshuffled[i]) {
			t.Errorf("hand does not contain %s", unshuffled[i])
		}
	}
}

func TestUnshuffled(t *testing.T) {
	if n, exp := len(unshuffled), unshuffledSize; n != exp {
		t.Fatalf("expected len(unshuffled) == %d, got: %d", exp, n)
	}
	for _, r := range "23456789TJQKA" {
		for _, s := range "shdc" {
			c := FromString(string(r) + string(s))
			if c == InvalidCard {
				t.Fatalf("expected valid card for %c%c", r, s)
			}
			if !contains(unshuffled, c) {
				t.Errorf("unshuffled does not contain %s", c)
			}
		}
	}
	if n, exp := len(unshuffledShort), unshuffledShortSize; n != exp {
		t.Fatalf("expected len(unshuffledShort) == %d, got: %d", exp, n)
	}
	for _, r := range "6789TJQKA" {
		for _, s := range "shdc" {
			c := FromString(string(r) + string(s))
			if c == InvalidCard {
				t.Fatalf("expected valid card for %c%c", r, s)
			}
			if !contains(unshuffled, c) {
				t.Errorf("unshuffled does not contain %s", c)
			}
		}
	}
	if n, exp := len(unshuffledRoyal), unshuffledRoyalSize; n != exp {
		t.Fatalf("expected len(unshuffledRoyal) == %d, got: %d", exp, n)
	}
	for _, r := range "TJQKA" {
		for _, s := range "shdc" {
			c := FromString(string(r) + string(s))
			if c == InvalidCard {
				t.Fatalf("expected valid card for %c%c", r, s)
			}
			if !contains(unshuffled, c) {
				t.Errorf("unshuffled does not contain %s", c)
			}
		}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var v []Card
	for i, n := 0, 13+r.Intn(26); i < n; i++ {
		if c := New(Rank(r.Intn(13)), Suit(1<<r.Intn(4))); !contains(v, c) {
			v = append(v, c)
		}
	}
	if n, exp := len(v), 2; n < exp {
		t.Fatalf("expected len(v) >= %d, got: %d", exp, n)
	}
	hand := UnshuffledExclude(v)
	for i, exp := range v {
		if contains(hand, exp) {
			t.Errorf("test %d expected hand to not contain %s", i, exp)
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
	t.Logf("Deck (%d):", len(d.d.v))
	for i := 0; i < len(d.d.v); i += 8 {
		t.Logf("  %v", d.d.v[i:min(uint16(i+8), uint16(len(d.d.v)))])
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
		t.Logf("  % 2d: %s %v %v %d", i, h1[i].Description(), h1[i].HiBest, h1[i].HiUnused, h1[i].HiRank)
		switch {
		case double:
			t.Logf("      %s %v %v %d", h2[i].Description(), h2[i].HiBest, h2[i].HiUnused, h2[i].HiRank)
		case low:
			t.Logf("      %s %v %v %d", h1[i].LowDescription(), h1[i].LoBest, h1[i].LoUnused, h1[i].LoRank)
		}
	}
	win := NewWin(h1, h2, low)
	t.Logf("  %s", win.Describe(func(_, i int) string {
		return strconv.Itoa(i)
	}))
	if !win.Scoop() && (double || low) {
		t.Logf("  %s", win.LowDescribe(func(_, i int) string {
			return strconv.Itoa(i)
		}))
	}
}

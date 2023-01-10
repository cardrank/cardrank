package cardrank

import (
	"math/rand"
	"reflect"
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
	if n, exp := len(d1.v), UnshuffledSize; n != exp {
		t.Fatalf("expected len(d1.v) == %d, got: %d", exp, n)
	}
	if n, exp := len(d2.v), UnshuffledSize; n != exp {
		t.Fatalf("expected len(d2.v) == %d, got: %d", exp, n)
	}
	for i := 0; i < UnshuffledSize; i++ {
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
	if n, exp := len(d.v), UnshuffledSize*decks; n != exp {
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
	limit := 5 * UnshuffledSize
	d.SetLimit(limit)
	v := d.Draw(decks * UnshuffledSize)
	if n, exp := len(v), limit; n != exp {
		t.Errorf("expected len(v) == %d, got: %d", exp, n)
	}
	if n, exp := d.Remaining(), 0; n != exp {
		t.Errorf("expected d.Remaining() == %d, got: %d", exp, n)
	}
}

func TestDeckDraw(t *testing.T) {
	for exp := 1; exp < UnshuffledSize; exp++ {
		d := NewDeck()
		hand := d.Draw(exp)
		if len(hand) != exp {
			t.Fatalf("expected len(hand) == %d, got: %d", exp, len(hand))
		}
		if d.Empty() {
			t.Fatalf("expected d to not be empty")
		}
		d.Draw(UnshuffledSize - exp)
		if !d.Empty() {
			t.Errorf("expected d to be empty")
		}
	}
}

func TestDeckDrawAll(t *testing.T) {
	d := NewDeck()
	hand := d.Draw(100)
	if n, exp := len(hand), UnshuffledSize; n != exp {
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
	hand := d.Draw(UnshuffledSize - 1)
	if d.Empty() {
		t.Fatalf("expected d to not be empty")
	}
	if n, exp := len(hand), UnshuffledSize-1; n != exp {
		t.Errorf("expected len(hand) == %d, got: %d", exp, n)
	}
	hand = append(hand, d.Draw(1)...)
	if !d.Empty() {
		t.Errorf("expected d to be empty, remaining: %d", d.Remaining())
	}
	if n, exp := len(hand), UnshuffledSize; n != exp {
		t.Fatalf("expected len(hand) == %d, got: %d", exp, n)
	}
	for i := 0; i < UnshuffledSize; i++ {
		if !contains(hand, unshuffled[i]) {
			t.Errorf("hand does not contain %s", unshuffled[i])
		}
	}
}

func TestUnshuffled(t *testing.T) {
	if n, exp := len(unshuffled), UnshuffledSize; n != exp {
		t.Fatalf("expected len(unshuffled) == %d, got: %d", exp, n)
	}
	for _, r := range "23456789TJQKA" {
		for _, s := range "shdc" {
			c, err := FromString(string(r) + string(s))
			if err != nil {
				t.Fatalf("expected no error for %c%c, got: %v", r, s, err)
			}
			if !contains(unshuffled, c) {
				t.Errorf("unshuffled does not contain %s", c)
			}
		}
	}
	if n, exp := len(unshuffledShort), UnshuffledShortSize; n != exp {
		t.Fatalf("expected len(unshuffledShort) == %d, got: %d", exp, n)
	}
	for _, r := range "6789TJQKA" {
		for _, s := range "shdc" {
			c, err := FromString(string(r) + string(s))
			if err != nil {
				t.Fatalf("expected no error for %c%c, got: %v", r, s, err)
			}
			if !contains(unshuffled, c) {
				t.Errorf("unshuffled does not contain %s", c)
			}
		}
	}
	if n, exp := len(unshuffledRoyal), UnshuffledRoyalSize; n != exp {
		t.Fatalf("expected len(unshuffledRoyal) == %d, got: %d", exp, n)
	}
	for _, r := range "TJQKA" {
		for _, s := range "shdc" {
			c, err := FromString(string(r) + string(s))
			if err != nil {
				t.Fatalf("expected no error for %c%c, got: %v", r, s, err)
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

func TestDeckDeal(t *testing.T) {
	for _, typ := range []Type{Holdem, Short, Royal, Omaha, OmahaHiLo, Stud, StudHiLo, Razz, Badugi} {
		for i := 2; i <= typ.MaxPlayers(); i++ {
			checkDeal(t, i, typ, time.Now().UnixNano())
		}
	}
}

func checkDeal(t *testing.T, i int, typ Type, seed int64) {
	r := rand.New(rand.NewSource(seed))
	d := typ.Deck()
	d.Shuffle(r)
	var f func(int) ([][]Card, []Card)
	var p, b int
	switch typ {
	case Holdem, Short, Royal:
		f, p, b = d.Holdem, 2, 5
	case Omaha, OmahaHiLo:
		f, p, b = d.Omaha, 4, 5
	case Stud, StudHiLo, Razz:
		f, p, b = d.Stud, 7, 0
	case Badugi:
		f, p, b = d.Badugi, 4, 0
	default:
		t.Fatalf("unknown type %q", typ)
	}
	pockets, board := f(i)
	if n, exp := len(board), b; n != exp {
		t.Fatalf("expected %d board cards, got: %d", exp, n)
	}
	m := make(map[Card]bool)
	for _, c := range board {
		if _, ok := m[c]; ok {
			t.Errorf("board card %b already encountered!", c)
		}
		m[c] = true
	}
	if n, exp := len(pockets), i; n != exp {
		t.Fatalf("expected %d pockets, got: %d", exp, n)
	}
	for j := 0; j < i; j++ {
		if n, exp := len(pockets[j]), p; n != exp {
			t.Errorf("pocket %d expected %d cards, got: %d", j, exp, n)
		}
		for k := 0; k < p; k++ {
			if _, ok := m[pockets[j][k]]; ok {
				t.Errorf("pocket card %b already encountered!", pockets[j][k])
			}
			m[pockets[j][k]] = true
		}
	}
}

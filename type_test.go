package cardrank

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestMax(t *testing.T) {
	rnd := rand.New(rand.NewSource(0))
	for _, v := range Types() {
		typ := v
		for n := 2; n <= typ.Max(); n++ {
			count := n
			t.Run(fmt.Sprintf("%s/%d", typ, count), func(t *testing.T) {
				pockets, _ := typ.Deal(rnd, 1, count)
				if l := len(pockets); l != count {
					t.Fatalf("expected %d, got: %d", count, l)
				}
				exp := typ.Pocket()
				for i := 0; i < count; i++ {
					if l := len(pockets[i]); l != exp {
						t.Errorf("expected %d, got: %d", exp, l)
					}
					t.Logf("%d: %v", i, pockets[i])
				}
			})
		}
	}
}

func TestShort(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
		s string
	}{
		{"9d 8d 7d 6d Ad Tc Jc", "9d 8d 7d 6d Ad", "Jc Tc", StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ A♦]"},
		{"9c 8c 7c 6c Ac Th Qh", "9c 8c 7c 6c Ac", "Qh Th", StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣]"},
		{"9h 8h 7h 6h Ah Tc Kd", "9h 8h 7h 6h Ah", "Kd Tc", StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♥ 8♥ 7♥ 6♥ A♥]"},
		{"9s 8s 7s 6s As 9c 8c", "9s 8s 7s 6s As", "9c 8c", StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♠ 8♠ 7♠ 6♠ A♠]"},
		{"9d 8d 7d 6d Ac 8h 8s", "9d 8d 7d 6d Ac", "8h 8s", Straight, "Straight, Nine-high [9♦ 8♦ 7♦ 6♦ A♣]"},
		{"9c 8c 7c 6c Ad Kh Jh", "9c 8c 7c 6c Ad", "Kh Jh", Straight, "Straight, Nine-high [9♣ 8♣ 7♣ 6♣ A♦]"},
		{"9h 8h 7h 6h As Kd Jd", "9h 8h 7h 6h As", "Kd Jd", Straight, "Straight, Nine-high [9♥ 8♥ 7♥ 6♥ A♠]"},
		{"9s 8s 7s 6s Ah Qh Jd", "9s 8s 7s 6s Ah", "Qh Jd", Straight, "Straight, Nine-high [9♠ 8♠ 7♠ 6♠ A♥]"},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Short.New(pocket, nil)
		if r, exp := ev.HiRank.Fixed(), test.r; r != exp {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, exp, r)
		}
		if !equals(ev.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !equals(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf("%s %b", desc, desc.Best), test.s; s != exp {
			t.Errorf("test %d expected description %q, got: %q", i, exp, s)
		}
	}
}

func TestRazz(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
	}{
		{"Kh Qh Jh Th 9h Ks Qs", "Kh Qh Jh Th 9h", "Ks Qs", 7936},
		{"Ah Kh Qh Jh Th Ks Qs", "Kh Qh Jh Th Ah", "Ks Qs", 7681},
		{"2h 2c 2d 2s As Ks Qs", "2h 2c As Ks Qs", "2d 2s", 59569},
		{"Ah Ac Ad Ks Kh Ks Qs", "Ah Ac Ks Kh Qs", "Ad Ks", 63067},
		{"Ah Ac Ad Ks Qh Ks Qs", "Ks Ks Qh Qs Ah", "Ac Ad", 62935},
		{"Kh Kd Qd Qs Jh Ks Js", "Qd Qs Jh Js Kh", "Kd Ks", 62813},
		{"3h 3c Kh Qd Jd Ks Qs", "3h 3c Kh Qd Jd", "Ks Qs", 59734},
		{"2h 2c Kh Qd Jd Ks Qs", "2h 2c Kh Qd Jd", "Ks Qs", 59514},
		{"3h 2c Kh Qd Jd Ks Qs", "Kh Qd Jd 3h 2c", "Ks Qs", 7174},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Razz.New(pocket, nil)
		if ev.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.r, ev.HiRank)
		}
		if !equals(ev.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !equals(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
	}
}

func TestBadugi(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
	}{
		{"Kh Qh Jh Th", "Th", "Kh Qh Jh", 25088},
		{"Kh Qh Jd Th", "Jd Th", "Kh Qh", 17920},
		{"Kh Qc Jd Th", "Qc Jd Th", "Kh", 11776},
		{"Ks Qc Jd Th", "Ks Qc Jd Th", "", 7680},
		{"2h 2c 2d 2s", "2s", "2h 2d 2c", 24578},
		{"Ah Kh Qh Jh", "Ah", "Kh Qh Jh", 24577},
		{"Kh Kd Qd Qs", "Kh Qs", "Kd Qd", 22528},
		{"Ah Ac Ad Ks", "Ks Ah", "Ad Ac", 20481},
		{"3h 3c Kh Qd", "Kh Qd 3c", "3h", 14340},
		{"2h 2c Kh Qd", "Kh Qd 2c", "2h", 14338},
		{"3h 2c Kh Ks", "Ks 3h 2c", "Kh", 12294},
		{"3h 2c Kh Qd", "Qd 3h 2c", "Kh", 10246},
		{"Ah 2c 4s 6d", "6d 4s 2c Ah", "", 43},
		{"Ac 2h 4d 6s", "6s 4d 2h Ac", "", 43},
		{"Ah 2c 3s 6d", "6d 3s 2c Ah", "", 39},
		{"Ah 2c 4s 5d", "5d 4s 2c Ah", "", 27},
		{"Ah 2c 3s 5d", "5d 3s 2c Ah", "", 23},
		{"Ah 2c 3s 4d", "4d 3s 2c Ah", "", 15},
		{"Ac 2h 3s 4d", "4d 3s 2h Ac", "", 15},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Badugi.New(pocket, nil)
		if ev.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.r, ev.HiRank)
		}
		if !equals(ev.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !equals(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
	}
}

func TestLowball(t *testing.T) {
	tests := []struct {
		v string
		r EvalRank
	}{
		{"7h 5h 4h 3h 2c", 1},
		{"7h 6h 4h 3h 2c", 2},
		{"7h 6h 5h 3h 2c", 3},
		{"7h 6h 5h 4h 2c", 4},
		{"8h 5h 4h 3h 2c", 5},
		{"8h 6h 4h 3h 2c", 6},
		{"8h 6h 5h 3h 2c", 7},
		{"8h 6h 5h 4h 2c", 8},
		{"8h 6h 5h 4h 3c", 9},
		{"8h 7h 4h 3h 2c", 10},
		{"8h 7h 5h 3h 2c", 11},
		{"8h 7h 5h 4h 2c", 12},
		{"8h 7h 5h 4h 3c", 13},
		{"8h 7h 6h 3h 2c", 14},
		{"8h 7h 6h 4h 2c", 15},
		{"8h 7h 6h 4h 3c", 16},
		{"8h 7h 6h 5h 2c", 17},
		{"8h 7h 6h 5h 3c", 18},
		{"9h 5h 4h 3h 2c", 19},
	}
	for i, test := range tests {
		pocket := Must(test.v)
		ev := Lowball.New(pocket, nil)
		if ev.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.r, ev.HiRank)
		}
	}
}

func TestTypeComp(t *testing.T) {
	tests := []struct {
		typ   Type
		board string
		a     string
		b     string
		j     EvalRank
		k     EvalRank
		exp   int
	}{
		{Short, "As 7d Ad 6s 6d", "8d Td", "Ac Qh", Flush, FullHouse, -1},
		{Short, "As 7d Ad 6s 6d", "Ac Qh", "8d Td", FullHouse, Flush, +1},
		{Short, "Kc Qh Jc Td 8d", "Ac Qh", "Ah 6c", Straight, Straight, 0},
		{Short, "Kc Qh Jc Td 8d", "Ah 6c", "Ac Qh", Straight, Straight, 0},
		{Short, "9c 7d 8d As Qs", "Ac 6s", "Tc Ts", Straight, Pair, -1},
		{Short, "9c 7d 8d As Qs", "Tc Ts", "Ac 6s", Pair, Straight, +1},
		{Short, "9s 7s 8s Ac Qs", "As 6s", "Tc Ts", StraightFlush, Flush, -1},
		{Short, "9s 7s 8s Ac Qs", "Tc Ts", "As 6s", Flush, StraightFlush, +1},
		{Manila, "As 8d Ad 7s 7d", "9d Jd", "Ac Qh", Flush, FullHouse, -1},
		{Manila, "As 8d Ad 7s 7d", "Ac Qh", "9d Jd", FullHouse, Flush, +1},
		{Manila, "Kc Qh Jc Td 9d", "Ac Qh", "Ah 7c", Straight, Straight, 0},
		{Manila, "Kc Qh Jc Td 9d", "Ah 7c", "Ac Qh", Straight, Straight, 0},
		{Manila, "10c 8d 9d As Qs", "Ac 7s", "Kc Ks", Straight, Pair, -1},
		{Manila, "10c 8d 9d As Qs", "Kc Ks", "Ac 7s", Pair, Straight, +1},
		{Manila, "10s 8s 9s Ac Qs", "As 7s", "Kc Ks", StraightFlush, Flush, -1},
		{Manila, "10s 8s 9s Ac Qs", "Kc Ks", "As 7s", Flush, StraightFlush, +1},
		{Omaha, "Td 2c Jd 4c 5c", "As Ah Qh 3s", "Ad Ac 7d 4d", Straight, Pair, -1},
		{Omaha, "Td 2c Jd 4c 5c", "Ad Ac 7d 4d", "As Ah Qh 3s", Pair, Straight, +1},
		{Omaha, "Kc Qh Jc 8d 4s", "Ac Td 3h 6c", "Ah Tc 2c 3c", Straight, Straight, 0},
		{Omaha, "Kc Qh Jc 8d 4s", "Ah Tc 2c 3c", "Ac Td 3h 6c", Straight, Straight, 0},
		{Omaha, "2d 3h 8s 8h 2s", "Kd Ts Td 4h", "Jd 7d 7c 4c", TwoPair, TwoPair, -1},
		{Omaha, "2d 3h 8s 8h 2s", "Jd 7d 7c 4c", "Kd Ts Td 4h", TwoPair, TwoPair, +1},
		{Omaha, "Tc 6c 2s 3s As", "Kd Qs Js 8h", "9h 9d 4h 4d", Flush, Pair, -1},
		{Omaha, "Tc 6c 2s 3s As", "9h 9d 4h 4d", "Kd Qs Js 8h", Pair, Flush, +1},
		{Omaha, "4s 3h 6c 2d Kd", "Kh Qs 5h 2c", "7s 7c 4h 2s", Straight, TwoPair, -1},
		{Omaha, "4s 3h 6c 2d Kd", "7s 7c 4h 2s", "Kh Qs 5h 2c", TwoPair, Straight, +1},
	}
	for i, test := range tests {
		board := Must(test.board)
		a, b := test.typ.New(Must(test.a), board), test.typ.New(Must(test.b), board)
		if r, exp := a.HiRank.Fixed(), test.j; r != exp {
			t.Errorf("test %d %s expected %d, got: %d", i, test.typ, exp, r)
		}
		if r, exp := b.HiRank.Fixed(), test.k; r != exp {
			t.Errorf("test %d %s expected %d, got: %d", i, test.typ, exp, r)
		}
		if n := a.Comp(b, false); n != test.exp {
			t.Errorf("test %d %s compare expected %d, got: %d", i, test.typ, test.exp, n)
		}
	}
}

func TestNumberedStreets(t *testing.T) {
	exp := []string{
		"Ante", "1st", "2nd", "3rd", "4th", "5th",
		"6th", "7th", "8th", "9th", "10th", "11th",
		"101st", "102nd", "River",
	}
	streets := NumberedStreets(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 90, 1, 1)
	for i := 0; i < len(streets); i++ {
		if s, exp := streets[i].Name, exp[i]; s != exp {
			t.Errorf("expected %q, got: %v", exp, s)
		}
	}
}

func TestTypeUnmarshal(t *testing.T) {
	tests := []struct {
		s   string
		exp Type
	}{
		{"HOLDEM", Holdem},
		{"Hh", Holdem},
		{"omaha", Omaha},
		{"studHiLo", StudHiLo},
		{"razz", Razz},
		{"BaDUGI", Badugi},
		{"fusIon", Fusion},
	}
	for i, test := range tests {
		var typ Type
		if err := typ.UnmarshalText([]byte(test.s)); err != nil {
			t.Fatalf("test %d expected no error, got: %v", i, err)
		}
		if typ != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, typ)
		}
	}
}

package cardrank

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestOrderHands(t *testing.T) {
	tests := []struct {
		r   int64
		n   int
		p   int
		exp []int
	}{
		{52, 2, 1, []int{0, 1}},
		{72, 3, 1, []int{2, 0, 1}},
		{99, 3, 1, []int{0, 2, 1}},
		{583, 6, 1, []int{1, 4, 0, 2, 5, 3}},
		{660, 2, 1, []int{1, 0}},
		{791, 6, 2, []int{1, 5, 2, 3, 4, 0}},
		{1109, 6, 3, []int{1, 2, 3, 5, 4, 0}},
		{1173, 6, 4, []int{0, 1, 2, 5, 4, 3}},
		{3521, 2, 1, []int{0, 1}},
		{5162, 6, 6, []int{0, 1, 2, 3, 4, 5}},
		{26076, 4, 1, []int{0, 3, 2, 1}},
		{56867, 2, 1, []int{0, 1}},
		{91981, 6, 6, []int{0, 1, 2, 3, 4, 5}},
	}
	for n, tt := range tests {
		i, test := n, tt
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := NewDeck()
			// note: use a real random source
			rnd := rand.New(rand.NewSource(test.r))
			d.Shuffle(rnd)
			board := d.Draw(5)
			t.Logf("board: %b", board)
			var hands []*Hand
			for i := 0; i < test.n; i++ {
				h := Holdem.RankHand(d.Draw(2), board)
				t.Logf("player %d: %b", i, h.Pocket)
				t.Logf("  hand: %b %b", h.Best(), h.Unused())
				t.Logf("  desc: %s", h.Description())
				hands = append(hands, h)
			}
			v, pivot := Order(hands)
			if pivot != test.p {
				t.Errorf("test %d expected pivot %d, got: %d", i, test.p, pivot)
			}
			for j := len(v); j > 0; j-- {
				typ := "shows "
				switch {
				case j <= pivot && pivot != 1:
					typ = "pushes"
				case j <= pivot:
					typ = "wins  "
				}
				h := hands[v[j-1]]
				t.Logf("player %d %s %b %b %s", v[j-1], typ, h.Best(), h.Unused(), h.Description())
			}
			if !reflect.DeepEqual(v, test.exp) {
				t.Errorf("test %d expected %v, got: %v", i, test.exp, v)
			}
		})
	}
}

func TestHandRankString(t *testing.T) {
	tests := []struct {
		r   HandRank
		exp string
	}{
		{0x1c35, "Nothing"},
		{0x1c1c, "Nothing"},
		{0x1aa7, "Nothing"},
		{0x1981, "Nothing"},
		{0x186c, "Nothing"},
		{0x1856, "Nothing"},
		{0x1854, "Nothing"},
		{0x0fec, "Pair"},
		{0x0d78, "Pair"},
		{0x0a6d, "Two Pair"},
		{0x0a69, "Two Pair"},
		{0x09c1, "Two Pair"},
		{0x0664, "Three of a Kind"},
		{0x0640, "Straight"},
		{0x0606, "Flush"},
		{0x018e, "Flush"},
		{0x012a, "Full House"},
		{0x0013, "Four of a Kind"},
		{0x0001, "Straight Flush"},
	}
	for i, test := range tests {
		if s := test.r.String(); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
	}
}

func TestHandCompare(t *testing.T) {
	tests := []struct {
		a   string
		b   string
		exp int
		r   HandRank
	}{
		{"As Ks Jc 7h 5d 2d 3c", "As Ks Jc 7h 5d 2d 3c", +0, Nothing},
		{"As Ks Jc 7h 4d 2d 3c", "As Ks Jc 7h 5d 2d 3c", +1, Nothing},
		{"As Ks Jc 7h 5d 2d 3c", "As Ks Jc 7h 4d 2d 3c", -1, Nothing},
		{"As Ac Ad Ah Kd 2d 3c", "As Ac Ad Ah Qd 2d 3c", -1, FourOfAKind},
		{"As Ac Ad Ah Qd 2d 3c", "As Ac Ad Ah Kd 2d 3c", +1, FourOfAKind},
		{"As Ks Qs Ts 9s 2s 3s", "Ks Qs Js Ts 9s 2d 3c", +1, StraightFlush},
		{"6s 6c 6d 5d 5c 4s 4s", "5s 5c 5d 6d 6c 4s 4s", -1, FullHouse},
		{"Ks Qs Js Ts 9s 2s 3s", "Kd Qd Jd Td 9d 2d 3d", +0, StraightFlush},
		{"Ks Qs Js 9s 3s Ad Kd", "Kd Qd Jd 9d 2d Ac Kc", -1, Flush},
		{"Kd Qd Jd 9d 2d Ac Kc", "Ks Qs Js 9s 3s Ad Kd", +1, Flush},
	}
	for i, test := range tests {
		a := Holdem.RankHand(Must(test.a), nil)
		b := Holdem.RankHand(Must(test.b), nil)
		switch r := a.Compare(b); {
		case r != test.exp:
			t.Errorf("test %d expected r == %d, got: %d", i, test.exp, r)
		case r == +0:
			if a.Fixed() != b.Fixed() && a.Fixed() != test.r {
				t.Errorf("test %d expected a == b == r", i)
			}
		case r == -1:
			if r := a.Fixed(); r != test.r {
				t.Errorf("test %d expected a to be %s, got: %s", i, test.r, r)
			}
		case r == +1:
			if r := b.Fixed(); r != test.r {
				t.Errorf("test %d expected b to be %s, got: %s", i, test.r, r)
			}
		}
	}
}

func TestShortDeckRankHand(t *testing.T) {
	tests := []struct {
		h   string
		typ Type
		r   HandRank
		s   string
	}{
		{"9d 8d 7d 6d Ad", Short, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ A♦]"},
		{"9c 8c 7c 6c Ac", Short, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣]"},
		{"9h 8h 7h 6h Ah", Short, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♥ 8♥ 7♥ 6♥ A♥]"},
		{"9s 8s 7s 6s As", Short, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♠ 8♠ 7♠ 6♠ A♠]"},
		{"9d 8d 7d 6d Ac", Short, Straight, "Straight, Nine-high [9♦ 8♦ 7♦ 6♦ A♣]"},
		{"9c 8c 7c 6c Ad", Short, Straight, "Straight, Nine-high [9♣ 8♣ 7♣ 6♣ A♦]"},
		{"9h 8h 7h 6h As", Short, Straight, "Straight, Nine-high [9♥ 8♥ 7♥ 6♥ A♠]"},
		{"9s 8s 7s 6s Ah", Short, Straight, "Straight, Nine-high [9♠ 8♠ 7♠ 6♠ A♥]"},
	}
	for i, test := range tests {
		hand := test.typ.RankHand(Must(test.h), nil)
		if r, exp := hand.Rank().Fixed(), test.r; r != exp {
			t.Errorf("test %d expected rank %s, got: %s -- %d", i, exp, r, hand.Rank())
		}
		if s, exp := fmt.Sprintf("%s %b", hand.Description(), hand.Best()), test.s; s != exp {
			t.Errorf("test %d expected description %q, got: %q", i, exp, s)
		}
	}
}

func TestHandRankCompare(t *testing.T) {
	tests := []struct {
		typ   Type
		board string
		a     string
		b     string
		j     HandRank
		k     HandRank
		exp   int
	}{
		{Short, "As 7d Ad 6s 6d", "8d Td", "Ac 5h", Flush, FullHouse, -1},
		{Short, "As 7d Ad 6s 6d", "Ac 5h", "8d Td", FullHouse, Flush, +1},
		{Short, "Kc Qh Jc Td 8d", "Ac 5h", "Ah 6c", Straight, Straight, 0},
		{Short, "Kc Qh Jc Td 8d", "Ah 6c", "Ac 5h", Straight, Straight, 0},
		{Short, "9c 7d 8d As Qs", "Ac 6s", "Tc Ts", Straight, Pair, -1},
		{Short, "9c 7d 8d As Qs", "Tc Ts", "Ac 6s", Pair, Straight, +1},
		{Short, "9s 7s 8s Ac Qs", "As 6s", "Tc Ts", StraightFlush, Flush, -1},
		{Short, "9s 7s 8s Ac Qs", "Tc Ts", "As 6s", Flush, StraightFlush, +1},
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
		a := test.typ.RankHand(Must(test.a), board)
		b := test.typ.RankHand(Must(test.b), board)
		if r := a.Fixed(); r != test.j {
			t.Errorf("test %d %s expected rank %s, got: %s", i, test.typ, test.j, r)
		}
		if r := b.Fixed(); r != test.k {
			t.Errorf("test %d %s expected rank %s, got: %s", i, test.typ, test.k, r)
		}
		if n := a.Compare(b); n != test.exp {
			t.Errorf("test %d %s compare expected %d, got: %d", i, test.typ, test.exp, n)
		}
	}
}

func TestMax(t *testing.T) {
	rnd := rand.New(rand.NewSource(0))
	for typ := Holdem; typ <= Badugi; typ++ {
		max := typ.Max()
		for i := 2; i <= max; i++ {
			pockets, _ := typ.Deal(rnd, i)
			if len(pockets) != i {
				t.Errorf("%s was not able to deal pockets for %d players", typ, i)
			}
		}
	}
}

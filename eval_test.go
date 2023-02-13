package cardrank

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHiOrder(t *testing.T) {
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
			d.Shuffle(rand.New(rand.NewSource(test.r)), 1)
			board := d.Draw(5)
			t.Logf("board: %b", board)
			var evals []*Eval
			for i := 0; i < test.n; i++ {
				pocket := d.Draw(2)
				ev := Holdem.New(pocket, board)
				t.Logf("player %d: %b", i, pocket)
				t.Logf("  best: %b", ev.HiBest)
				t.Logf("  unused: %b", ev.HiUnused)
				t.Logf("  desc: %s", ev.HiDesc())
				evals = append(evals, ev)
			}
			v, pivot := HiOrder(evals)
			if pivot != test.p {
				t.Errorf("test %d expected pivot %d, got: %d", i, test.p, pivot)
			}
			for j := len(v); 0 < j; j-- {
				typ := "shows "
				switch {
				case pivot != 1 && j <= pivot:
					typ = "pushes"
				case j <= pivot:
					typ = "wins  "
				}
				ev := evals[v[j-1]]
				t.Logf("player %d %s %b %b %s", v[j-1], typ, ev.HiBest, ev.HiUnused, ev.HiDesc())
			}
			if !equals(v, test.exp) {
				t.Errorf("test %d expected %v, got: %v", i, test.exp, v)
			}
		})
	}
}

func TestEvalHiComp(t *testing.T) {
	tests := []struct {
		a   string
		b   string
		exp int
		r   EvalRank
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
		h1 := Holdem.New(Must(test.a), nil)
		h2 := Holdem.New(Must(test.b), nil)
		switch r := h1.HiComp(h2); {
		case r != test.exp:
			t.Errorf("test %d expected r == %d, got: %d", i, test.exp, r)
		case r == +0:
			if h1.HiRank.Fixed() != h2.HiRank.Fixed() && h1.HiRank.Fixed() != test.r {
				t.Errorf("test %d expected a == b == r", i)
			}
		case r == -1:
			if r := h1.HiRank.Fixed(); r != test.r {
				t.Errorf("test %d expected a to be %s, got: %s", i, test.r, r)
			}
		case r == +1:
			if r := h2.HiRank.Fixed(); r != test.r {
				t.Errorf("test %d expected b to be %s, got: %s", i, test.r, r)
			}
		}
	}
}

func TestEval(t *testing.T) {
	for _, rr := range cactusTests(true) {
		for i, f := range []func() []cardTest{
			fiveCardTests,
			sixCardTests,
			sevenCardTests,
		} {
			r, tests := rr, f()
			t.Run(fmt.Sprintf("%s/%d", r.name, i+5), func(t *testing.T) {
				for j, test := range tests {
					v := Must(test.v)
					rank := r.eval(v)
					if rank != test.r {
						t.Errorf("test %d %d expected %d, got: %d", i, j, test.r, rank)
					}
					if fixed := rank.Fixed(); fixed != test.exp {
						t.Errorf("test %d %d expected %s, got: %s", i, j, test.exp, fixed)
					}
					ev := EvalOf(Holdem).Eval(v[:5], v[5:])
					if s := fmt.Sprintf("%b %b", ev, ev.HiUnused); s != test.desc {
						t.Errorf("test %d %d expected %q, got: %q", i, j, test.desc, s)
					}
				}
			})
		}
	}
}

func TestEvalRank(t *testing.T) {
	tests := []struct {
		v string
		r EvalRank
		f RankFunc
	}{
		{"Kh Qh Jh Th 9h", 7936, RankRazz},
		{"9h 7h 6h 5h 4h", 33144, RankEightOrBetter},
	}
	for i, test := range tests {
		f := NewRankFunc(test.f)
		if e, exp := f(Must(test.v)), test.r; e != exp {
			t.Errorf("test %d expected rank %d, got: %d", i, exp, e)
		}
	}
}

func TestRankEightOrBetter(t *testing.T) {
	p0 := Must("Ah 2h 3h 4h 5h 6h 7h 8h")
	for i := Nine; i <= King; i++ {
		p1 := Must(i.String() + "h 4h 3h 2h Ah")
		r1 := RankEightOrBetter(p1[0], p1[1], p1[2], p1[3], p1[4])
		for c0 := 0; c0 < len(p0); c0++ {
			for c1 := c0 + 1; c1 < len(p0); c1++ {
				for c2 := c1 + 1; c2 < len(p0); c2++ {
					for c3 := c2 + 1; c3 < len(p0); c3++ {
						for c4 := c3 + 1; c4 < len(p0); c4++ {
							h0 := []Card{p0[c0], p0[c1], p0[c2], p0[c3], p0[c4]}
							r0 := RankEightOrBetter(h0[0], h0[1], h0[2], h0[3], h0[4])
							if r1 <= r0 {
								t.Errorf("%s does not have lower rank than %s", h0, p1)
							}
						}
					}
				}
			}
		}
	}
}

func TestCactus(t *testing.T) {
	if !strings.Contains(os.Getenv("TESTS"), "cactus") {
		t.Skipf("skipping: $ENV{TESTS} does not contain 'cactus'")
	}
	if cactus == nil {
		t.Skipf("skipping: cactus is not available")
	}
	v, f, tests := shuffled(DeckFrench), NewRankFunc(cactus), cactusTests(false)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								vv := []Card{v[c0], v[c1], v[c2], v[c3], v[c4], v[c5], v[c6]}
								exp := f(vv)
								for _, test := range tests {
									if r := test.eval(vv); r != exp {
										t.Errorf("test %s(%b) expected %d (%s), got: %d (%s)", test.name, vv, exp, exp.Fixed(), r, r.Fixed())
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestEvalRankString(t *testing.T) {
	tests := []struct {
		r   EvalRank
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

func shuffled(typ DeckType) []Card {
	v := typ.Unshuffled()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(v), func(i, j int) {
		v[i], v[j] = v[j], v[i]
	})
	return v
}

type cardTest struct {
	v    string
	r    EvalRank
	exp  EvalRank
	desc string
}

func fiveCardTests() []cardTest {
	return []cardTest{
		{"As Ks Jc 7h 5d", 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] []"},
		{"As Ac Jc 7h 5d", 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] []"},
		{"Jd 6s 6c 5c 5d", 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] []"},
		{"6s 6c Jc Jd 5d", 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] []"},
		{"As Ac Jc Jd 5d", 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] []"},
		{"As Ac Ad Jd 5d", 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] []"},
		{"4s 5s 2d 3h Ac", 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] []"},
		{"9s Ks Qd Jh Td", 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] []"},
		{"As Ks Qd Jh Td", 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] []"},
		{"Ts 7s 4s 3s 2s", 0x0606, Flush, "Flush, Ten-high, kickers Seven, Four, Three, Two [T♠ 7♠ 4♠ 3♠ 2♠] []"},
		{"4s 4c 4d 2s 2h", 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♥ 2♠] []"},
		{"5s 5c 5d 6s 6h", 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] []"},
		{"6s 6c 6d 5s 5h", 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] []"},
		{"As Ac Ad Ah 5h", 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] []"},
		{"3d 5d 2d 4d Ad", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] []"},
		{"6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] []"},
		{"9♦ 6♦ 8♦ 5♦ 7♦", 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] []"},
		{"As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] []"},
	}
}

func sixCardTests() []cardTest {
	return []cardTest{
		{"3d As Ks Jc 7h 5d", 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦]"},
		{"3d As Ac Jc 7h 5d", 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] [3♦]"},
		{"9d Jd 6s 6c 5c 5d", 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] [9♦]"},
		{"3d 6s 6c Jc Jd 5d", 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] [3♦]"},
		{"3d As Ac Jc Jd 5d", 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] [3♦]"},
		{"3d As Ac Ad Jd 5d", 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] [3♦]"},
		{"4s 5s 2d 3h Ac Jd", 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] [J♦]"},
		{"3d 9s Ks Qd Jh Td", 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] [3♦]"},
		{"3d As Ks Qd Jh Td", 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] [3♦]"},
		{"3d Ts 7s 4s 3s 2s", 0x0606, Flush, "Flush, Ten-high, kickers Seven, Four, Three, Two [T♠ 7♠ 4♠ 3♠ 2♠] [3♦]"},
		{"3d 4s 4c 4d 2s 2h", 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♥ 2♠] [3♦]"},
		{"3d 5s 5c 5d 6s 6h", 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] [3♦]"},
		{"3d 6s 6c 6d 5s 5h", 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] [3♦]"},
		{"3d As Ac Ad Ah 5h", 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] [3♦]"},
		{"3d 5d 2d 4d Ad 3s", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [3♠]"},
		{"T♦ 6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] [T♦]"},
		{"J♦ 9♦ 6♦ 8♦ 5♦ 7♦", 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{"7♦ J♦ 9♦ 6♦ 8♦ 5♦", 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{"3d As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦]"},
	}
}

func sevenCardTests() []cardTest {
	return []cardTest{
		{"2d 3d As Ks Jc 7h 5d", 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{"2d 3d As Ac Jc 7h 5d", 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{"9d Jd 6s 6c 5c 5d 4d", 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] [9♦ 4♦]"},
		{"2d 3d 6s 6c Jc Jd 5d", 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] [3♦ 2♦]"},
		{"2d 3d As Ac Jc Jd 5d", 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] [3♦ 2♦]"},
		{"2c 3d As Ac Ad Jd 5d", 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] [3♦ 2♣]"},
		{"4s 5s 2d 3h Ac Jd Qs", 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] [Q♠ J♦]"},
		{"2d 3d 9s Ks Qd Jh Td", 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] [3♦ 2♦]"},
		{"2d 3d As Ks Qd Jh Td", 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] [3♦ 2♦]"},
		{"2d 3d Ts 7s 4s 3s 2s", 0x0606, Flush, "Flush, Ten-high, kickers Seven, Four, Three, Two [T♠ 7♠ 4♠ 3♠ 2♠] [3♦ 2♦]"},
		{"2d 3d 4s 4c 4d 2s 2h", 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♦ 2♥] [2♠ 3♦]"},
		{"4d 3d 5s 5c 5d 6s 6h", 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] [4♦ 3♦]"},
		{"4d 3d 6s 6c 6d 5s 5h", 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] [4♦ 3♦]"},
		{"2d 3d As Ac Ad Ah 5h", 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] [3♦ 2♦]"},
		{"3d 5d 2d 4d Ad 3s 4s", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [4♠ 3♠]"},
		{"J♦ T♦ 6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] [J♦ T♦]"},
		{"7♦ J♦ 9♦ 6♦ 8♦ 5♦ 2♦", 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦ 2♦]"},
		{"2d 3d As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦ 2♦]"},
	}
}

type cactusTest struct {
	name string
	eval EvalRankFunc
}

func cactusTests(base bool) []cactusTest {
	var tests []cactusTest
	if base && cactus != nil {
		tests = append(tests, cactusTest{"Cactus", NewRankFunc(cactus)})
	}
	if cactusFast != nil {
		tests = append(tests, cactusTest{"CactusFast", NewRankFunc(cactusFast)})
	}
	if twoPlusTwo != nil {
		tests = append(tests, cactusTest{"TwoPlusTwo", twoPlusTwo})
	}
	if cactusFast != nil && twoPlusTwo != nil {
		tests = append(tests, cactusTest{"Hybrid", NewHybrid(cactusFast, twoPlusTwo)})
	}
	return tests
}

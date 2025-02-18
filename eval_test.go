package cardrank

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOrder(t *testing.T) {
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
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := NewDeck()
			// note: use a real random source
			d.Shuffle(rand.New(rand.NewSource(test.r)), 1)
			board := d.Draw(5)
			t.Logf("board: %b", board)
			var evals []*Eval
			for i := range test.n {
				pocket := d.Draw(2)
				ev := Holdem.Eval(pocket, board)
				desc := ev.Desc(false)
				t.Logf("player %d: %b", i, pocket)
				t.Logf("  best: %b", desc.Best)
				t.Logf("  unused: %b", desc.Unused)
				t.Logf("  desc: %s", desc)
				evals = append(evals, ev)
			}
			v, pivot := Order(evals, false)
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
				t.Logf("player %d %s %b %b %s", v[j-1], typ, ev.HiBest, ev.HiUnused, ev.Desc(false))
			}
			if !slices.Equal(v, test.exp) {
				t.Errorf("test %d expected %v, got: %v", i, test.exp, v)
			}
		})
	}
}

func TestEvalComp(t *testing.T) {
	tests := []struct {
		typ Type
		a   string
		b   string
		exp int
		r   EvalRank
		s   string
	}{
		{Holdem, "As Ks Jc 7h 5d 2d 3c", "As Ks Jc 7h 5d 2d 3c", +0, 6252, "Ace-high, kickers King, Jack, Seven, Five [As Ks Jc 7h 5d]"},
		{Holdem, "As Ks Jc 7h 4d 2d 3c", "As Ks Jc 7h 5d 2d 3c", +1, 6252, "Ace-high, kickers King, Jack, Seven, Five [As Ks Jc 7h 5d]"},
		{Holdem, "As Ks Jc 7h 5d 2d 3c", "As Ks Jc 7h 4d 2d 3c", -1, 6252, "Ace-high, kickers King, Jack, Seven, Five [As Ks Jc 7h 5d]"},
		{Holdem, "As Ac Ad Ah Kd 2d 3c", "As Ac Ad Ah Qd 2d 3c", -1, 11, "Four of a Kind, Aces, kicker King [Ac Ad Ah As Kd]"},
		{Holdem, "As Ac Ad Ah Qd 2d 3c", "As Ac Ad Ah Kd 2d 3c", +1, 11, "Four of a Kind, Aces, kicker King [Ac Ad Ah As Kd]"},
		{Holdem, "As Ks Qs Ts 9s 2s 3s", "Ks Qs Js Ts 9s 2d 3c", +1, 2, "Straight Flush, King-high, Platinum Oxide [Ks Qs Js Ts 9s]"},
		{Holdem, "6s 6c 6d 5d 5c 4s 4s", "5s 5c 5d 6d 6c 4s 4s", -1, 271, "Full House, Sixes full of Fives [6c 6d 6s 5c 5d]"},
		{Holdem, "Ks Qs Js Ts 9s 2s 3s", "Kd Qd Jd Td 9d 2d 3d", +0, 2, "Straight Flush, King-high, Platinum Oxide [Ks Qs Js Ts 9s]"},
		{Holdem, "Ks Qs Js 9s 3s Ad Kd", "Kd Qd Jd 9d 2d Ac Kc", -1, 828, "Flush, King-high, kickers Queen, Jack, Nine, Three [Ks Qs Js 9s 3s]"},
		{Holdem, "Kd Qd Jd 9d 2d Ac Kc", "Ks Qs Js 9s 3s Ad Kd", +1, 828, "Flush, King-high, kickers Queen, Jack, Nine, Three [Ks Qs Js 9s 3s]"},
		{Soko, "Ah Th 9h 8h 6c Tc 4c", "Th Tc 8h 6h Kh Qc 2s", -1, 5098, "Four Flush, Ace-high, kickers Ten, Nine, Eight, Ten [Ah Th 9h 8h Tc]"},
		{Soko, "Th Tc 8h 6h Kh Qc 2s", "Ah Th 9h 8h 6c Tc 4c", +1, 5098, "Four Flush, Ace-high, kickers Ten, Nine, Eight, Ten [Ah Th 9h 8h Tc]"},
		{Soko, "Ah Kh Th Jh 6c 9c 4s", "Ac Kc Tc Jc 6s 9s 4d", 0, 3461, "Four Flush, Ace-high, kickers King, Jack, Ten, Nine [Ah Kh Jh Th 9c]"},
		{Soko, "Ac Kc Tc Jc 6s 9s 4d", "Ah Kh Th Jh 6c 9c 4s", 0, 3461, "Four Flush, Ace-high, kickers King, Jack, Ten, Nine [Ac Kc Jc Tc 9s]"},
		{Soko, "Ah Kh Qc Jh 6c 9c 4s", "Th Tc 8h 6c Kh Qc 2s", -1, 12626, "Four Straight, Ace-high, kicker Nine [Ah Kh Qc Jh 9c]"},
		{Soko, "Th Tc 8h 6c Kh Qc 2s", "Ah Kh Qc Jh 6c 9c 4s", +1, 12626, "Four Straight, Ace-high, kicker Nine [Ah Kh Qc Jh 9c]"},
		{Soko, "Jd Tc 9c 9h 8c Ah 2d", "Kh Kd Ts 9d 8d 2c 3c", -1, 12660, "Four Straight, Jack-high, kicker Ace [Jd Tc 9c 8c Ah]"},
		{Soko, "Kh Kd Ts 9d 8d 2c 3c", "Jd Tc 9c 9h 8c Ah 2d", +1, 12660, "Four Straight, Jack-high, kicker Ace [Jd Tc 9c 8c Ah]"},
		{Short, "5c 3c Ah Th 9h 8h 7h", "J♣ J♥ 6♣ 6♦ 6♥ 5♥ 3♣", -1, 535, "Flush, Ace-high, kickers Ten, Nine, Eight, Seven [Ah Th 9h 8h 7h]"},
		{Short, "5♥ 3♣ 6♣ 6♦ 6♥ J♣ J♥", "8h 7h Ah Th 9h 5c 3c", +1, 535, "Flush, Ace-high, kickers Ten, Nine, Eight, Seven [Ah Th 9h 8h 7h]"},
	}
	for i, test := range tests {
		a, b := test.typ.Eval(Must(test.a), nil), test.typ.Eval(Must(test.b), nil)
		var ev *Eval
		r := a.Comp(b, false)
		if r != test.exp {
			t.Errorf("test %d expected comp of %d, got: %d", i, test.exp, r)
		}
		switch {
		case r == +0:
			if a.HiRank != b.HiRank || a.HiRank != test.r {
				t.Errorf("test %d expected %d == %d == %d", i, a.HiRank, b.HiRank, test.r)
			}
			ev = a
		case r == -1:
			if b.HiRank <= a.HiRank {
				t.Errorf("test %d expected %d < %d", i, a.HiRank, b.HiRank)
			}
			ev = a
		case r == +1:
			if a.HiRank <= b.HiRank {
				t.Errorf("test %d expected %d < %d", i, b.HiRank, a.HiRank)
			}
			ev = b
		}
		if ev.HiRank != test.r {
			t.Errorf("test %d expected %d, got: %d", i, test.r, ev.HiRank)
		}
		if s := fmt.Sprintf("%s", ev); s != test.s {
			t.Errorf("test %d expected %q, got: %q", i, test.s, s)
		}
	}
}

func TestEval(t *testing.T) {
	for _, r := range cactusTests(true, true) {
		for i, f := range []func() []cardTest{
			fiveCardTests,
			sixCardTests,
			sevenCardTests,
		} {
			t.Run(fmt.Sprintf("%s/%d", r.name, i+5), func(t *testing.T) {
				for j, test := range f() {
					v := Must(test.v)
					ev := EvalOf(0)
					r.eval(ev, v[:2], v[2:])
					if r, exp := ev.HiRank, test.r; r != exp {
						t.Errorf("test %d %d expected %d, got: %d", i, j, exp, r)
					}
					if s := fmt.Sprintf("%b %b", ev, ev.HiUnused); s != test.desc {
						t.Errorf("test %d %d expected %q, got: %q", i, j, test.desc, s)
					}
				}
			})
		}
	}
}

func TestNewSplitEval(t *testing.T) {
	tests := []struct {
		f   RankFunc
		max EvalRank
		v   string
		hi  EvalRank
		lo  EvalRank
	}{
		{RankLowball, Invalid, "Ah Kh Qh Jh Th", 1, 7462},
		{RankLowball, Invalid, "Ah Kh Qh Jh Th 9h 8h", 1, 6528},
		{RankRazz, Invalid, "Kh Qh Jh Th 9h", 2, 7936},
		{RankRazz, Invalid, "Kh Qh Jh Th 9h 8h 7h", 2, 1984},
		{RankShort, Invalid, "Kh Qh Jh Th 9h", 2, 2},
		{RankRazz, Invalid, "Kh Qh Qc Jc Jh Th 9h", 2, 7936},
		{RankEightOrBetter, eightOrBetterMax, "5h 4h 3h 2h Ah", 10, 31},
		{RankEightOrBetter, eightOrBetterMax, "8h 7h 6h 5h 4h", 7, 248},
		{RankEightOrBetter, eightOrBetterMax, "9h Th 8h 7h 6h 5h 4h", 5, 248},
		{RankEightOrBetter, eightOrBetterMax, "9h 7h 6h 5h 4h", 1567, Invalid},
		{RankEightOrBetter, eightOrBetterMax, "Ah Kh Qh Jh Th", 1, Invalid},
	}
	for i, test := range tests {
		p, f := Must(test.v), NewSplitEval(RankCactus, test.f, test.max)
		ev := EvalOf(0)
		f(ev, p, nil)
		if r, exp := ev.HiRank, test.hi; r != exp {
			t.Errorf("test %d expected rank %d, got: %d", i, exp, r)
		}
		if r, exp := ev.LoRank, test.lo; r != exp {
			t.Errorf("test %d expected rank %d, got: %d", i, exp, r)
		}
	}
}

func TestRankEightOrBetter(t *testing.T) {
	p0 := Must("Ah 2h 3h 4h 5h 6h 7h 8h")
	for i := Nine; i <= King; i++ {
		p1 := Must(i.String() + "h 4h 3h 2h Ah")
		r1 := RankEightOrBetter(p1[0], p1[1], p1[2], p1[3], p1[4])
		for c0 := range p0 {
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

func TestEvalRankToFrom(t *testing.T) {
	for i := EvalRank(1); i <= Nothing; i++ {
		a := i.ToFlushOver()
		if b := a.FromFlushOver(); b != i {
			t.Errorf("expected %d, got: %d", i, b)
		}
	}
	for i := EvalRank(1); i <= Nothing; i++ {
		a := i.ToLowball()
		if b := a.FromLowball(); b != i {
			t.Errorf("expected %d, got: %d", i, b)
		}
	}
}

func TestEvalRankTitle(t *testing.T) {
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
		if s := test.r.Title(); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
	}
}

func TestSoko(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"Jd Tc 9c 9h 8c", "Four Straight, Jack-high, kicker Nine [Jd Tc 9c 8c 9h]"},
		{"Jd Jc Tc 8c 9c", "Four Flush, Jack-high, kickers Ten, Nine, Eight, Jack [Jc Tc 9c 8c Jd]"},
		{"Ac Qh 4h 3c 2c", "Ace-high, kickers Queen, Four, Three, Two [Ac Qh 4h 3c 2c]"},
		{"Ac Kc Qc Jc 9s", "Four Flush, Ace-high, kickers King, Queen, Jack, Nine [Ac Kc Qc Jc 9s]"},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := Must(test.s)
			ev := Soko.Eval(v, nil)
			if s := fmt.Sprintf("%s", ev); s != test.exp {
				t.Errorf("expected %q, got: %q", test.exp, s)
			}
		})
	}
}

func TestLowballCards(t *testing.T) {
	if s := os.Getenv("TESTS"); !strings.Contains(s, "lowball") && !strings.Contains(s, "all") {
		t.Skip("skipping: $ENV{TESTS} does not contain 'lowball' or 'all'")
	}
	t.Parallel()
	u, c, l, ev, uv := shuffled(DeckFrench), NewCactusEval(0, false, false), NewLowballEval(false), EvalOf(Holdem), EvalOf(Lowball)
	for c0 := range 52 {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						v := []Card{u[c0], u[c1], u[c2], u[c3], u[c4]}
						c(ev, v, nil)
						l(uv, v, nil)
						if r, exp := uv.HiRank.FromLowball(), ev.HiRank; r != exp {
							t.Fatalf("expected equal ranks for %v %d, got: %d", v, exp, r)
						}
					}
				}
			}
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

type cactusTest struct {
	name string
	eval EvalFunc
}

func cactusTests(base, normalize bool) []cactusTest {
	var tests []cactusTest
	if base && cactus != nil {
		tests = append(tests, cactusTest{"Cactus", wrapCactus(cactus, normalize)})
	}
	if cactusFast != nil {
		tests = append(tests, cactusTest{"CactusFast", wrapCactus(cactusFast, normalize)})
	}
	if twoPlusTwo != nil {
		tests = append(tests, cactusTest{"TwoPlusTwo", wrapTwoPlusTwo(normalize)})
	}
	if cactusFast != nil && twoPlusTwo != nil {
		tests = append(tests, cactusTest{"Hybrid", NewHybridEval(normalize, false)})
	}
	return tests
}

func wrapCactus(f RankFunc, normalize bool) EvalFunc {
	if !normalize {
		return NewEval(f)
	}
	g := NewEval(f)
	return func(ev *Eval, p, b []Card) {
		g(ev, p, b)
		bestCactus(ev.HiRank, ev.HiBest, ev.HiUnused, 0, nil)
	}
}

func wrapTwoPlusTwo(normalize bool) EvalFunc {
	if !normalize {
		return func(ev *Eval, p, b []Card) {
			n, m := len(p), len(b)
			v := make([]Card, n+m)
			copy(v, p)
			copy(v[n:], b)
			ev.HiRank = twoPlusTwo(v)
		}
	}
	return func(ev *Eval, p, b []Card) {
		n, m := len(p), len(b)
		v := make([]Card, n+m)
		copy(v, p)
		copy(v[n:], b)
		ev.HiRank = twoPlusTwo(v)
		ev.HiBest, ev.HiUnused = bestCactusSplit(ev.HiRank, v, 0)
	}
}

type cardTest struct {
	v    string
	r    EvalRank
	exp  EvalRank
	desc string
}

func fiveCardTests() []cardTest {
	return []cardTest{
		{"As Ks Jc 7h 5d", 0x186c, Nothing, "Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] []"},
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
		{"As Ac Ad Ah Th", 0x000e, FourOfAKind, "Four of a Kind, Aces, kicker Ten [A♣ A♦ A♥ A♠ T♥] []"},
		{"3d 5d 2d 4d Ad", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] []"},
		{"6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high, Aluminum Window [6♦ 5♦ 4♦ 3♦ 2♦] []"},
		{"9♦ 6♦ 8♦ 5♦ 7♦", 0x0006, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ 5♦] []"},
		{"As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] []"},
	}
}

func sixCardTests() []cardTest {
	return []cardTest{
		{"3d As Ks Jc 7h 5d", 0x186c, Nothing, "Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦]"},
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
		{"4d As Ac Ad Ah Th", 0x000e, FourOfAKind, "Four of a Kind, Aces, kicker Ten [A♣ A♦ A♥ A♠ T♥] [4♦]"},
		{"3d 5d 2d 4d Ad 3s", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [3♠]"},
		{"T♦ 6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high, Aluminum Window [6♦ 5♦ 4♦ 3♦ 2♦] [T♦]"},
		{"J♦ 9♦ 6♦ 8♦ 5♦ 7♦", 0x0006, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{"7♦ J♦ 9♦ 6♦ 8♦ 5♦", 0x0006, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{"3d As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦]"},
	}
}

func sevenCardTests() []cardTest {
	return []cardTest{
		{"2d 3d As Ks Jc 7h 5d", 0x186c, Nothing, "Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{"2d 3d As Ac Jc 7h 5d", 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{"9d Jd 6s 6c 5c 5d 4d", 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] [9♦ 4♦]"},
		{"2d 3d 6s 6c Jc Jd 5d", 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] [3♦ 2♦]"},
		{"2d 3d As Ac Jc Jd 5d", 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] [3♦ 2♦]"},
		{"9c 7d 4s 7c Js Qd 4d", 3185, TwoPair, "Two Pair, Sevens over Fours, kicker Queen [7♣ 7♦ 4♦ 4♠ Q♦] [J♠ 9♣]"},
		{"2c 3d As Ac Ad Jd 5d", 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] [3♦ 2♣]"},
		{"4s 5s 2d 3h Ac Jd Qs", 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] [Q♠ J♦]"},
		{"2d 3d 9s Ks Qd Jh Td", 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] [3♦ 2♦]"},
		{"2d 3d As Ks Qd Jh Td", 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] [3♦ 2♦]"},
		{"2d 3d Ts 7s 4s 3s 2s", 0x0606, Flush, "Flush, Ten-high, kickers Seven, Four, Three, Two [T♠ 7♠ 4♠ 3♠ 2♠] [3♦ 2♦]"},
		{"2d 3d 4s 4c 4d 2s 2h", 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♦ 2♥] [3♦ 2♠]"},
		{"4d 3d 5s 5c 5d 6s 6h", 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] [4♦ 3♦]"},
		{"4d 3d 6s 6c 6d 5s 5h", 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] [4♦ 3♦]"},
		{"2d 3d As Ac Ad Ah 5h", 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] [3♦ 2♦]"},
		{"4d 4s As Ac Ad Ah Th", 0x000e, FourOfAKind, "Four of a Kind, Aces, kicker Ten [A♣ A♦ A♥ A♠ T♥] [4♦ 4♠]"},
		{"3d 5d 2d 4d Ad 3s 4s", 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [4♠ 3♠]"},
		{"J♦ T♦ 6♦ 5♦ 4♦ 3♦ 2♦", 0x0009, StraightFlush, "Straight Flush, Six-high, Aluminum Window [6♦ 5♦ 4♦ 3♦ 2♦] [J♦ T♦]"},
		{"7♦ J♦ 9♦ 6♦ 8♦ 5♦ 2♦", 0x0006, StraightFlush, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ 5♦] [J♦ 2♦]"},
		{"2d 3d As Ks Qs Js Ts", 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦ 2♦]"},
	}
}

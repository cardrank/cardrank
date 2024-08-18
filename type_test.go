package cardrank

import (
	"fmt"
	"math/rand"
	"slices"
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

func TestHoldem(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
		s string
	}{
		{"5d 4d 3d 2d Ad Tc Jc", "5d 4d 3d 2d Ad", "Jc Tc", 10, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦]"},
		{"5c 4c 3c 2c Ac Th Qh", "5c 4c 3c 2c Ac", "Qh Th", 10, "Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣]"},
		{"5h 4h 3h 2h Ah Tc Kd", "5h 4h 3h 2h Ah", "Kd Tc", 10, "Straight Flush, Five-high, Steel Wheel [5♥ 4♥ 3♥ 2♥ A♥]"},
		{"5s 4s 3s 2s As 9c 8c", "5s 4s 3s 2s As", "9c 8c", 10, "Straight Flush, Five-high, Steel Wheel [5♠ 4♠ 3♠ 2♠ A♠]"},
		{"7d Jc 5d 4d 3d 2d Ad", "5d 4d 3d 2d Ad", "Jc 7d", 10, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦]"},
		{"7c Qh 5c 4c 3c 2c Ac", "5c 4c 3c 2c Ac", "Qh 7c", 10, "Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣]"},
		{"7h Kd 5h 4h 3h 2h Ah", "5h 4h 3h 2h Ah", "Kd 7h", 10, "Straight Flush, Five-high, Steel Wheel [5♥ 4♥ 3♥ 2♥ A♥]"},
		{"7s 8c 5s 4s 3s 2s As", "5s 4s 3s 2s As", "8c 7s", 10, "Straight Flush, Five-high, Steel Wheel [5♠ 4♠ 3♠ 2♠ A♠]"},
		{"5d 4d 3d 2d Ac 8h 8s", "5d 4d 3d 2d Ac", "8h 8s", 1609, "Straight, Five-high [5♦ 4♦ 3♦ 2♦ A♣]"},
		{"5c 4c 3c 2c Ad Kh Jh", "5c 4c 3c 2c Ad", "Kh Jh", 1609, "Straight, Five-high [5♣ 4♣ 3♣ 2♣ A♦]"},
		{"5h 4h 3h 2h As Kd Jd", "5h 4h 3h 2h As", "Kd Jd", 1609, "Straight, Five-high [5♥ 4♥ 3♥ 2♥ A♠]"},
		{"5s 4s 3s 2s Ah Qh Jd", "5s 4s 3s 2s Ah", "Qh Jd", 1609, "Straight, Five-high [5♠ 4♠ 3♠ 2♠ A♥]"},
		{"9d 8d 7d 6d Ad Tc Jc", "Ad 9d 8d 7d 6d", "Jc Tc", 747, "Flush, Ace-high, kickers Nine, Eight, Seven, Six [A♦ 9♦ 8♦ 7♦ 6♦]"},
		{"9c 8c 7c 6c Ac Th Qh", "Ac 9c 8c 7c 6c", "Qh Th", 747, "Flush, Ace-high, kickers Nine, Eight, Seven, Six [A♣ 9♣ 8♣ 7♣ 6♣]"},
		{"9h 8h 7h 6h Ah Tc Kd", "Ah 9h 8h 7h 6h", "Kd Tc", 747, "Flush, Ace-high, kickers Nine, Eight, Seven, Six [A♥ 9♥ 8♥ 7♥ 6♥]"},
		{"9s 8s 7s 6s As 9c 8c", "As 9s 8s 7s 6s", "9c 8c", 747, "Flush, Ace-high, kickers Nine, Eight, Seven, Six [A♠ 9♠ 8♠ 7♠ 6♠]"},
		{"9d 8d 7d 6d Ac 8h 8s", "8d 8h 8s Ac 9d", "7d 6d", 2010, "Three of a Kind, Eights, kickers Ace, Nine [8♦ 8♥ 8♠ A♣ 9♦]"},
		{"9c 8c 7c 6c Ad Kh Jh", "Ad Kh Jh 9c 8c", "7c 6c", 6238, "Ace-high, kickers King, Jack, Nine, Eight [A♦ K♥ J♥ 9♣ 8♣]"},
		{"9h 8h 7h 6h As Kd Jd", "As Kd Jd 9h 8h", "7h 6h", 6238, "Ace-high, kickers King, Jack, Nine, Eight [A♠ K♦ J♦ 9♥ 8♥]"},
		{"9s 8s 7s 6s Ah Qh Jd", "Ah Qh Jd 9s 8s", "7s 6s", 6358, "Ace-high, kickers Queen, Jack, Nine, Eight [A♥ Q♥ J♦ 9♠ 8♠]"},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Holdem.Eval(pocket, nil)
		if r, exp := ev.HiRank, test.r; r != exp {
			t.Errorf("test %d %v expected %d, got: %d", i, pocket, exp, r)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf("%s %b", desc, desc.Best), test.s; s != exp {
			t.Errorf("test %d expected %q, got: %q", i, exp, s)
		}
	}
}

func TestDallas(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
		s string
	}{
		{"2d 2h 3c 5c 5s Jh 7c", "5c 5s 2d 2h Jh", "7c 3c", 3285, "Two Pair, Fives over Twos, kicker Jack [5♣ 5♠ 2♦ 2♥ J♥]"},
		{"2d 3c 2h 5c 5s Jh 7c", "5c 5s 2d 2h 3c", "Jh 7c", 3292, "Two Pair, Fives over Twos, kicker Three [5♣ 5♠ 2♦ 2♥ 3♣]"},
		{"Jh 7c 2d 3c 2h 5c 5s", "5c 5s Jh 7c 3c", "2d 2h", 5462, "Pair, Fives, kickers Jack, Seven, Three [5♣ 5♠ J♥ 7♣ 3♣]"},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Dallas.Eval(pocket[:2], pocket[2:])
		if r, exp := ev.HiRank, test.r; r != exp {
			t.Errorf("test %d %v expected %d, got: %d", i, pocket, exp, r)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf("%s %b", desc, desc.Best), test.s; s != exp {
			t.Errorf("test %d expected %q, got: %q", i, exp, s)
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
		{"9d 8d 7d 6d Ad Tc Jc", "9d 8d 7d 6d Ad", "Jc Tc", 6, "Straight Flush, Nine-high, Iron Maiden [9♦ 8♦ 7♦ 6♦ A♦]"},
		{"9c 8c 7c 6c Ac Th Qh", "9c 8c 7c 6c Ac", "Qh Th", 6, "Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣]"},
		{"9h 8h 7h 6h Ah Tc Kd", "9h 8h 7h 6h Ah", "Kd Tc", 6, "Straight Flush, Nine-high, Iron Maiden [9♥ 8♥ 7♥ 6♥ A♥]"},
		{"9s 8s 7s 6s As 9c 8c", "9s 8s 7s 6s As", "9c 8c", 6, "Straight Flush, Nine-high, Iron Maiden [9♠ 8♠ 7♠ 6♠ A♠]"},
		{"9d 8d 7d 6d Ac 8h 8s", "9d 8d 7d 6d Ac", "8h 8s", 1605, "Straight, Nine-high [9♦ 8♦ 7♦ 6♦ A♣]"},
		{"9c 8c 7c 6c Ad Kh Jh", "9c 8c 7c 6c Ad", "Kh Jh", 1605, "Straight, Nine-high [9♣ 8♣ 7♣ 6♣ A♦]"},
		{"9h 8h 7h 6h As Kd Jd", "9h 8h 7h 6h As", "Kd Jd", 1605, "Straight, Nine-high [9♥ 8♥ 7♥ 6♥ A♠]"},
		{"9s 8s 7s 6s Ah Qh Jd", "9s 8s 7s 6s Ah", "Qh Jd", 1605, "Straight, Nine-high [9♠ 8♠ 7♠ 6♠ A♥]"},
		{"2♣ 2♥ 2♠ 3♦ 3♠ J♥ 7♦", "2c 2h 2s 3d 3s", "Jh 7d", 1599, "Full House, Twos full of Threes [2♣ 2♥ 2♠ 3♦ 3♠]"},
		{"T♣ T♦ 8♦ 8♥ 8♠ Q♠ J♣", "8d 8h 8s Tc Td", "Qs Jc", 1520, "Full House, Eights full of Tens [8♦ 8♥ 8♠ T♣ T♦]"},
		{"A♣ A♥ A♠ K♦ K♠ J♥ 7♦", "Ac Ah As Kd Ks", "Jh 7d", 1444, "Full House, Aces full of Kings [A♣ A♥ A♠ K♦ K♠]"},
		{"A♦ Q♦ T♦ 9♦ 8♦ A♣ 9♣", "Ad Qd Td 9d 8d", "Ac 9c", 367, "Flush, Ace-high, kickers Queen, Ten, Nine, Eight [A♦ Q♦ T♦ 9♦ 8♦]"},
		{"A♦ K♦ T♦ 9♦ 8♦ A♣ 9♣", "Ad Kd Td 9d 8d", "Ac 9c", 247, "Flush, Ace-high, kickers King, Ten, Nine, Eight [A♦ K♦ T♦ 9♦ 8♦]"},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Short.Eval(pocket, nil)
		if r, exp := ev.HiRank, test.r; r != exp {
			t.Errorf("test %d %v expected %d, got: %d", i, pocket, exp, r)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf("%s %b", desc, desc.Best), test.s; s != exp {
			t.Errorf("test %d expected %q, got: %q", i, exp, s)
		}
	}
}

func TestManila(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r EvalRank
		s string
	}{
		{"Td 9d 8d 7d Ad Tc Jc", "Td 9d 8d 7d Ad", "Jc Tc", 5, "Straight Flush, Ten-high, Golden Ratio [T♦ 9♦ 8♦ 7♦ A♦]"},
		{"Tc 9c 8c 7c Ac Th Qh", "Tc 9c 8c 7c Ac", "Qh Th", 5, "Straight Flush, Ten-high, Golden Ratio [T♣ 9♣ 8♣ 7♣ A♣]"},
		{"Th 9h 8h 7h Ah Tc Kd", "Th 9h 8h 7h Ah", "Kd Tc", 5, "Straight Flush, Ten-high, Golden Ratio [T♥ 9♥ 8♥ 7♥ A♥]"},
		{"Ts 9s 8s 7s As 9c 8c", "Ts 9s 8s 7s As", "9c 8c", 5, "Straight Flush, Ten-high, Golden Ratio [T♠ 9♠ 8♠ 7♠ A♠]"},
		{"Td 9d 8d 7d Ac 8h 8s", "Td 9d 8d 7d Ac", "8h 8s", 1604, "Straight, Ten-high [T♦ 9♦ 8♦ 7♦ A♣]"},
		{"Tc 9c 8c 7c Ad Kh Qh", "Tc 9c 8c 7c Ad", "Kh Qh", 1604, "Straight, Ten-high [T♣ 9♣ 8♣ 7♣ A♦]"},
		{"Th 9h 8h 7h As Kd Qd", "Th 9h 8h 7h As", "Kd Qd", 1604, "Straight, Ten-high [T♥ 9♥ 8♥ 7♥ A♠]"},
		{"Ts 9s 8s 7s Ah Kh Qd", "Ts 9s 8s 7s Ah", "Kh Qd", 1604, "Straight, Ten-high [T♠ 9♠ 8♠ 7♠ A♥]"},
		{"Ts 7h Th 7s Kh Qd Tc", "Tc Th Ts 7h 7s", "Kh Qd", 1498, "Full House, Tens full of Sevens [T♣ T♥ T♠ 7♥ 7♠]"},
		{"Th 7h 6h 8h Kh Qd Tc", "Kh Th 8h 7h 6h", "Qd Tc", 884, "Flush, King-high, kickers Ten, Eight, Seven, Six [K♥ T♥ 8♥ 7♥ 6♥]"},
	}
	for i, test := range tests {
		v, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Manila.Eval(v[:2], v[2:])
		if r, exp := ev.HiRank, test.r; r != exp {
			t.Errorf("test %d %v expected %d, got: %d", i, v, exp, r)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected %v, got: %v", i, v, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected %v, got: %v", i, v, unused, ev.HiUnused)
		}
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf("%s %b", desc, desc.Best), test.s; s != exp {
			t.Errorf("test %d expected %q, got: %q", i, exp, s)
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
		{"2h 2c 2d 2s As Ks Qs", "2c 2h As Ks Qs", "2d 2s", 59569},
		{"Ah Ac Ad Ks Kh Ks Qs", "Ac Ah Kh Ks Qs", "Ad Ks", 63067},
		{"Ah Ac Ad Ks Qh Ks Qs", "Ks Ks Qh Qs Ah", "Ac Ad", 62935},
		{"Kh Kd Qd Qs Jh Ks Js", "Qd Qs Jh Js Kh", "Kd Ks", 62813},
		{"3h 3c Kh Qd Jd Ks Qs", "3c 3h Kh Qd Jd", "Ks Qs", 59734},
		{"2h 2c Kh Qd Jd Ks Qs", "2c 2h Kh Qd Jd", "Ks Qs", 59514},
		{"3h 2c Kh Qd Jd Ks Qs", "Kh Qd Jd 3h 2c", "Ks Qs", 7174},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Razz.Eval(pocket, nil)
		if ev.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.r, ev.HiRank)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
	}
}

func TestBadugi(t *testing.T) {
	tests := []struct {
		v   string
		b   string
		u   string
		exp EvalRank
		s   string
	}{
		{"Kh Qh Jh Th", "Th", "Kh Qh Jh", 25088, "Ten-low [Th]"},
		{"Kh Qh Jd Th", "Jd Th", "Kh Qh", 17920, "Jack, Ten-low [Jd Th]"},
		{"Kh Qc Jd Th", "Qc Jd Th", "Kh", 11776, "Queen, Jack, Ten-low [Qc Jd Th]"},
		{"Ks Qc Jd Th", "Ks Qc Jd Th", "", 7680, "King, Queen, Jack, Ten-low [Ks Qc Jd Th]"},
		{"2h 2c 2d 2s", "2s", "2c 2d 2h", 24578, "Two-low [2s]"},
		{"Ah Kh Qh Jh", "Ah", "Kh Qh Jh", 24577, "Ace-low [Ah]"},
		{"Kh Kd Qd Qs", "Kh Qs", "Kd Qd", 22528, "King, Queen-low [Kh Qs]"},
		{"Ah Ac Ad Ks", "Ks Ah", "Ac Ad", 20481, "King, Ace-low [Ks Ah]"},
		{"3h 3c Kh Qd", "Kh Qd 3c", "3h", 14340, "King, Queen, Three-low [Kh Qd 3c]"},
		{"2h 2c Kh Qd", "Kh Qd 2c", "2h", 14338, "King, Queen, Two-low [Kh Qd 2c]"},
		{"3h 2c Kh Ks", "Ks 3h 2c", "Kh", 12294, "King, Three, Two-low [Ks 3h 2c]"},
		{"3h 2c Kh Qd", "Qd 3h 2c", "Kh", 10246, "Queen, Three, Two-low [Qd 3h 2c]"},
		{"Ah 2c 4s 6d", "6d 4s 2c Ah", "", 43, "Six, Four, Two, Ace-low [6d 4s 2c Ah]"},
		{"Ac 2h 4d 6s", "6s 4d 2h Ac", "", 43, "Six, Four, Two, Ace-low [6s 4d 2h Ac]"},
		{"Ah 2c 3s 6d", "6d 3s 2c Ah", "", 39, "Six, Three, Two, Ace-low [6d 3s 2c Ah]"},
		{"Ah 2c 4s 5d", "5d 4s 2c Ah", "", 27, "Five, Four, Two, Ace-low [5d 4s 2c Ah]"},
		{"Ah 2c 3s 5d", "5d 3s 2c Ah", "", 23, "Five, Three, Two, Ace-low [5d 3s 2c Ah]"},
		{"Ah 2c 3s 4d", "4d 3s 2c Ah", "", 15, "Four, Three, Two, Ace-low [4d 3s 2c Ah]"},
		{"Ac 2h 3s 4d", "4d 3s 2h Ac", "", 15, "Four, Three, Two, Ace-low [4d 3s 2h Ac]"},
	}
	for i, test := range tests {
		pocket, best, unused := Must(test.v), Must(test.b), Must(test.u)
		ev := Badugi.Eval(pocket, nil)
		if ev.HiRank != test.exp {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.exp, ev.HiRank)
		}
		if !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, pocket, best, ev.HiBest)
		}
		if !slices.Equal(ev.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, pocket, unused, ev.HiUnused)
		}
		if s := fmt.Sprintf("%s", ev); s != test.s {
			t.Errorf("test %d %v expected %q, got: %q", i, pocket, test.s, s)
		}
	}
}

func TestLowball(t *testing.T) {
	tests := []struct {
		v   string
		b   string
		exp EvalRank
		s   string
	}{
		{"3h 5h 7h 4h 2c", "7h 5h 4h 3h 2c", 1, "Seven, Five, Four, Three, Two-low, No. 1"},
		{"3h 6h 7h 4h 2c", "7h 6h 4h 3h 2c", 2, "Seven, Six, Four, Three, Two-low, No. 2"},
		{"3h 6h 7h 5h 2c", "7h 6h 5h 3h 2c", 3, "Seven, Six, Five, Three, Two-low, No. 3"},
		{"4h 6h 7h 5h 2c", "7h 6h 5h 4h 2c", 4, "Seven, Six, Five, Four, Two-low, No. 4"},
		{"3h 5h 8h 4h 2c", "8h 5h 4h 3h 2c", 5, "Eight, Five, Four, Three, Two-low, No. 5"},
		{"3h 6h 8h 4h 2c", "8h 6h 4h 3h 2c", 6, "Eight, Six, Four, Three, Two-low, No. 6"},
		{"3h 6h 8h 5h 2c", "8h 6h 5h 3h 2c", 7, "Eight, Six, Five, Three, Two-low, No. 7"},
		{"4h 6h 8h 5h 2c", "8h 6h 5h 4h 2c", 8, "Eight, Six, Five, Four, Two-low, No. 8"},
		{"4h 6h 8h 5h 3c", "8h 6h 5h 4h 3c", 9, "Eight, Six, Five, Four, Three-low, No. 9"},
		{"3h 7h 8h 4h 2c", "8h 7h 4h 3h 2c", 10, "Eight, Seven, Four, Three, Two-low, No. 10"},
		{"3h 7h 8h 5h 2c", "8h 7h 5h 3h 2c", 11, "Eight, Seven, Five, Three, Two-low"},
		{"4h 7h 8h 5h 2c", "8h 7h 5h 4h 2c", 12, "Eight, Seven, Five, Four, Two-low"},
		{"4h 7h 8h 5h 3c", "8h 7h 5h 4h 3c", 13, "Eight, Seven, Five, Four, Three-low"},
		{"3h 7h 8h 6h 2c", "8h 7h 6h 3h 2c", 14, "Eight, Seven, Six, Three, Two-low"},
		{"4h 7h 8h 6h 2c", "8h 7h 6h 4h 2c", 15, "Eight, Seven, Six, Four, Two-low"},
		{"4h 7h 8h 6h 3c", "8h 7h 6h 4h 3c", 16, "Eight, Seven, Six, Four, Three-low"},
		{"5h 7h 8h 6h 2c", "8h 7h 6h 5h 2c", 17, "Eight, Seven, Six, Five, Two-low"},
		{"5h 7h 8h 6h 3c", "8h 7h 6h 5h 3c", 18, "Eight, Seven, Six, Five, Three-low"},
		{"3h 5h 9h 4h 2c", "9h 5h 4h 3h 2c", 19, "Nine, Five, Four, Three, Two-low"},
		{"Ks Qd Jd Tc 8d", "Ks Qd Jd Tc 8d", 784, "King, Queen, Jack, Ten, Eight-low"},
		{"3c 5c As 4s 2d", "As 5c 4s 3c 2d", 785, "Ace, Five, Four, Three, Two-low"},
		{"3s 6d Ac 4d 2c", "Ac 6d 4d 3s 2c", 786, "Ace, Six, Four, Three, Two-low"},
		{"Ah Kh Jc Ac Ad", "Ac Ad Ah Kh Jc", 5853, "Three of a Kind, Aces, kickers King, Jack"},
		{"Ah Kh Qh Ac Ad", "Ac Ad Ah Kh Qh", 5854, "Three of a Kind, Aces, kickers King, Queen"},
		{"6c 4d 3d 2d 5d", "6c 5d 4d 3d 2d", 5855, "Straight, Six-high"},
		{"Jd Td Kd Qd 8d", "Kd Qd Jd Td 8d", 6647, "Flush, King-high, kickers Queen, Jack, Ten, Eight"},
		{"3d 6d Ad 4d 2d", "Ad 6d 4d 3d 2d", 6648, "Flush, Ace-high, kickers Six, Four, Three, Two"},
		{"5c 3c 2c Ac 6c", "Ac 6c 5c 3c 2c", 6649, "Flush, Ace-high, kickers Six, Five, Three, Two"},
		{"Kh Ac Ad As Kd", "Ac Ad As Kd Kh", 7297, "Full House, Aces full of Kings"},
	}
	for i, test := range tests {
		pocket := Must(test.v)
		ev := Lowball.Eval(pocket, nil)
		if ev.HiRank != test.exp {
			t.Errorf("test %d %v expected rank %d, got: %d", i, pocket, test.exp, ev.HiRank)
		}
		if best := Must(test.b); !slices.Equal(ev.HiBest, best) {
			t.Errorf("test %d expected %v, got: %v", i, best, ev.HiBest)
		}
		if s, exp := fmt.Sprintf("%s", ev.Desc(false)), test.s; s != exp {
			t.Errorf("test %d expected %q, got: %q", i, exp, s)
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
		{Short, "As 7d Ad 6s 6d", "8d Td", "Ac Qh", 556, 1451, -1},
		{Short, "As 7d Ad 6s 6d", "Ac Qh", "8d Td", 1451, 556, +1},
		{Short, "Kc Qh Jc Td 8d", "Ac Qh", "Ah 6c", 1600, 1600, 0},
		{Short, "Kc Qh Jc Td 8d", "Ah 6c", "Ac Qh", 1600, 1600, 0},
		{Short, "9c 7d 8d As Qs", "Ac 6s", "Tc Ts", 1605, 4217, -1},
		{Short, "9c 7d 8d As Qs", "Tc Ts", "Ac 6s", 4217, 1605, +1},
		{Short, "9s 7s 8s Ac Qs", "As 6s", "Tc Ts", 6, 1072, -1},
		{Short, "9s 7s 8s Ac Qs", "Tc Ts", "As 6s", 1072, 6, +1},
		{Manila, "As 8d Ad 7s 7d", "9d Jd", "Ac Qh", 479, 1624, -1},
		{Manila, "As 8d Ad 7s 7d", "Ac Qh", "9d Jd", 1624, 479, +1},
		{Manila, "Kc Qh Jc Td 9d", "Ac Qh", "Ah 7c", 1600, 6188, -1},
		{Manila, "Kc Qh Jc Td 9d", "Ah 7c", "Ac Qh", 6188, 1600, +1},
		{Manila, "10c 8d 9d As Qs", "Ac 7s", "Kc Ks", 1604, 3547, -1},
		{Manila, "10c 8d 9d As Qs", "Kc Ks", "Ac 7s", 3547, 1604, +1},
		{Manila, "10s 8s 9s Ac Qs", "As 7s", "Kc Ks", 5, 3547, -1},
		{Manila, "10s 8s 9s Ac Qs", "Kc Ks", "As 7s", 3547, 5, +1},
		{Dallas, "4h Tc 4d 6s 6h", "4s 7h", "Jc Ts", 2310, 2966, -1},
		{Dallas, "4h Tc 4d 6s 6h", "Jc Ts", "4s 7h", 2966, 2310, +1},
		{Houston, "4h Tc 4d 6s", "4s 7h 6s", "Jc Ts 2h", 295, 2988, -1},
		{Houston, "4h Tc 4d 6s", "Jc Ts 2h", "4s 7h 6s", 2988, 295, +1},
		{Omaha, "Td 2c Jd 4c 5c", "As Ah Qh 3s", "Ad Ac 7d 4d", 1609, 3430, -1},
		{Omaha, "Td 2c Jd 4c 5c", "Ad Ac 7d 4d", "As Ah Qh 3s", 3430, 1609, +1},
		{Omaha, "Kc Qh Jc 8d 4s", "Ac Td 3h 6c", "Ah Tc 2c 3c", 1600, 1600, 0},
		{Omaha, "Kc Qh Jc 8d 4s", "Ah Tc 2c 3c", "Ac Td 3h 6c", 1600, 1600, 0},
		{Omaha, "2d 3h 8s 8h 2s", "Kd Ts Td 4h", "Jd 7d 7c 4c", 2950, 3104, -1},
		{Omaha, "2d 3h 8s 8h 2s", "Jd 7d 7c 4c", "Kd Ts Td 4h", 3104, 2950, +1},
		{Omaha, "Tc 6c 2s 3s As", "Kd Qs Js 8h", "9h 9d 4h 4d", 522, 4455, -1},
		{Omaha, "Tc 6c 2s 3s As", "9h 9d 4h 4d", "Kd Qs Js 8h", 4455, 522, +1},
		{Omaha, "4s 3h 6c 2d Kd", "Kh Qs 5h 2c", "7s 7c 4h 2s", 1608, 3305, -1},
		{Omaha, "4s 3h 6c 2d Kd", "7s 7c 4h 2s", "Kh Qs 5h 2c", 3305, 1608, +1},
		{OmahaFive, "Qs Ts 8s 10c 9s", "As Kc Jh Jd Th", "Ks Js 3h 4d 6h", 1601, 2, +1},
		{OmahaFive, "Qs Ts 8s 10c 9s", "Ks Js 3h 4d 6h", "As Kc Jh Jd Th", 2, 1601, -1},
		{OmahaSix, "Qs Ts 8s 10c 9s", "As Kc Jh Jd Th Jc", "Ks Js 3h 4d 6h 5c", 1601, 2, +1},
		{OmahaSix, "Qs Ts 8s 10c 9s", "Ks Js 3h 4d 6h 5c", "As Kc Jh Jd Th Jc", 2, 1601, -1},
	}
	for i, test := range tests {
		board := Must(test.board)
		a, b := test.typ.Eval(Must(test.a), board), test.typ.Eval(Must(test.b), board)
		if r, exp := a.HiRank, test.j; r != exp {
			t.Errorf("test %d %s expected %d, got: %d", i, test.typ, exp, r)
		}
		if r, exp := b.HiRank, test.k; r != exp {
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
		"21st", "22nd", "93rd", "94th",
		"101st", "102nd", "River",
	}
	streets := NumberedStreets(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 10, 1, 71, 1, 7, 1, 1)
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

func TestIdToType(t *testing.T) {
	for _, desc := range DefaultTypes() {
		s := desc.Type.Id()
		typ, err := IdToType(s)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if desc.Type != typ {
			t.Errorf("expected types to match")
		}
	}
}

func TestDescType(t *testing.T) {
	tests := []struct {
		typ Type
		v   string
		s   string
		exp string
	}{
		{Holdem, "Ah Kh Qh Jh Th", "%s", "Straight Flush, Ace-high, Royal"},
		{Holdem, "Ah Kh Qh Jh Th", "%S", "Straight Flush, Ace-high"},
		{Holdem, "Ah Kh Qh Jh Th", "%e", "Straight Flush"},
		{Short, "Qh Qd Qc Qh Ah", "%s", "Four of a Kind, Queens, kicker Ace"},
		{Short, "Qh Qd Qc Qh Ah", "%S", "Four of a Kind, Queens"},
		{Short, "Qh Qd Qc Qh Ah", "%e", "Four of a Kind"},
		{Holdem, "Ah Ac As Th Tc", "%s", "Full House, Aces full of Tens"},
		{Holdem, "Ah Ac As Th Tc", "%S", "Full House"},
		{Holdem, "Ah Ac As Th Tc", "%e", "Full House"},
		{Short, "Ah Kh Qh 9h 8h", "%s", "Flush, Ace-high, kickers King, Queen, Nine, Eight"},
		{Short, "Ah Kh Qh 9h 8h", "%S", "Flush, Ace-high"},
		{Short, "Ah Kh Qh 9h 8h", "%e", "Flush"},
		{Short, "7h 2h 2c 7s 7c", "%s", "Full House, Sevens full of Twos"},
		{Short, "7h 2h 2c 7s 7c", "%S", "Full House"},
		{Short, "7h 2h 2c 7s 7c", "%e", "Full House"},
		{Short, "Ah 6c 7c 8s 9d", "%s", "Straight, Nine-high"},
		{Short, "Ah 6c 7c 8s 9d", "%S", "Straight, Nine-high"},
		{Short, "Ah 6c 7c 8s 9d", "%e", "Straight"},
		{Holdem, "6h As Qc Qd Qs", "%s", "Three of a Kind, Queens, kickers Ace, Six"},
		{Holdem, "6h As Qc Qd Qs", "%S", "Three of a Kind, Queens"},
		{Holdem, "6h As Qc Qd Qs", "%e", "Three of a Kind"},
		{Holdem, "6h As Ac Qd Qs", "%s", "Two Pair, Aces over Queens, kicker Six"},
		{Holdem, "6h As Ac Qd Qs", "%S", "Two Pair, Aces over Queens"},
		{Holdem, "6h As Ac Qd Qs", "%e", "Two Pair"},
		{Holdem, "6h 6s Ac Qd Ks", "%s", "Pair, Sixes, kickers Ace, King, Queen"},
		{Holdem, "6h 6s Ac Qd Ks", "%S", "Pair, Sixes"},
		{Holdem, "6h 6s Ac Qd Ks", "%e", "Pair"},
		{Holdem, "7h 5c 4h 2h As", "%s", "Ace-high, kickers Seven, Five, Four, Two"},
		{Holdem, "7h 5c 4h 2h As", "%S", "Ace-high"},
		{Holdem, "7h 5c 4h 2h As", "%e", "Ace-high"},
		{Lowball, "7h 2h 3c 4h 6h", "%s", "Seven, Six, Four, Three, Two-low, No. 2"},
		{Lowball, "7h 2h 3c 4h 6h", "%S", "Seven-low, No. 2"},
		{Lowball, "7h 2h 3c 4h 6h", "%e", "Seven-low"},
		{Lowball, "6h 5c 3s 2h 4h", "%s", "Straight, Six-high"},
		{Lowball, "6h 5c 3s 2h 4h", "%S", "Straight, Six-high"},
		{Lowball, "6h 5c 3s 2h 4h", "%e", "Straight"},
		{Lowball, "7s 4h 6h 5c 3s", "%s", "Straight, Seven-high"},
		{Lowball, "7s 4h 6h 5c 3s", "%S", "Straight, Seven-high"},
		{Lowball, "7s 4h 6h 5c 3s", "%e", "Straight"},
		{Razz, "5h 4h 3h 2h Ah", "%s", "Five, Four, Three, Two, Ace-low"},
		{Razz, "5h 4h 3h 2h Ah", "%S", "Five-low"},
		{Razz, "5h 4h 3h 2h Ah", "%e", "Five-low"},
		{Soko, "4h Th 6h 9c 7h", "%s", "Four Flush, Ten-high, kickers Seven, Six, Four, Nine"},
		{Soko, "4h Th 6h 9c 7h", "%S", "Four Flush, Ten-high"},
		{Soko, "4h Th 6h 9c 7h", "%e", "Four Flush"},
		{Soko, "5c Qh 4h 3c 2c", "%s", "Four Straight, Five-high, kicker Queen"},
		{Soko, "5c Qh 4h 3c 2c", "%S", "Four Straight, Five-high"},
		{Soko, "5c Qh 4h 3c 2c", "%e", "Four Straight"},
	}
	for i, test := range tests {
		ev := test.typ.Eval(Must(test.v), nil)
		desc := ev.Desc(false)
		if s, exp := fmt.Sprintf(test.s, desc), test.exp; s != test.exp {
			t.Errorf("test %d %s %q expected %q, got: %q", i, test.typ, test.v, exp, s)
		}
	}
}

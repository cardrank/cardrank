package cardrank

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSokoCards(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if s := os.Getenv("TESTS"); !strings.Contains(s, "soko") && !strings.Contains(s, "all") {
		t.Skip("skipping: $ENV{TESTS} does not contain 'soko' or 'all'")
	}
	u, f, ev, uv := shuffled(DeckFrench), NewSokoEval(false), EvalOf(Soko), EvalOf(Soko)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						v := []Card{u[c0], u[c1], u[c2], u[c3], u[c4]}
						f(ev, v, nil)
						switch r := ev.HiRank; {
						case r == 0, r == Invalid:
							t.Fatalf("%v expected valid rank, got: %d", v, r)
						case r <= TwoPair:
						case hasFlush4(v):
							if r <= TwoPair || sokoFlush < r {
								t.Errorf("%v expected four flush %d < r <= %d, got: %d", v, TwoPair, sokoFlush, r)
							}
						case hasStraight4(v):
							if r <= sokoFlush || sokoStraight < r {
								t.Errorf("%v expected four straight %d < r <= %d, got: %d", v, sokoFlush, sokoStraight, r)
							}
						case sokoNothing < r:
							t.Errorf("%v expected nothing r <= %d, got: %d", v, sokoNothing, r)
						}
						u := make([]Card, 5)
						copy(u, v)
						for k := 0; k < 3; k++ {
							r.Shuffle(5, func(i, j int) {
								u[i], u[j] = u[j], u[i]
							})
						}
						f(uv, u, nil)
						if ev.HiRank != uv.HiRank {
							t.Fatalf("expected equal ranks for %v %v, got: %d", v, u, uv.HiRank)
						}
						if s, z := fmt.Sprintf("%s", ev), fmt.Sprintf("%s", uv); s != z {
							t.Errorf("expected %q == %q %v %v", s, z, v, u)
						}
					}
				}
			}
		}
	}
}

func TestRankSoko(t *testing.T) {
	t.Logf("flush4: %d straight4: %d", len(sokoFlush4), len(sokoStraight4))
	tests := []struct {
		a   string
		b   string
		exp EvalRank
	}{
		{"Ah Kh Ks Qh Jh", "Ad Kd Kh Qd Jd", 3327},
		{"Ah Qd Ks Jh As", "Ad Qh Kh Jd Ac", 12621},
		{"Ah Qd Jh Th 8c", "8d Ac Qh Jc Tc", 15777},
	}
	for i, test := range tests {
		a, b := Must(test.a), Must(test.b)
		if r, exp := RankSoko(a[0], a[1], a[2], a[3], a[4]), test.exp; r != exp {
			t.Errorf("test %d expected %d, got: %d", i, exp, r)
		}
		if r, exp := RankSoko(b[0], b[1], b[2], b[3], b[4]), test.exp; r != exp {
			t.Errorf("test %d expected %d, got: %d", i, exp, r)
		}
	}
}

func hasFlush4(v []Card) bool {
	for i := 0; i < 5; i++ {
		c0, c1, c2, c3 := v[i%5], v[(i+1)%5], v[(i+2)%5], v[(i+3)%5]
		if c0&c1&c2&c3&0xf000 != 0 {
			return true
		}
	}
	return false
}

var straight4 map[Card]bool

func init() {
	straight4 = make(map[Card]bool)
	for i, r := 0, 9; r >= 0; i, r = i+1, r-1 {
		straight4[0xf<<r] = true
	}
}

func hasStraight4(v []Card) bool {
	for i := 0; i < 5; i++ {
		c0, c1, c2, c3 := v[i%5], v[(i+1)%5], v[(i+2)%5], v[(i+3)%5]
		if straight4[(c0|c1|c2|c3)>>16] {
			return true
		}
	}
	return false
}

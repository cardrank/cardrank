package cardrank

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCactus(t *testing.T) {
	if s := os.Getenv("TESTS"); !strings.Contains(s, "cactus") && !strings.Contains(s, "all") {
		t.Skip("skipping: $ENV{TESTS} does not contain 'cactus' or 'all'")
	}
	if cactus == nil {
		t.Skip("skipping: cactus is not available")
	}
	const total = 133784560
	t.Parallel()
	u, f, tests, exp, ev := shuffled(DeckFrench), NewEval(cactus), cactusTests(false, true), EvalOf(Holdem), EvalOf(Holdem)
	v := make([]Card, 7)
	var i int
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								v[0], v[1], v[2], v[3], v[4], v[5], v[6] = u[c0], u[c1], u[c2], u[c3], u[c4], u[c5], u[c6]
								exp.HiRank = Invalid
								f(exp, v, nil)
								if r := exp.HiRank; r == 0 || r == Invalid {
									t.Fatalf("test cactus %v expected valid rank, got: %d", v, r)
								}
								bestCactus(exp.HiRank, exp.HiBest, exp.HiUnused, 0, nil)
								expDesc := fmt.Sprintf("%s", exp.Desc(false))
								for _, test := range tests {
									ev.HiRank = Invalid
									test.eval(ev, v, nil)
									desc := fmt.Sprintf("%s", ev.Desc(false))
									switch r, exp := ev.HiRank, exp.HiRank; {
									case r == 0, r == Invalid:
										t.Errorf("test %s %v expected valid rank, got: %d", test.name, v, r)
									case r != exp:
										t.Errorf("test %s %v expected %d, got: %d", test.name, v, exp, r)
									case desc != expDesc:
										t.Errorf("test %s %v expected %q, got: %q", test.name, v, expDesc, desc)
									}
								}
								i++
								if i%1000000 == 0 {
									t.Logf("%.0f%%", float64(i)/total*100)
								}
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d/%d", i, total)
}

func TestNextBitPermutation(t *testing.T) {
	n := uint32(31)
	for _, exp := range []uint32{47, 55, 59, 61, 62} {
		if n = nextBitPermutation(n); n != exp {
			t.Errorf("expected n == %d, got: %d", exp, n)
		}
	}
}

func TestPrimeProduct(t *testing.T) {
	tests := []struct {
		v   []Card
		exp uint32
	}{
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x21103, 0x22103}, 0x2ccbb},
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x21103, 0x00001}, 0x0eee9},
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x00001, 0x00001}, 0x04fa3},
	}
	for i, test := range tests {
		if n, exp := primeProduct(test.v[0], test.v[1], test.v[2], test.v[3], test.v[4]), test.exp; n != exp {
			t.Errorf("test %d %v expected %d == %d", i, test.v, exp, n)
		}
	}
}

func TestPrimeProductBits(t *testing.T) {
	tests := []struct {
		bits uint32
		exp  uint32
	}{
		{0x079, 0x84f2},
		{0x158, 0x759b},
		{0x037, 0x10c2},
		{0x01f, 0x0906},
		{0x04e, 0x06f9},
		{0x063, 0x052e},
		{0x015, 0x006e},
		{0x00b, 0x002a},
		{0x001, 0x0002},
	}
	for i, test := range tests {
		if n, exp := primeProductBits(test.bits), test.exp; n != exp {
			t.Errorf("test %d expected primeProductBits(%d) == %d, got: %d", i, test.bits, exp, n)
		}
	}
}

func TestCactusMaps(t *testing.T) {
	flushes, unique5 := cactusMaps()
	if n, exp := len(flushes), 1287; n != exp {
		t.Fatalf("expected len(flush) == %d, got: %d", exp, n)
	}
	flushesTests := []struct {
		r   uint32
		exp EvalRank
	}{
		{0x005ffe37, 0x0184},
		{0x003d1623, 0x0185},
		{0x003c6619, 0x0340},
		{0x003a5e5d, 0x0334},
		{0x00345631, 0x01b1},
		{0x0029f659, 0x047b},
		{0x0017ae13, 0x0167},
		{0x00166e15, 0x01b3},
		{0x0014b621, 0x03ba},
		{0x00104625, 0x026f},
		{0x000b2e27, 0x01aa},
		{0x0009ee29, 0x01d5},
		{0x00092e22, 0x0274},
		{0x00044e32, 0x016e},
		{0x0003f7af, 0x05cf},
		{0x0003a61e, 0x04a5},
		{0x00026e35, 0x01e3},
		{0x00021f1b, 0x04f7},
		{0x0001f617, 0x0307},
		{0x0001ce36, 0x01dc},
		{0x0000b3b2, 0x031d},
		{0x0000725a, 0x0320},
		{0x00006f4a, 0x0475},
		{0x00004e2a, 0x0580},
		{0x0000310e, 0x0605},
	}
	for i, test := range flushesTests {
		if n := flushes[test.r]; n != test.exp {
			t.Fatalf("test %d flush[%d] to be %d, got: %d", i, test.r, test.exp, n)
		}
	}
	if n, exp := len(unique5), 6175; n != exp {
		t.Fatalf("expected len(unique5) == %d, got: %d", exp, n)
	}
	unique5Tests := []struct {
		r   uint32
		exp EvalRank
	}{
		{0x01c51151, 0x0a3e},
		{0x005f112f, 0x0695},
		{0x00529143, 0x0d0f},
		{0x0021110f, 0x0ee4},
		{0x001f912d, 0x1af7},
		{0x0017912b, 0x0826},
		{0x00171a57, 0x0c19},
		{0x0010914d, 0x0059},
		{0x000c1147, 0x0787},
		{0x000b1101, 0x0beb},
		{0x00049123, 0x13b1},
		{0x0003915a, 0x112b},
		{0x000382e3, 0x110f},
		{0x00029135, 0x1cc6},
		{0x0002112e, 0x1bc6},
		{0x0001913e, 0x1c60},
		{0x00019113, 0x0a89},
		{0x00011132, 0x0c2c},
		{0x00002d8e, 0x08d3},
		{0x0000115e, 0x1733},
		{0x00001144, 0x1818},
		{0x00001138, 0x0981},
		{0x00001117, 0x008c},
		{0x000010fe, 0x1639},
		{0x000010fb, 0x094a},
	}
	for i, test := range unique5Tests {
		if n, exp := unique5[test.r], test.exp; n != exp {
			t.Fatalf("test %d unique5[%d] to be %d, got: %d", i, test.r, exp, n)
		}
	}
}

func TestSokoCards(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if s := os.Getenv("TESTS"); !strings.Contains(s, "soko") && !strings.Contains(s, "all") {
		t.Skip("skipping: $ENV{TESTS} does not contain 'soko' or 'all'")
	}
	t.Parallel()
	u, f, ev, uv := shuffled(DeckFrench), NewSokoEval(true, false), EvalOf(Soko), EvalOf(Soko)
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
	for r := 9; 0 <= r; r-- {
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

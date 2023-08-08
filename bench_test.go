package cardrank

import (
	"fmt"
	"testing"
)

func BenchmarkCactus(b *testing.B) {
	for _, test := range cactusTests(true, false) {
		for i, f := range []func(*testing.B, EvalFunc, int){
			bench5, bench6, bench7,
		} {
			b.Run(fmt.Sprintf("%s/%d", test.name, i+5), func(b *testing.B) {
				f(b, test.eval, b.N)
			})
		}
	}
}

func bench5(b *testing.B, f EvalFunc, n int) {
	b.Helper()
	u, count, ev, v := shuffled(DeckFrench), 0, EvalOf(Holdem), make([]Card, 5)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						ev.HiRank, ev.LoRank = Invalid, Invalid
						v[0], v[1], v[2], v[3], v[4] = u[c0], u[c1], u[c2], u[c3], u[c4]
						f(ev, v[:2], v[2:])
						if benchR = ev.HiRank; benchR == 0 || benchR == Invalid || HighCard < benchR {
							b.Fail()
						}
						count++
						if n <= count {
							b.Logf("count: %d", count)
							return
						}
					}
				}
			}
		}
	}
}

func bench6(b *testing.B, f EvalFunc, n int) {
	b.Helper()
	u, count, ev, v := shuffled(DeckFrench), 0, EvalOf(Holdem), make([]Card, 6)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							ev.HiRank, ev.LoRank = Invalid, Invalid
							v[0], v[1], v[2], v[3], v[4], v[5] = u[c0], u[c1], u[c2], u[c3], u[c4], u[c5]
							f(ev, v[:2], v[2:])
							if benchR = ev.HiRank; benchR == 0 || benchR == Invalid || HighCard < benchR {
								b.Fail()
							}
							count++
							if n <= count {
								b.Logf("count: %d", count)
								return
							}
						}
					}
				}
			}
		}
	}
}

func bench7(b *testing.B, f EvalFunc, n int) {
	b.Helper()
	u, count, ev, v := shuffled(DeckFrench), 0, EvalOf(Holdem), make([]Card, 7)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								ev.HiRank, ev.LoRank = Invalid, Invalid
								v[0], v[1], v[2], v[3], v[4], v[5], v[6] = u[c0], u[c1], u[c2], u[c3], u[c4], u[c5], u[c6]
								f(ev, v[:2], v[2:])
								if benchR = ev.HiRank; benchR == 0 || benchR == Invalid || HighCard < benchR {
									b.Fail()
								}
								count++
								if n <= count {
									b.Logf("count: %d", count)
									return
								}
							}
						}
					}
				}
			}
		}
	}
}

func BenchmarkType(b *testing.B) {
	e, l := make(map[EvalType]bool), make(map[EvalType]bool)
	for _, typ := range Types() {
		low, etyp := typ.Low(), typ.Desc().Eval
		switch {
		// skip already evaluated types and Jacks-or-better
		case typ.Double(),
			low && l[etyp],
			!low && e[etyp],
			etyp == EvalJacksOrBetter:
			continue
		}
		e[etyp] = true
		if low {
			l[etyp] = true
		}
		name := etyp.Name()
		if !low {
			name += "/Hi"
		} else {
			name += "/HiLo"
		}
		b.Run(name+"/eval", func(b *testing.B) {
			benchType(b, typ, b.N, true)
		})
		b.Run(name+"/calc", func(b *testing.B) {
			benchType(b, typ, b.N, false)
		})
	}
}

func benchType(b *testing.B, typ Type, n int, eval bool) {
	b.Helper()
	v, p, m, count, ev := shuffled(typ.DeckType()), typ.Pocket(), typ.Board(), 0, EvalOf(typ)
	var f EvalFunc
	if eval {
		f = evals[typ]
	} else {
		f = calcs[typ]
	}
	for {
		for i := 0; i < len(v)-p-m; i++ {
			ev.HiRank, ev.LoRank = Invalid, Invalid
			f(ev, v[i:i+p], v[i+p:i+p+m])
			if benchE = ev.HiRank; benchE == 0 || benchE == Invalid {
				b.Fail()
			}
		}
		count++
		if n <= count {
			b.Logf("count: %d", count)
			return
		}
	}
}

var (
	benchR EvalRank
	benchE EvalRank
)

package cardrank

import (
	"fmt"
	"testing"
)

func BenchmarkCactus(b *testing.B) {
	for _, t := range cactusTests(true, false) {
		for i, benchf := range []func(*testing.B, EvalFunc, int){
			bench5, bench6, bench7,
		} {
			test, f := t, benchf
			b.Run(fmt.Sprintf("%s/%d", test.name, i+5), func(b *testing.B) {
				f(b, test.eval, b.N)
			})
		}
	}
}

func bench5(b *testing.B, f EvalFunc, n int) {
	b.Helper()
	v, count, ev := shuffled(DeckFrench), 0, EvalOf(Holdem)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						f(ev, []Card{v[c0], v[c1], v[c2], v[c3], v[c4]}, nil)
						if benchR = ev.HiRank; benchR == 0 || benchR == Invalid || HighCard < benchR {
							b.Fail()
						}
						count++
						if n <= count {
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
	v, count, ev := shuffled(DeckFrench), 0, EvalOf(Holdem)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							f(ev, []Card{v[c0], v[c1], v[c2], v[c3], v[c4], v[c5]}, nil)
							if benchR = ev.HiRank; benchR == 0 || benchR == Invalid || HighCard < benchR {
								b.Fail()
							}
							count++
							if n <= count {
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
	v, count, ev := shuffled(DeckFrench), 0, EvalOf(Holdem)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								ev.HiRank, ev.LoRank = Invalid, Invalid
								f(ev, []Card{v[c0], v[c1], v[c2], v[c3], v[c4], v[c5], v[c6]}, nil)
								if benchR == 0 || benchR == Invalid || HighCard < benchR {
									b.Fail()
								}
								count++
								if n <= count {
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
	for _, t := range Types() {
		typ := t
		low, etyp := typ.Low(), typ.Desc().Eval
		switch {
		case typ.Double(), low && l[etyp], !low && e[etyp], etyp == EvalJacksOrBetter:
			// skip already evaluated types and Jacks-or-better
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
			name += "/Lo"
		}
		b.Run(name, func(b *testing.B) {
			benchType(b, typ, b.N)
		})
	}
}

func benchType(b *testing.B, typ Type, n int) {
	b.Helper()
	v, p, m, count, ev := shuffled(typ.DeckType()), typ.Pocket(), typ.Board(), 0, EvalOf(typ)
	f := evals[typ]
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
			return
		}
	}
}

var (
	benchR EvalRank
	benchE EvalRank
)

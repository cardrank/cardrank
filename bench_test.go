package cardrank

import (
	"fmt"
	"testing"
)

func BenchmarkRank(b *testing.B) {
	for _, t := range cactusTests(true) {
		for i, benchf := range []func(*testing.B, CactusFunc, int){
			bench5, bench6, bench7,
		} {
			test, f := t, benchf
			b.Run(fmt.Sprintf("%s/%d", test.name, i+5), func(b *testing.B) {
				f(b, test.eval, b.N)
			})
		}
	}
}

func BenchmarkEval(b *testing.B) {
	e, l := make(map[EvalType]bool), make(map[EvalType]bool)
	for _, t := range Types() {
		typ := t
		low, etyp := typ.Low(), typ.TypeDesc().Eval
		switch {
		case typ.Double(), low && l[etyp], !low && e[etyp]:
			// skip already evaluated types
			continue
		}
		e[etyp] = true
		if low {
			l[etyp] = true
		}
		name := etyp.Name() + "/Hi"
		if low {
			name += "Lo"
		}
		b.Run(name, func(b *testing.B) {
			benchEval(b, typ, b.N)
		})
	}
}

func bench5(b *testing.B, f CactusFunc, n int) {
	b.Helper()
	v, count := shuffled(DeckFrench), 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						benchR = f([]Card{v[c0], v[c1], v[c2], v[c3], v[c4]})
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

func bench6(b *testing.B, f CactusFunc, n int) {
	b.Helper()
	v, count := shuffled(DeckFrench), 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							benchR = f([]Card{v[c0], v[c1], v[c2], v[c3], v[c4], v[c5]})
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

func bench7(b *testing.B, f CactusFunc, n int) {
	b.Helper()
	v, count := shuffled(DeckFrench), 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								benchR = f([]Card{v[c0], v[c1], v[c2], v[c3], v[c4], v[c5], v[c6]})
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

func benchEval(b *testing.B, typ Type, n int) {
	b.Helper()
	v, p, m, count := shuffled(typ.DeckType()), typ.Pocket(), typ.Board(), 0
	for {
		for i := 0; i < len(v)-p-m; i++ {
			benchH = EvalOf(typ).Eval(v[i:i+p], v[i+p:i+p+m])
			switch {
			case benchH.HiRank == 0,
				benchH.HiRank == Invalid,
				benchH.LoRank == 0:
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
	benchH *Eval
)

package cardrank

import (
	"fmt"
	"testing"
)

func Benchmark(b *testing.B) {
	for _, evalTest := range evalTests(true) {
		for i, benchf := range []func(*testing.B, HandRankFunc, int){
			bench5, bench6, bench7,
		} {
			test, f := evalTest, benchf
			b.Run(fmt.Sprintf("%s/%d", test.name, i+5), func(b *testing.B) {
				f(b, test.eval, b.N)
			})
		}
	}
}

func bench5(b *testing.B, f HandRankFunc, n int) {
	count := 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						benchR = f([]Card{allCards[c0], allCards[c1], allCards[c2], allCards[c3], allCards[c4]})
						if benchR > HighCard {
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

func bench6(b *testing.B, f HandRankFunc, n int) {
	count := 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							benchR = f([]Card{allCards[c0], allCards[c1], allCards[c2], allCards[c3], allCards[c4], allCards[c5]})
							if benchR > HighCard {
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

func bench7(b *testing.B, f HandRankFunc, n int) {
	count := 0
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								benchR = f([]Card{allCards[c0], allCards[c1], allCards[c2], allCards[c3], allCards[c4], allCards[c5], allCards[c6]})
								if benchR > HighCard {
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

var benchR HandRank

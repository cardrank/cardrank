package cardrank

// Soko maps.
var (
	sokoFlush4    map[uint32]EvalRank
	sokoStraight4 map[uint32]EvalRank
)

func init() {
	sokoFlush4, sokoStraight4 = sokoMaps()
}

// sokoMaps generates Soko flush4 and straight4 maps.
//
// See: https://www.denexa.com/blog/soko-canadian-stud/
func sokoMaps() (map[uint32]EvalRank, map[uint32]EvalRank) {
	flush4, straight4 := make(map[uint32]EvalRank), make(map[uint32]EvalRank)
	// calculate flush rank offset
	for i, r0 := 0, 12; r0 >= 0; r0-- {
		for r1 := r0 - 1; r1 >= 0; r1-- {
			for r2 := r1 - 1; r2 >= 0; r2-- {
				for r3 := r2 - 1; r3 >= 0; i, r3 = i+1, r3-1 {
					flush4[1<<r0|1<<r1|1<<r2|1<<r3] = 1 + TwoPair + 13*EvalRank(i)
				}
			}
		}
	}
	// calculate straight rank offset
	// only 10 straights, as there is no 4 card ace low straight
	for i, r := 0, 9; r >= 0; i, r = i+1, r-1 {
		straight4[0xf<<r] = 1 + TwoPair + EvalRank(13*len(flush4)) + 13*EvalRank(i)
	}
	return flush4, straight4
}

// RankSoko is a Soko eval rank func.
//
// Has similar orders to Cactus, adding a Four Flush and Four Straight that
// beat Pairs and Nothing:
//
//	Straight Flush
//	Four of a Kind
//	Full House
//	Flush
//	Straight
//	Three of a Kind
//	Two Pair
//	Four Flush
//	Four Straight
//	Pair
//	Nothing
func RankSoko(c0, c1, c2, c3, c4 Card) EvalRank {
	rank := RankCactus(c0, c1, c2, c3, c4)
	if rank <= TwoPair {
		return rank
	}
	r, v := Invalid, []Card{c0, c1, c2, c3, c4}
	for c, i := EvalRank(0), 0; i < 5; i++ {
		c0, c1, c2, c3, c4 = v[i%5], v[(i+1)%5], v[(i+2)%5], v[(i+3)%5], v[(i+4)%5]
		if c0&c1&c2&c3&0xf000 != 0 {
			// four flush
			if c = sokoFlush4[uint32(c0|c1|c2|c3)>>16] + EvalRank(Ace-c4.Rank()); c < r {
				r = c
			}
		} else if c, ok := sokoStraight4[uint32(c0|c1|c2|c3)>>16]; ok {
			// four straight
			if c += EvalRank(Ace - c4.Rank()); c < r {
				r = c
			}
		}
	}
	if r != Invalid {
		return r
	}
	return 1 + sokoStraight - TwoPair + rank
}

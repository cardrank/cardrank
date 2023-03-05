package cardrank

var (
	flush5        map[uint32]EvalRank
	unique5       map[uint32]EvalRank
	sokoFlush4    map[uint32]EvalRank
	sokoStraight4 map[uint32]EvalRank
)

func init() {
	flush5, unique5 = cactusMaps()
	sokoFlush4, sokoStraight4 = sokoMaps()
	cactus = Cactus
}

// Cactus is a Cactus Kev rank eval func, using lookup maps generated on the
// fly.
//
// See: https://archive.is/G6GZg
func Cactus(c0, c1, c2, c3, c4 Card) EvalRank {
	if c0&c1&c2&c3&c4&0xf000 != 0 {
		return flush5[primeProductBits(uint32(c0|c1|c2|c3|c4)>>16)]
	}
	return unique5[primeProduct(c0, c1, c2, c3, c4)]
}

// cactusMaps builds the cactus flush and unique5 maps.
func cactusMaps() (map[uint32]EvalRank, map[uint32]EvalRank) {
	flush5, unique5 := make(map[uint32]EvalRank), make(map[uint32]EvalRank)
	// straight orders
	orders := [10]uint32{
		0x1f00, // royal
		0x0f80, // king
		0x07c0, // queen
		0x03e0, // jack
		0x01f0, // ten
		0x00f8, // nine
		0x007c, // eight
		0x003e, // seven
		0x001f, // six
		0x100f, // steel wheel
	}
	var r []uint32
	for i, n := 0, uint32(0x1f); i < 1286; i++ { // 1276 + len(orders)
		n = nextBitPermutation(n)
		var sflush bool
		for _, j := range orders {
			if n^j == 0 {
				sflush = true
				break
			}
		}
		if !sflush {
			r = append(r, n)
		}
	}
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	for i := 0; i < len(orders); i++ {
		// straight flush
		flush5[primeProductBits(orders[i])] = 1 + EvalRank(i)
		// straight
		unique5[primeProductBits(orders[i])] = 1 + Flush + EvalRank(i)
	}
	for i := 0; i < len(r); i++ {
		// flush
		flush5[primeProductBits(r[i])] = 1 + FullHouse + EvalRank(i)
		// nothing (high cards)
		unique5[primeProductBits(r[i])] = 1 + Pair + EvalRank(i)
	}
	v := [13]int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	kickers := func(z []int, n int) []int {
		k := make([]int, len(z))
		copy(k, z)
		for i := 0; i < len(k); i++ {
			if k[i] == v[n] {
				k = append(k[:i], k[i+1:]...)
				break
			}
		}
		return k
	}
	for i, r3, r2, r1 := 0, 1+Straight, 1+ThreeOfAKind, 1+TwoPair; i < 13; i++ {
		k := kickers(v[:], i)
		for j, n := range k {
			// four of a kind
			unique5[primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[n]] = 1 + StraightFlush + EvalRank(i*len(k)+j)
			// full house
			unique5[primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[n]*primes[n]] = 1 + FourOfAKind + EvalRank(i*len(k)+j)
		}
		// three of a kind
		for j := 0; j < len(k)-1; j++ {
			for l := j + 1; l < len(k); l++ {
				unique5[primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[k[j]]*primes[k[l]]] = r3
				r3++
			}
		}
		// two pair
		for j := i + 1; j < 13; j++ {
			for _, n := range kickers(k, j) {
				unique5[primes[v[i]]*primes[v[i]]*primes[v[j]]*primes[v[j]]*primes[n]] = r2
				r2++
			}
		}
		// pair
		for l := 0; l < len(k)-2; l++ {
			for m := l + 1; m < len(k)-1; m++ {
				for n := m + 1; n < len(k); n++ {
					unique5[primes[v[i]]*primes[v[i]]*primes[k[l]]*primes[k[m]]*primes[k[n]]] = r1
					r1++
				}
			}
		}
	}
	return flush5, unique5
}

// RankSoko is a [Soko] rank eval func.
//
// Has ranks to [Cactus], adding a Four Flush and Four Straight that beat
// [Pair]'s and [Nothing]:
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

// sokoMaps generates [Soko] flush4 and straight4 maps.
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

// nextBitPermutation calculates the lexicographical next bit permutation.
//
// See: https://graphics.stanford.edu/~seander/bithacks.html#NextBitPermutation.
func nextBitPermutation(bits uint32) uint32 {
	i := (bits | (bits - 1)) + 1
	return i | ((((i & -i) / (bits & -bits)) >> 1) - 1)
}

// primeProduct returns the prime product of the cards.
func primeProduct(c0, c1, c2, c3, c4 Card) uint32 {
	i := uint32(c0) & 0xff
	i *= uint32(c1) & 0xff
	i *= uint32(c2) & 0xff
	i *= uint32(c3) & 0xff
	i *= uint32(c4) & 0xff
	return i
}

// primeProductBits returns the prime product of the rank bits.
func primeProductBits(bits uint32) uint32 {
	i := uint32(1)
	for j := 0; j < 13; j++ {
		if bits&(1<<j) != 0 {
			i *= primes[j]
		}
	}
	return i
}

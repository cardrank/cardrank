package cardrank

func init() {
	flushes, unique5 = cactusMaps()
	cactus = Cactus
}

// flushes is the flush map.
var flushes map[uint32]HandRank

// unique5 is the unique5 map.
var unique5 map[uint32]HandRank

// Cactus is a Cactus Kev hand rank func that generates the lookup maps on the
// fly.
//
// See: https://archive.is/G6GZg
func Cactus(c0, c1, c2, c3, c4 Card) HandRank {
	if c0&c1&c2&c3&c4&0xf000 != 0 {
		return flushes[primeProductBits(uint32(c0|c1|c2|c3|c4)>>16)]
	}
	return unique5[primeProduct(c0, c1, c2, c3, c4)]
}

// cactusMaps builds the cactus flush and unique5 maps.
func cactusMaps() (map[uint32]HandRank, map[uint32]HandRank) {
	flushes, unique5 := make(map[uint32]HandRank), make(map[uint32]HandRank)
	// rank orders
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
		flushes[primeProductBits(orders[i])] = 1 + HandRank(i)
		// straight
		unique5[primeProductBits(orders[i])] = 1 + Flush + HandRank(i)
	}
	for i := 0; i < len(r); i++ {
		// flush
		flushes[primeProductBits(r[i])] = 1 + FullHouse + HandRank(i)
		// nothing (high cards)
		unique5[primeProductBits(r[i])] = 1 + Pair + HandRank(i)
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
			unique5[primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[n]] = 1 + StraightFlush + HandRank(i*len(k)+j)
			// full house
			unique5[primes[v[i]]*primes[v[i]]*primes[v[i]]*primes[n]*primes[n]] = 1 + FourOfAKind + HandRank(i*len(k)+j)
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
	return flushes, unique5
}

// nextBitPermutation calculates the lexicographical next bit permutation.
//
// See: https://graphics.stanford.edu/~seander/bithacks.html#NextBitPermutation.
func nextBitPermutation(bits uint32) uint32 {
	i := (bits | (bits - 1)) + 1
	return i | ((((i & -i) / (bits & -bits)) >> 1) - 1)
}

// primeProduct returns the prime product of the hand.
func primeProduct(c0, c1, c2, c3, c4 Card) uint32 {
	i := uint32(c0) & 0xff
	i *= uint32(c1) & 0xff
	i *= uint32(c2) & 0xff
	i *= uint32(c3) & 0xff
	i *= uint32(c4) & 0xff
	return i
}

// primeProductBits returns the prime product of the hand's rank bits.
func primeProductBits(bits uint32) uint32 {
	i := uint32(1)
	for j := 0; j < 13; j++ {
		if bits&(1<<j) != 0 {
			i *= uint32(primes[j])
		}
	}
	return i
}

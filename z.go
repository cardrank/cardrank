//go:build !noinit

package cardrank

func init() {
	switch {
	case twoPlus != nil && cactusFast != nil:
		DefaultRanker = HybridRanker(cactusFast, twoPlus)
	case cactusFast != nil:
		DefaultRanker = HandRanker(cactusFast)
	case cactus != nil:
		DefaultRanker = HandRanker(cactus)
	}
	switch {
	case cactusFast != nil:
		DefaultSixPlusRanker = HandRanker(SixPlusRanker(cactusFast))
	case cactus != nil:
		DefaultSixPlusRanker = HandRanker(SixPlusRanker(cactus))
	}
}

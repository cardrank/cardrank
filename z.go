//go:build !noinit

package cardrank

func init() {
	switch {
	case twoPlusTwo != nil && cactusFast != nil:
		DefaultRanker = HybridRanker(cactusFast, twoPlusTwo)
	case cactusFast != nil:
		DefaultRanker = HandRanker(cactusFast)
	case cactus != nil:
		DefaultRanker = HandRanker(cactus)
	}
	switch {
	case cactusFast != nil:
		DefaultCactus, DefaultSixPlusRanker = cactusFast, HandRanker(SixPlusRanker(cactusFast))
	case cactus != nil:
		DefaultCactus, DefaultSixPlusRanker = cactus, HandRanker(SixPlusRanker(cactus))
	}
}

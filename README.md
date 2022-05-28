# cardrank.io/cardrank

Package `cardrank.io/cardrank` provides a library of types, funcs, and
utilities for working with playing cards, card decks, and evaluating poker hand
ranks. Supports [Texas Holdem][holdem-example], [Texas Holdem Short Deck
(6-plus)][short-deck-example], [Omaha][omaha-example], [Omaha
Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud
Hi/Lo][stud-hi-lo-example], and [Razz][razz-example] (everything needed for
HORSE).

[![GoDoc](https://godoc.org/cardrank.io/cardrank?status.svg)](https://godoc.org/cardrank.io/cardrank)
[![Tests on Linux, MacOS and Windows](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/cardrank.io/cardrank)](https://goreportcard.com/report/cardrank.io/cardrank)

## Overview

High-level types, funcs, and standardized interfaces are included in the
package to deal and evaluate hands of poker, including all necessary types for
representing and working with [cards][card], [card suits][suit], [card
ranks][rank], [card decks][deck], and [hands of cards][hand]. Hand evaluation
is achieved with pure Go implementations of [common poker hand rank
evaluators][hand-ranking].

Hands of [Texas Holdem][holdem-example], [Texas Holdem Short Deck
(6-Plus)][short-deck-example], [Omaha][omaha-example], [Omaha
Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud
Hi/Lo][stud-hi-lo-example], and [Razz][razz-example] are easily created and
dealt using standardized interfaces and logic, with winners [being easily
determined and ordered][order].

[Development of additional poker variants][future], such as Badugi and Candian
Stud (Sökö), is planned.

## Using

See [Go documentation][pkg].

```sh
go get cardrank.io/cardrank
```

### Examples

Complete examples for [Texas Holdem][holdem-example], [Texas Holdem Short Deck
(6-plus)][short-deck-example], [Omaha][omaha-example], [Omaha
Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud
Hi/Lo][stud-hi-lo-example] are available in the source repository. [Further
examples][examples] are available in the [Go package documentation][pkg] for
overviews of using the package's types, funcs and interfaces.

Below are quick examples for Texas Holdem and Omaha Hi/Lo:

#### Texas Holdem

```go
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cardrank.io/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.Holdem.Deal(rnd.Shuffle, players)
	hands := cardrank.Holdem.RankHands(pockets, board)
	fmt.Printf("------ Holdem %d ------\n", seed)
	fmt.Printf("Board:    %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b %s %b %b\n", i+1, hands[i].Pocket(), hands[i].Description(), hands[i].Best(), hands[i].Unused())
	}
	h, pivot := cardrank.Order(hands)
	if pivot == 1 {
		fmt.Printf("Result:   Player %d wins with %s %b\n", h[0]+1, hands[h[0]].Description(), hands[h[0]].Best())
	} else {
		var s, b []string
		for j := 0; j < pivot; j++ {
			s = append(s, strconv.Itoa(h[j]+1))
			b = append(b, fmt.Sprintf("%b", hands[h[j]].Best()))
		}
		fmt.Printf("Result:   Players %s push with %s %s\n", strings.Join(s, ", "), hands[h[0]].Description(), strings.Join(b, ", "))
	}
}
```

Output:

```txt
------ Holdem 1653086388531833787 ------
Board:    [A♣ 3♥ Q♥ J♣ J♠]
Player 1: [5♣ 5♦] Two Pair, Jacks over Fives, kicker Ace [J♣ J♠ 5♣ 5♦ A♣] [Q♥ 3♥]
Player 2: [A♥ 7♠] Two Pair, Aces over Jacks, kicker Queen [A♣ A♥ J♣ J♠ Q♥] [7♠ 3♥]
Player 3: [8♥ 3♣] Two Pair, Jacks over Threes, kicker Ace [J♣ J♠ 3♣ 3♥ A♣] [Q♥ 8♥]
Player 4: [9♣ T♠] Pair, Jacks, kickers Ace, Queen, Ten [J♣ J♠ A♣ Q♥ T♠] [9♣ 3♥]
Player 5: [6♣ J♥] Three of a Kind, Jacks, kickers Ace, Queen [J♣ J♥ J♠ A♣ Q♥] [6♣ 3♥]
Player 6: [2♣ T♦] Pair, Jacks, kickers Ace, Queen, Ten [J♣ J♠ A♣ Q♥ T♦] [3♥ 2♣]
Result:   Player 5 wins with Three of a Kind, Jacks, kickers Ace, Queen [J♣ J♥ J♠ A♣ Q♥]
```

#### Omaha Hi/Lo

```go
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cardrank.io/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.OmahaHiLo.Deal(rnd.Shuffle, players)
	hands := cardrank.OmahaHiLo.RankHands(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b\n", i+1, pockets[i])
		fmt.Printf("  Hi: %s %b %b\n", hands[i].Description(), hands[i].Best(), hands[i].Unused())
		if hands[i].LowValid() {
			fmt.Printf("  Lo: %s %b %b\n", hands[i].LowDescription(), hands[i].LowBest(), hands[i].LowUnused())
		} else {
			fmt.Printf("  Lo: None\n")
		}
	}
	h, hPivot := cardrank.Order(hands)
	l, lPivot := cardrank.LowOrder(hands)
	typ := "wins"
	if lPivot == 0 {
		typ = "scoops"
	}
	if hPivot == 1 {
		fmt.Printf("Result (Hi): Player %d %s with %s %b\n", h[0]+1, typ, hands[h[0]].Description(), hands[h[0]].Best())
	} else {
		var s, b []string
		for i := 0; i < hPivot; i++ {
			s = append(s, strconv.Itoa(h[i]+1))
			b = append(b, fmt.Sprintf("%b", hands[h[i]].Best()))
		}
		fmt.Printf("Result (Hi): Players %s push with %s %s\n", strings.Join(s, ", "), hands[h[0]].Description(), strings.Join(b, ", "))
	}
	if lPivot == 1 {
		fmt.Printf("Result (Lo): Player %d wins with %s %b\n", l[0]+1, hands[l[0]].LowDescription(), hands[l[0]].LowBest())
	} else if lPivot > 1 {
		var s, b []string
		for j := 0; j < lPivot; j++ {
			s = append(s, strconv.Itoa(l[j]+1))
			b = append(b, fmt.Sprintf("%b", hands[l[j]].LowBest()))
		}
		fmt.Printf("Result (Lo): Players %s push with %s %s\n", strings.Join(s, ", "), hands[l[0]].LowDescription(), strings.Join(b, ", "))
	} else {
		fmt.Printf("Result (Lo): no player made a low hand\n")
	}
}
```

Output:

```txt
------ OmahaHiLo 1653086518494356973 ------
Board: [2♣ K♥ 6♠ 5♣ 8♠]
Player 1: [A♣ 9♦ J♦ T♣]
  Hi: Nothing, Ace-high, kickers King, Jack, Eight, Six [A♣ K♥ J♦ 8♠ 6♠] [9♦ T♣ 2♣ 5♣]
  Lo: None
Player 2: [7♣ 5♥ 6♣ T♦]
  Hi: Two Pair, Sixes over Fives, kicker King [6♣ 6♠ 5♣ 5♥ K♥] [7♣ T♦ 2♣ 8♠]
  Lo: Eight-low [8♠ 7♣ 6♠ 5♥ 2♣] [6♣ T♦ K♥ 5♣]
Player 3: [4♣ Q♥ K♣ Q♦]
  Hi: Pair, Kings, kickers Queen, Eight, Six [K♣ K♥ Q♥ 8♠ 6♠] [4♣ Q♦ 2♣ 5♣]
  Lo: None
Player 4: [5♦ 3♦ 9♠ 9♣]
  Hi: Pair, Nines, kickers King, Eight, Six [9♣ 9♠ K♥ 8♠ 6♠] [5♦ 3♦ 2♣ 5♣]
  Lo: Eight-low [8♠ 6♠ 5♦ 3♦ 2♣] [9♠ 9♣ K♥ 5♣]
Player 5: [2♠ K♦ 2♥ 8♦]
  Hi: Three of a Kind, Twos, kickers King, Eight [2♣ 2♥ 2♠ K♥ 8♠] [K♦ 8♦ 6♠ 5♣]
  Lo: None
Player 6: [J♠ 3♣ K♠ J♥]
  Hi: Pair, Kings, kickers Jack, Eight, Six [K♥ K♠ J♠ 8♠ 6♠] [3♣ J♥ 2♣ 5♣]
  Lo: None
Result (Hi): Player 5 wins with Three of a Kind, Twos, kickers King, Eight [2♣ 2♥ 2♠ K♥ 8♠]
Result (Lo): Player 4 wins with Eight-low [8♠ 6♠ 5♦ 3♦ 2♣]
```

### Hand Ranking

A `HandRank` type is used to determine the relative rank of a [`Hand`][hand],
on a low-to-high basis. Higher hands will have a lower value than low hands.
For example, a Straight Flush will have a lower `HandRank` than Full House.

#### Rankers

For regular poker hands (ie, Holdem, Omaha, and Stud), pure Go implementations
for the well-known [Cactus Kev (`CactusRanker`)][cactus-ranker], [Fast Cactus
(`CactusFastRanker`)][cactus-fast-ranker], and [Two-Plus-Two
(`TwoPlusTwoRanker`)][two-plus-two-ranker] poker hand evaluators are provided.
Additionally a [`SixPlusRanker`][six-plus-ranker], a [`EightOrBetterRanker`][eight-or-better-ranker],
and a [`RazzRanker`][razz-ranker] are available for use with Texas Holdem Short
Deck and Omaha/Stud Lo evaluation, respectively.

##### Default and Hybrid Rankers

The package's [`DefaultRanker`][default-ranker] is a [`HybridRanker`][hybrid-ranker]
using either the [`CactusFastRanker`][cactus-fast-ranker] or [`TwoPlusTwoRanker`][two-plus-two-ranker]
depending on the [`Hand`][hand] having 5, 6, or 7 cards. The `HybridRanker`
provides the best possible evaluation speed in most cases.

#### Ordering and Winner Determination

Hands can be compared to each other using `Compare` or can be [ordered][order]
using the package level `Order` and `LowOrder` funcs. See [the examples][examples]
for overviews on winner determination.

### Build Tags

Package level build tags are used to change the build configuration of the
package:

#### `portable`

The `portable` build tag can be used to disable the `TwoPlusTwoRanker`, which
requires embedding a large (approximately 130 Mib) look-up table.

This is useful when using this package in a portable or embedded application.
For example, when targetting a WASM build, the following can be used to create
slimmer WASM binaries:

```sh
GOOS=js GOARCH=wasm go build -tags portable
```

#### `embedded`

The `embedded` tag can be used to disable the `CactusFastRanker` and the
`TwoPlusTwoRanker`, creating the smallest possible binaries:

```sh
GOOS=js GOARCH=wasm go build -tags embedded
```

#### `noinit`

The `noinit` tag enables a slightly faster startup time by disabling
initialization of package level variables `DefaultRanker` and
`DefaultSixPlusRanker` until needed.

```sh
GOOS=js GOARCH=wasm go build -tags 'embedded noinit'
```

When using the `noinit` build tag, the user will need to manually set the
`DefaultCactus`, `DefaultRanker` and `DefaultSixPlusRanker` variables:

```go
cardrank.DefaultCactus = cardrank.CactusRanker
cardrank.DefaultRanker = cardrank.HandRanker(cardrank.CactusRanker)
cardrank.DefaultSixPlusRanker = cardrank.HandRanker(cardrank.SixPlusRanker(cardrank.CactusRanker))
```

[See `z.go` for initialization logic.](/z.go)

## Future Development

Rankers for Badugi, Sökö, and other poker variants will be added to this
package in addition to standardized interfaces for managing poker tables and
games.

## Links

* [Overview of Cactus Kev][cactus-kev] - original Cactus Kev article
* [Coding the Wheel][coding-the-wheel] - article covering various poker hand evaluators
* [Paul Senzee Perfect Hash for Cactus Kev][senzee] - overview of the "Fast Cactus" perfect hash by Paul Senszee
* [TwoPlusTwoHandEvaluator][tangentforks] - original implementation of the Two-Plus-Two evaluator

[cactus-kev]: https://archive.is/G6GZg
[coding-the-wheel]: https://www.codingthewheel.com/archives/poker-hand-evaluator-roundup/
[senzee]: http://senzee.blogspot.com/2006/06/some-perfect-hash.html
[tangentforks]: https://github.com/tangentforks/TwoPlusTwoHandEvaluator

[pkg]: https://pkg.go.dev/cardrank.io/cardrank
[examples]: https://pkg.go.dev/cardrank.io/cardrank#pkg-examples
[hand-ranking]: #hand-ranking
[future]: #future-development

[card]: https://pkg.go.dev/cardrank.io/cardrank#Card
[suit]: https://pkg.go.dev/cardrank.io/cardrank#Suit
[rank]: https://pkg.go.dev/cardrank.io/cardrank#Rank
[deck]: https://pkg.go.dev/cardrank.io/cardrank#Deck
[hand]: https://pkg.go.dev/cardrank.io/cardrank#Hand
[order]: https://pkg.go.dev/cardrank.io/cardrank#Order

[holdem-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Holdem
[short-deck-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-ShortDeck
[omaha-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Omaha
[omaha-hi-lo-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-OmahaHiLo
[stud-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Stud
[stud-hi-lo-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-StudHiLo
[razz-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Razz

[default-ranker]: https://pkg.go.dev/cardrank.io/cardrank#DefaultRanker
[hand-ranker]: https://pkg.go.dev/cardrank.io/cardrank#HandRanker
[cactus-ranker]: https://pkg.go.dev/cardrank.io/cardrank#CactusRanker
[cactus-fast-ranker]: https://pkg.go.dev/cardrank.io/cardrank#CactusFastRanker
[two-plus-two-ranker]: https://pkg.go.dev/cardrank.io/cardrank#TwoPlusTwoRanker
[hybrid-ranker]: https://pkg.go.dev/cardrank.io/cardrank#HybridRanker
[six-plus-ranker]: https://pkg.go.dev/cardrank.io/cardrank#SixPlusRanker
[eight-or-better-ranker]: https://pkg.go.dev/cardrank.io/cardrank#EightOrBetterRanker
[razz-ranker]: https://pkg.go.dev/cardrank.io/cardrank#RazzRanker

# cardrank.io/cardrank

Package `cardrank.io/cardrank` provides a library of types, funcs, and
utilities for working with playing cards, decks, and evaluating poker hands.
Supports [Texas Holdem][holdem], [Texas Holdem Short Deck (6-plus)][short-deck],
[Omaha][omaha], [Omaha Hi/Lo][omaha-hi-lo], [Stud][stud], and [Stud Hi/Lo][stud-hi-lo].

## Overview

High-level types, funcs, and standardized interfaces are included in the
package to deal and evaluate hands of poker, including all necessary types for
representing and working with [cards][card], [card suits][suit], [card
ranks][rank], [card decks][deck], and [hands of cards][hand]. Hand evaluation
is achieved with pure Go implementations of [common poker hand rank
evaluators][hand-ranking].

[Hands][hand] of [Texas Holdem][holdem], [Texas Holdem Short Deck (aka
6-Plus)][short-deck], [Omaha][omaha], and [Omaha Hi/Lo][omaha-hi-lo],
[Stud][stud], and [Stud Hi/Lo][stud-hi-lo] are easily created and dealt using
standardized interfaces and logic, with winners [being easily determined and ordered][order].

[Future development](#future) of Razz and Badugi is planned, as well as for
other poker variants.

## Using

See [Go documentation][pkg].

```sh
go get cardrank.io/cardrank
```

### Examples

Complete examples for [Holdem][holdem], [6-plus (ie, short deck
Holdem)][short-deck], [Omaha][omaha], [Omaha Hi/Lo][omaha-hi-lo], [Stud][stud],
[Stud Hi/Lo][stud-hi-lo] are available in the source repository. [Further
examples][examples] are available in the [Go package documentation][pkg] for
overviews of using most commonly used types, funcs and interfaces.

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

### Hand Ranking

A `HandRank` type is used to determine the relative rank of a [`Hand`][hand],
on a low-to-high basis.Higher hands will have a lower value than low hands. For
example, a Straight Flush will have a lower `HandRank` than Full House.

For regular poker hands (ie, Holdem, Omaha, and Stud), pure Go implementations
for the well-known [Cactus Kev][cactus], [Fast Cactus][cactus-fast], and
[Two-Plus][two-plus] poker hand evaluators are provided for hand evaluation.

#### Rankers

The package's [`DefaultRanker`][default] is a [`HybridRanker`][hybrid] using
either the [`CactusFastRanker`][cactus-fast] or [`TwoPlusRanker`][two-plus]
depending on the [`Hand`][hand] having 5, 6, or 7 cards, providing the best
possible evaluation speed in most cases.

Additionally a [`SixPlusRanker`][six-plus] and a [`EightOrBetterRanker`][eight-or-better]
rankers are provided, used for Short Deck and Omaha/Stud Lo evaluation
respectively.

#### Ordering and Winner Determination

Hands can be compared to each other using `Compare` or can be [ordered][order]
using the package level `Order` and `LowOrder` funcs. See [the examples][examples]
for overviews on winner determination.

#### Portablility

The [Two-Plus][two-plus] ranker implementation requires embedding the
`handranks*.dat` files, which adds approximately 130 MiB to any Go binary. This
can be disabled by using the `portable` build tag:

```sh
go build -tags portable
```

This is useful when using this package in a portable or embedded application.
For example, when targetting a WASM build, the following can be used to create
slimmer WASM binaries:

```sh
GOOS=js GOARCH=wasm go build -tags portable
```

Alternately, the `embedded` tag can be used to create even slimmer builds:

```sh
go build -tags 'portable embedded'
```

This has the effect of disaling the Cactus Fast ranker in favor of a generative
Cactus ranker.

## Development Status

A partially complete Razz ranker is available, however it currently does not
work with hands having a set (i.e., pair, two pair, three-of-a-kind, full
house, or four-of-a-kind).

### Future

Rankers for Badugi, and other poker variants will be added to this package in
addition to standardized interfaces for managing poker tables and games.

[pkg]: https://pkg.go.dev/cardrank.io/cardrank
[examples]: https://pkg.go.dev/cardrank.io/cardrank#pkg-examples
[hand-ranking]: #hand-ranking

[card]: https://pkg.go.dev/cardrank.io/cardrank#Card
[suit]: https://pkg.go.dev/cardrank.io/cardrank#Suit
[rank]: https://pkg.go.dev/cardrank.io/cardrank#Rank
[deck]: https://pkg.go.dev/cardrank.io/cardrank#Deck
[hand]: https://pkg.go.dev/cardrank.io/cardrank#Hand
[order]: https://pkg.go.dev/cardrank.io/cardrank#Order
[default]: https://pkg.go.dev/cardrank.io/cardrank#DefaultRanker

[holdem]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Holdem
[short-deck]: https://pkg.go.dev/cardrank.io/cardrank#example-package-ShortDeck
[omaha]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Omaha
[omaha-hi-lo]: https://pkg.go.dev/cardrank.io/cardrank#example-package-OmahaHiLo
[stud]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Stud
[stud-hi-lo]: https://pkg.go.dev/cardrank.io/cardrank#example-package-StudHiLo

[hand-ranker]: https://pkg.go.dev/cardrank.io/cardrank#HandRanker
[cactus]: https://pkg.go.dev/cardrank.io/cardrank#CactusRanker
[cactus-fast]: https://pkg.go.dev/cardrank.io/cardrank#CactusFastRanker
[two-plus]: https://pkg.go.dev/cardrank.io/cardrank#TwoPlusRanker
[hybrid]: https://pkg.go.dev/cardrank.io/cardrank#HybridRanker
[six-plus]: https://pkg.go.dev/cardrank.io/cardrank#SixPlusRanker
[eight-or-better]: https://pkg.go.dev/cardrank.io/cardrank#EightOrBetter

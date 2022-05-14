# cardrank.io/cardrank

Package `cardrank.io/cardrank` provides a library of types, funcs, and
utilities for working with playing cards, decks, and evaluating poker hands.

Supports Texas Holdem, Texas Holdem Short Deck (aka 6-plus), Omaha, and Omaha
Hi/Lo.

## Overview

High-level types, funcs, and standardized interfaces are included in the
package to deal and evaluate hands of poker, including all necessary types for
representing and working with [card suits][suit], [card ranks][rank], and [card
decks][deck].

Additionally, full, complete, pure Go implementations of [poker hand rank
evaluators][ranker] are provided.

Through high-level APIs, games of Texas Holdem, Texas Holdem Short Deck (aka
6-Plus), Omaha, and Omaha Hi/Lo are easily created using standardized
interfaces and logic.

[Future development](#future) will build out support for most common
poker variants, such as Razz and Badugi.

## Using

See [Go documentation][pkg].

```sh
go get cardrank.io/cardrank
```

### Examples

See all [examples][] in the Go documentation for quick overviews of using the
high-level APIs.

#### Texas Holdem

Dealing a game of Texas Holdem and determining the winner(s):

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
	h, pivot := cardrank.OrderHands(hands)
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

Dealing a game of Omaha Hi/Lo and determining the winner(s):

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
	highs := cardrank.OmahaHiLo.RankHands(pockets, board)
	lows := cardrank.OmahaHiLo.LowRankHands(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b\n", i+1, pockets[i])
		fmt.Printf("  Hi: %s %b %b\n", highs[i].Description(), highs[i].Best(), highs[i].Unused())
		if lows[i].Rank() < 31 {
			fmt.Printf("  Lo: %s %b %b\n", lows[i].LowDescription(), lows[i].LowBest(), lows[i].LowUnused())
		} else {
			fmt.Printf("  Lo: None\n")
		}
	}
	h, hPivot := cardrank.OrderHands(highs)
	l, lPivot := cardrank.LowOrderHands(lows)
	typ := "wins"
	if lPivot == 0 {
		typ = "scoops"
	}
	if hPivot == 1 {
		fmt.Printf("Result (Hi): Player %d %s with %s %b\n", h[0]+1, typ, highs[h[0]].Description(), highs[h[0]].Best())
	} else {
		var s, b []string
		for i := 0; i < hPivot; i++ {
			s = append(s, strconv.Itoa(h[i]+1))
			b = append(b, fmt.Sprintf("%b", highs[h[i]].Best()))
		}
		fmt.Printf("Result (Hi): Players %s push with %s %s\n", strings.Join(s, ", "), highs[h[0]].Description(), strings.Join(b, ", "))
	}
	if lPivot == 1 {
		fmt.Printf("Result (Lo): Player %d wins with %s %b\n", l[0]+1, lows[l[0]].LowDescription(), lows[l[0]].LowBest())
	} else if lPivot > 1 {
		var s, b []string
		for j := 0; j < lPivot; j++ {
			s = append(s, strconv.Itoa(l[j]+1))
			b = append(b, fmt.Sprintf("%b", lows[l[j]].LowBest()))
		}
		fmt.Printf("Result (Lo): Players %s push with %s %s\n", strings.Join(s, ", "), lows[l[0]].LowDescription(), strings.Join(b, ", "))
	} else {
		fmt.Printf("Result (Lo): no player made a low hand\n")
	}
}
```

### Rankers

The `cardrank.io/cardrank` package includes pure Go implementations for the
well-known [Cactus Kev][cactus], [Fast Cactus][cactus-fast], and
[Two-Plus][two-plus] poker hand evaluators. A [hybrid][hybrid] implementation
is provided that uses will use either the Fast Cactus or Two-Plus evaluators
dependent on if the evaluated poker hand has 5, 6, or 7 cards.

Additionally a [Six Plus][six-plus] and a [Eight-or-Better][eight-or-better]
hand evaluator is provided, useful for Short Deck and Eight-or-Better (Omaha
Lo) evaluations.

All poker hand rank evaluators (ie, a `Ranker` or `RankerFunc`) rank hands
based from low-to-high value, meaning that lower hand values beat hands with
higher ranks.

## Portablility

The [Two Plus][two-plus] ranker implementation requires embedding the
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

### Future

Eventually, rankers for Stud, Stud Hi/Lo, Razz, Badugi, and other poker
variants will be added to this package in addition to standardized interfaces
for managing poker tables and games.

## Links

[pkg]: https://pkg.go.dev/cardrank.io/cardrank
[examples]: https://pkg.go.dev/cardrank.io/cardrank#pkg-examples

[cactus]: https://pkg.go.dev/cardrank.io/cardrank#NewCactusRanker
[cactus-fast]: https://pkg.go.dev/cardrank.io/cardrank#NewCactusFastRanker
[two-plus]: https://pkg.go.dev/cardrank.io/cardrank#NewTwoPlusRanker
[hybrid]: https://pkg.go.dev/cardrank.io/cardrank#NewHybridRanker

[six-plus]: https://pkg.go.dev/cardrank.io/cardrank#NewCactusFastSixPlusRanker
[eight-or-better]: https://pkg.go.dev/cardrank.io/cardrank#NewEightOrBetterRanker

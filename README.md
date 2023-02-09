# About

Package `github.com/cardrank/cardrank` is a library of types, utilities, and interfaces
for working with playing cards, card decks, and evaluating poker hand ranks.

Supports [Texas Holdem][holdem-example], [Texas Holdem Short (6+)][short-example],
[Texas Holdem Royal (10+)][royal-example], [Omaha][omaha-example],
[Omaha Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud Hi/Lo][stud-hi-lo-example],
[Razz][razz-example], and [Badugi][badugi-example].

[![Tests](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/cardrank/cardrank)](https://goreportcard.com/report/github.com/cardrank/cardrank)
[![Reference](https://pkg.go.dev/badge/github.com/cardrank/cardrank.svg)](https://pkg.go.dev/github.com/cardrank/cardrank)
[![Releases](https://img.shields.io/github/v/release/cardrank/cardrank?display_name=tag&sort=semver)](https://github.com/cardrank/cardrank/releases)

## Overview

The `github.com/cardrank/cardrank` package contains types for [cards][card], [card
suits][suit], [card ranks][rank], [card decks][deck], and [hands of
cards][hand].

A single [package level type][type] provides a standard interface for dealing
cards for, and evaluating [poker hands][hand] of the following:

* [Texas Holdem][holdem-example]
* [Texas Holdem Short (6+)][short-example]
* [Texas Holdem Royal (10+)][royal-example]
* [Omaha][omaha-example]
* [Omaha Hi/Lo][omaha-hi-lo-example]
* [Stud][stud-example]
* [Stud Hi/Lo][stud-hi-lo-example]
* [Razz][razz-example]
* [Badugi][badugi-example]

[Hand evaluation and ranking][hand-ranking] of the different [poker hand
types][type] is accomplished through pure Go implementations of [well-known
poker rank evaluation algorithms](#cactus-kev). [Poker hands][hand] can be
compared and [ordered to determine the hand's winner(s)][hand-ordering].

[Development of additional poker variants][future], such as Kansas City Lowball
and Candian Stud (Sökö), is planned.

## Using

To use within a Go package or module:

```sh
go get github.com/cardrank/cardrank
```

See package level [Go documentation][pkg].

### Quickstart

Complete examples for [Texas Holdem][holdem-example], [Texas Holdem Short (6+)][short-example],
[Texas Holdem Royal (10+)][royal-example], [Omaha][omaha-example], [Omaha Hi/Lo][omaha-hi-lo-example],
[Stud][stud-example], [Stud Hi/Lo][stud-hi-lo-example], [Razz][razz-example],
and [Badugi][badugi-example] are available in the source repository. [Other examples][examples]
are available in the [Go package documentation][pkg] showing use of the
package's types, utilities, and interfaces.

##### Texas Holdem

The following showcases a simple game of Texas Holdem:

```go
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.Holdem.Deal(rnd, players)
	hands := cardrank.Holdem.RankHands(pockets, board)
	fmt.Printf("------ Holdem %d ------\n", seed)
	fmt.Printf("Board:    %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b %s %b %b\n", i+1, hands[i].Pocket, hands[i].Description(), hands[i].Best(), hands[i].Unused())
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
------ Holdem 1676110749917132920 ------
Board:    [8♦ 9♠ 2♦ 2♣ J♦]
Player 1: [J♣ 4♠] Two Pair, Jacks over Twos, kicker Nine [J♣ J♦ 2♣ 2♦ 9♠] [8♦ 4♠]
Player 2: [3♣ T♣] Pair, Twos, kickers Jack, Ten, Nine [2♣ 2♦ J♦ T♣ 9♠] [8♦ 3♣]
Player 3: [6♦ 5♣] Pair, Twos, kickers Jack, Nine, Eight [2♣ 2♦ J♦ 9♠ 8♦] [6♦ 5♣]
Player 4: [9♣ 4♣] Two Pair, Nines over Twos, kicker Jack [9♣ 9♠ 2♣ 2♦ J♦] [8♦ 4♣]
Player 5: [7♠ 2♥] Three of a Kind, Twos, kickers Jack, Nine [2♣ 2♦ 2♥ J♦ 9♠] [8♦ 7♠]
Player 6: [T♠ 3♠] Pair, Twos, kickers Jack, Ten, Nine [2♣ 2♦ J♦ T♠ 9♠] [8♦ 3♠]
Result:   Player 5 wins with Three of a Kind, Twos, kickers Jack, Nine [2♣ 2♦ 2♥ J♦ 9♠]
```

##### Omaha Hi/Lo

The following showcases a simple game of Omaha Hi/Lo, highlighting Hi/Lo winner
determination:

```go
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cardrank/cardrank"
)

func main() {
	const players = 6
	seed := time.Now().UnixNano()
	// note: use a better pseudo-random number generator
	rnd := rand.New(rand.NewSource(seed))
	pockets, board := cardrank.OmahaHiLo.Deal(rnd, players)
	hands := cardrank.OmahaHiLo.RankHands(pockets, board)
	fmt.Printf("------ OmahaHiLo %d ------\n", seed)
	fmt.Printf("Board: %b\n", board)
	for i := 0; i < players; i++ {
		fmt.Printf("Player %d: %b\n", i+1, pockets[i])
		fmt.Printf("  Hi: %s %b %b\n", hands[i].Description(), hands[i].Best(), hands[i].Unused())
		fmt.Printf("  Lo: %s %b %b\n", hands[i].LowDescription(), hands[i].LowBest(), hands[i].LowUnused())
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
------ OmahaHiLo 1676110711435292197 ------
Board: [9♥ 4♦ A♣ 9♦ 2♣]
Player 1: [J♦ Q♠ J♠ 7♦]
  Hi: Two Pair, Jacks over Nines, kicker Ace [J♦ J♠ 9♦ 9♥ A♣] [Q♠ 7♦ 4♦ 2♣]
  Lo: None [] []
Player 2: [K♥ 2♠ T♥ 3♦]
  Hi: Two Pair, Nines over Twos, kicker King [9♦ 9♥ 2♣ 2♠ K♥] [T♥ 3♦ 4♦ A♣]
  Lo: None [] []
Player 3: [2♦ T♦ K♣ 6♥]
  Hi: Two Pair, Nines over Twos, kicker King [9♦ 9♥ 2♣ 2♦ K♣] [T♦ 6♥ 4♦ A♣]
  Lo: None [] []
Player 4: [9♣ 5♥ J♥ 8♦]
  Hi: Three of a Kind, Nines, kickers Ace, Jack [9♣ 9♦ 9♥ A♣ J♥] [5♥ 8♦ 4♦ 2♣]
  Lo: Eight, Five, Four, Two, Ace-low [8♦ 5♥ 4♦ 2♣ A♣] [9♣ J♥ 9♥ 9♦]
Player 5: [8♠ 6♣ 4♣ Q♥]
  Hi: Two Pair, Nines over Fours, kicker Queen [9♦ 9♥ 4♣ 4♦ Q♥] [8♠ 6♣ A♣ 2♣]
  Lo: Eight, Six, Four, Two, Ace-low [8♠ 6♣ 4♦ 2♣ A♣] [4♣ Q♥ 9♥ 9♦]
Player 6: [T♠ 6♠ 5♠ 5♦]
  Hi: Two Pair, Nines over Fives, kicker Ace [9♦ 9♥ 5♦ 5♠ A♣] [T♠ 6♠ 4♦ 2♣]
  Lo: Six, Five, Four, Two, Ace-low [6♠ 5♠ 4♦ 2♣ A♣] [T♠ 5♦ 9♥ 9♦]
Result (Hi): Player 4 wins with Three of a Kind, Nines, kickers Ace, Jack [9♣ 9♦ 9♥ A♣ J♥]
Result (Lo): Player 6 wins with Six, Five, Four, Two, Ace-low [6♠ 5♠ 4♦ 2♣ A♣]
```

### Hand Ranking

Poker [`Hand`][hand]'s make use of a [`HandRank`][hand-rank] to determine the
relative rank of a `Hand`, on a low-to-high basis. Higher poker hands
have a lower value `HandRank` than lower poker hands. For example, a
[`StraightFlush`][hand-rank] will have a lower `HandRank` than a
[`FullHouse`][hand-rank].

When a [`Hand`][hand] is created, a Hi and Lo (if applicable) `HandRank` is
evaluated, and made available via [`Hand.Rank`][hand.rank] and
[`Hand.LowRank`][hand.low-rank] methods, respectively. The Hi and Lo
`HandRank`'s are evaluated by a `EvalRankFunc`, dependent on the `Hand`'s
[`Type`][type]:

#### Cactus Kev

For regular poker hand types ([`Holdem`][type], [`Royal`][type],
[`Omaha`][type], and [`Stud`][type]), poker hand rank is determined by Go
implementations of different [Cactus Kev][cactus-kev] evaluators:

* [`Cactus`][cactus] - the original [Cactus Kev][cactus-kev] poker hand evaluator
* [`CactusFast`][cactus-fast] - the [Fast Cactus][senzee] poker hand evaluator, using Paul Senzee's perfect hash lookup
* [`TwoPlusTwo`][two-plus-two] - the [2+2 forum][tangentforks] poker hand evaluator, using a 130 MiB lookup table

See [below for more information](#default-rank-func) on the default rank func in
use by the package, and for information on [using build tags][build-tags] to
enable/disable functionality for different target runtime environments.

#### Default Rank Func

The package-level [`DefaultRank`][default-rank] variable is used for regular
poker evaluation. For most scenarios, the [`Hybrid`][hybrid] rank func provides
the best possible evaluation performance, and is used by default when no [build
tags][build-tags] have been specified.

#### Hybrid

The [`Hybrid`][hybrid] rank func uses either the [`CactusFast`][cactus-fast] or
an instance of the [`TwoPlusTwo`][two-plus-two] depending on the [`Hand`][hand]
having 5, 6, or 7 cards.

#### Two-Plus-Two

The [`TwoPlusTwo`][two-plus-two] makes use of a large (approximately 130 MiB)
lookup table to accomplish extremely fast 5, 6 and 7 card hand rank evaluation.
Due to the large size of the lookup table, the `TwoPlusTwo` will be excluded
when using the using the [`portable` or `embedded` build tags][build-tags].

The `TwoPlusTwo` is disabled by default for `GOOS=js` (ie, WASM) builds, but
can be enabled using the [`forcefat` build tag][build-tags].

### Winner Determination

Winner(s) are determined by the lowest possible [`HandRank`][hand-rank] when
comparing a `Hand`'s [`Rank`][hand.rank] or [`LowRank`][hand.low-rank]
against another hand. Two or more hands having a `HandRank` of equal value
indicate that the hands have equivalent `HandRank`, and thus have both won.

As such, when hands are sorted (low-to-high) by `Rank` or `LowRank`, the
winner(s) of a hand will be all the hands in the lowest position and having the
same `HandRank`.

#### Comparing Hands

A [`Hand`][hand] can be compared to another `Hand` using [`Compare`][hand.compare]
and [`LowCompare`][hand.low-compare].

`Compare` and `LowCompare` return `-1`, `0`, or `+1`, making it easy to compare
or sort hands:

```go
// Compare hands:
if hand1.Compare(hand2) < 0 {
	fmt.Printf("%s is a winner!", hand1)
}

// Compare low hands:
// (applicable only for a hand types that supports low hands)
if hand1.LowCompare(hand2) == 0 {
	fmt.Printf("%s and %s are equal!", hand1, hand2)
}

// Sort hands:
sort.Slice(hands, func(i, j int) bool {
	return hands[i].Compare(hands[j]) < 0
})
```

#### Ordering Hands

[`Order`][order] and [`LowOrder`][low-order] determine the winner(s) of a hand
by ordering the indexes of a [`[]*Hand`][hand] and returning the list of
ordered hands as a `[]int` and an `int` pivot indicating the position within
the returned list demarcating winning and losing hands.

A `Hand` whose index is in position `i < pivot` is considered to be the
winner(s) of the hand. Hi hands are guaranteed to have 1 or more winner(s),
while Lo hands have 0 or more winner(s):

```go
// Order hands by lowest hand rank, low to high:
h, pivot := cardrank.Order(hands)
for i := 0; i < pivot; i++ {
	fmt.Printf("%s is a Hi winner!", hands[h[i]])
}

// Order low hands by lowest hand rank, low to high:
// (applicable only for a hand types that supports low hands)
l, pivot := cardrank.LowOrder(hands)
for i := 0; i < pivot; i++ {
	fmt.Printf("%s is a Lo winner!", hands[h[i]])
}
```

See [the examples][examples] for an overview of using the package APIs for
winner determination for the different [`Type`][type] of poker hands.

### Build Tags

Build tags can be used with `go build` to change the package's build
configuration. Available tags:

#### `portable`

The `portable` tag disables the `TwoPlusTwo`, in effect excluding the the
[large lookup table](#two-plus-two), and creating significantly smaller
binaries but at the cost of more expensive poker hand rank evaluation. Useful
when building for portable or embedded environments, such as a client
application:

```sh
go build -tags portable
```

#### `embedded`

The `embedded` tag disables the `CactusFast` and the `TwoPlusTwo`, creating the
smallest possible binaries. Useful when either embedding the package in another
application, or in constrained runtime environments such as WASM:

```sh
GOOS=js GOARCH=wasm go build -tags embedded
```

#### `noinit`

The `noinit` tag disables the package level initialization. Useful when
applications need the fastest possible startup times and can defer
initialization, or when using a third-party algorithm:

```sh
GOOS=js GOARCH=wasm go build -tags 'embedded noinit' -o cardrank.wasm
```

When using the `noinit` build tag, the user will need to call the [`Init`
func][init] to set `DefaultCactus`, `DefaultRank` and to register the default
types automatically:

```go
// Set DefaultCactus, DefaultRank based on available implementations:
cardrank.Init()
```

Alternatively, the `DefaultCactus` and `DefaultRank` can be set manually. After
`DefaultCactus` and `DefaultRank` have been set, call `RegisterDefaultTypes` to
register built in types:

```go
// Set manually (such as when using a third-party implementation):
cardrank.DefaultCactus = cardrank.Cactus
cardrank.DefaultRank = cardrank.NewHandRank(cardrank.Cactus)

// Then call RegisterDefaultTypes to register default types
if err := cardrank.RegisterDefaultTypes(); err != nil {
	panic(err)
}
```

#### `forcefat`

The `forcefat` tag forces a "fat" binary build, including the `TwoPlusTwo`'s
large lookup table, irrespective of other build tags:

```sh
GOOS=js GOARCH=wasm go build -tags 'forcefat' -o cardrank.wasm
```

## Future Development

Rank funcs for Kansas City Lowball, Sökö, and other poker variants will be
added to this package in addition to standardized interfaces for managing poker
tables and games.

## Links

* [Overview of Cactus Kev][cactus-kev] - original Cactus Kev article
* [Coding the Wheel][coding-the-wheel] - article covering various poker hand evaluators
* [Paul Senzee Perfect Hash for Cactus Kev][senzee] - overview of the "Fast Cactus" perfect hash by Paul Senzee
* [TwoPlusTwoHandEvaluator][tangentforks] - original implementation of the Two-Plus-Two evaluator

[cactus-kev]: https://archive.is/G6GZg
[coding-the-wheel]: https://www.codingthewheel.com/archives/poker-hand-evaluator-roundup/
[senzee]: http://senzee.blogspot.com/2006/06/some-perfect-hash.html
[tangentforks]: https://github.com/tangentforks/TwoPlusTwoHandEvaluator

[pkg]: https://pkg.go.dev/github.com/cardrank/cardrank
[examples]: https://pkg.go.dev/github.com/cardrank/cardrank#pkg-examples
[hand-ranking]: #hand-ranking
[build-tags]: #build-tags
[future]: #future-development
[hand-ordering]: #winner-determination

[card]: https://pkg.go.dev/github.com/cardrank/cardrank#Card
[suit]: https://pkg.go.dev/github.com/cardrank/cardrank#Suit
[rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Rank
[deck]: https://pkg.go.dev/github.com/cardrank/cardrank#Deck
[hand]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand
[hand-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#HandRank
[type]: https://pkg.go.dev/github.com/cardrank/cardrank#Type
[order]: https://pkg.go.dev/github.com/cardrank/cardrank#Order
[init]: https://pkg.go.dev/github.com/cardrank/cardrank#Init
[low-order]: https://pkg.go.dev/github.com/cardrank/cardrank#LowOrder
[hand.compare]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand.Compare
[hand.low-compare]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand.LowCompare
[hand.rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand.Rank
[hand.low-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand.LowRank

[holdem-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Holdem
[short-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Short
[royal-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Royal
[omaha-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Omaha
[omaha-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-OmahaHiLo
[stud-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Stud
[stud-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-StudHiLo
[razz-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Razz
[badugi-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Badugi

[default-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#DefaultRank
[cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#Cactus
[cactus-fast]: https://pkg.go.dev/github.com/cardrank/cardrank#CactusFast
[two-plus-two]: https://pkg.go.dev/github.com/cardrank/cardrank#TwoPlusTwo
[hybrid]: https://pkg.go.dev/github.com/cardrank/cardrank#Hybrid

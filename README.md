# cardrank.io/cardrank

Package `cardrank.io/cardrank` is a library of types, utilities, and interfaces
for working with playing cards, card decks, and evaluating poker hand ranks.

Supports [Texas Holdem][holdem-example], [Texas Holdem Short (6+)][short-example],
[Texas Holdem Royal (10+)][royal-example], [Omaha][omaha-example],
[Omaha Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud Hi/Lo][stud-hi-lo-example],
[Razz][razz-example], and [Badugi][badugi-example].

[![GoDoc](https://godoc.org/cardrank.io/cardrank?status.svg)](https://godoc.org/cardrank.io/cardrank)
[![Tests on Linux, MacOS and Windows](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/cardrank.io/cardrank)](https://goreportcard.com/report/cardrank.io/cardrank)

## Overview

The `cardrank.io/cardrank` package contains types for [cards][card], [card
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
poker rank evaluation algorithms](#rankers). [Poker hands][hand] can be
compared and [ordered to determine the hand's winner(s)][hand-ordering].

[Development of additional poker variants][future], such as Kansas City Lowball
and Candian Stud (Sökö), is planned.

## Using

To use within a Go package or module:

```sh
go get cardrank.io/cardrank
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

Poker [`Hand`][hand]'s make use of a [`HandRank`][hand-rank] to determine the
relative rank of a `Hand`, on a low-to-high basis. Higher poker hands
have a lower value `HandRank` than lower poker hands. For example, a
[`StraightFlush`][hand-rank] will have a lower `HandRank` than a
[`FullHouse`][hand-rank].

When a [`Hand`][hand] is created, a Hi and Lo (if applicable) `HandRank` is
evaluated, and made available via [`Hand.Rank`][hand.rank] and
[`Hand.LowRank`][hand.low-rank] methods, respectively. The Hi and Lo
`HandRank`'s are evaluated by a `Ranker`, dependent on the `Hand`'s
[`Type`][type]:

#### Cactus Kev

For regular poker hand types ([`Holdem`][type], [`Royal`][type],
[`Omaha`][type], and [`Stud`][type]), poker hand rank is determined by Go
implementations of different [Cactus Kev][cactus-kev] evaluators:

* [`CactusRanker`][cactus-ranker] - the original [Cactus Kev][cactus-kev] poker hand evaluator
* [`CactusFastRanker`][cactus-fast-ranker] - the [Fast Cactus][senzee] poker hand evaluator, using Paul Senzee's perfect hash lookup
* [`TwoPlusTwoRanker`][two-plus-two-ranker] - the [2+2 forum][tangentforks] poker hand evaluator, using a 130 MiB lookup table

See [below for more information](#default-ranker) on the default ranker in use
by the package, and for information on [using build tags to disable][build-tags]
for different scenarios.

#### Default Ranker

The package-level [`DefaultRanker`][default-ranker] variable is used for regular poker
evaluation. For most scenarios, the [`HybridRanker`][hybrid-ranker] provides
the best possible evaluation performance, and is used by default when no [build
tags][build-tags] have been specified.

#### Hybrid

The [`HybridRanker`][hybrid-ranker] uses either the [`CactusFastRanker`][cactus-fast-ranker]
or an instance of the [`TwoPlusTwoRanker`][two-plus-two-ranker] depending on
the [`Hand`][hand] having 5, 6, or 7 cards.

#### Two-Plus-Two

The [`TwoPlusTwoRanker`][two-plus-two-ranker] makes use of a large
(approximately 130 MiB) lookup table to accomplish extremely fast 5, 6 and 7
card hand rank evaluation. Due to the large size of the lookup table, the
`TwoPlusTwoRanker` will be excluded when using the using the [`portable` or
`embedded` build tags][build-tags].

#### Other Variants

For the [`Short`][type], [`OmahaHiLo`][type], [`StudHiLo`][type], [`Razz`][type],
and [`Badugi`][type] poker hand types, a [6-plus][six-plus-ranker],
an [8-or-better][eight-or-better-ranker], a [Razz][razz-ranker], and a
[Badugi][badugi-ranker] poker hand rank evaluators are used.

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

The `portable` tag disables the `TwoPlusTwoRanker`, in effect excluding the the
[large lookup table](#two-plus-two-ranker), and creating significantly smaller
binaries but at the cost of more expensive poker hand rank evaluation. Useful
when building for portable or embedded environments, such as a WASM
application:

```sh
GOOS=js GOARCH=wasm go build -tags portable
```

#### `embedded`

The `embedded` tag disables the `CactusFastRanker` and the `TwoPlusTwoRanker`,
creating the smallest possible binaries:

```sh
GOOS=js GOARCH=wasm go build -tags embedded
```

#### `noinit`

The `noinit` tag disables the packagelevel initialization of variables
`DefaultRanker` and `DefaultSixPlusRanker`. Useful when applications need the
fastest possible startup times and can defer initialization, or when using a
third-party `Ranker` algorithm:

```sh
GOOS=js GOARCH=wasm go build -tags 'embedded noinit'
```

When using the `noinit` build tag, the user will need to [`Init` func][init] to
set `DefaultCactus`, `DefaultRanker` and `DefaultSixPlusRanker` automatically
or by manually specifying the variables:

```go
// Set DefaultCactus, DefaultRanker and DefaultSixPlusRanker based on
// available implementations:
cardrank.Init()

// Set manually (such as when using a third-party implementation):
cardrank.DefaultCactus = cardrank.CactusRanker
cardrank.DefaultRanker = cardrank.HandRanker(cardrank.CactusRanker)
cardrank.DefaultSixPlusRanker = cardrank.HandRanker(cardrank.SixPlusRanker(cardrank.CactusRanker))
```

## Future Development

Rankers for Kansas City Lowball, Sökö, and other poker variants will be added
to this package in addition to standardized interfaces for managing poker
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

[pkg]: https://pkg.go.dev/cardrank.io/cardrank
[examples]: https://pkg.go.dev/cardrank.io/cardrank#pkg-examples
[hand-ranking]: #hand-ranking
[build-tags]: #build-tags
[future]: #future-development
[hand-ordering]: #winner-determination-and-hand-ordering

[card]: https://pkg.go.dev/cardrank.io/cardrank#Card
[suit]: https://pkg.go.dev/cardrank.io/cardrank#Suit
[rank]: https://pkg.go.dev/cardrank.io/cardrank#Rank
[deck]: https://pkg.go.dev/cardrank.io/cardrank#Deck
[hand]: https://pkg.go.dev/cardrank.io/cardrank#Hand
[hand-rank]: https://pkg.go.dev/cardrank.io/cardrank#HandRank
[type]: https://pkg.go.dev/cardrank.io/cardrank#Type
[order]: https://pkg.go.dev/cardrank.io/cardrank#Order
[init]: https://pkg.go.dev/cardrank.io/cardrank#Init
[low-order]: https://pkg.go.dev/cardrank.io/cardrank#LowOrder
[hand.compare]: https://pkg.go.dev/cardrank.io/cardrank#Hand.Compare
[hand.low-compare]: https://pkg.go.dev/cardrank.io/cardrank#Hand.LowCompare
[hand.rank]: https://pkg.go.dev/cardrank.io/cardrank#Hand.Rank
[hand.low-rank]: https://pkg.go.dev/cardrank.io/cardrank#Hand.LowRank

[holdem-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Holdem
[short-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Short
[royal-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Royal
[omaha-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Omaha
[omaha-hi-lo-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-OmahaHiLo
[stud-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Stud
[stud-hi-lo-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-StudHiLo
[razz-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Razz
[badugi-example]: https://pkg.go.dev/cardrank.io/cardrank#example-package-Badugi

[default-ranker]: https://pkg.go.dev/cardrank.io/cardrank#DefaultRanker
[cactus-ranker]: https://pkg.go.dev/cardrank.io/cardrank#CactusRanker
[cactus-fast-ranker]: https://pkg.go.dev/cardrank.io/cardrank#CactusFastRanker
[two-plus-two-ranker]: https://pkg.go.dev/cardrank.io/cardrank#TwoPlusTwoRanker
[hybrid-ranker]: https://pkg.go.dev/cardrank.io/cardrank#HybridRanker
[six-plus-ranker]: https://pkg.go.dev/cardrank.io/cardrank#SixPlusRanker
[eight-or-better-ranker]: https://pkg.go.dev/cardrank.io/cardrank#EightOrBetterRanker
[razz-ranker]: https://pkg.go.dev/cardrank.io/cardrank#RazzRanker
[badugi-ranker]: https://pkg.go.dev/cardrank.io/cardrank#BadugiRanker

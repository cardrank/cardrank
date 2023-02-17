# About

Package `github.com/cardrank/cardrank` is a library of types, utilities, and
interfaces for working with playing cards, card decks, and evaluating poker
hand ranks.

<!-- START supports -->
Supports [Texas Holdem][holdem-example], [Texas Holdem Short (6+)][short-example],
[Texas Holdem Royal (10+)][royal-example], [Omaha][omaha-example],
[Omaha Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud Hi/Lo][stud-hi-lo-example],
[Razz][razz-example], and [Badugi][badugi-example].
<!-- END -->

[![Tests](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/cardrank/cardrank)](https://goreportcard.com/report/github.com/cardrank/cardrank)
[![Reference](https://pkg.go.dev/badge/github.com/cardrank/cardrank.svg)](https://pkg.go.dev/github.com/cardrank/cardrank)
[![Releases](https://img.shields.io/github/v/release/cardrank/cardrank?display_name=tag&sort=semver)](https://github.com/cardrank/cardrank/releases)

## Overview

The `github.com/cardrank/cardrank` package contains types for [cards][card],
[card suits][suit], [card ranks][rank], [card decks][deck], and [evaluating
cards][eval].

A single [package level type][type] provides a standard interface for dealing
cards for, and [evaluating ranks of 5, 6, and 7 poker hands][eval] of the
following:

<!-- START overview -->
* [Texas Holdem][holdem-example]
* [Texas Holdem Short (6+)][short-example]
* [Texas Holdem Royal (10+)][royal-example]
* [Omaha][omaha-example]
* [Omaha Hi/Lo][omaha-hi-lo-example]
* [Stud][stud-example]
* [Stud Hi/Lo][stud-hi-lo-example]
* [Razz][razz-example]
* [Badugi][badugi-example]
<!-- END -->

[Evaluation and ranking][eval-ranking] of the supported [poker hand
types][type] is accomplished through pure Go implementations of [well-known
poker rank evaluation algorithms](#cactus-kev). [Poker hands][eval] can be
compared and [ordered to determine winner(s)][winners].

## Using

To use within a Go package:

```sh
go get github.com/cardrank/cardrank
```

See package level [Go documentation][pkg].

### Quickstart

Complete [examples for all `Type`'s][examples] are available, including
additional examples showing use of various types, utilities, and interfaces is
available in the [Go package documentation][pkg].

Below are quick examples for [`Dealer`][dealer] [`Holdem`][type] and
[`OmahaHiLo`][type] included in the [example](/_example) directory:

##### Dealer

The following showcases a [`Dealer`][dealer]:

<!-- START dealer -->
<!-- END -->

##### Texas Holdem

The following showcases a simple game of [Texas Holdem][type]:

<!-- START holdem -->
<!-- END -->

##### Omaha Hi/Lo

The following showcases a simple game of [Omaha Hi/Lo][type]:

<!-- START omahahilo -->
<!-- END -->

### Eval Ranking

A pocket and optional board of [`Card`'s][card] can be passed to a
[`Type`'s][type] `New` method, which in turn uses the `Type`'s registered
[`EvalFunc`][eval-func] and creating a new [`Eval`][eval]:

```go
v := cardrank.Must("Ah Kh Qh Jh Th")
ev := cardrank.Holdem.New(v, nil)
fmt.Printf("%s\n", ev)

// Output:
// Straight Flush, Ace-high, Royal [Ah Kh Qh Jh Th]
```

When evaluating cards, one usually passes 5, 6, or 7 cards, but some `Type`'s
are capable of evaluating fewer `Card`'s:

```go
v := cardrank.Must("2h 3s 4c")
ev := cardrank.Badugi.New(v, nil)
fmt.Printf("%s\n", ev)

// Output:
// Four, Three, Two-low [4c 3s 2h]
```

A returned [`Eval`][eval] of a [`HandRank`][hand-rank] to determine the
relative rank of a `Hand`, on a low-to-high basis. Higher poker hands
have a lower value `HandRank` than lower poker hands. For example, a
[`StraightFlush`][hand-rank] will have a lower `HandRank` than a
[`FullHouse`][hand-rank].

If an invalid number of cards is passed to a `Type`'s `EvalFunc`, the `Eval`'s
[`HiRank`][eval.hi-rank] and [`LoRank`][eval.lo-rank] values will be set to
[`Invalid`][invalid].

Currently, no `Type`'s supports evaluating hands containing more than 7
`Card`'s.

When a [`Eval`][hand] is created, a Hi and Lo (if applicable) `HandRank` is
evaluated, and made available via [`Eval.HiRank`][eval.hi-rank] and
[`Eval.LoRank`][eval.lo-rank] methods, respectively. The Hi and Lo
`HandRank`'s are evaluated by a `EvalRankFunc`, dependent on the `Hand`'s
[`Type`][type]:

#### Cactus Kev

For regular poker hand types, poker hand rank is determined by Go
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
comparing a [`Hand`'s][hand] [`HiRank`][eval.hi-rank] or [`LoRank`][eval.lo-rank]
against another hand's. Two or more hands having a `HandRank` of equal value
indicate that the hands have equivalent ranks, and thus have both won.

Thus, when `Hand`'s are sorted (low-to-high) by `HiRank` or `LoRank`, the
winner(s) of a hand will be the hands in the lowest position and having
equivalent `HiRank`'s or `LoRank`'s.

#### Comparing Hands

A [`Hand`][hand] can be compared to another `Hand` using
[`HiComp`][eval.hi-comp] and [`LoComp`][eval.lo-comp].

`HiComp` and `LowComp` return `-1`, `0`, or `+1`, making it easy to compare
or sort hands:

```go
// Compare hi hands:
if hand1.HiComp()(hand1, hand2) < 0 {
	fmt.Printf("%s is a winner!", hand1)
}

// Compare lo hands:
if hand1.LoComp()(hand1, hand2) == 0 {
	fmt.Printf("%s and %s are equal!", hand1, hand2)
}

// Sort hi hands:
hi := hands[0].HiComp()
sort.Slice(hands, func(i, j int) bool {
	return hi(hands[i], hands[j]) < 0
})

// Sort lo hands:
lo := hands[0].LoComp()
sort.Slice(hands, func(i, j int) bool {
	return lo(hands[i], hands[j]) < 0
})
```

#### Ordering Hands

[`HiOrder`][hi-order] and [`LoOrder`][lo-order] determine the winner(s) of a
hand by ordering the indexes of a [`[]*Hand`][hand] and returning the list of
ordered hands as a `[]int` and an `int` pivot indicating the position within
the returned list demarcating winning and losing hands.

A `Hand` whose index is in position `i < pivot` is considered to be the
winner(s) of the eval. Hi hands are guaranteed to have 1 or more winner(s),
while Lo hands have 0 or more winner(s):

```go
// Order hi hands by lowest hand rank, low to high:
h, pivot := cardrank.Order(hands)
for i := 0; i < pivot; i++ {
	fmt.Printf("%s is a Hi winner!", hands[h[i]])
}

// Order lo hands by lowest hand rank, low to high:
// (applicable only for a hand types that supports lo hands)
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
[eval-ranking]: #eval-ranking
[build-tags]: #build-tags
[winners]: #winner-determination

[card]: https://pkg.go.dev/github.com/cardrank/cardrank#Card
[suit]: https://pkg.go.dev/github.com/cardrank/cardrank#Suit
[rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Rank
[deck]: https://pkg.go.dev/github.com/cardrank/cardrank#Deck
[hand]: https://pkg.go.dev/github.com/cardrank/cardrank#Hand
[hand-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#HandRank
[type]: https://pkg.go.dev/github.com/cardrank/cardrank#Type
[init]: https://pkg.go.dev/github.com/cardrank/cardrank#Init
[hi-order]: https://pkg.go.dev/github.com/cardrank/cardrank#HiOrder
[lo-order]: https://pkg.go.dev/github.com/cardrank/cardrank#LoOrder
[eval.hi-comp]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.HiComp
[eval.lo-comp]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.LoComp
[eval.hi-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.HiRank
[eval.lo-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.LoRank
[default-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#DefaultRank
[cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#Cactus
[cactus-fast]: https://pkg.go.dev/github.com/cardrank/cardrank#CactusFast
[two-plus-two]: https://pkg.go.dev/github.com/cardrank/cardrank#TwoPlusTwo
[hybrid]: https://pkg.go.dev/github.com/cardrank/cardrank#Hybrid

<!-- START links -->
[holdem-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Holdem
[short-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Short
[royal-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Royal
[omaha-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Omaha
[omaha-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-OmahaHiLo
[stud-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Stud
[stud-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-StudHiLo
[razz-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Razz
[badugi-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Badugi
<!-- END -->

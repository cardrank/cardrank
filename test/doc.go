// Package test benchmarks [cardrank] against other well-known, pure Go poker
// hand evaluators, and verifies that they all rank hands identically.
//
// Evaluators are benchmarked with 5, 6, and 7 card hands, and with Omaha hands
// of 4, 5, and 6 pocket cards and 5 board cards. None of the other evaluators
// support Omaha, so their Omaha evals rank every combination of 2 pocket and 3
// board cards with their 5 card eval.
//
// It is a separate module so that the comparison libraries never become
// dependencies of cardrank itself. Run from this directory:
//
//	go test -bench .
//
// Use 'go mod tidy -e' to tidy this module: the riverboat module predates
// module graph pruning, so tidy also walks its tests, which import a
// notnil/joker package that no longer exists.
//
// [cardrank]: https://pkg.go.dev/github.com/cardrank/cardrank
package test

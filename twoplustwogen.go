//go:build ignore

// go implementation of the two-plus-two handrank table generator.
// from https://github.com/tangentforks/TwoPlusTwoHandEvaluator
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/cardrank/cardrank"
)

// references:
// http://archives1.twoplustwo.com/showflat.php?Cat=0&Number=8513906
// https://web.archive.org/web/20111103160502/http://www.codingthewheel.com/archives/poker-hand-evaluator-roundup#2p2

func main() {
	verbose := flag.Bool("v", true, "verbose")
	out := flag.String("out", "twoplustwo%02d.dat", "out")
	sum := flag.String("sum", "5de2fa6f53f4340d7d91ad605a6400fb", "md5 sum")
	flag.Parse()
	if err := run(*verbose, *out, *sum); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(verbose bool, out, sum string) error {
	logf := func(string, ...interface{}) {}
	if verbose {
		logf = func(s string, v ...interface{}) {
			fmt.Fprintf(os.Stdout, s, v...)
		}
	}
	// step through the ID array - always shifting the current ID and adding 52
	// cards to the end of the array. when I am at 7 cards put the Hand Rank
	// in!! stepping through the ID array is perfect!!
	tbl := NewTwoPlusTwoGenerator(logf)
	counts := make([]int, 10)
	// Store the total of each type of hand (One Pair, Flush, etc)
	// another algorithm right off the thread
	// var c0, c1, c2, c3, c4, c5, c6 int
	// var u0, u1, u2, u3, u4, u5 int
	var total int
	for c0 := uint32(1); c0 < 53; c0++ {
		u0 := tbl[53+c0]
		for c1 := c0 + 1; c1 < 53; c1++ {
			u1 := tbl[u0+c1]
			for c2 := c1 + 1; c2 < 53; c2++ {
				u2 := tbl[u1+c2]
				for c3 := c2 + 1; c3 < 53; c3++ {
					u3 := tbl[u2+c3]
					for c4 := c3 + 1; c4 < 53; c4++ {
						u4 := tbl[u3+c4]
						for c5 := c4 + 1; c5 < 53; c5++ {
							u5 := tbl[u4+c5]
							for c6 := c5 + 1; c6 < 53; c6++ {
								counts[tbl[u5+c6]>>12]++
								total++
							}
						}
					}
				}
			}
		}
	}
	exp := []struct {
		r     cardrank.EvalRank
		count int
	}{
		{cardrank.Invalid, 0},
		{cardrank.HighCard, 23294460},
		{cardrank.Pair, 58627800},
		{cardrank.TwoPair, 31433400},
		{cardrank.ThreeOfAKind, 6461620},
		{cardrank.Straight, 6180020},
		{cardrank.Flush, 4047644},
		{cardrank.FullHouse, 3473184},
		{cardrank.FourOfAKind, 224848},
		{cardrank.StraightFlush, 41584},
	}
	for i := 0; i <= 9; i++ {
		if exp[i].count != counts[i] {
			return fmt.Errorf("expected %s to have count %d, got: %d", exp[i].r, exp[i].count, counts[i])
		}
		logf("%16s: %d\n", exp[i].r, counts[i])
	}
	if total != 133784560 {
		return fmt.Errorf("expected total count of %d, got: %d", 133784560, total)
	}
	logf("%16s: %d\n", "Total", total)
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, tbl); err != nil {
		return err
	}
	b := buf.Bytes()
	hash := fmt.Sprintf("%x", md5.Sum(b))
	if hash != sum {
		return fmt.Errorf("expected hash of %s, got: %s", sum, hash)
	}
	logf("%16s: matches!\n", "Hash")
	for i := 0; i < len(b); i += TenMiB {
		n, name := min(len(b), i+TenMiB), fmt.Sprintf(out, i/TenMiB)
		if err := ioutil.WriteFile(name, b[i:n], 0o644); err != nil {
			return fmt.Errorf("unable to write %s: %w", name, err)
		}
	}
	return nil
}

const TenMiB = 10 * (1 << 20)

type TwoPlusTwoGenerator struct {
	f     cardrank.EvalFunc
	ev    *cardrank.Eval
	ids   []int64
	tbl   []uint32
	count uint32
	max   int64
}

func NewTwoPlusTwoGenerator(logf func(string, ...interface{})) []uint32 {
	g := &TwoPlusTwoGenerator{
		f:     cardrank.NewEval(cardrank.RankCactus),
		ev:    cardrank.EvalOf(cardrank.Holdem),
		ids:   make([]int64, 612978),
		tbl:   make([]uint32, 32487834),
		count: 1,
		max:   0,
	}
	// Jmd: Okay, this loop is going to fill up the IDs[] array which has
	// 612,967 slots. as this loops through and find new combinations it adds
	// them to the end. I need this list to be stable when I set the handranks
	// (next set)  (I do the insertion sort on new IDs these) so I had to get
	// the IDs first and then set the handranks
	for i := 0; g.ids[i] != 0 || i == 0; i++ {
		for j := 0; j < 52; j++ {
			// the ids above contain cards upto the current card.  Now add a
			// new card get the new ID for it and save it in the list if I am
			// not on the 7th card
			if n, id := g.id(g.ids[i], uint32(j)); n < 7 {
				_ = g.insert(id)
			}
		}
		logf("\r%16s: %6d ", "Generating", i) // show progress -- this counts up to 612976
	}
	logf("(done)\n")
	// this is as above, but will not add anything to the ID list, so it is stable
	var max uint32
	for i := uint32(0); g.ids[i] != 0 || i == 0; i++ {
		var n int
		var id int64
		for j := uint32(0); j < 52; j++ {
			var pos uint32
			if n, id = g.id(g.ids[i], j); n < 7 {
				// when in the index mode (< 7 cards) get the id to save
				pos = g.insert(id)*53 + 53
			} else {
				// if I am at the 7th card, get the equivalence class ("hand rank") to save
				pos = uint32(g.eval(id))
			}
			// start at 1 so I have a zero catching entry (just in case)
			max = i*53 + j + 54 // find where to put it
			g.tbl[max] = pos    // and save the pointer to the next card or the handrank
		}
		if n == 6 || n == 7 {
			// an extra, If you want to know what the handrank when there is 5 or 6 cards
			// you can just do HR[u3] or HR[u4] from below code for Handrank of the 5 or
			// 6 card hand
			// this puts the above handrank into the array
			g.tbl[i*53+53] = uint32(g.eval(g.ids[i]))
		}
		logf("\r%16s: %6d ", "Evaluating", i) // show the progress -- counts to 612976 again
	}
	logf("(done)\n%16s: %d\n%16s: %d\n", "ID Count", g.count, "Max", max)
	return g.tbl
}

// id creates an id for card returning the number of cards and created id.
// generated id is a 64 bit value with each card represented by 8 bits.
func (g *TwoPlusTwoGenerator) id(id int64, card uint32) (int, int64) {
	v := make([]uint32, 8) // intentionally keeping one as a 0 end
	// add first card. formats card to rrrr00ss
	v[0] = (((card >> 2) + 1) << 4) + (card & 3) + 1
	// can't have more than 6 cards!
	for i := 0; i < 6; i++ {
		// leave the 0 hole for new card
		v[i+1] = uint32((id >> (8 * i)) & 0xff)
	}
	ranks, suits, dupe := make([]int, 13+1), make([]int, 4+1), false
	var n int
	for n = 0; v[n] != 0; n++ {
		suits[v[n]&0xf]++
		ranks[(v[n]>>4)&0xf]++
		if n != 0 && v[0] == v[n] {
			// can't have the same card twice, so need to bail
			dupe = true
		}
	}
	// has duplicate card (ignore this one)
	if dupe {
		return n, 0
	}
	if n > 4 {
		for rank := 1; rank < 14; rank++ {
			// if I have more than 4 of a rank then I shouldn't do this one!!
			// can't have more than 4 of a rank so return an ID that can't be!
			if ranks[rank] > 4 {
				return n, 0
			}
		}
	}
	// However in the ID process I preferred that 2s = 0x21, 3s = 0x31,....
	// Kc = 0xD4, Ac = 0xE4 This allows me to sort in Rank then Suit order

	// for suit to be significant, need to have n-2 of same suit if we don't
	// have at least 2 cards of the same suit for 4, we make this card suit 0.
	if required := n - 2; required > 1 {
		for i := 0; i < n; i++ { // for each card
			if suits[v[i]&0xf] < required {
				// check suitcount to the number I need to have suits
				// significant if not enough - 0 out the suit - now this suit
				// would be a 0 vs 1-4
				v[i] &= 0xf0
			}
		}
	}

	// sort
	swap := func(i, j int) {
		if v[i] < v[j] {
			v[i], v[j] = v[j], v[i]
		}
	}
	swap(0, 4)
	swap(1, 5)
	swap(2, 6)
	swap(0, 2)
	swap(1, 3)
	swap(4, 6)
	swap(2, 4)
	swap(3, 5)
	swap(0, 1)
	swap(2, 3)
	swap(4, 5)
	swap(1, 4)
	swap(3, 6)
	swap(1, 2)
	swap(3, 4)
	swap(5, 6)

	// put the pieces into a int64 -- cards in bytes -- 66554433221100
	// id is a 64 bit value with each card represented by 8 bits.
	return n, int64(v[0]) +
		(int64(v[1]) << 8) +
		(int64(v[2]) << 16) +
		(int64(v[3]) << 24) +
		(int64(v[4]) << 32) +
		(int64(v[5]) << 40) +
		(int64(v[6]) << 48)
}

// insert inserts a hand ID into ids.
func (g *TwoPlusTwoGenerator) insert(id int64) uint32 {
	switch {
	case id == 0:
		// don't use up a record for a 0!
		return 0
	case id >= g.max:
		// take care of the most likely first goes on the end...
		if id > g.max { // greater than create new else it was the last one!
			g.ids[g.count] = id // add the new ID
			g.count++
			g.max = id
		}
		return g.count - 1
	}
	// find the slot (by a pseudo bsearch algorithm)
	i, n := uint32(0), g.count-1
	for n-i > 1 {
		j := (n + i + 1) / 2
		switch k := g.ids[j] - id; {
		case k > 0:
			n = j
		case k < 0:
			i = j
		default:
			return j
		}
	}
	// it couldn't be found so must be added to the current location (high)
	// make space...  don't expect this much!
	copy(g.ids[n+1:], g.ids[n:])
	g.ids[n] = id // do the insert into the hole created
	g.count++
	return n
}

// eval converts a 64bit handID to an absolute ranking.
//
// I guess I have some explaining to do here... I used the Cactus Kevs Eval ref
// http://www.suffecool.net/poker/evaluator.html I Love the pokersource for
// speed, but I needed to do some tweaking to get it my way and Cactus Kevs
// stuff was easy to tweak ;-)
func (g *TwoPlusTwoGenerator) eval(id int64) cardrank.EvalRank {
	// bail if bad id
	if id == 0 {
		return 0
	}
	v, n, suit := make([]uint32, 8), 0, uint32(20)
	for i := 0; i < 7; i, n = i+1, n+1 {
		// convert all 7 cards (0s are ok)
		if v[i] = uint32((id >> (8 * i)) & 0xff); v[i] == 0 {
			// once I hit a 0 I know I am done
			break
		}
		// if not 0 then count the card
		if s := v[i] & 0xf; s != 0 {
			// if suit is significant, save
			suit = s
		}
	}
	// intentionally keeping one with a 0 end
	p := make([]cardrank.Card, 8)
	// changed as per Ray Wotton's comment at http://archives1.twoplustwo.com/showflat.php?Cat=0&Number=8513906&page=0&fpart=18&vc=1
	for i, j := 0, uint32(1); i < n; i++ {
		// convert to cactus kev way
		// ref http://www.suffecool.net/poker/evaluator.html
		// +--------+--------+--------+--------+
		// |xxxbbbbb|bbbbbbbb|cdhsrrrr|xxpppppp|
		// +--------+--------+--------+--------+
		// p = prime number of rank (deuce=2,trey=3,four=5,five=7,...,ace=41)
		// r = rank of card (deuce=0,trey=1,four=2,five=3,...,ace=12)
		// cdhs = suit of card
		// b = bit turned on depending on rank of card
		// rank is top 4 bits 1-13 so convert
		// suit is bottom 4 bits 1-4, order is different, but who cares?
		r, s := (v[i]>>4)-1, v[i]&0xf
		if s == 0 {
			// if suit is not significant
			s = j
			// loop through available suits
			if j = j + 1; j == 5 {
				j = 1
			}
			if s == suit { // if it was the sigificant suit...  Don't want extras!!
				// skip it
				s = j
				if j = j + 1; j == 5 { // roll 1-4
					j = 1
				}
			}
		}
		// Cactus Kev's value
		p[i] = cardrank.Card(primes[r] | (r << 8) | (1 << (s + 11)) | (1 << (16 + r)))
	}

	if n != 5 && n != 6 && n != 7 {
		// problem!!  shouldn't hit this...
		panic("invalid number of cards " + strconv.Itoa(n))
	}

	// (eval)
	// I would like to change the format of Catus Kev's ret value to:
	// hhhhrrrrrrrrrrrr   hhhh = 1 high card -> 9 straight flush
	// r..r = rank within the above  1 to max of 2861
	// now the worst hand = 1
	g.f(g.ev, p[:n], nil)
	result := cardrank.Nothing - g.ev.HiRank + 1
	switch {
	case result < 1278:
		// 1277 high card
		result = result - 0 + 4096*1
	case result < 4138:
		// 2860 one pair
		result = result - 1277 + 4096*2
	case result < 4996:
		// 858 two pair
		result = result - 4137 + 4096*3
	case result < 5854:
		// 858 three-kind
		result = result - 4995 + 4096*4
	case result < 5864:
		// 10 straights
		result = result - 5853 + 4096*5
	case result < 7141:
		// 1277 flushes
		result = result - 5863 + 4096*6
	case result < 7297:
		// 156 full house
		result = result - 7140 + 4096*7
	case result < 7453:
		// 156 four-kind
		result = result - 7296 + 4096*8
	default:
		// 10 str.flushes
		result = result - 7452 + 4096*9
	}
	// now a handrank that I like
	return result
}

// ordered is the ordered constraint.
type ordered interface {
	~float32 | ~float64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// min returns the min of a, b.
func min[T ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

var primes = [...]uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}

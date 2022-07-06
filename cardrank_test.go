package cardrank

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRanker(t *testing.T) {
	for _, r := range rankers(true) {
		for i, f := range []func() []test{
			fiveCardTests,
			sixCardTests,
			sevenCardTests,
		} {
			ranker, tests := r, f()
			t.Run(fmt.Sprintf("%s/%d", ranker.name, i+5), func(t *testing.T) {
				for j, test := range tests {
					hand := make([]Card, len(test.hand))
					copy(hand, test.hand)
					rank := ranker.ranker(hand)
					if rank != test.r {
						t.Errorf("test %d %d expected %d, got: %d", i, j, test.r, rank)
					}
					if fixed := rank.Fixed(); fixed != test.exp {
						t.Errorf("test %d %d expected %s, got: %s", i, j, test.exp, fixed)
					}
					h := NewHand(Holdem, test.hand[:5], test.hand[5:])
					if s := fmt.Sprintf("%b %b", h, h.Unused()); s != test.v {
						t.Errorf("test %d %d expected %q, got: %q", i, j, test.v, s)
					}
				}
			})
		}
	}
}

func TestLow(t *testing.T) {
	tests := []struct {
		v string
		r HandRank
		f RankFiveFunc
	}{
		{"Kh Qh Jh Th 9h", 7936, LowRanker},
		{"9h 7h 6h 5h 4h", 33144, EightOrBetterRanker},
	}
	for i, test := range tests {
		ranker := HandRanker(test.f)
		if n, exp := ranker(Must(test.v)), test.r; n != exp {
			t.Errorf("test %d expected rank %d, got: %d", i, exp, n)
		}
	}
}

func TestEightOrBetterRanker(t *testing.T) {
	p0 := Must("Ah 2h 3h 4h 5h 6h 7h 8h")
	for i := Nine; i <= King; i++ {
		p1 := Must(i.String() + "h 4h 3h 2h Ah")
		r1 := EightOrBetterRanker(p1[0], p1[1], p1[2], p1[3], p1[4])
		for c0 := 0; c0 < len(p0); c0++ {
			for c1 := c0 + 1; c1 < len(p0); c1++ {
				for c2 := c1 + 1; c2 < len(p0); c2++ {
					for c3 := c2 + 1; c3 < len(p0); c3++ {
						for c4 := c3 + 1; c4 < len(p0); c4++ {
							h0 := []Card{p0[c0], p0[c1], p0[c2], p0[c3], p0[c4]}
							r0 := EightOrBetterRanker(h0[0], h0[1], h0[2], h0[3], h0[4])
							if r0 >= r1 {
								t.Errorf("%s does not have lower rank than %s", h0, p1)
							}
						}
					}
				}
			}
		}
	}
}

func TestRazzRanker(t *testing.T) {
	tests := []struct {
		v string
		r HandRank
	}{
		{"Kh Qh Jh Th 9h", 7936},
		{"Ah Kh Qh Jh Th", 7681},
		{"2h 2c 2d 2s As", 65380},
		{"Ah Ac Ad Ks Kh", 65368},
		{"Ah Ac Ad Ks Qh", 63925},
		{"Kh Kd Qd Qs Jh", 62934},
		{"3h 3c Kh Qd Jd", 59734},
		{"2h 2c Kh Qd Jd", 59514},
		{"3h 2c Kh Qd Jd", 7174},
	}
	ranker := HandRanker(RazzRanker)
	for i, test := range tests {
		if n, exp := ranker(Must(test.v)), test.r; n != exp {
			t.Errorf("test %d expected rank %d, got: %d", i, exp, n)
		}
	}
}

func TestAllCards(t *testing.T) {
	if !strings.Contains(os.Getenv("TESTS"), "allCards") {
		t.Logf("skipping: $ENV{TESTS} does not contain 'allCards'")
		return
	}
	if cactus == nil {
		t.Logf("skipping: Cactus ranker is not available")
		return
	}
	ranker, rankers := HandRanker(cactus), rankers(false)
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								hand := []Card{allCards[c0], allCards[c1], allCards[c2], allCards[c3], allCards[c4], allCards[c5], allCards[c6]}
								exp := ranker(hand)
								for _, rnk := range rankers {
									if r := rnk.ranker(hand); r != exp {
										t.Errorf("test %s(%b) expected %d (%s), got: %d (%s)", rnk.name, hand, exp, exp.Fixed(), r, r.Fixed())
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

var allCards []Card

func init() {
	allCards = make([]Card, 52)
	copy(allCards, unshuffled)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(52, func(i, j int) {
		allCards[i], allCards[j] = allCards[j], allCards[i]
	})
}

type test struct {
	hand []Card
	r    HandRank
	exp  HandRank
	v    string
}

func fiveCardTests() []test {
	return []test{
		{Must("As Ks Jc 7h 5d"), 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] []"},
		{Must("As Ac Jc 7h 5d"), 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] []"},
		{Must("Jd 6s 6c 5c 5d"), 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] []"},
		{Must("6s 6c Jc Jd 5d"), 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] []"},
		{Must("As Ac Jc Jd 5d"), 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] []"},
		{Must("As Ac Ad Jd 5d"), 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] []"},
		{Must("4s 5s 2d 3h Ac"), 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] []"},
		{Must("9s Ks Qd Jh Td"), 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] []"},
		{Must("As Ks Qd Jh Td"), 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] []"},
		{Must("Ts 7s 4s 3s 2s"), 0x0606, Flush, "Flush, Ten-high [T♠ 7♠ 4♠ 3♠ 2♠] []"},
		{Must("4s 4c 4d 2s 2h"), 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♥ 2♠] []"},
		{Must("5s 5c 5d 6s 6h"), 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] []"},
		{Must("6s 6c 6d 5s 5h"), 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] []"},
		{Must("As Ac Ad Ah 5h"), 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] []"},
		{Must("3d 5d 2d 4d Ad"), 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] []"},
		{Must("6♦ 5♦ 4♦ 3♦ 2♦"), 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] []"},
		{Must("9♦ 6♦ 8♦ 5♦ 7♦"), 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] []"},
		{Must("As Ks Qs Js Ts"), 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] []"},
	}
}

func sixCardTests() []test {
	return []test{
		{Must("3d As Ks Jc 7h 5d"), 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦]"},
		{Must("3d As Ac Jc 7h 5d"), 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] [3♦]"},
		{Must("9d Jd 6s 6c 5c 5d"), 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] [9♦]"},
		{Must("3d 6s 6c Jc Jd 5d"), 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] [3♦]"},
		{Must("3d As Ac Jc Jd 5d"), 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] [3♦]"},
		{Must("3d As Ac Ad Jd 5d"), 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] [3♦]"},
		{Must("4s 5s 2d 3h Ac Jd"), 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] [J♦]"},
		{Must("3d 9s Ks Qd Jh Td"), 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] [3♦]"},
		{Must("3d As Ks Qd Jh Td"), 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] [3♦]"},
		{Must("3d Ts 7s 4s 3s 2s"), 0x0606, Flush, "Flush, Ten-high [T♠ 7♠ 4♠ 3♠ 2♠] [3♦]"},
		{Must("3d 4s 4c 4d 2s 2h"), 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♥ 2♠] [3♦]"},
		{Must("3d 5s 5c 5d 6s 6h"), 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] [3♦]"},
		{Must("3d 6s 6c 6d 5s 5h"), 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] [3♦]"},
		{Must("3d As Ac Ad Ah 5h"), 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] [3♦]"},
		{Must("3d 5d 2d 4d Ad 3s"), 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [3♠]"},
		{Must("T♦ 6♦ 5♦ 4♦ 3♦ 2♦"), 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] [T♦]"},
		{Must("J♦ 9♦ 6♦ 8♦ 5♦ 7♦"), 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{Must("7♦ J♦ 9♦ 6♦ 8♦ 5♦"), 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦]"},
		{Must("3d As Ks Qs Js Ts"), 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦]"},
	}
}

func sevenCardTests() []test {
	return []test{
		{Must("2d 3d As Ks Jc 7h 5d"), 0x186c, Nothing, "Nothing, Ace-high, kickers King, Jack, Seven, Five [A♠ K♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{Must("2d 3d As Ac Jc 7h 5d"), 0x0d78, Pair, "Pair, Aces, kickers Jack, Seven, Five [A♣ A♠ J♣ 7♥ 5♦] [3♦ 2♦]"},
		{Must("9d Jd 6s 6c 5c 5d 4d"), 0x0c93, TwoPair, "Two Pair, Sixes over Fives, kicker Jack [6♣ 6♠ 5♣ 5♦ J♦] [9♦ 4♦]"},
		{Must("2d 3d 6s 6c Jc Jd 5d"), 0x0b42, TwoPair, "Two Pair, Jacks over Sixes, kicker Five [J♣ J♦ 6♣ 6♠ 5♦] [3♦ 2♦]"},
		{Must("2d 3d As Ac Jc Jd 5d"), 0x09c1, TwoPair, "Two Pair, Aces over Jacks, kicker Five [A♣ A♠ J♣ J♦ 5♦] [3♦ 2♦]"},
		{Must("2c 3d As Ac Ad Jd 5d"), 0x0664, ThreeOfAKind, "Three of a Kind, Aces, kickers Jack, Five [A♣ A♦ A♠ J♦ 5♦] [3♦ 2♣]"},
		{Must("4s 5s 2d 3h Ac Jd Qs"), 0x0649, Straight, "Straight, Five-high [5♠ 4♠ 3♥ 2♦ A♣] [Q♠ J♦]"},
		{Must("2d 3d 9s Ks Qd Jh Td"), 0x0641, Straight, "Straight, King-high [K♠ Q♦ J♥ T♦ 9♠] [3♦ 2♦]"},
		{Must("2d 3d As Ks Qd Jh Td"), 0x0640, Straight, "Straight, Ace-high [A♠ K♠ Q♦ J♥ T♦] [3♦ 2♦]"},
		{Must("2d 3d Ts 7s 4s 3s 2s"), 0x0606, Flush, "Flush, Ten-high [T♠ 7♠ 4♠ 3♠ 2♠] [3♦ 2♦]"},
		{Must("2d 3d 4s 4c 4d 2s 2h"), 0x012a, FullHouse, "Full House, Fours full of Twos [4♣ 4♦ 4♠ 2♦ 2♥] [2♠ 3♦]"},
		{Must("4d 3d 5s 5c 5d 6s 6h"), 0x011b, FullHouse, "Full House, Fives full of Sixes [5♣ 5♦ 5♠ 6♥ 6♠] [4♦ 3♦]"},
		{Must("4d 3d 6s 6c 6d 5s 5h"), 0x010f, FullHouse, "Full House, Sixes full of Fives [6♣ 6♦ 6♠ 5♥ 5♠] [4♦ 3♦]"},
		{Must("2d 3d As Ac Ad Ah 5h"), 0x0013, FourOfAKind, "Four of a Kind, Aces, kicker Five [A♣ A♦ A♥ A♠ 5♥] [3♦ 2♦]"},
		{Must("3d 5d 2d 4d Ad 3s 4s"), 0x000a, StraightFlush, "Straight Flush, Five-high, Steel Wheel [5♦ 4♦ 3♦ 2♦ A♦] [4♠ 3♠]"},
		{Must("J♦ T♦ 6♦ 5♦ 4♦ 3♦ 2♦"), 0x0009, StraightFlush, "Straight Flush, Six-high [6♦ 5♦ 4♦ 3♦ 2♦] [J♦ T♦]"},
		{Must("7♦ J♦ 9♦ 6♦ 8♦ 5♦ 2♦"), 0x0006, StraightFlush, "Straight Flush, Nine-high [9♦ 8♦ 7♦ 6♦ 5♦] [J♦ 2♦]"},
		{Must("2d 3d As Ks Qs Js Ts"), 0x0001, StraightFlush, "Straight Flush, Ace-high, Royal [A♠ K♠ Q♠ J♠ T♠] [3♦ 2♦]"},
	}
}

type ranker struct {
	name   string
	ranker RankerFunc
}

func rankers(base bool) []ranker {
	var rankers []ranker
	if base && cactus != nil {
		rankers = append(rankers, ranker{"Cactus", HandRanker(cactus)})
	}
	if cactusFast != nil {
		rankers = append(rankers, ranker{"CactusFast", HandRanker(cactusFast)})
	}
	if twoPlusTwo != nil {
		rankers = append(rankers, ranker{"TwoPlusTwo", twoPlusTwo})
	}
	if cactusFast != nil && twoPlusTwo != nil {
		rankers = append(rankers, ranker{"HybridRanker", HybridRanker(cactusFast, twoPlusTwo)})
	}
	return rankers
}

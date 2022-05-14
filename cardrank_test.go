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
	for _, r := range []Ranker{Cactus, CactusFast, TwoPlus, Hybrid} {
		if !r.Available() {
			continue
		}
		for i, f := range []func() []test{
			fiveCardTests,
			sixCardTests,
			sevenCardTests,
		} {
			ranker, tests := r, f()
			t.Run(fmt.Sprintf("%s/%d", ranker, i+5), func(t *testing.T) {
				for j, test := range tests {
					h := NewHand(test.hand[:5], test.hand[5:], r.Rank)
					rank := h.Rank()
					if rank != test.r {
						t.Errorf("test %s %d %d expected %d, got: %d", ranker, i, j, test.r, rank)
					}
					if fixed := rank.Fixed(); fixed != test.exp {
						t.Errorf("test %s %d %d expected %s, got: %s", ranker, i, j, test.exp, fixed)
					}
					if s := fmt.Sprintf("%b %b", h, h.Unused()); s != test.v {
						t.Errorf("test %s %d %d expected %q, got: %q", ranker, i, j, test.v, s)
					}
				}
			})
		}
	}
}

func TestEightOrBetter_Rank(t *testing.T) {
	v := Must("Ah 2h 3h 4h 5h 6h 7h 8h")
	for c0 := 0; c0 < len(v); c0++ {
		for c1 := c0 + 1; c1 < len(v); c1++ {
			for c2 := c1 + 1; c2 < len(v); c2++ {
				for c3 := c2 + 1; c3 < len(v); c3++ {
					for c4 := c3 + 1; c4 < len(v); c4++ {
						hand := []Card{v[c0], v[c1], v[c2], v[c3], v[c4]}
						r := EightOrBetter.Rank(hand)
						t.Logf("%b: %d", hand, r)
					}
				}
			}
		}
	}
	for i := Eight; i <= King; i++ {
		r := EightOrBetter.Rank(Must(i.String() + "h 4h 3h 2h Ah"))
		t.Logf("%d", r)
	}
}

func TestRanker_allCards(t *testing.T) {
	if !strings.Contains(os.Getenv("TESTS"), "allCards") {
		t.Logf("skipping: $ENV{TESTS} does not contain 'allCards'")
		return
	}
	if !Cactus.Available() {
		t.Logf("skipping: Cactus ranker is not available")
		return
	}
	for c0 := 0; c0 < 52; c0++ {
		for c1 := c0 + 1; c1 < 52; c1++ {
			for c2 := c1 + 1; c2 < 52; c2++ {
				for c3 := c2 + 1; c3 < 52; c3++ {
					for c4 := c3 + 1; c4 < 52; c4++ {
						for c5 := c4 + 1; c5 < 52; c5++ {
							for c6 := c5 + 1; c6 < 52; c6++ {
								hand := []Card{allCards[c0], allCards[c1], allCards[c2], allCards[c3], allCards[c4], allCards[c5], allCards[c6]}
								exp := Cactus.Rank(hand)
								for _, ranker := range []Ranker{CactusFast, TwoPlus, Hybrid} {
									if r := ranker.Rank(hand); r != exp {
										t.Errorf("test %s.Rank(%b) expected %d (%s), got: %d (%s)", ranker, hand, exp, exp.Fixed(), r, r.Fixed())
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

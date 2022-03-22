package cardrank_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"cardrank.io/cardrank"
)

func ExampleFromRune() {
	c, err := cardrank.FromRune('🂡')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%b\n", c)
	// Output:
	// A♠
}

func ExampleMustCard() {
	c := cardrank.MustCard("Ah")
	fmt.Printf("%N of %L (%b)\n", c, c, c)
	// Output:
	// Ace of Hearts (A♥)
}

func ExampleCard_unmarshal() {
	var hand []cardrank.Card
	if err := json.Unmarshal([]byte(`["3s", "4c", "5c", "Ah", "2d"]`), &hand); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", hand)
	// Output:
	// [3s 4c 5c Ah 2d]
}

func ExampleDeck_Draw() {
	d := cardrank.NewDeck()
	// note: use a real random source
	rnd := rand.New(rand.NewSource(52))
	d.Shuffle(rnd.Shuffle)
	hand := d.Draw(7)
	fmt.Printf("%b\n", hand)
	// Output:
	// [9♣ 6♥ Q♠ 3♠ J♠ 9♥ K♣]
}

func ExampleNewHand() {
	d := cardrank.NewDeck()
	// note: use a real random source
	rnd := rand.New(rand.NewSource(6265))
	d.Shuffle(rnd.Shuffle)
	hand := d.Draw(5)
	h := cardrank.NewHand(hand, nil, cardrank.CactusFast.Rank)
	fmt.Printf("%b\n", h)
	// Output:
	// Four of a Kind, Eights, kicker Seven [8♣ 8♦ 8♥ 8♠ 7♠]
}

func ExampleHoldem_RankHand() {
	d := cardrank.NewDeck()
	// note: use a real random source
	rnd := rand.New(rand.NewSource(26076))
	d.Shuffle(rnd.Shuffle)
	h := cardrank.Holdem.RankHand(d.Draw(5), d.Draw(2))
	fmt.Printf("%b\n", h)
	// Output:
	// Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣]
}

func Example_holdem() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{3, 2},
		{278062, 2},
		{1928, 6},
		{6151, 6},
		{5680, 6},
		{23965, 2},
		{13959, 2},
		{23366, 6},
		{29555, 3},
		{472600, 3},
		{107, 10},
	} {
		// note: use a real random source
		rnd := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Holdem.Deal(rnd.Shuffle, game.players)
		hands := cardrank.Holdem.RankHands(pockets, board)
		fmt.Printf("------ Holdem %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, hands[j].Pocket(), hands[j].Description(), hands[j].Best(), hands[j].Unused())
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
	// Output:
	// ------ Holdem 1 ------
	// Board:    [J♠ T♠ 2♦ 2♠ Q♥]
	// Player 1: [6♦ 8♠] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [8♠ 6♦]
	// Player 2: [7♠ 4♣] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [7♠ 4♣]
	// Result:   Players 1, 2 push with Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠], [2♦ 2♠ Q♥ J♠ T♠]
	// ------ Holdem 2 ------
	// Board:    [8♠ 9♠ J♠ 9♣ T♠]
	// Player 1: [7♠ T♣] Straight Flush, Jack-high [J♠ T♠ 9♠ 8♠ 7♠] [T♣ 9♣]
	// Player 2: [6♦ Q♠] Straight Flush, Queen-high [Q♠ J♠ T♠ 9♠ 8♠] [9♣ 6♦]
	// Result:   Player 2 wins with Straight Flush, Queen-high [Q♠ J♠ T♠ 9♠ 8♠]
	// ------ Holdem 3 ------
	// Board:    [A♠ T♣ K♠ J♣ 6♥]
	// Player 1: [T♥ 2♠] Pair, Tens, kickers Ace, King, Jack [T♣ T♥ A♠ K♠ J♣] [6♥ 2♠]
	// Player 2: [Q♣ J♠] Straight, Ace-high [A♠ K♠ Q♣ J♣ T♣] [J♠ 6♥]
	// Player 3: [4♥ Q♠] Straight, Ace-high [A♠ K♠ Q♠ J♣ T♣] [6♥ 4♥]
	// Player 4: [5♦ K♦] Pair, Kings, kickers Ace, Jack, Ten [K♦ K♠ A♠ J♣ T♣] [6♥ 5♦]
	// Player 5: [Q♥ 7♣] Straight, Ace-high [A♠ K♠ Q♥ J♣ T♣] [7♣ 6♥]
	// Player 6: [6♠ 3♣] Pair, Sixes, kickers Ace, King, Jack [6♥ 6♠ A♠ K♠ J♣] [T♣ 3♣]
	// Result:   Players 2, 3, 5 push with Straight, Ace-high [A♠ K♠ Q♣ J♣ T♣], [A♠ K♠ Q♠ J♣ T♣], [A♠ K♠ Q♥ J♣ T♣]
	// ------ Holdem 4 ------
	// Board:    [9♦ J♣ A♥ 9♥ J♠]
	// Player 1: [K♠ 7♦] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [K♠ 7♦]
	// Player 2: [A♦ 4♥] Two Pair, Aces over Jacks, kicker Nine [A♦ A♥ J♣ J♠ 9♦] [9♥ 4♥]
	// Player 3: [3♥ T♣] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [T♣ 3♥]
	// Player 4: [8♦ 9♠] Full House, Nines full of Jacks [9♦ 9♥ 9♠ J♣ J♠] [A♥ 8♦]
	// Player 5: [8♥ 6♣] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [8♥ 6♣]
	// Player 6: [5♥ J♦] Full House, Jacks full of Nines [J♣ J♦ J♠ 9♦ 9♥] [A♥ 5♥]
	// Result:   Player 6 wins with Full House, Jacks full of Nines [J♣ J♦ J♠ 9♦ 9♥]
	// ------ Holdem 5 ------
	// Board:    [3♠ 9♥ A♦ 6♥ Q♦]
	// Player 1: [T♦ 8♦] Nothing, Ace-high, kickers Queen, Ten, Nine, Eight [A♦ Q♦ T♦ 9♥ 8♦] [6♥ 3♠]
	// Player 2: [K♠ T♣] Nothing, Ace-high, kickers King, Queen, Ten, Nine [A♦ K♠ Q♦ T♣ 9♥] [6♥ 3♠]
	// Player 3: [7♥ 8♣] Nothing, Ace-high, kickers Queen, Nine, Eight, Seven [A♦ Q♦ 9♥ 8♣ 7♥] [6♥ 3♠]
	// Player 4: [4♥ 7♦] Nothing, Ace-high, kickers Queen, Nine, Seven, Six [A♦ Q♦ 9♥ 7♦ 6♥] [4♥ 3♠]
	// Player 5: [K♥ 5♦] Nothing, Ace-high, kickers King, Queen, Nine, Six [A♦ K♥ Q♦ 9♥ 6♥] [5♦ 3♠]
	// Player 6: [T♥ 5♣] Nothing, Ace-high, kickers Queen, Ten, Nine, Six [A♦ Q♦ T♥ 9♥ 6♥] [5♣ 3♠]
	// Result:   Player 2 wins with Nothing, Ace-high, kickers King, Queen, Ten, Nine [A♦ K♠ Q♦ T♣ 9♥]
	// ------ Holdem 6 ------
	// Board:    [T♥ 6♥ 7♥ 2♥ 7♣]
	// Player 1: [6♣ 6♠] Full House, Sixes full of Sevens [6♣ 6♥ 6♠ 7♣ 7♥] [T♥ 2♥]
	// Player 2: [K♥ 5♥] Flush, King-high [K♥ T♥ 7♥ 6♥ 5♥] [2♥ 7♣]
	// Result:   Player 1 wins with Full House, Sixes full of Sevens [6♣ 6♥ 6♠ 7♣ 7♥]
	// ------ Holdem 7 ------
	// Board:    [4♦ A♥ A♣ 4♠ A♦]
	// Player 1: [T♥ T♠] Full House, Aces full of Tens [A♣ A♦ A♥ T♥ T♠] [4♦ 4♠]
	// Player 2: [9♣ A♠] Four of a Kind, Aces, kicker Four [A♣ A♦ A♥ A♠ 4♦] [4♠ 9♣]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker Four [A♣ A♦ A♥ A♠ 4♦]
	// ------ Holdem 8 ------
	// Board:    [Q♥ T♥ T♠ J♥ K♥]
	// Player 1: [A♥ 9♠] Straight Flush, Ace-high, Royal [A♥ K♥ Q♥ J♥ T♥] [T♠ 9♠]
	// Player 2: [Q♣ 2♠] Two Pair, Queens over Tens, kicker King [Q♣ Q♥ T♥ T♠ K♥] [J♥ 2♠]
	// Player 3: [6♥ 3♦] Flush, King-high [K♥ Q♥ J♥ T♥ 6♥] [T♠ 3♦]
	// Player 4: [8♥ 8♦] Flush, King-high [K♥ Q♥ J♥ T♥ 8♥] [T♠ 8♦]
	// Player 5: [4♦ Q♦] Two Pair, Queens over Tens, kicker King [Q♦ Q♥ T♥ T♠ K♥] [J♥ 4♦]
	// Player 6: [A♦ T♣] Straight, Ace-high [A♦ K♥ Q♥ J♥ T♣] [T♥ T♠]
	// Result:   Player 1 wins with Straight Flush, Ace-high, Royal [A♥ K♥ Q♥ J♥ T♥]
	// ------ Holdem 9 ------
	// Board:    [A♣ 2♣ 4♣ 5♣ 9♥]
	// Player 1: [T♣ J♦] Flush, Ace-high [A♣ T♣ 5♣ 4♣ 2♣] [J♦ 9♥]
	// Player 2: [4♥ 6♠] Pair, Fours, kickers Ace, Nine, Six [4♣ 4♥ A♣ 9♥ 6♠] [5♣ 2♣]
	// Player 3: [3♣ T♠] Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣] [T♠ 9♥]
	// Result:   Player 3 wins with Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣]
	// ------ Holdem 10 ------
	// Board:    [8♣ J♣ 8♥ 7♥ 9♥]
	// Player 1: [8♦ 8♠] Four of a Kind, Eights, kicker Jack [8♣ 8♦ 8♥ 8♠ J♣] [9♥ 7♥]
	// Player 2: [6♥ T♥] Straight Flush, Ten-high [T♥ 9♥ 8♥ 7♥ 6♥] [J♣ 8♣]
	// Player 3: [3♣ K♥] Pair, Eights, kickers King, Jack, Nine [8♣ 8♥ K♥ J♣ 9♥] [7♥ 3♣]
	// Result:   Player 2 wins with Straight Flush, Ten-high [T♥ 9♥ 8♥ 7♥ 6♥]
	// ------ Holdem 11 ------
	// Board:    [5♥ 3♣ J♥ 6♦ 6♣]
	// Player 1: [8♥ 4♥] Pair, Sixes, kickers Jack, Eight, Five [6♣ 6♦ J♥ 8♥ 5♥] [4♥ 3♣]
	// Player 2: [T♣ 3♥] Two Pair, Sixes over Threes, kicker Jack [6♣ 6♦ 3♣ 3♥ J♥] [T♣ 5♥]
	// Player 3: [A♠ 6♠] Three of a Kind, Sixes, kickers Ace, Jack [6♣ 6♦ 6♠ A♠ J♥] [5♥ 3♣]
	// Player 4: [J♠ 8♠] Two Pair, Jacks over Sixes, kicker Eight [J♥ J♠ 6♣ 6♦ 8♠] [5♥ 3♣]
	// Player 5: [6♥ 2♣] Three of a Kind, Sixes, kickers Jack, Five [6♣ 6♦ 6♥ J♥ 5♥] [3♣ 2♣]
	// Player 6: [T♥ Q♣] Pair, Sixes, kickers Queen, Jack, Ten [6♣ 6♦ Q♣ J♥ T♥] [5♥ 3♣]
	// Player 7: [Q♠ 5♦] Two Pair, Sixes over Fives, kicker Queen [6♣ 6♦ 5♦ 5♥ Q♠] [J♥ 3♣]
	// Player 8: [T♠ 2♠] Pair, Sixes, kickers Jack, Ten, Five [6♣ 6♦ J♥ T♠ 5♥] [3♣ 2♠]
	// Player 9: [5♣ 9♦] Two Pair, Sixes over Fives, kicker Jack [6♣ 6♦ 5♣ 5♥ J♥] [9♦ 3♣]
	// Player 10: [J♣ A♣] Two Pair, Jacks over Sixes, kicker Ace [J♣ J♥ 6♣ 6♦ A♣] [5♥ 3♣]
	// Result:   Player 3 wins with Three of a Kind, Sixes, kickers Ace, Jack [6♣ 6♦ 6♠ A♠ J♥]
}

func Example_shortDeck() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{155, 4},
		{384, 8},
		{880, 4},
		{3453, 3},
		{5662, 3},
		{65481, 2},
		{27947, 4},
	} {
		// note: use a real random source
		rnd := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.ShortDeck.Deal(rnd.Shuffle, game.players)
		hands := cardrank.ShortDeck.RankHands(pockets, board)
		fmt.Printf("------ ShortDeck %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, hands[j].Pocket(), hands[j].Description(), hands[j].Best(), hands[j].Unused())
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
	// Output:
	// ------ ShortDeck 1 ------
	// Board:    [9♥ A♦ A♥ 8♣ A♣]
	// Player 1: [8♥ 7♥] Full House, Aces full of Eights [A♣ A♦ A♥ 8♣ 8♥] [9♥ 7♥]
	// Player 2: [A♠ J♦] Four of a Kind, Aces, kicker Jack [A♣ A♦ A♥ A♠ J♦] [9♥ 8♣]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker Jack [A♣ A♦ A♥ A♠ J♦]
	// ------ ShortDeck 2 ------
	// Board:    [9♣ 6♦ A♠ J♠ 6♠]
	// Player 1: [T♥ 6♣] Three of a Kind, Sixes, kickers Ace, Jack [6♣ 6♦ 6♠ A♠ J♠] [T♥ 9♣]
	// Player 2: [6♥ 9♥] Full House, Sixes full of Nines [6♦ 6♥ 6♠ 9♣ 9♥] [A♠ J♠]
	// Player 3: [A♣ 7♣] Two Pair, Aces over Sixes, kicker Jack [A♣ A♠ 6♦ 6♠ J♠] [9♣ 7♣]
	// Player 4: [T♠ K♠] Flush, Ace-high [A♠ K♠ J♠ T♠ 6♠] [9♣ 6♦]
	// Result:   Player 4 wins with Flush, Ace-high [A♠ K♠ J♠ T♠ 6♠]
	// ------ ShortDeck 3 ------
	// Board:    [T♥ J♣ 7♥ 9♥ K♣]
	// Player 1: [8♥ T♠] Straight, Jack-high [J♣ T♥ 9♥ 8♥ 7♥] [K♣ T♠]
	// Player 2: [J♠ 6♣] Pair, Jacks, kickers King, Ten, Nine [J♣ J♠ K♣ T♥ 9♥] [7♥ 6♣]
	// Player 3: [7♦ 8♠] Straight, Jack-high [J♣ T♥ 9♥ 8♠ 7♦] [K♣ 7♥]
	// Player 4: [9♣ A♥] Pair, Nines, kickers Ace, King, Jack [9♣ 9♥ A♥ K♣ J♣] [T♥ 7♥]
	// Player 5: [T♣ Q♠] Straight, King-high [K♣ Q♠ J♣ T♣ 9♥] [T♥ 7♥]
	// Player 6: [7♣ Q♦] Straight, King-high [K♣ Q♦ J♣ T♥ 9♥] [7♣ 7♥]
	// Player 7: [6♠ 8♦] Straight, Jack-high [J♣ T♥ 9♥ 8♦ 7♥] [K♣ 6♠]
	// Player 8: [K♥ K♦] Three of a Kind, Kings, kickers Jack, Ten [K♣ K♦ K♥ J♣ T♥] [9♥ 7♥]
	// Result:   Players 5, 6 push with Straight, King-high [K♣ Q♠ J♣ T♣ 9♥], [K♣ Q♦ J♣ T♥ 9♥]
	// ------ ShortDeck 4 ------
	// Board:    [T♦ 9♣ 9♦ Q♦ 8♦]
	// Player 1: [J♠ T♥] Straight, Queen-high [Q♦ J♠ T♦ 9♣ 8♦] [T♥ 9♦]
	// Player 2: [6♣ A♣] Pair, Nines, kickers Ace, Queen, Ten [9♣ 9♦ A♣ Q♦ T♦] [8♦ 6♣]
	// Player 3: [9♥ 8♠] Full House, Nines full of Eights [9♣ 9♦ 9♥ 8♦ 8♠] [Q♦ T♦]
	// Player 4: [J♦ A♦] Straight Flush, Queen-high [Q♦ J♦ T♦ 9♦ 8♦] [9♣ A♦]
	// Result:   Player 4 wins with Straight Flush, Queen-high [Q♦ J♦ T♦ 9♦ 8♦]
	// ------ ShortDeck 5 ------
	// Board:    [6♠ A♣ 7♦ A♠ 6♦]
	// Player 1: [9♣ T♠] Two Pair, Aces over Sixes, kicker Ten [A♣ A♠ 6♦ 6♠ T♠] [9♣ 7♦]
	// Player 2: [J♥ T♦] Two Pair, Aces over Sixes, kicker Jack [A♣ A♠ 6♦ 6♠ J♥] [T♦ 7♦]
	// Player 3: [K♠ A♥] Full House, Aces full of Sixes [A♣ A♥ A♠ 6♦ 6♠] [K♠ 7♦]
	// Result:   Player 3 wins with Full House, Aces full of Sixes [A♣ A♥ A♠ 6♦ 6♠]
	// ------ ShortDeck 6 ------
	// Board:    [A♣ 6♣ 9♣ T♦ 8♣]
	// Player 1: [6♥ 7♣] Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣] [T♦ 6♥]
	// Player 2: [6♠ 9♠] Two Pair, Nines over Sixes, kicker Ace [9♣ 9♠ 6♣ 6♠ A♣] [T♦ 8♣]
	// Player 3: [J♥ Q♠] Straight, Queen-high [Q♠ J♥ T♦ 9♣ 8♣] [A♣ 6♣]
	// Result:   Player 1 wins with Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣]
	// ------ ShortDeck 7 ------
	// Board:    [K♥ K♦ K♠ K♣ J♣]
	// Player 1: [7♦ T♦] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [T♦ 7♦]
	// Player 2: [8♦ 6♥] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [8♦ 6♥]
	// Result:   Players 1, 2 push with Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣], [K♣ K♦ K♥ K♠ J♣]
	// ------ ShortDeck 8 ------
	// Board:    [8♦ 8♥ 8♠ Q♠ T♦]
	// Player 1: [J♦ T♣] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♣ T♦] [Q♠ J♦]
	// Player 2: [K♠ T♠] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♠] [K♠ Q♠]
	// Player 3: [9♣ J♣] Straight, Queen-high [Q♠ J♣ T♦ 9♣ 8♦] [8♥ 8♠]
	// Player 4: [T♥ 7♥] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♥] [Q♠ 7♥]
	// Result:   Players 1, 2, 4 push with Full House, Eights full of Tens [8♦ 8♥ 8♠ T♣ T♦], [8♦ 8♥ 8♠ T♦ T♠], [8♦ 8♥ 8♠ T♦ T♥]
}

func Example_omaha() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		rnd := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Omaha.Deal(rnd.Shuffle, game.players)
		hands := cardrank.Omaha.RankHands(pockets, board)
		fmt.Printf("------ Omaha %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, hands[j].Pocket(), hands[j].Description(), hands[j].Best(), hands[j].Unused())
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
	// Output:
	// ------ Omaha 1 ------
	// Board:    [3♥ 5♥ 4♥ 7♥ K♣]
	// Player 1: [K♥ 7♣ J♣ 4♣] Two Pair, Kings over Sevens, kicker Five [K♣ K♥ 7♣ 7♥ 5♥] [J♣ 4♣ 3♥ 4♥]
	// Player 2: [A♥ 5♠ Q♠ 2♠] Straight, Five-high [5♥ 4♥ 3♥ 2♠ A♥] [5♠ Q♠ 7♥ K♣]
	// Result:   Player 2 wins with Straight, Five-high [5♥ 4♥ 3♥ 2♠ A♥]
	// ------ Omaha 2 ------
	// Board:    [3♥ 7♣ 3♣ 9♠ 9♣]
	// Player 1: [3♠ 6♦ Q♦ K♦] Three of a Kind, Threes, kickers King, Nine [3♣ 3♥ 3♠ K♦ 9♠] [6♦ Q♦ 7♣ 9♣]
	// Player 2: [J♦ 3♦ Q♣ K♠] Three of a Kind, Threes, kickers King, Nine [3♣ 3♦ 3♥ K♠ 9♠] [J♦ Q♣ 7♣ 9♣]
	// Player 3: [T♦ 2♥ T♠ 8♥] Two Pair, Tens over Nines, kicker Seven [T♦ T♠ 9♣ 9♠ 7♣] [2♥ 8♥ 3♥ 3♣]
	// Player 4: [8♣ 8♦ Q♥ Q♠] Two Pair, Queens over Nines, kicker Seven [Q♥ Q♠ 9♣ 9♠ 7♣] [8♣ 8♦ 3♥ 3♣]
	// Player 5: [6♣ A♥ 4♥ 6♠] Two Pair, Nines over Sixes, kicker Seven [9♣ 9♠ 6♣ 6♠ 7♣] [A♥ 4♥ 3♥ 3♣]
	// Result:   Players 1, 2 push with Three of a Kind, Threes, kickers King, Nine [3♣ 3♥ 3♠ K♦ 9♠], [3♣ 3♦ 3♥ K♠ 9♠]
	// ------ Omaha 3 ------
	// Board:    [J♣ T♥ 4♥ K♣ Q♣]
	// Player 1: [K♠ J♠ 3♠ 5♣] Two Pair, Kings over Jacks, kicker Queen [K♣ K♠ J♣ J♠ Q♣] [3♠ 5♣ T♥ 4♥]
	// Player 2: [7♠ 4♠ Q♠ 3♣] Two Pair, Queens over Fours, kicker King [Q♣ Q♠ 4♥ 4♠ K♣] [7♠ 3♣ J♣ T♥]
	// Player 3: [T♠ 5♥ 3♥ 8♦] Pair, Tens, kickers King, Queen, Eight [T♥ T♠ K♣ Q♣ 8♦] [5♥ 3♥ J♣ 4♥]
	// Player 4: [4♣ 8♥ 2♣ T♦] Flush, King-high [K♣ Q♣ J♣ 4♣ 2♣] [8♥ T♦ T♥ 4♥]
	// Player 5: [6♠ K♦ J♦ 2♠] Two Pair, Kings over Jacks, kicker Queen [K♣ K♦ J♣ J♦ Q♣] [6♠ 2♠ T♥ 4♥]
	// Player 6: [Q♦ 2♦ A♣ T♣] Straight Flush, Ace-high, Royal [A♣ K♣ Q♣ J♣ T♣] [Q♦ 2♦ T♥ 4♥]
	// Result:   Player 6 wins with Straight Flush, Ace-high, Royal [A♣ K♣ Q♣ J♣ T♣]
	// ------ Omaha 4 ------
	// Board:    [2♦ 6♦ 6♣ Q♣ 7♣]
	// Player 1: [6♠ Q♥ 2♣ 9♠] Full House, Sixes full of Queens [6♣ 6♦ 6♠ Q♣ Q♥] [2♣ 9♠ 2♦ 7♣]
	// Player 2: [3♦ T♣ K♥ 4♥] Pair, Sixes, kickers King, Queen, Ten [6♣ 6♦ K♥ Q♣ T♣] [3♦ 4♥ 2♦ 7♣]
	// Player 3: [6♥ J♥ 4♦ Q♦] Full House, Sixes full of Queens [6♣ 6♦ 6♥ Q♣ Q♦] [J♥ 4♦ 2♦ 7♣]
	// Player 4: [A♣ J♣ 5♣ K♠] Flush, Ace-high [A♣ Q♣ J♣ 7♣ 6♣] [5♣ K♠ 2♦ 6♦]
	// Player 5: [K♣ A♠ 8♣ 5♥] Flush, King-high [K♣ Q♣ 8♣ 7♣ 6♣] [A♠ 5♥ 2♦ 6♦]
	// Player 6: [Q♠ J♠ 8♦ 7♥] Two Pair, Queens over Sevens, kicker Six [Q♣ Q♠ 7♣ 7♥ 6♦] [J♠ 8♦ 2♦ 6♣]
	// Result:   Players 1, 3 push with Full House, Sixes full of Queens [6♣ 6♦ 6♠ Q♣ Q♥], [6♣ 6♦ 6♥ Q♣ Q♦]
	// ------ Omaha 5 ------
	// Board:    [4♣ K♣ 6♦ 9♦ 5♠]
	// Player 1: [3♦ T♥ A♣ 7♦] Straight, Seven-high [7♦ 6♦ 5♠ 4♣ 3♦] [T♥ A♣ K♣ 9♦]
	// Player 2: [5♣ 6♠ 4♦ J♠] Two Pair, Sixes over Fives, kicker King [6♦ 6♠ 5♣ 5♠ K♣] [4♦ J♠ 4♣ 9♦]
	// Player 3: [9♠ 3♣ Q♠ 7♠] Straight, Seven-high [7♠ 6♦ 5♠ 4♣ 3♣] [9♠ Q♠ K♣ 9♦]
	// Player 4: [5♦ K♠ T♠ 8♠] Two Pair, Kings over Fives, kicker Nine [K♣ K♠ 5♦ 5♠ 9♦] [T♠ 8♠ 4♣ 6♦]
	// Player 5: [J♥ 7♥ J♣ 2♣] Pair, Jacks, kickers King, Nine, Six [J♣ J♥ K♣ 9♦ 6♦] [7♥ 2♣ 4♣ 5♠]
	// Player 6: [3♠ 7♣ 2♠ 2♥] Straight, Seven-high [7♣ 6♦ 5♠ 4♣ 3♠] [2♠ 2♥ K♣ 9♦]
	// Result:   Players 1, 3, 6 push with Straight, Seven-high [7♦ 6♦ 5♠ 4♣ 3♦], [7♠ 6♦ 5♠ 4♣ 3♣], [7♣ 6♦ 5♠ 4♣ 3♠]
}

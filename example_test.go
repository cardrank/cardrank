package cardrank_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cardrank/cardrank"
)

func ExampleFromRune() {
	c := cardrank.FromRune('🂡')
	fmt.Printf("%b\n", c)
	// Output:
	// A♠
}

func ExampleFromString() {
	c := cardrank.FromString("Ah")
	fmt.Printf("%N of %L (%b)\n", c, c, c)
	// Output:
	// Ace of Hearts (A♥)
}

func ExampleMust() {
	hand := cardrank.Must("Ah K♠ 🃍 J♤ 10h")
	fmt.Printf("%b", hand)
	// Output:
	// [A♥ K♠ Q♦ J♠ T♥]
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
	r := rand.New(rand.NewSource(52))
	d.Shuffle(r, 1)
	hand := d.Draw(7)
	fmt.Printf("%b\n", hand)
	// Output:
	// [9♣ 6♥ Q♠ 3♠ J♠ 9♥ K♣]
}

func ExampleHoldem_New() {
	d := cardrank.NewDeck()
	// note: use a real random source
	r := rand.New(rand.NewSource(26076))
	d.Shuffle(r, 1)
	ev := cardrank.Holdem.New(d.Draw(2), d.Draw(5))
	fmt.Printf("%b\n", ev)
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Holdem.Deal(r, 1, game.players)
		evs := cardrank.Holdem.Eval(pockets, board)
		fmt.Printf("------ Holdem %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Holdem 1 ------
	// Board:    [J♠ T♠ 2♦ 2♠ Q♥]
	// Player 1: [6♦ 7♠] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [7♠ 6♦]
	// Player 2: [8♠ 4♣] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [8♠ 4♣]
	// Result:   Players 1, 2 push with Pair, Twos, kickers Queen, Jack, Ten
	// ------ Holdem 2 ------
	// Board:    [8♠ 9♠ J♠ 9♣ T♠]
	// Player 1: [7♠ 6♦] Straight Flush, Jack-high [J♠ T♠ 9♠ 8♠ 7♠] [9♣ 6♦]
	// Player 2: [T♣ Q♠] Straight Flush, Queen-high [Q♠ J♠ T♠ 9♠ 8♠] [T♣ 9♣]
	// Result:   Player 2 wins with Straight Flush, Queen-high
	// ------ Holdem 3 ------
	// Board:    [A♠ T♣ K♠ J♣ 6♥]
	// Player 1: [T♥ 5♦] Pair, Tens, kickers Ace, King, Jack [T♣ T♥ A♠ K♠ J♣] [6♥ 5♦]
	// Player 2: [2♠ K♦] Pair, Kings, kickers Ace, Jack, Ten [K♦ K♠ A♠ J♣ T♣] [6♥ 2♠]
	// Player 3: [Q♣ Q♥] Straight, Ace-high [A♠ K♠ Q♣ J♣ T♣] [Q♥ 6♥]
	// Player 4: [J♠ 7♣] Pair, Jacks, kickers Ace, King, Ten [J♣ J♠ A♠ K♠ T♣] [7♣ 6♥]
	// Player 5: [4♥ 6♠] Pair, Sixes, kickers Ace, King, Jack [6♥ 6♠ A♠ K♠ J♣] [T♣ 4♥]
	// Player 6: [Q♠ 3♣] Straight, Ace-high [A♠ K♠ Q♠ J♣ T♣] [6♥ 3♣]
	// Result:   Players 3, 6 push with Straight, Ace-high
	// ------ Holdem 4 ------
	// Board:    [9♦ J♣ A♥ 9♥ J♠]
	// Player 1: [K♠ 8♦] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [K♠ 8♦]
	// Player 2: [7♦ 9♠] Full House, Nines full of Jacks [9♦ 9♥ 9♠ J♣ J♠] [A♥ 7♦]
	// Player 3: [A♦ 8♥] Two Pair, Aces over Jacks, kicker Nine [A♦ A♥ J♣ J♠ 9♦] [9♥ 8♥]
	// Player 4: [4♥ 6♣] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [6♣ 4♥]
	// Player 5: [3♥ 5♥] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [5♥ 3♥]
	// Player 6: [T♣ J♦] Full House, Jacks full of Nines [J♣ J♦ J♠ 9♦ 9♥] [A♥ T♣]
	// Result:   Player 6 wins with Full House, Jacks full of Nines
	// ------ Holdem 5 ------
	// Board:    [3♠ 9♥ A♦ 6♥ Q♦]
	// Player 1: [T♦ 4♥] Nothing, Ace-high, kickers Queen, Ten, Nine, Six [A♦ Q♦ T♦ 9♥ 6♥] [4♥ 3♠]
	// Player 2: [8♦ 7♦] Nothing, Ace-high, kickers Queen, Nine, Eight, Seven [A♦ Q♦ 9♥ 8♦ 7♦] [6♥ 3♠]
	// Player 3: [K♠ K♥] Pair, Kings, kickers Ace, Queen, Nine [K♥ K♠ A♦ Q♦ 9♥] [6♥ 3♠]
	// Player 4: [T♣ 5♦] Nothing, Ace-high, kickers Queen, Ten, Nine, Six [A♦ Q♦ T♣ 9♥ 6♥] [5♦ 3♠]
	// Player 5: [7♥ T♥] Nothing, Ace-high, kickers Queen, Ten, Nine, Seven [A♦ Q♦ T♥ 9♥ 7♥] [6♥ 3♠]
	// Player 6: [8♣ 5♣] Nothing, Ace-high, kickers Queen, Nine, Eight, Six [A♦ Q♦ 9♥ 8♣ 6♥] [5♣ 3♠]
	// Result:   Player 3 wins with Pair, Kings, kickers Ace, Queen, Nine
	// ------ Holdem 6 ------
	// Board:    [T♥ 6♥ 7♥ 2♥ 7♣]
	// Player 1: [6♣ K♥] Flush, King-high, kickers Ten, Seven, Six, Two [K♥ T♥ 7♥ 6♥ 2♥] [7♣ 6♣]
	// Player 2: [6♠ 5♥] Flush, Ten-high, kickers Seven, Six, Five, Two [T♥ 7♥ 6♥ 5♥ 2♥] [7♣ 6♠]
	// Result:   Player 1 wins with Flush, King-high, kickers Ten, Seven, Six, Two
	// ------ Holdem 7 ------
	// Board:    [4♦ A♥ A♣ 4♠ A♦]
	// Player 1: [T♥ 9♣] Full House, Aces full of Fours [A♣ A♦ A♥ 4♦ 4♠] [T♥ 9♣]
	// Player 2: [T♠ A♠] Four of a Kind, Aces, kicker Four [A♣ A♦ A♥ A♠ 4♦] [4♠ T♠]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker Four
	// ------ Holdem 8 ------
	// Board:    [Q♥ T♥ T♠ J♥ K♥]
	// Player 1: [A♥ 8♥] Straight Flush, Ace-high, Royal [A♥ K♥ Q♥ J♥ T♥] [8♥ T♠]
	// Player 2: [9♠ 8♦] Straight, King-high [K♥ Q♥ J♥ T♥ 9♠] [T♠ 8♦]
	// Player 3: [Q♣ 4♦] Two Pair, Queens over Tens, kicker King [Q♣ Q♥ T♥ T♠ K♥] [J♥ 4♦]
	// Player 4: [2♠ Q♦] Two Pair, Queens over Tens, kicker King [Q♦ Q♥ T♥ T♠ K♥] [J♥ 2♠]
	// Player 5: [6♥ A♦] Flush, King-high, kickers Queen, Jack, Ten, Six [K♥ Q♥ J♥ T♥ 6♥] [A♦ T♠]
	// Player 6: [3♦ T♣] Three of a Kind, Tens, kickers King, Queen [T♣ T♥ T♠ K♥ Q♥] [J♥ 3♦]
	// Result:   Player 1 wins with Straight Flush, Ace-high, Royal
	// ------ Holdem 9 ------
	// Board:    [A♣ 2♣ 4♣ 5♣ 9♥]
	// Player 1: [T♣ 6♠] Flush, Ace-high, kickers Ten, Five, Four, Two [A♣ T♣ 5♣ 4♣ 2♣] [9♥ 6♠]
	// Player 2: [J♦ 3♣] Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣] [J♦ 9♥]
	// Player 3: [4♥ T♠] Pair, Fours, kickers Ace, Ten, Nine [4♣ 4♥ A♣ T♠ 9♥] [5♣ 2♣]
	// Result:   Player 2 wins with Straight Flush, Five-high, Steel Wheel
	// ------ Holdem 10 ------
	// Board:    [8♣ J♣ 8♥ 7♥ 9♥]
	// Player 1: [8♦ T♥] Straight, Jack-high [J♣ T♥ 9♥ 8♣ 7♥] [8♦ 8♥]
	// Player 2: [8♠ 3♣] Three of a Kind, Eights, kickers Jack, Nine [8♣ 8♥ 8♠ J♣ 9♥] [7♥ 3♣]
	// Player 3: [6♥ K♥] Flush, King-high, kickers Nine, Eight, Seven, Six [K♥ 9♥ 8♥ 7♥ 6♥] [J♣ 8♣]
	// Result:   Player 3 wins with Flush, King-high, kickers Nine, Eight, Seven, Six
	// ------ Holdem 11 ------
	// Board:    [5♥ 3♣ J♥ 6♦ 6♣]
	// Player 1: [8♥ T♥] Pair, Sixes, kickers Jack, Ten, Eight [6♣ 6♦ J♥ T♥ 8♥] [5♥ 3♣]
	// Player 2: [4♥ Q♣] Pair, Sixes, kickers Queen, Jack, Five [6♣ 6♦ Q♣ J♥ 5♥] [4♥ 3♣]
	// Player 3: [T♣ Q♠] Pair, Sixes, kickers Queen, Jack, Ten [6♣ 6♦ Q♠ J♥ T♣] [5♥ 3♣]
	// Player 4: [3♥ 5♦] Two Pair, Sixes over Fives, kicker Jack [6♣ 6♦ 5♦ 5♥ J♥] [3♣ 3♥]
	// Player 5: [A♠ T♠] Pair, Sixes, kickers Ace, Jack, Ten [6♣ 6♦ A♠ J♥ T♠] [5♥ 3♣]
	// Player 6: [6♠ 2♠] Three of a Kind, Sixes, kickers Jack, Five [6♣ 6♦ 6♠ J♥ 5♥] [3♣ 2♠]
	// Player 7: [J♠ 5♣] Two Pair, Jacks over Sixes, kicker Five [J♥ J♠ 6♣ 6♦ 5♣] [5♥ 3♣]
	// Player 8: [8♠ 9♦] Pair, Sixes, kickers Jack, Nine, Eight [6♣ 6♦ J♥ 9♦ 8♠] [5♥ 3♣]
	// Player 9: [6♥ J♣] Full House, Sixes full of Jacks [6♣ 6♦ 6♥ J♣ J♥] [5♥ 3♣]
	// Player 10: [2♣ A♣] Pair, Sixes, kickers Ace, Jack, Five [6♣ 6♦ A♣ J♥ 5♥] [3♣ 2♣]
	// Result:   Player 9 wins with Full House, Sixes full of Jacks
}

func Example_short() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Short.Deal(r, 1, game.players)
		evs := cardrank.Short.Eval(pockets, board)
		fmt.Printf("------ Short %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Short 1 ------
	// Board:    [9♥ A♦ A♥ 8♣ A♣]
	// Player 1: [8♥ A♠] Four of a Kind, Aces, kicker Eight [A♣ A♦ A♥ A♠ 8♣] [8♥ 9♥]
	// Player 2: [7♥ J♦] Three of a Kind, Aces, kickers Jack, Nine [A♣ A♦ A♥ J♦ 9♥] [8♣ 7♥]
	// Result:   Player 1 wins with Four of a Kind, Aces, kicker Eight
	// ------ Short 2 ------
	// Board:    [9♣ 6♦ A♠ J♠ 6♠]
	// Player 1: [T♥ A♣] Two Pair, Aces over Sixes, kicker Jack [A♣ A♠ 6♦ 6♠ J♠] [T♥ 9♣]
	// Player 2: [6♣ 7♣] Three of a Kind, Sixes, kickers Ace, Jack [6♣ 6♦ 6♠ A♠ J♠] [9♣ 7♣]
	// Player 3: [6♥ T♠] Three of a Kind, Sixes, kickers Ace, Jack [6♦ 6♥ 6♠ A♠ J♠] [T♠ 9♣]
	// Player 4: [9♥ K♠] Two Pair, Nines over Sixes, kicker Ace [9♣ 9♥ 6♦ 6♠ A♠] [K♠ J♠]
	// Result:   Players 2, 3 push with Three of a Kind, Sixes, kickers Ace, Jack
	// ------ Short 3 ------
	// Board:    [T♥ J♣ 7♥ 9♥ K♣]
	// Player 1: [8♥ T♣] Straight, Jack-high [J♣ T♣ 9♥ 8♥ 7♥] [K♣ T♥]
	// Player 2: [T♠ Q♠] Straight, King-high [K♣ Q♠ J♣ T♥ 9♥] [T♠ 7♥]
	// Player 3: [J♠ 7♣] Two Pair, Jacks over Sevens, kicker King [J♣ J♠ 7♣ 7♥ K♣] [T♥ 9♥]
	// Player 4: [6♣ Q♦] Straight, King-high [K♣ Q♦ J♣ T♥ 9♥] [7♥ 6♣]
	// Player 5: [7♦ 6♠] Pair, Sevens, kickers King, Jack, Ten [7♦ 7♥ K♣ J♣ T♥] [9♥ 6♠]
	// Player 6: [8♠ 8♦] Straight, Jack-high [J♣ T♥ 9♥ 8♦ 7♥] [K♣ 8♠]
	// Player 7: [9♣ K♥] Two Pair, Kings over Nines, kicker Jack [K♣ K♥ 9♣ 9♥ J♣] [T♥ 7♥]
	// Player 8: [A♥ K♦] Pair, Kings, kickers Ace, Jack, Ten [K♣ K♦ A♥ J♣ T♥] [9♥ 7♥]
	// Result:   Players 2, 4 push with Straight, King-high
	// ------ Short 4 ------
	// Board:    [T♦ 9♣ 9♦ Q♦ 8♦]
	// Player 1: [J♠ 9♥] Straight, Queen-high [Q♦ J♠ T♦ 9♣ 8♦] [9♦ 9♥]
	// Player 2: [T♥ 8♠] Two Pair, Tens over Nines, kicker Queen [T♦ T♥ 9♣ 9♦ Q♦] [8♦ 8♠]
	// Player 3: [6♣ J♦] Straight Flush, Queen-high [Q♦ J♦ T♦ 9♦ 8♦] [9♣ 6♣]
	// Player 4: [A♣ A♦] Flush, Ace-high, kickers Queen, Ten, Nine, Eight [A♦ Q♦ T♦ 9♦ 8♦] [A♣ 9♣]
	// Result:   Player 3 wins with Straight Flush, Queen-high
	// ------ Short 5 ------
	// Board:    [6♠ A♣ 7♦ A♠ 6♦]
	// Player 1: [9♣ T♦] Two Pair, Aces over Sixes, kicker Ten [A♣ A♠ 6♦ 6♠ T♦] [9♣ 7♦]
	// Player 2: [T♠ K♠] Two Pair, Aces over Sixes, kicker King [A♣ A♠ 6♦ 6♠ K♠] [T♠ 7♦]
	// Player 3: [J♥ A♥] Full House, Aces full of Sixes [A♣ A♥ A♠ 6♦ 6♠] [J♥ 7♦]
	// Result:   Player 3 wins with Full House, Aces full of Sixes
	// ------ Short 6 ------
	// Board:    [A♣ 6♣ 9♣ T♦ 8♣]
	// Player 1: [6♥ 9♠] Two Pair, Nines over Sixes, kicker Ace [9♣ 9♠ 6♣ 6♥ A♣] [T♦ 8♣]
	// Player 2: [7♣ J♥] Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣] [J♥ T♦]
	// Player 3: [6♠ Q♠] Pair, Sixes, kickers Ace, Queen, Ten [6♣ 6♠ A♣ Q♠ T♦] [9♣ 8♣]
	// Result:   Player 2 wins with Straight Flush, Nine-high, Iron Maiden
	// ------ Short 7 ------
	// Board:    [K♥ K♦ K♠ K♣ J♣]
	// Player 1: [7♦ 8♦] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [8♦ 7♦]
	// Player 2: [T♦ 6♥] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [T♦ 6♥]
	// Result:   Players 1, 2 push with Four of a Kind, Kings, kicker Jack
	// ------ Short 8 ------
	// Board:    [8♦ 8♥ 8♠ Q♠ T♦]
	// Player 1: [J♦ 9♣] Straight, Queen-high [Q♠ J♦ T♦ 9♣ 8♦] [8♥ 8♠]
	// Player 2: [T♣ J♣] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♣ T♦] [Q♠ J♣]
	// Player 3: [K♠ T♥] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♥] [K♠ Q♠]
	// Player 4: [T♠ 7♥] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♠] [Q♠ 7♥]
	// Result:   Players 2, 3, 4 push with Full House, Eights full of Tens
}

func Example_royal() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{155, 3},
		{384, 4},
		{880, 5},
		{3453, 2},
		{5662, 3},
		{65481, 4},
		{27947, 5},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Royal.Deal(r, 1, game.players)
		evs := cardrank.Royal.Eval(pockets, board)
		fmt.Printf("------ Royal %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Royal 1 ------
	// Board:    [K♦ A♦ T♥ T♣ J♠]
	// Player 1: [A♠ T♠] Full House, Tens full of Aces [T♣ T♥ T♠ A♦ A♠] [K♦ J♠]
	// Player 2: [A♥ K♠] Two Pair, Aces over Kings, kicker Jack [A♦ A♥ K♦ K♠ J♠] [T♣ T♥]
	// Result:   Player 1 wins with Full House, Tens full of Aces
	// ------ Royal 2 ------
	// Board:    [A♣ K♠ J♦ Q♣ J♣]
	// Player 1: [A♠ Q♠] Two Pair, Aces over Queens, kicker King [A♣ A♠ Q♣ Q♠ K♠] [J♣ J♦]
	// Player 2: [T♠ J♥] Straight, Ace-high [A♣ K♠ Q♣ J♣ T♠] [J♦ J♥]
	// Player 3: [K♣ T♥] Straight, Ace-high [A♣ K♣ Q♣ J♣ T♥] [K♠ J♦]
	// Result:   Players 2, 3 push with Straight, Ace-high
	// ------ Royal 3 ------
	// Board:    [K♠ T♦ T♣ Q♦ A♥]
	// Player 1: [T♠ T♥] Four of a Kind, Tens, kicker Ace [T♣ T♦ T♥ T♠ A♥] [K♠ Q♦]
	// Player 2: [J♣ Q♣] Straight, Ace-high [A♥ K♠ Q♣ J♣ T♣] [Q♦ T♦]
	// Player 3: [A♦ K♦] Two Pair, Aces over Kings, kicker Queen [A♦ A♥ K♦ K♠ Q♦] [T♣ T♦]
	// Player 4: [K♥ K♣] Full House, Kings full of Tens [K♣ K♥ K♠ T♣ T♦] [A♥ Q♦]
	// Result:   Player 1 wins with Four of a Kind, Tens, kicker Ace
	// ------ Royal 4 ------
	// Board:    [J♥ A♠ T♥ T♣ K♠]
	// Player 1: [Q♦ T♠] Straight, Ace-high [A♠ K♠ Q♦ J♥ T♣] [T♥ T♠]
	// Player 2: [K♥ T♦] Full House, Tens full of Kings [T♣ T♦ T♥ K♥ K♠] [A♠ J♥]
	// Player 3: [A♣ Q♠] Straight, Ace-high [A♣ K♠ Q♠ J♥ T♣] [A♠ T♥]
	// Player 4: [A♦ J♠] Two Pair, Aces over Jacks, kicker King [A♦ A♠ J♥ J♠ K♠] [T♣ T♥]
	// Player 5: [K♦ J♦] Two Pair, Kings over Jacks, kicker Ace [K♦ K♠ J♦ J♥ A♠] [T♣ T♥]
	// Result:   Player 2 wins with Full House, Tens full of Kings
	// ------ Royal 5 ------
	// Board:    [J♣ K♥ K♠ J♥ Q♣]
	// Player 1: [A♥ T♦] Straight, Ace-high [A♥ K♥ Q♣ J♣ T♦] [K♠ J♥]
	// Player 2: [J♦ Q♠] Full House, Jacks full of Kings [J♣ J♦ J♥ K♥ K♠] [Q♣ Q♠]
	// Result:   Player 2 wins with Full House, Jacks full of Kings
	// ------ Royal 6 ------
	// Board:    [K♥ A♠ K♦ K♠ A♣]
	// Player 1: [J♥ J♠] Full House, Kings full of Aces [K♦ K♥ K♠ A♣ A♠] [J♥ J♠]
	// Player 2: [Q♦ A♥] Full House, Aces full of Kings [A♣ A♥ A♠ K♦ K♥] [K♠ Q♦]
	// Player 3: [Q♠ T♣] Full House, Kings full of Aces [K♦ K♥ K♠ A♣ A♠] [Q♠ T♣]
	// Result:   Player 2 wins with Full House, Aces full of Kings
	// ------ Royal 7 ------
	// Board:    [J♥ T♦ Q♠ K♣ K♥]
	// Player 1: [K♦ J♣] Full House, Kings full of Jacks [K♣ K♦ K♥ J♣ J♥] [Q♠ T♦]
	// Player 2: [T♥ T♠] Full House, Tens full of Kings [T♦ T♥ T♠ K♣ K♥] [Q♠ J♥]
	// Player 3: [A♠ A♥] Straight, Ace-high [A♥ K♣ Q♠ J♥ T♦] [A♠ K♥]
	// Player 4: [Q♣ A♦] Straight, Ace-high [A♦ K♣ Q♣ J♥ T♦] [K♥ Q♠]
	// Result:   Player 1 wins with Full House, Kings full of Jacks
	// ------ Royal 8 ------
	// Board:    [A♠ K♦ Q♦ A♦ A♣]
	// Player 1: [Q♠ J♠] Full House, Aces full of Queens [A♣ A♦ A♠ Q♦ Q♠] [K♦ J♠]
	// Player 2: [T♦ A♥] Four of a Kind, Aces, kicker King [A♣ A♦ A♥ A♠ K♦] [Q♦ T♦]
	// Player 3: [J♥ K♠] Full House, Aces full of Kings [A♣ A♦ A♠ K♦ K♠] [Q♦ J♥]
	// Player 4: [Q♥ J♦] Full House, Aces full of Queens [A♣ A♦ A♠ Q♦ Q♥] [K♦ J♦]
	// Player 5: [K♣ T♥] Full House, Aces full of Kings [A♣ A♦ A♠ K♣ K♦] [Q♦ T♥]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker King
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Omaha.Deal(r, 1, game.players)
		evs := cardrank.Omaha.Eval(pockets, board)
		fmt.Printf("------ Omaha %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Omaha 1 ------
	// Board:    [3♥ 5♥ 4♥ 7♥ K♣]
	// Player 1: [K♥ J♣ A♥ Q♠] Flush, Ace-high, kickers King, Seven, Five, Four [A♥ K♥ 7♥ 5♥ 4♥] [J♣ Q♠ 3♥ K♣]
	// Player 2: [7♣ 4♣ 5♠ 2♠] Two Pair, Sevens over Fives, kicker King [7♣ 7♥ 5♥ 5♠ K♣] [4♣ 2♠ 3♥ 4♥]
	// Result:   Player 1 wins with Flush, Ace-high, kickers King, Seven, Five, Four
	// ------ Omaha 2 ------
	// Board:    [3♥ 7♣ 3♣ 9♠ 9♣]
	// Player 1: [3♠ 3♦ T♠ Q♠] Four of a Kind, Threes, kicker Nine [3♣ 3♦ 3♥ 3♠ 9♠] [T♠ Q♠ 7♣ 9♣]
	// Player 2: [6♦ Q♣ 8♥ 6♣] Flush, Queen-high, kickers Nine, Seven, Six, Three [Q♣ 9♣ 7♣ 6♣ 3♣] [6♦ 8♥ 3♥ 9♠]
	// Player 3: [Q♦ K♠ 8♣ A♥] Pair, Nines, kickers Ace, King, Seven [9♣ 9♠ A♥ K♠ 7♣] [Q♦ 8♣ 3♥ 3♣]
	// Player 4: [K♦ T♦ 8♦ 4♥] Pair, Nines, kickers King, Ten, Seven [9♣ 9♠ K♦ T♦ 7♣] [8♦ 4♥ 3♥ 3♣]
	// Player 5: [J♦ 2♥ Q♥ 6♠] Pair, Nines, kickers Queen, Jack, Seven [9♣ 9♠ Q♥ J♦ 7♣] [2♥ 6♠ 3♥ 3♣]
	// Result:   Player 1 wins with Four of a Kind, Threes, kicker Nine
	// ------ Omaha 3 ------
	// Board:    [J♣ T♥ 4♥ K♣ Q♣]
	// Player 1: [K♠ Q♠ 4♣ J♦] Two Pair, Kings over Queens, kicker Jack [K♣ K♠ Q♣ Q♠ J♣] [4♣ J♦ T♥ 4♥]
	// Player 2: [J♠ 3♣ 8♥ 2♠] Pair, Jacks, kickers King, Queen, Eight [J♣ J♠ K♣ Q♣ 8♥] [3♣ 2♠ T♥ 4♥]
	// Player 3: [3♠ T♠ 2♣ Q♦] Two Pair, Queens over Tens, kicker King [Q♣ Q♦ T♥ T♠ K♣] [3♠ 2♣ J♣ 4♥]
	// Player 4: [5♣ 5♥ T♦ 2♦] Pair, Tens, kickers King, Queen, Five [T♦ T♥ K♣ Q♣ 5♣] [5♥ 2♦ J♣ 4♥]
	// Player 5: [7♠ 3♥ 6♠ A♣] Nothing, Ace-high, kickers King, Queen, Jack, Seven [A♣ K♣ Q♣ J♣ 7♠] [3♥ 6♠ T♥ 4♥]
	// Player 6: [4♠ 8♦ K♦ T♣] Two Pair, Kings over Tens, kicker Queen [K♣ K♦ T♣ T♥ Q♣] [4♠ 8♦ J♣ 4♥]
	// Result:   Player 1 wins with Two Pair, Kings over Queens, kicker Jack
	// ------ Omaha 4 ------
	// Board:    [2♦ 6♦ 6♣ Q♣ 7♣]
	// Player 1: [6♠ K♥ A♣ 8♣] Flush, Ace-high, kickers Queen, Eight, Seven, Six [A♣ Q♣ 8♣ 7♣ 6♣] [6♠ K♥ 2♦ 6♦]
	// Player 2: [Q♥ 4♥ J♣ 5♥] Two Pair, Queens over Sixes, kicker Jack [Q♣ Q♥ 6♣ 6♦ J♣] [4♥ 5♥ 2♦ 7♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠] Full House, Sixes full of Queens [6♣ 6♦ 6♥ Q♣ Q♠] [2♣ 5♣ 2♦ 7♣]
	// Player 4: [9♠ J♥ K♠ J♠] Two Pair, Jacks over Sixes, kicker Queen [J♥ J♠ 6♣ 6♦ Q♣] [9♠ K♠ 2♦ 7♣]
	// Player 5: [3♦ 4♦ K♣ 8♦] Pair, Sixes, kickers King, Queen, Eight [6♣ 6♦ K♣ Q♣ 8♦] [3♦ 4♦ 2♦ 7♣]
	// Player 6: [T♣ Q♦ A♠ 7♥] Two Pair, Queens over Sevens, kicker Six [Q♣ Q♦ 7♣ 7♥ 6♦] [T♣ A♠ 2♦ 6♣]
	// Result:   Player 3 wins with Full House, Sixes full of Queens
	// ------ Omaha 5 ------
	// Board:    [4♣ K♣ 6♦ 9♦ 5♠]
	// Player 1: [3♦ 4♦ 5♦ J♣] Two Pair, Fives over Fours, kicker King [5♦ 5♠ 4♣ 4♦ K♣] [3♦ J♣ 6♦ 9♦]
	// Player 2: [T♥ J♠ K♠ 2♣] Pair, Kings, kickers Jack, Nine, Six [K♣ K♠ J♠ 9♦ 6♦] [T♥ 2♣ 4♣ 5♠]
	// Player 3: [A♣ 9♠ T♠ 3♠] Pair, Nines, kickers Ace, King, Six [9♦ 9♠ A♣ K♣ 6♦] [T♠ 3♠ 4♣ 5♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣] Straight, Nine-high [9♦ 8♠ 7♦ 6♦ 5♠] [3♣ 7♣ 4♣ K♣]
	// Player 5: [5♣ Q♠ J♥ 2♠] Pair, Fives, kickers King, Queen, Nine [5♣ 5♠ K♣ Q♠ 9♦] [J♥ 2♠ 4♣ 6♦]
	// Player 6: [6♠ 7♠ 7♥ 2♥] Pair, Sevens, kickers King, Nine, Six [7♥ 7♠ K♣ 9♦ 6♦] [6♠ 2♥ 4♣ 5♠]
	// Result:   Player 4 wins with Straight, Nine-high
}

func Example_omahaHiLo() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.OmahaHiLo.Deal(r, 1, game.players)
		evs := cardrank.OmahaHiLo.Eval(pockets, board)
		fmt.Printf("------ OmahaHiLo %d ------\n", i+1)
		fmt.Printf("Board: %b\n", board)
		for j := 0; j < game.players; j++ {
			hi, lo := evs[j].Desc(false), evs[j].Desc(true)
			fmt.Printf("Player %d: %b\n", j+1, pockets[j])
			fmt.Printf("  Hi: %s %b %b\n", hi, hi.Best, hi.Unused)
			fmt.Printf("  Lo: %s %b %b\n", lo, lo.Best, lo.Unused)
		}
		hiOrder, hiPivot := cardrank.Order(evs, false)
		loOrder, loPivot := cardrank.Order(evs, true)
		typ := "wins"
		if loPivot == 0 {
			typ = "scoops"
		}
		desc := evs[hiOrder[0]].Desc(false)
		if hiPivot == 1 {
			fmt.Printf("Result (Hi): Player %d %s with %s\n", hiOrder[0]+1, typ, desc)
		} else {
			var s []string
			for j := 0; j < hiPivot; j++ {
				s = append(s, strconv.Itoa(hiOrder[j]+1))
			}
			fmt.Printf("Result (Hi): Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
		if loPivot == 1 {
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Player %d wins with %s\n", loOrder[0]+1, desc)
		} else if loPivot > 1 {
			var s []string
			for j := 0; j < loPivot; j++ {
				s = append(s, strconv.Itoa(loOrder[j]+1))
			}
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Players %s push with %s\n", strings.Join(s, ", "), desc)
		} else {
			fmt.Printf("Result (Lo): no player made a low hand\n")
		}
	}
	// Output:
	// ------ OmahaHiLo 1 ------
	// Board: [3♥ 5♥ 4♥ 7♥ K♣]
	// Player 1: [K♥ J♣ A♥ Q♠]
	//   Hi: Flush, Ace-high, kickers King, Seven, Five, Four [A♥ K♥ 7♥ 5♥ 4♥] [J♣ Q♠ 3♥ K♣]
	//   Lo: None [] []
	// Player 2: [7♣ 4♣ 5♠ 2♠]
	//   Hi: Two Pair, Sevens over Fives, kicker King [7♣ 7♥ 5♥ 5♠ K♣] [4♣ 2♠ 3♥ 4♥]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♥ 2♠] [4♣ 5♠ 7♥ K♣]
	// Result (Hi): Player 1 wins with Flush, Ace-high, kickers King, Seven, Five, Four
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ OmahaHiLo 2 ------
	// Board: [3♥ 7♣ 3♣ 9♠ 9♣]
	// Player 1: [3♠ 3♦ T♠ Q♠]
	//   Hi: Four of a Kind, Threes, kicker Nine [3♣ 3♦ 3♥ 3♠ 9♠] [T♠ Q♠ 7♣ 9♣]
	//   Lo: None [] []
	// Player 2: [6♦ Q♣ 8♥ 6♣]
	//   Hi: Flush, Queen-high, kickers Nine, Seven, Six, Three [Q♣ 9♣ 7♣ 6♣ 3♣] [6♦ 8♥ 3♥ 9♠]
	//   Lo: None [] []
	// Player 3: [Q♦ K♠ 8♣ A♥]
	//   Hi: Pair, Nines, kickers Ace, King, Seven [9♣ 9♠ A♥ K♠ 7♣] [Q♦ 8♣ 3♥ 3♣]
	//   Lo: None [] []
	// Player 4: [K♦ T♦ 8♦ 4♥]
	//   Hi: Pair, Nines, kickers King, Ten, Seven [9♣ 9♠ K♦ T♦ 7♣] [8♦ 4♥ 3♥ 3♣]
	//   Lo: None [] []
	// Player 5: [J♦ 2♥ Q♥ 6♠]
	//   Hi: Pair, Nines, kickers Queen, Jack, Seven [9♣ 9♠ Q♥ J♦ 7♣] [2♥ 6♠ 3♥ 3♣]
	//   Lo: None [] []
	// Result (Hi): Player 1 scoops with Four of a Kind, Threes, kicker Nine
	// Result (Lo): no player made a low hand
	// ------ OmahaHiLo 3 ------
	// Board: [J♣ T♥ 4♥ K♣ Q♣]
	// Player 1: [K♠ Q♠ 4♣ J♦]
	//   Hi: Two Pair, Kings over Queens, kicker Jack [K♣ K♠ Q♣ Q♠ J♣] [4♣ J♦ T♥ 4♥]
	//   Lo: None [] []
	// Player 2: [J♠ 3♣ 8♥ 2♠]
	//   Hi: Pair, Jacks, kickers King, Queen, Eight [J♣ J♠ K♣ Q♣ 8♥] [3♣ 2♠ T♥ 4♥]
	//   Lo: None [] []
	// Player 3: [3♠ T♠ 2♣ Q♦]
	//   Hi: Two Pair, Queens over Tens, kicker King [Q♣ Q♦ T♥ T♠ K♣] [3♠ 2♣ J♣ 4♥]
	//   Lo: None [] []
	// Player 4: [5♣ 5♥ T♦ 2♦]
	//   Hi: Pair, Tens, kickers King, Queen, Five [T♦ T♥ K♣ Q♣ 5♣] [5♥ 2♦ J♣ 4♥]
	//   Lo: None [] []
	// Player 5: [7♠ 3♥ 6♠ A♣]
	//   Hi: Nothing, Ace-high, kickers King, Queen, Jack, Seven [A♣ K♣ Q♣ J♣ 7♠] [3♥ 6♠ T♥ 4♥]
	//   Lo: None [] []
	// Player 6: [4♠ 8♦ K♦ T♣]
	//   Hi: Two Pair, Kings over Tens, kicker Queen [K♣ K♦ T♣ T♥ Q♣] [4♠ 8♦ J♣ 4♥]
	//   Lo: None [] []
	// Result (Hi): Player 1 scoops with Two Pair, Kings over Queens, kicker Jack
	// Result (Lo): no player made a low hand
	// ------ OmahaHiLo 4 ------
	// Board: [2♦ 6♦ 6♣ Q♣ 7♣]
	// Player 1: [6♠ K♥ A♣ 8♣]
	//   Hi: Flush, Ace-high, kickers Queen, Eight, Seven, Six [A♣ Q♣ 8♣ 7♣ 6♣] [6♠ K♥ 2♦ 6♦]
	//   Lo: Eight, Seven, Six, Two, Ace-low [8♣ 7♣ 6♦ 2♦ A♣] [6♠ K♥ 6♣ Q♣]
	// Player 2: [Q♥ 4♥ J♣ 5♥]
	//   Hi: Two Pair, Queens over Sixes, kicker Jack [Q♣ Q♥ 6♣ 6♦ J♣] [4♥ 5♥ 2♦ 7♣]
	//   Lo: Seven, Six, Five, Four, Two-low [7♣ 6♦ 5♥ 4♥ 2♦] [Q♥ J♣ 6♣ Q♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠]
	//   Hi: Full House, Sixes full of Queens [6♣ 6♦ 6♥ Q♣ Q♠] [2♣ 5♣ 2♦ 7♣]
	//   Lo: None [] []
	// Player 4: [9♠ J♥ K♠ J♠]
	//   Hi: Two Pair, Jacks over Sixes, kicker Queen [J♥ J♠ 6♣ 6♦ Q♣] [9♠ K♠ 2♦ 7♣]
	//   Lo: None [] []
	// Player 5: [3♦ 4♦ K♣ 8♦]
	//   Hi: Pair, Sixes, kickers King, Queen, Eight [6♣ 6♦ K♣ Q♣ 8♦] [3♦ 4♦ 2♦ 7♣]
	//   Lo: Seven, Six, Four, Three, Two-low [7♣ 6♦ 4♦ 3♦ 2♦] [K♣ 8♦ 6♣ Q♣]
	// Player 6: [T♣ Q♦ A♠ 7♥]
	//   Hi: Two Pair, Queens over Sevens, kicker Six [Q♣ Q♦ 7♣ 7♥ 6♦] [T♣ A♠ 2♦ 6♣]
	//   Lo: None [] []
	// Result (Hi): Player 3 wins with Full House, Sixes full of Queens
	// Result (Lo): Player 5 wins with Seven, Six, Four, Three, Two-low
	// ------ OmahaHiLo 5 ------
	// Board: [4♣ K♣ 6♦ 9♦ 5♠]
	// Player 1: [3♦ 4♦ 5♦ J♣]
	//   Hi: Two Pair, Fives over Fours, kicker King [5♦ 5♠ 4♣ 4♦ K♣] [3♦ J♣ 6♦ 9♦]
	//   Lo: None [] []
	// Player 2: [T♥ J♠ K♠ 2♣]
	//   Hi: Pair, Kings, kickers Jack, Nine, Six [K♣ K♠ J♠ 9♦ 6♦] [T♥ 2♣ 4♣ 5♠]
	//   Lo: None [] []
	// Player 3: [A♣ 9♠ T♠ 3♠]
	//   Hi: Pair, Nines, kickers Ace, King, Six [9♦ 9♠ A♣ K♣ 6♦] [T♠ 3♠ 4♣ 5♠]
	//   Lo: Six, Five, Four, Three, Ace-low [6♦ 5♠ 4♣ 3♠ A♣] [9♠ T♠ K♣ 9♦]
	// Player 4: [7♦ 3♣ 8♠ 7♣]
	//   Hi: Straight, Nine-high [9♦ 8♠ 7♦ 6♦ 5♠] [3♣ 7♣ 4♣ K♣]
	//   Lo: Seven, Six, Five, Four, Three-low [7♦ 6♦ 5♠ 4♣ 3♣] [8♠ 7♣ K♣ 9♦]
	// Player 5: [5♣ Q♠ J♥ 2♠]
	//   Hi: Pair, Fives, kickers King, Queen, Nine [5♣ 5♠ K♣ Q♠ 9♦] [J♥ 2♠ 4♣ 6♦]
	//   Lo: None [] []
	// Player 6: [6♠ 7♠ 7♥ 2♥]
	//   Hi: Pair, Sevens, kickers King, Nine, Six [7♥ 7♠ K♣ 9♦ 6♦] [6♠ 2♥ 4♣ 5♠]
	//   Lo: Seven, Six, Five, Four, Two-low [7♠ 6♦ 5♠ 4♣ 2♥] [6♠ 7♥ K♣ 9♦]
	// Result (Hi): Player 4 wins with Straight, Nine-high
	// Result (Lo): Player 3 wins with Six, Five, Four, Three, Ace-low
}

func Example_stud() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Stud.Deal(r, 1, game.players)
		evs := cardrank.Stud.Eval(pockets, nil)
		fmt.Printf("------ Stud %d ------\n", i+1)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Stud 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦] Pair, Queens, kickers Ace, King, Jack [Q♦ Q♠ A♥ K♥ J♣] [6♣ 5♥]
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥] Two Pair, Sevens over Fours, kicker Five [7♣ 7♥ 4♣ 4♥ 5♠] [3♥ 2♠]
	// Result:   Player 2 wins with Two Pair, Sevens over Fours, kicker Five
	// ------ Stud 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥] Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♠ 9♠]
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥] Pair, Sixes, kickers Queen, Ten, Eight [6♣ 6♦ Q♣ T♣ 8♥] [7♥ 3♥]
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣] Nothing, Ace-high, kickers King, Queen, Nine, Eight [A♥ K♠ Q♦ 9♣ 8♣] [7♣ 2♣]
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦] Nothing, King-high, kickers Jack, Ten, Eight, Four [K♦ J♠ T♦ 8♦ 4♥] [3♣ 2♦]
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦] Nothing, Ace-high, kickers Queen, Jack, Seven, Six [A♦ Q♥ J♦ 7♠ 6♠] [5♦ 2♥]
	// Result:   Player 1 wins with Two Pair, Tens over Threes, kicker King
	// ------ Stud 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥] Two Pair, Jacks over Sevens, kicker King [J♦ J♥ 7♣ 7♥ K♠] [Q♠ 4♣]
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦] Pair, Jacks, kickers Queen, Eight, Seven [J♣ J♠ Q♣ 8♥ 7♦] [3♣ 2♠]
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦] Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♦ 2♣]
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥] Two Pair, Fives over Twos, kicker Ten [5♣ 5♥ 2♦ 2♥ T♦] [9♦ 4♥]
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦] Two Pair, Aces over Sixes, kicker Eight [A♣ A♦ 6♦ 6♠ 8♠] [7♠ 3♥]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣] Pair, Kings, kickers Ten, Nine, Eight [K♣ K♦ T♣ 9♣ 8♦] [5♠ 4♠]
	// Result:   Player 5 wins with Two Pair, Aces over Sixes, kicker Eight
	// ------ Stud 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥] Pair, Aces, kickers King, Eight, Six [A♣ A♥ K♥ 8♣ 6♠] [5♦ 2♠]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠] Nothing, Queen-high, kickers Jack, Seven, Five, Four [Q♥ J♣ 7♣ 5♥ 4♥] [3♠ 2♦]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣] Pair, Sixes, kickers Queen, Nine, Five [6♦ 6♥ Q♠ 9♥ 5♣] [3♣ 2♣]
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠] Two Pair, Kings over Jacks, kicker Ten [K♦ K♠ J♥ J♠ T♠] [9♠ 6♣]
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥] Pair, Eights, kickers King, Ten, Nine [8♦ 8♥ K♣ T♥ 9♣] [4♦ 3♦]
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥] Two Pair, Queens over Sevens, kicker Ace [Q♣ Q♦ 7♦ 7♥ A♠] [T♣ 2♥]
	// Result:   Player 4 wins with Two Pair, Kings over Jacks, kicker Ten
	// ------ Stud 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣] Pair, Fours, kickers King, Jack, Eight [4♦ 4♥ K♥ J♣ 8♣] [5♦ 3♦]
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦] Pair, Twos, kickers King, Jack, Ten [2♣ 2♦ K♠ J♠ T♥] [5♠ 4♣]
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥] Pair, Aces, kickers King, Ten, Nine [A♣ A♥ K♣ T♠ 9♠] [8♦ 3♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣] Full House, Sixes full of Sevens [6♣ 6♦ 6♥ 7♣ 7♦] [8♠ 3♣]
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠] Nothing, Ace-high, kickers Queen, Jack, Eight, Five [A♠ Q♠ J♥ 8♥ 5♣] [4♠ 2♠]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦] Pair, Sevens, kickers King, Ten, Nine [7♥ 7♠ K♦ T♦ 9♦] [6♠ 2♥]
	// Result:   Player 4 wins with Full House, Sixes full of Sevens
}

func Example_studHiLo() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.StudHiLo.Deal(r, 1, game.players)
		evs := cardrank.StudHiLo.Eval(pockets, nil)
		fmt.Printf("------ StudHiLo %d ------\n", i+1)
		for j := 0; j < game.players; j++ {
			hi, lo := evs[j].Desc(false), evs[j].Desc(true)
			fmt.Printf("Player %d: %b\n", j+1, pockets[j])
			fmt.Printf("  Hi: %s %b %b\n", hi, hi.Best, hi.Unused)
			fmt.Printf("  Lo: %s %b %b\n", lo, lo.Best, lo.Unused)
		}
		hiOrder, hiPivot := cardrank.Order(evs, false)
		loOrder, loPivot := cardrank.Order(evs, true)
		typ := "wins"
		if loPivot == 0 {
			typ = "scoops"
		}
		desc := evs[hiOrder[0]].Desc(false)
		if hiPivot == 1 {
			fmt.Printf("Result (Hi): Player %d %s with %s\n", hiOrder[0]+1, typ, desc)
		} else {
			var s []string
			for j := 0; j < hiPivot; j++ {
				s = append(s, strconv.Itoa(hiOrder[j]+1))
			}
			fmt.Printf("Result (Hi): Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
		if loPivot == 1 {
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Player %d wins with %s\n", loOrder[0]+1, desc)
		} else if loPivot > 1 {
			var s []string
			for j := 0; j < loPivot; j++ {
				s = append(s, strconv.Itoa(loOrder[j]+1))
			}
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Players %s push with %s\n", strings.Join(s, ", "), desc)
		} else {
			fmt.Printf("Result (Lo): no player made a low hand\n")
		}
	}
	// Output:
	// ------ StudHiLo 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦]
	//   Hi: Pair, Queens, kickers Ace, King, Jack [Q♦ Q♠ A♥ K♥ J♣] [6♣ 5♥]
	//   Lo: None [] []
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥]
	//   Hi: Two Pair, Sevens over Fours, kicker Five [7♣ 7♥ 4♣ 4♥ 5♠] [3♥ 2♠]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♠ 4♣ 3♥ 2♠] [4♥ 7♥]
	// Result (Hi): Player 2 wins with Two Pair, Sevens over Fours, kicker Five
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ StudHiLo 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥]
	//   Hi: Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♠ 9♠]
	//   Lo: None [] []
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥]
	//   Hi: Pair, Sixes, kickers Queen, Ten, Eight [6♣ 6♦ Q♣ T♣ 8♥] [7♥ 3♥]
	//   Lo: None [] []
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣]
	//   Hi: Nothing, Ace-high, kickers King, Queen, Nine, Eight [A♥ K♠ Q♦ 9♣ 8♣] [7♣ 2♣]
	//   Lo: None [] []
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦]
	//   Hi: Nothing, King-high, kickers Jack, Ten, Eight, Four [K♦ J♠ T♦ 8♦ 4♥] [3♣ 2♦]
	//   Lo: None [] []
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦]
	//   Hi: Nothing, Ace-high, kickers Queen, Jack, Seven, Six [A♦ Q♥ J♦ 7♠ 6♠] [5♦ 2♥]
	//   Lo: Seven, Six, Five, Two, Ace-low [7♠ 6♠ 5♦ 2♥ A♦] [J♦ Q♥]
	// Result (Hi): Player 1 wins with Two Pair, Tens over Threes, kicker King
	// Result (Lo): Player 5 wins with Seven, Six, Five, Two, Ace-low
	// ------ StudHiLo 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥]
	//   Hi: Two Pair, Jacks over Sevens, kicker King [J♦ J♥ 7♣ 7♥ K♠] [Q♠ 4♣]
	//   Lo: None [] []
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦]
	//   Hi: Pair, Jacks, kickers Queen, Eight, Seven [J♣ J♠ Q♣ 8♥ 7♦] [3♣ 2♠]
	//   Lo: None [] []
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦]
	//   Hi: Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♦ 2♣]
	//   Lo: None [] []
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥]
	//   Hi: Two Pair, Fives over Twos, kicker Ten [5♣ 5♥ 2♦ 2♥ T♦] [9♦ 4♥]
	//   Lo: None [] []
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦]
	//   Hi: Two Pair, Aces over Sixes, kicker Eight [A♣ A♦ 6♦ 6♠ 8♠] [7♠ 3♥]
	//   Lo: Eight, Seven, Six, Three, Ace-low [8♠ 7♠ 6♠ 3♥ A♣] [6♦ A♦]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣]
	//   Hi: Pair, Kings, kickers Ten, Nine, Eight [K♣ K♦ T♣ 9♣ 8♦] [5♠ 4♠]
	//   Lo: None [] []
	// Result (Hi): Player 5 wins with Two Pair, Aces over Sixes, kicker Eight
	// Result (Lo): Player 5 wins with Eight, Seven, Six, Three, Ace-low
	// ------ StudHiLo 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥]
	//   Hi: Pair, Aces, kickers King, Eight, Six [A♣ A♥ K♥ 8♣ 6♠] [5♦ 2♠]
	//   Lo: Eight, Six, Five, Two, Ace-low [8♣ 6♠ 5♦ 2♠ A♣] [K♥ A♥]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠]
	//   Hi: Nothing, Queen-high, kickers Jack, Seven, Five, Four [Q♥ J♣ 7♣ 5♥ 4♥] [3♠ 2♦]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♠ 2♦] [Q♥ J♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣]
	//   Hi: Pair, Sixes, kickers Queen, Nine, Five [6♦ 6♥ Q♠ 9♥ 5♣] [3♣ 2♣]
	//   Lo: None [] []
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠]
	//   Hi: Two Pair, Kings over Jacks, kicker Ten [K♦ K♠ J♥ J♠ T♠] [9♠ 6♣]
	//   Lo: None [] []
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥]
	//   Hi: Pair, Eights, kickers King, Ten, Nine [8♦ 8♥ K♣ T♥ 9♣] [4♦ 3♦]
	//   Lo: None [] []
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥]
	//   Hi: Two Pair, Queens over Sevens, kicker Ace [Q♣ Q♦ 7♦ 7♥ A♠] [T♣ 2♥]
	//   Lo: None [] []
	// Result (Hi): Player 4 wins with Two Pair, Kings over Jacks, kicker Ten
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ StudHiLo 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣]
	//   Hi: Pair, Fours, kickers King, Jack, Eight [4♦ 4♥ K♥ J♣ 8♣] [5♦ 3♦]
	//   Lo: None [] []
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦]
	//   Hi: Pair, Twos, kickers King, Jack, Ten [2♣ 2♦ K♠ J♠ T♥] [5♠ 4♣]
	//   Lo: None [] []
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥]
	//   Hi: Pair, Aces, kickers King, Ten, Nine [A♣ A♥ K♣ T♠ 9♠] [8♦ 3♠]
	//   Lo: None [] []
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣]
	//   Hi: Full House, Sixes full of Sevens [6♣ 6♦ 6♥ 7♣ 7♦] [8♠ 3♣]
	//   Lo: None [] []
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠]
	//   Hi: Nothing, Ace-high, kickers Queen, Jack, Eight, Five [A♠ Q♠ J♥ 8♥ 5♣] [4♠ 2♠]
	//   Lo: Eight, Five, Four, Two, Ace-low [8♥ 5♣ 4♠ 2♠ A♠] [Q♠ J♥]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦]
	//   Hi: Pair, Sevens, kickers King, Ten, Nine [7♥ 7♠ K♦ T♦ 9♦] [6♠ 2♥]
	//   Lo: None [] []
	// Result (Hi): Player 4 wins with Full House, Sixes full of Sevens
	// Result (Lo): Player 5 wins with Eight, Five, Four, Two, Ace-low
}

func Example_razz() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Razz.Deal(r, 1, game.players)
		evs := cardrank.Razz.Eval(pockets, nil)
		fmt.Printf("------ Razz %d ------\n", i+1)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Razz 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦] Queen, Jack, Six, Five, Ace-low [Q♠ J♣ 6♣ 5♥ A♥] [K♥ Q♦]
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥] Seven, Five, Four, Three, Two-low [7♣ 5♠ 4♣ 3♥ 2♠] [4♥ 7♥]
	// Result:   Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ Razz 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥] King, Queen, Ten, Nine, Three-low [K♥ Q♠ T♠ 9♠ 3♠] [3♦ T♥]
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥] Ten, Eight, Seven, Six, Three-low [T♣ 8♥ 7♥ 6♦ 3♥] [Q♣ 6♣]
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣] Nine, Eight, Seven, Two, Ace-low [9♣ 8♣ 7♣ 2♣ A♥] [Q♦ K♠]
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦] Ten, Eight, Four, Three, Two-low [T♦ 8♦ 4♥ 3♣ 2♦] [K♦ J♠]
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦] Seven, Six, Five, Two, Ace-low [7♠ 6♠ 5♦ 2♥ A♦] [J♦ Q♥]
	// Result:   Player 5 wins with Seven, Six, Five, Two, Ace-low
	// ------ Razz 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥] King, Queen, Jack, Seven, Four-low [K♠ Q♠ J♦ 7♥ 4♣] [7♣ J♥]
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦] Jack, Eight, Seven, Three, Two-low [J♠ 8♥ 7♦ 3♣ 2♠] [J♣ Q♣]
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦] King, Queen, Ten, Three, Two-low [K♥ Q♦ T♠ 3♠ 2♣] [T♥ 3♦]
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥] Ten, Nine, Five, Four, Two-low [T♦ 9♦ 5♣ 4♥ 2♦] [5♥ 2♥]
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦] Eight, Seven, Six, Three, Ace-low [8♠ 7♠ 6♠ 3♥ A♣] [6♦ A♦]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣] Ten, Nine, Eight, Five, Four-low [T♣ 9♣ 8♦ 5♠ 4♠] [K♦ K♣]
	// Result:   Player 5 wins with Eight, Seven, Six, Three, Ace-low
	// ------ Razz 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥] Eight, Six, Five, Two, Ace-low [8♣ 6♠ 5♦ 2♠ A♣] [K♥ A♥]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠] Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♠ 2♦] [Q♥ J♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣] Nine, Six, Five, Three, Two-low [9♥ 6♥ 5♣ 3♣ 2♣] [Q♠ 6♦]
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠] King, Jack, Ten, Nine, Six-low [K♠ J♥ T♠ 9♠ 6♣] [J♠ K♦]
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥] Ten, Nine, Eight, Four, Three-low [T♥ 9♣ 8♦ 4♦ 3♦] [K♣ 8♥]
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥] Queen, Ten, Seven, Two, Ace-low [Q♦ T♣ 7♥ 2♥ A♠] [Q♣ 7♦]
	// Result:   Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ Razz 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣] Jack, Eight, Five, Four, Three-low [J♣ 8♣ 5♦ 4♦ 3♦] [4♥ K♥]
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦] Jack, Ten, Five, Four, Two-low [J♠ T♥ 5♠ 4♣ 2♣] [K♠ 2♦]
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥] Ten, Nine, Eight, Three, Ace-low [T♠ 9♠ 8♦ 3♠ A♣] [K♣ A♥]
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣] Pair, Sixes, kickers Eight, Seven, Three [6♦ 6♥ 8♠ 7♦ 3♣] [7♣ 6♣]
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠] Eight, Five, Four, Two, Ace-low [8♥ 5♣ 4♠ 2♠ A♠] [Q♠ J♥]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦] Ten, Nine, Seven, Six, Two-low [T♦ 9♦ 7♠ 6♠ 2♥] [7♥ K♦]
	// Result:   Player 5 wins with Eight, Five, Four, Two, Ace-low
}

func Example_badugi() {
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
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Badugi.Deal(r, 1, game.players)
		evs := cardrank.Badugi.Eval(pockets, nil)
		fmt.Printf("------ Badugi %d ------\n", i+1)
		for j := 0; j < game.players; j++ {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := 0; j < pivot; j++ {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Badugi 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠] Queen, Jack, Ace-low [Q♠ J♣ A♥] [K♥]
	// Player 2: [7♣ 4♣ 5♠ 2♠] Four, Two-low [4♣ 2♠] [7♣ 5♠]
	// Result:   Player 1 wins with Queen, Jack, Ace-low
	// ------ Badugi 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠] Ten, Three-low [T♠ 3♦] [Q♠ 3♠]
	// Player 2: [6♦ Q♣ 8♥ 6♣] Queen, Eight, Six-low [Q♣ 8♥ 6♦] [6♣]
	// Player 3: [Q♦ K♠ 8♣ A♥] King, Queen, Eight, Ace-low [K♠ Q♦ 8♣ A♥] []
	// Player 4: [K♦ T♦ 8♦ 4♥] Eight, Four-low [8♦ 4♥] [K♦ T♦]
	// Player 5: [J♦ 2♥ Q♥ 6♠] Jack, Six, Two-low [J♦ 6♠ 2♥] [Q♥]
	// Result:   Player 3 wins with King, Queen, Eight, Ace-low
	// ------ Badugi 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦] Queen, Jack, Four-low [Q♠ J♦ 4♣] [K♠]
	// Player 2: [J♠ 3♣ 8♥ 2♠] Eight, Three, Two-low [8♥ 3♣ 2♠] [J♠]
	// Player 3: [3♠ T♠ 2♣ Q♦] Queen, Three, Two-low [Q♦ 3♠ 2♣] [T♠]
	// Player 4: [5♣ 5♥ T♦ 2♦] Five, Two-low [5♥ 2♦] [T♦ 5♣]
	// Player 5: [7♠ 3♥ 6♠ A♣] Six, Three, Ace-low [6♠ 3♥ A♣] [7♠]
	// Player 6: [4♠ 8♦ K♦ T♣] Ten, Eight, Four-low [T♣ 8♦ 4♠] [K♦]
	// Result:   Player 5 wins with Six, Three, Ace-low
	// ------ Badugi 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣] King, Six, Ace-low [K♥ 6♠ A♣] [8♣]
	// Player 2: [Q♥ 4♥ J♣ 5♥] Jack, Four-low [J♣ 4♥] [Q♥ 5♥]
	// Player 3: [2♣ 6♥ 5♣ Q♠] Queen, Six, Two-low [Q♠ 6♥ 2♣] [5♣]
	// Player 4: [9♠ J♥ K♠ J♠] Jack, Nine-low [J♥ 9♠] [K♠ J♠]
	// Player 5: [3♦ 4♦ K♣ 8♦] King, Three-low [K♣ 3♦] [8♦ 4♦]
	// Player 6: [T♣ Q♦ A♠ 7♥] Queen, Ten, Seven, Ace-low [Q♦ T♣ 7♥ A♠] []
	// Result:   Player 6 wins with Queen, Ten, Seven, Ace-low
	// ------ Badugi 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣] Jack, Three-low [J♣ 3♦] [5♦ 4♦]
	// Player 2: [T♥ J♠ K♠ 2♣] Jack, Ten, Two-low [J♠ T♥ 2♣] [K♠]
	// Player 3: [A♣ 9♠ T♠ 3♠] Three, Ace-low [3♠ A♣] [T♠ 9♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣] Eight, Seven, Three-low [8♠ 7♦ 3♣] [7♣]
	// Player 5: [5♣ Q♠ J♥ 2♠] Jack, Five, Two-low [J♥ 5♣ 2♠] [Q♠]
	// Player 6: [6♠ 7♠ 7♥ 2♥] Six, Two-low [6♠ 2♥] [7♠ 7♥]
	// Result:   Player 4 wins with Eight, Seven, Three-low
}

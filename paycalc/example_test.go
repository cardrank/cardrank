package paycalc_test

import (
	"fmt"

	"github.com/cardrank/cardrank/paycalc"
)

func Example() {
	const typ, entries, buyin, guaranteed, rake = paycalc.Top10, 110, 1000, 60000, 0.15
	levels, amounts, payouts := typ.Stakes(entries, buyin, guaranteed, rake)
	for i := 0; i < len(levels); i++ {
		title := paycalc.LevelTitle(levels[i][0], levels[i][1])
		fmt.Printf("%s (%0.2f%%): %d\n", title, amounts[i]*100, payouts[i])
	}
	// Output:
	// 1st (28.00%): 26180
	// 2nd (17.00%): 15895
	// 3rd (10.60%): 9911
	// 4th (8.60%): 8041
	// 5th (7.60%): 7106
	// 6th (5.30%): 4955
	// 7th (4.30%): 4020
	// 8th (3.30%): 3085
	// 9th (2.70%): 2524
	// 10th (2.10%): 1963
	// 11-15 (2.10%): 1963
}

func Example_payout() {
	const typ, entries, buyin, guaranteed, rake = paycalc.Top10, 110, 1000, 60000, 0.15
	for i := 0; i < 15; i++ {
		payout := typ.Payout(i, entries, buyin, guaranteed, rake)
		fmt.Printf("%d: %d\n", i+1, payout)
	}
	// Output:
	// 1: 26180
	// 2: 15895
	// 3: 9911
	// 4: 8041
	// 5: 7106
	// 6: 4955
	// 7: 4020
	// 8: 3085
	// 9: 2524
	// 10: 1963
	// 11: 1963
	// 12: 1963
	// 13: 1963
	// 14: 1963
	// 15: 1963
}

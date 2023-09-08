package paycalc_test

import (
	"fmt"

	"github.com/cardrank/cardrank/paycalc"
)

func Example() {
	const entries, buyin, guaranteed, rake = 88, 200000, 10000000, 0.15
	for _, typ := range []paycalc.Type{paycalc.Top10, paycalc.Top15, paycalc.Top20} {
		paid, row, col := typ.Paid(entries)
		unallocated := typ.Unallocated(paid, row, col)
		fmt.Printf("%t (%d of %d):\n", typ, paid, entries)
		fmt.Printf("  l/e:   %s/%s\n", typ.MaxLevelTitle(paid), typ.EntriesTitle(entries))
		fmt.Printf("  short: %.2f%%\n", unallocated*100.0)
		fmt.Printf("  buyin: %d\n", buyin)
		fmt.Printf("  rake:  %.0f%%\n", rake*100.0)
		fmt.Printf("  gtd:   %d\n", guaranteed)
		fmt.Printf("  prize: %d\n", paycalc.Prize(entries, buyin, guaranteed, rake))
		levels, amounts, payouts, total := typ.Stakes(entries, buyin, guaranteed, rake)
		fmt.Printf("  total: %d\n", total)
		var sum int64
		var gsum, fsum float64
		for i := 0; i < len(levels); i++ {
			title := paycalc.LevelTitle(levels[i][0], levels[i][1])
			g := amounts[i] * float64(levels[i][1]-levels[i][0])
			f := float64(payouts[i]) / float64(total) * float64(levels[i][1]-levels[i][0])
			fmt.Printf("  %5s: %8d (%6.2f%% -> %6.2f%%)\n", title, payouts[i], g*100.0, f*100.0)
			sum, gsum, fsum = sum+payouts[i]*int64(levels[i][1]-levels[i][0]), gsum+g, fsum+f
		}
		fmt.Printf("  Total: %8d (%6.2f%%    %6.2f%%)\n\n", sum, gsum*100.0, fsum*100.0)
	}
	// Output:
	// Top 10% (9 of 88):
	//   l/e:   9th/76-100
	//   short: 2.50%
	//   buyin: 200000
	//   rake:  15%
	//   gtd:   10000000
	//   prize: 14960000
	//   total: 14960004
	//     1st:  4603077 ( 30.00% ->  30.77%)
	//     2nd:  3068718 ( 20.00% ->  20.51%)
	//     3rd:  1841231 ( 12.00% ->  12.31%)
	//     4th:  1457642 (  9.50% ->   9.74%)
	//     5th:  1227488 (  8.00% ->   8.21%)
	//     6th:   920616 (  6.00% ->   6.15%)
	//     7th:   767180 (  5.00% ->   5.13%)
	//     8th:   613744 (  4.00% ->   4.10%)
	//     9th:   460308 (  3.00% ->   3.08%)
	//   Total: 14960004 ( 97.50%    100.00%)
	//
	// Top 15% (14 of 88):
	//   l/e:   11-14/76-100
	//   short: 1.86%
	//   buyin: 200000
	//   rake:  15%
	//   gtd:   10000000
	//   prize: 14960000
	//   total: 14960007
	//     1st:  4344406 ( 28.50% ->  29.04%)
	//     2nd:  2515183 ( 16.50% ->  16.81%)
	//     3rd:  1554841 ( 10.20% ->  10.39%)
	//     4th:  1295701 (  8.50% ->   8.66%)
	//     5th:  1143265 (  7.50% ->   7.64%)
	//     6th:   914612 (  6.00% ->   6.11%)
	//     7th:   762177 (  5.00% ->   5.09%)
	//     8th:   533524 (  3.50% ->   3.57%)
	//     9th:   381089 (  2.50% ->   2.55%)
	//    10th:   381089 (  2.50% ->   2.55%)
	//   11-14:   283530 (  7.44% ->   7.58%)
	//   Total: 14960007 ( 98.14%    100.00%)
	//
	// Top 20% (18 of 88):
	//   l/e:   16-18/76-100
	//   short: 2.60%
	//   buyin: 200000
	//   rake:  15%
	//   gtd:   10000000
	//   prize: 14960000
	//   total: 14960010
	//     1st:  4300617 ( 28.00% ->  28.75%)
	//     2nd:  2611089 ( 17.00% ->  17.45%)
	//     3rd:  1290185 (  8.40% ->   8.62%)
	//     4th:  1151951 (  7.50% ->   7.70%)
	//     5th:   998358 (  6.50% ->   6.67%)
	//     6th:   844764 (  5.50% ->   5.65%)
	//     7th:   660452 (  4.30% ->   4.41%)
	//     8th:   445421 (  2.90% ->   2.98%)
	//     9th:   399343 (  2.60% ->   2.67%)
	//    10th:   276469 (  1.80% ->   1.85%)
	//   11-15:   276469 (  9.00% ->   9.24%)
	//   16-18:   199672 (  3.90% ->   4.00%)
	//   Total: 14960010 ( 97.40%    100.00%)
}

func Example_payouts() {
	const entries, buyin, guaranteed, rake = 110, 1000, 60000, 0.15
	for _, typ := range []paycalc.Type{paycalc.Top10, paycalc.Top15, paycalc.Top20} {
		paid, _, _ := typ.Paid(entries)
		fmt.Printf("%t (%d of %d):\n", typ, paid, entries)
		payouts, total := typ.Payouts(entries, buyin, guaranteed, rake)
		for i := 0; i < len(payouts); i++ {
			fmt.Printf("   %2d: %5d\n", i+1, payouts[i])
		}
		fmt.Printf("Total: %5d\n\n", total)
	}
	// Output:
	// Top 10% (11 of 110):
	//     1: 28581
	//     2: 17353
	//     3: 10820
	//     4:  8779
	//     5:  7758
	//     6:  5410
	//     7:  4390
	//     8:  3369
	//     9:  2757
	//    10:  2144
	//    11:  2144
	// Total: 93505
	//
	// Top 15% (17 of 110):
	//     1: 27500
	//     2: 15911
	//     3:  9134
	//     4:  7170
	//     5:  6188
	//     6:  5206
	//     7:  4125
	//     8:  2750
	//     9:  2063
	//    10:  1719
	//    11:  1719
	//    12:  1719
	//    13:  1719
	//    14:  1719
	//    15:  1719
	//    16:  1572
	//    17:  1572
	// Total: 93505
	//
	// Top 20% (22 of 110):
	//     1: 26067
	//     2: 15930
	//     3:  7917
	//     4:  6951
	//     5:  5986
	//     6:  5021
	//     7:  3862
	//     8:  2607
	//     9:  2317
	//    10:  1545
	//    11:  1545
	//    12:  1545
	//    13:  1545
	//    14:  1545
	//    15:  1545
	//    16:  1111
	//    17:  1111
	//    18:  1111
	//    19:  1111
	//    20:  1111
	//    21:  1014
	//    22:  1014
	// Total: 93511
}

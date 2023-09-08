// Package paycalc contains a tournament payout tables.
package paycalc

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// tables are tournament payout tables.
var tables = make(map[Type]*Table)

// Register registers a tournament payout table.
func Register(typ Type, t *Table) error {
	if _, ok := tables[typ]; ok {
		return fmt.Errorf("type %d already registered", typ)
	}
	tables[typ] = t
	return nil
}

// RegisterReader registers a tournament payout table read as CSV from the
// reader.
func RegisterReader(typ Type, r io.Reader, top float64, name string) error {
	t, err := LoadReader(r, top, name)
	if err != nil {
		return err
	}
	return Register(typ, t)
}

// RegisterBytes registers a tournament payout table read as CSV from buf.
func RegisterBytes(typ Type, buf []byte, top float64, name string) error {
	return RegisterReader(typ, bytes.NewReader(buf), top, name)
}

// Init initializes the tournament payout tables.
func Init() error {
	if err := RegisterBytes(Top10, top10, 0.10, "top10"); err != nil {
		return err
	}
	if err := RegisterBytes(Top15, top15, 0.15, "top15"); err != nil {
		return err
	}
	if err := RegisterBytes(Top20, top20, 0.20, "top20"); err != nil {
		return err
	}
	return nil
}

// Type is a tournament payout table type.
type Type int

// Top tournament payout table types.
const (
	Top00 Type = iota
	Top10
	Top15
	Top20
)

// Top returns the proportion of paid rankings for the tournament payout table.
func (typ Type) Top() float64 {
	if t, ok := tables[typ]; ok {
		return t.Top()
	}
	return 0.0
}

// Name returns the name of the tournament payout table.
func (typ Type) Name() string {
	if t, ok := tables[typ]; ok {
		return t.Name()
	}
	return ""
}

// Title returns the title of the tournament payout table.
func (typ Type) Title() string {
	if t, ok := tables[typ]; ok {
		return t.Title()
	}
	return ""
}

// Format satisfies the [fmt.Formatter] interface.
func (typ Type) Format(f fmt.State, verb rune) {
	if t, ok := tables[typ]; ok {
		t.Format(f, verb)
	}
}

// WriteTable writes a plain text version of the tournament payout table, along
// with optional title, to w.
func (typ Type) WriteTable(w io.Writer, title bool) error {
	if t, ok := tables[typ]; ok {
		return t.WriteTable(w, title)
	}
	return nil
}

// WriteCSV writes a CSV version of the tournament payout table to w.
func (typ Type) WriteCSV(w io.Writer) error {
	if t, ok := tables[typ]; ok {
		return t.WriteCSV(w)
	}
	return nil
}

// WriteMarkdown writes a Markdown formatted version of the tournament payout
// table to w.
func (typ Type) WriteMarkdown(w io.Writer, header int) error {
	if t, ok := tables[typ]; ok {
		return t.WriteMarkdown(w, header)
	}
	return nil
}

// WriteTo writes the tournament payout table to w, formatting cells with f,
// breaking line output when early is true.
func (typ Type) WriteTo(w io.Writer, f func(interface{}, int, bool) string, divider func(int, int) string, early bool) error {
	if t, ok := tables[typ]; ok {
		return t.WriteTo(w, f, divider, early)
	}
	return nil
}

// EntriesMax returns the max entries for the tournament payout table.
func (typ Type) EntriesMax() int {
	if t, ok := tables[typ]; ok {
		return t.EntriesMax()
	}
	return 0
}

// Entries returns the entries column in the tournament payout table.
func (typ Type) Entries(entries int) int {
	if t, ok := tables[typ]; ok {
		return t.Entries(entries)
	}
	return 0
}

// Level returns the level (row) in the tournament payout table.
func (typ Type) Level(level int) int {
	if t, ok := tables[typ]; ok {
		return t.Level(level)
	}
	return -1
}

// EntriesTitle returns the column title for the specified entries in the
// tournament payout table.
func (typ Type) EntriesTitle(entries int) string {
	if t, ok := tables[typ]; ok {
		return t.EntriesTitle(entries)
	}
	return ""
}

// LevelTitle returns the level (row) title for rank n in the tournament
// payout table.
func (typ Type) LevelTitle(level int) string {
	if t, ok := tables[typ]; ok {
		return t.LevelTitle(level)
	}
	return ""
}

// MaxLevelTitle returnss the row title for the max level in tournament payout
// table.
func (typ Type) MaxLevelTitle(level int) string {
	if t, ok := tables[typ]; ok {
		return t.MaxLevelTitle(level)
	}
	return ""
}

// Paid returns the paid rankings, level (row) and corresponding entries column
// for the tournament payout table.
func (typ Type) Paid(entries int) (int, int, int) {
	if t, ok := tables[typ]; ok {
		return t.Paid(entries)
	}
	return 0, 0, 0
}

// Unallocated returns the unallocated amount for the paid rankings, row, and
// col from the tournament payout table.
func (typ Type) Unallocated(paid, row, col int) float64 {
	if t, ok := tables[typ]; ok {
		return t.Unallocated(paid, row, col)
	}
	return 0.0
}

// Stakes returns the paid levels as [low, high), the tournament table value
// per level, the calculated payouts per level, and the total amount paid.
func (typ Type) Stakes(entries int, buyin, guaranteed int64, rake float64) ([][2]int, []float64, []int64, int64) {
	if t, ok := tables[typ]; ok {
		return t.Stakes(entries, buyin, guaranteed, rake)
	}
	return nil, nil, nil, 0
}

// Payouts returns the tournament paid rankings.
func (typ Type) Payouts(entries int, buyin, guaranteed int64, rake float64) ([]int64, int64) {
	if t, ok := tables[typ]; ok {
		return t.Payouts(entries, buyin, guaranteed, rake)
	}
	return nil, 0
}

// Amount returns the amount for the level and entries from the tournament
// payout table.
func (typ Type) Amount(level, entries int) float64 {
	if t, ok := tables[typ]; ok {
		return t.Amount(level, entries)
	}
	return 0.0
}

// At returns the amount at row, col of the tournament payout table.
func (typ Type) At(row, col int) float64 {
	if t, ok := tables[typ]; ok {
		return t.At(row, col)
	}
	return 0.0
}

// MarshalText satisfies the [encoding.TextMarshaler] interface.
func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.Name()), nil
}

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface.
func (typ *Type) UnmarshalText(buf []byte) error {
	name := strings.ToLower(string(buf))
	for t, tbl := range tables {
		if strings.ToLower(tbl.name) == name {
			*typ = t
			return nil
		}
	}
	return fmt.Errorf("invalid type %q", string(buf))
}

// EntriesTitle formats the entries title of last, n.
func EntriesTitle(last, entries int) string {
	switch {
	case last < 0 && entries < 0:
		return ""
	case last == 0:
		return strconv.Itoa(entries) + "+"
	case entries-last < 1:
		return strconv.Itoa(last)
	}
	return strconv.Itoa(last+1) + "-" + strconv.Itoa(entries)
}

// LevelTitle formats the levels title of last, n.
func LevelTitle(last, level int) string {
	if level-last < 2 {
		return strconv.Itoa(level) + ord(level)
	}
	return fmt.Sprintf("%d-%d", last+1, level)
}

// Equal returns true when a and b are within [Epsilon].
func Equal[S, T Ordered](a S, b T) bool {
	return math.Abs(float64(a)-float64(b)) <= Epsilon
}

// Epsilon is the epsilon value used for for [EpsilonEqual].
var Epsilon = 0.0000000001

// EqualEpsilon returns true when a and b are within epsilon.
func EqualEpsilon[R, S, T Ordered](a R, b S, epsilon T) bool {
	return math.Abs(float64(a)-float64(b)) <= float64(epsilon)
}

// Prize calculates the total payout based on entries, buyin, guaranteed
// amount, and rake. Uses [Round] to round to [Precision].
func Prize(entries int, buyin, guaranteed int64, rake float64) int64 {
	amt := int64(entries) * buyin
	if rake < 0.0 || 1.0 < rake {
		return amt
	}
	return max(guaranteed, amt-int64(Round(rake*float64(amt))))
}

// Calc calculates the amount scaled by unallocated. Uses [Round] to round to
// [Precision].
func Calc(f float64, total int64, unallocated float64) int64 {
	if unallocated < 0.0 || 1.0 < unallocated {
		unallocated = 0.0
	}
	return int64(Round(f * float64(total) / (1.0 - unallocated)))
}

// Round is the round implementation used by [Calc]. Rounds to [Precision].
var Round = func(f float64) float64 {
	return math.Ceil(math.Round(f*Precision) / Precision)
}

// Precision is the precision amount used by [Calc].
var Precision float64 = 10000

// Paid returns the number of paid rankings, uses [PaidRound].
var Paid = func(f float64, entries int) int {
	return int(PaidRound(f * float64(entries)))
}

// PaidRound by default is [math.Ceil] and is used by [PaidMax] to determine
// the number of paid rankings.
var PaidRound = math.Ceil

// ord returns the ordinal suffix for n.
func ord(n int) string {
	if 11 <= n && n <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}

// tournament payout tables.
var (
	//go:embed top10.csv
	top10 []byte
	//go:embed top15.csv
	top15 []byte
	//go:embed top20.csv
	top20 []byte
)

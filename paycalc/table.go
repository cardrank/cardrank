package paycalc

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Table is a tournament payout table.
type Table struct {
	name    string
	top     float64
	levels  []int
	entries []int
	amounts [][]float64
}

// New creates a new tournament payout table.
func New(name string, top float64, levels, entries []int, amounts [][]float64) (*Table, error) {
	rows, cols := len(levels), len(entries)
	if rows < 1 || cols < 1 {
		return nil, errors.New("invalid table")
	}
	// check levels
	for i, last := 0, 0; i < rows; i++ {
		switch n := len(amounts[i]); {
		case cols < n:
			return nil, fmt.Errorf("level %d has %d columns: more than %d", i, len(amounts[i]), cols)
		case levels[i] <= last:
			return nil, fmt.Errorf("levels are not ordered low to high: row %d value %d must be <= %d ", i, levels[i], last)
		case n < cols:
			amounts[i] = append(amounts[i], make([]float64, cols-n)...)
		}
		last = levels[i]
	}
	// check entries
	for i, last := 0, 9999; i < cols; i++ {
		if last <= entries[i] {
			return nil, fmt.Errorf("entries are not ordered high to low: col %d value %d must be <= %d", i, entries[i], last)
		}
		last = entries[i]
	}
	// calc sums
	sums := make([]float64, cols)
	for i := 0; i < rows; i++ {
		n := 1
		if 0 < i {
			n = levels[i] - levels[i-1]
		}
		for j := 0; j < cols; j++ {
			sums[j] += float64(n) * amounts[i][j]
		}
	}
	// check entries
	for i := 0; i < cols; i++ {
		// check column sum
		if !Equal(1.0, sums[i]) {
			return nil, fmt.Errorf("entries %s sum %f != 1.0", EntriesTitle(entries[i+1], entries[i]), sums[i])
		}
		// check levels exist for entries
		if i != 0 {
			paid := Paid(top, entries[i])
			if l := findLevel(paid-1, levels); l == -1 || amounts[l][i] == 0.0 {
				last := 0
				if 0 < l {
					last = levels[l-1]
				}
				return nil, fmt.Errorf(
					"top %0.2f%% needs %d levels for entries %s: level %s not defined",
					top*100.0, paid,
					EntriesTitle(entries[i+1], entries[i]),
					LevelTitle(last, levels[l]),
				)
			}
		}
	}
	return &Table{
		name:    name,
		top:     top,
		levels:  levels,
		entries: entries,
		amounts: amounts,
	}, nil
}

// LoadReader loads a CSV formatted tournament payout table from the reader.
func LoadReader(name string, top float64, rdr io.Reader) (*Table, error) {
	r := csv.NewReader(rdr)
	r.FieldsPerRecord = -1
	lines, err := r.ReadAll()
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to parse table: %w", err)
	case len(lines) < 2 || len(lines[0]) < 2:
		return nil, errors.New("invalid table")
	case lines[0][0] != reCell:
		return nil, fmt.Errorf("row 0 col 0 must be %q ", reCell)
	}
	rows, cols := len(lines)-1, len(lines[0])-1
	levels, entries, amounts := make([]int, rows), make([]int, cols), make([][]float64, rows)
	// parse levels
	for i := 1; i <= rows; i++ {
		switch n := len(lines[i]) - 1; {
		case cols < n:
			return nil, fmt.Errorf("row %d has %d columns: more than %d", i, n, cols)
		case n < cols:
			lines[i] = append(lines[i], make([]string, cols-n)...)
		}
		if levels[i-1], err = parseLevels(lines[i][0]); err != nil {
			return nil, fmt.Errorf("unable to parse row %d header %q: %w", i, lines[i][0], err)
		}
		amounts[i-1] = make([]float64, cols)
	}
	// parse entries
	for i := 1; i <= cols; i++ {
		if entries[i-1], err = parseEntries(lines[0][i]); err != nil {
			return nil, fmt.Errorf("unable to parse col %d header %q: %w", i, lines[0][i], err)
		}
	}
	// parse amounts
	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			cell := strings.TrimSpace(lines[i][j])
			if cell == "" {
				continue
			}
			if amounts[i-1][j-1], err = strconv.ParseFloat(cell, 64); err != nil {
				return nil, fmt.Errorf("unable to parse row %d col %d cell %q: %w", i, j, lines[i][j], err)
			}
			amounts[i-1][j-1] /= 100.0
		}
	}
	return New(name, top, levels, entries, amounts)
}

// LoadBytes loads CSV formatted tournament payout table from buf.
func LoadBytes(name string, top float64, buf []byte) (*Table, error) {
	return LoadReader(name, top, bytes.NewReader(buf))
}

// Top returns the proportion of paid rankings for the tournament payout table.
func (t *Table) Top() float64 {
	return t.top
}

// Name returns the name of the tournament payout table.
func (t *Table) Name() string {
	return t.name
}

// Title returns the title of the tournament payout table.
func (t *Table) Title() string {
	return fmt.Sprintf("Top %d%%", int(t.Top()*100))
}

// Format satisfies the [fmt.Formatter] interface.
func (t *Table) Format(f fmt.State, verb rune) {
	switch verb {
	case 'd':
		fmt.Fprintf(f, "%d", int(t.Top()*100))
	case 'f':
		fmt.Fprintf(f, "%f", t.Top())
	case 't':
		fmt.Fprint(f, t.Title())
	case 'n':
		fmt.Fprint(f, t.Name())
	case 'c':
		_ = t.WriteCSV(f)
	case 'm':
		_ = t.WriteMarkdown(f, 1)
	case 's', 'v':
		_ = t.WriteTable(f, true)
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, Table)", verb)
	}
}

// WriteTable writes a plain text version of the tournament payout table, along
// with optional title, to w.
func (t *Table) WriteTable(w io.Writer, title bool) error {
	if title {
		if _, err := fmt.Fprintln(w, t.Title()+"\n"); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	return t.WriteTo(w, tableFormat, nil, true)
}

// WriteCSV writes a CSV version of the tournament payout table to w.
func (t *Table) WriteCSV(w io.Writer) error {
	return t.WriteTo(w, csvFormat, nil, true)
}

// WriteMarkdown writes a Markdown formatted version of the tournament payout
// table to w.
func (t *Table) WriteMarkdown(w io.Writer, header int) error {
	if 0 < header && header <= 6 {
		if _, err := fmt.Fprintln(w, strings.Repeat("#", header), t.Title()+"\n"); err != nil {
			return fmt.Errorf("unable to write markdown: %w", err)
		}
	}
	return t.WriteTo(w, markdownFormat, markdownDivider, false)
}

// WriteTo writes the tournament payout table to w, formatting cells with f,
// breaking line output when early is true.
func (t *Table) WriteTo(w io.Writer, f func(interface{}, int, bool) string, divider func(int, int) string, early bool) error {
	if f == nil {
		f = tableFormat
	}
	maxlen := 7
	// generate headers
	v := make([]string, len(t.entries))
	for i, entries := range t.entries {
		s := t.EntriesTitle(entries)
		v[i], maxlen = s, max(maxlen, len(s))
	}
	// write 0,0 cell
	var err error
	if _, err = fmt.Fprint(w, f("", maxlen, false)); err != nil {
		return fmt.Errorf("unable to write table: %w", err)
	}
	// write headers
	for i := 0; i < len(t.entries); i++ {
		if _, err = fmt.Fprint(w, f(v[i], maxlen, i == len(t.entries)-1)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	if _, err = fmt.Fprintln(w); err != nil {
		return fmt.Errorf("unable to write table: %w", err)
	}
	// write divider
	if divider != nil {
		if _, err = fmt.Fprintln(w, divider(len(t.entries)+1, maxlen)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	// write levels/entries
	for i := 0; i < len(t.levels); i++ {
		// write ranking
		if _, err = fmt.Fprint(w, f(t.LevelTitle(t.levels[i]-1), maxlen, false)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
		// write entries
		for j := 0; j < len(t.entries); j++ {
			last := j == len(t.entries)-1 || (early && t.amounts[i][j+1] == 0)
			if _, err = fmt.Fprint(w, f(t.amounts[i][j], maxlen, last)); err != nil {
				return fmt.Errorf("unable to write table: %w", err)
			}
			if early && last {
				break
			}
		}
		if _, err = fmt.Fprintln(w); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	return nil
}

// EntriesMax returns the max entries for the tournament payout table.
func (t *Table) EntriesMax() int {
	return t.entries[0]
}

// Entries returns the column for the entries in the tournament payout table.
func (t *Table) Entries(entries int) int {
	if entries < 2 {
		return -1
	}
	for i := len(t.entries) - 1; 0 <= i; i-- {
		if entries <= t.entries[i] {
			return i
		}
	}
	return 0
}

// Level returns the row for the level in the tournament payout table.
func (t *Table) Level(level int) int {
	return findLevel(level, t.levels)
}

// EntriesTitle returns the column title for the specified entries in the
// tournament payout table.
func (t *Table) EntriesTitle(entries int) string {
	col := t.Entries(entries)
	switch {
	case col < 0:
		return ""
	case col == 0:
		return EntriesTitle(0, t.entries[col])
	case col == len(t.entries)-1:
		return EntriesTitle(t.entries[col], 0)
	}
	return EntriesTitle(t.entries[col+1], t.entries[col])
}

// LevelTitle returns the row title for the level in the tournament payout
// table.
func (t *Table) LevelTitle(level int) string {
	for i, last := 0, 0; i < len(t.levels); i++ {
		if level < t.levels[i] {
			return LevelTitle(last, t.levels[i])
		}
		last = t.levels[i]
	}
	return ""
}

// MaxLevelTitle returns the row title for the max level in tournament payout
// table.
func (t *Table) MaxLevelTitle(level int) string {
	for i, last := 0, 0; i < len(t.levels); i++ {
		if level <= t.levels[i] {
			return LevelTitle(last, level)
		}
		last = t.levels[i]
	}
	return ""
}

// Paid returns the paid rankings, level (row) and corresponding entries column
// for the tournament payout table.
func (t *Table) Paid(entries int) (int, int, int) {
	paid := Paid(t.top, entries)
	return paid, t.Level(paid - 1), t.Entries(entries)
}

// Unallocated returns the unallocated amount for the paid rankings, row, and
// col from the tournament payout table.
func (t *Table) Unallocated(paid, row, col int) float64 {
	if row <= 0 || col < 0 || len(t.levels) <= row || len(t.entries) <= col {
		return 0.0
	}
	f := float64(t.levels[row]-paid) * t.amounts[row][col]
	for i, last := row+1, t.levels[row]; i < len(t.levels) && t.amounts[i][col] != 0.0; i++ {
		f += float64(t.levels[i]-last) * t.amounts[i][col]
		last = t.levels[i]
	}
	return f
}

// Stakes returns the paid levels as [low, high), the tournament table value
// per level, the calculated payouts per level, and the total amount paid.
func (t *Table) Stakes(entries int, buyin, guaranteed int64, rake float64) ([][2]int, []float64, []int64, int64) {
	paid, row, col := t.Paid(entries)
	prize, unallocated := Prize(entries, buyin, guaranteed, rake), t.Unallocated(paid, row, col)
	levels, amounts, payouts, total := make([][2]int, row+1), make([]float64, row+1), make([]int64, row+1), int64(0)
	for i, last := 0, 0; i <= row; i++ {
		level := min(t.levels[i], paid)
		amounts[i] = t.amounts[i][col]
		levels[i], payouts[i] = [2]int{last, level}, Calc(amounts[i], prize, unallocated)
		total += payouts[i] * int64(level-last)
		last = level
	}
	return levels, amounts, payouts, total
}

// Payouts returns the tournament payouts for rankings from [start, end) based
// on the number of entries.
func (t *Table) Payouts(entries int, buyin, guaranteed int64, rake float64) ([]int64, int64) {
	levels, _, amounts, total := t.Stakes(entries, buyin, guaranteed, rake)
	payouts := make([]int64, 0, levels[len(levels)-1][1])
	for i, level := range levels {
		for j := level[0]; j < level[1]; j++ {
			payouts = append(payouts, amounts[i])
		}
	}
	return payouts, total
}

// Amount returns the amount for the level and entries from the tournament
// payout table.
func (t *Table) Amount(level, entries int) float64 {
	return t.At(t.Level(level), t.Entries(entries))
}

// At returns the amount at row, col of the tournament payout table.
func (t *Table) At(row, col int) float64 {
	if 0 <= row && row < len(t.levels) && 0 <= col && col < len(t.entries) {
		return t.amounts[row][col]
	}
	return 0.0
}

// tableFormat formats v in table format.
func tableFormat(v any, n int, last bool) string {
	var s string
	switch x := v.(type) {
	case string:
		s = fmt.Sprintf("%*s", n, x)
	case float64:
		if x != 0.0 {
			s = fmt.Sprintf("%*.2f", n, x*100)
		}
	}
	if last {
		return s
	}
	return s + " "
}

// csvFormat formats v in CSV format.
func csvFormat(v any, n int, last bool) string {
	var s string
	switch x := v.(type) {
	case string:
		s = x
		if s == "" {
			s = reCell
		}
	case float64:
		if x != 0.0 {
			s = strconv.FormatFloat(math.Round(x*1000000.0)/10000.0, 'f', -1, 64)
		}
	}
	if last {
		return s
	}
	return s + ","
}

// markdownFormat formats v in Markdown format.
func markdownFormat(v any, n int, last bool) string {
	s := "| "
	switch x := v.(type) {
	case string:
		s += fmt.Sprintf("%*s", n, x)
	case float64:
		var v string
		if x != 0.0 {
			v = strconv.FormatFloat(math.Round(x*10000)/100, 'f', -1, 64)
		}
		s += fmt.Sprintf("%*s", n, v)
	}
	if last {
		return s + " |"
	}
	return s + " "
}

// markdownDivider returns a divider for Markdown tables.
func markdownDivider(count, n int) string {
	s, div := "", strings.Repeat("-", n+2)
	for i := 0; i < count; i++ {
		s += "|" + div
	}
	return s + "|"
}

// parseLevels parses levels title.
func parseLevels(s string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1st":
		return 1, nil
	case "2nd":
		return 2, nil
	case "3rd":
		return 3, nil
	case "4th":
		return 4, nil
	case "5th":
		return 5, nil
	case "6th":
		return 6, nil
	case "7th":
		return 7, nil
	case "8th":
		return 8, nil
	case "9th":
		return 9, nil
	case "10th":
		return 10, nil
	}
	return parseEntries(s)
}

// parseEntries parses the entries title.
func parseEntries(s string) (int, error) {
	v := strings.SplitN(s, "-", 2)
	i, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(v[len(v)-1]), "+"))
	if err != nil {
		return 0, fmt.Errorf("unable to parse col %q: %w", v[len(v)-1], err)
	}
	return i, nil
}

// findLevel returns the row for the level in levels.
func findLevel(level int, levels []int) int {
	if 0 <= level {
		for i, l := range levels {
			if level < l {
				return i
			}
		}
	}
	return -1
}

// reCell is the upper left cell value in a csv tournament pay table.
const reCell = "r/e"

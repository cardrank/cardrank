package paycalc

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoadReader(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"simple", simpleTable},
		{"top10", string(top10)},
		{"top15", string(top15)},
		{"top20", string(top20)},
	}
	for _, test := range tests {
		t.Run(test.name, testLoadReader(test.s))
	}
}

func testLoadReader(data string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		tbl, err := LoadReader(strings.NewReader(data), 0.10, "")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		t.Logf("\n%s\n%c\n%m", tbl, tbl, tbl)
		buf := new(bytes.Buffer)
		if err := tbl.WriteCSV(buf); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if s, exp := buf.String(), data; s != exp {
			t.Errorf("expected generated table csv output to match input, expected:\n%s\ngot:\n%s", exp, s)
		}
	}
}

const simpleTable = `r/e,401+,351-400,301-350,251-300,201-250,151-200,101-150,76-100,61-75,51-60,41-50,31-40,11-30,3-10,2
1st,24.5,25,25.5,26,26.5,27,28,30,31,35,37,40,50,70,100
2nd,14.25,14.5,14.75,15,15.5,16,17,20,21,22,25,25,30,30
3rd,9,9.2,9.4,9.6,9.8,10,10.6,12,13,15,15,20,20
4th,7,7.2,7.4,7.6,7.8,8,8.6,9.5,10,11,12,15
5th,6,6.2,6.4,6.6,6.8,7,7.6,8,8.5,9,11
6th,4.2,4.3,4.4,4.5,4.6,4.9,5.3,6,6.5,8
7th,3.2,3.3,3.4,3.5,3.6,3.9,4.3,5,5.5
8th,2.2,2.3,2.4,2.6,2.8,2.9,3.3,4,4.5
9th,1.65,1.85,1.95,2.1,2.2,2.4,2.7,3
10th,1.25,1.4,1.4,1.5,1.65,1.9,2.1,2.5
11-15,1.25,1.4,1.4,1.5,1.65,1.9,2.1
16-20,0.85,0.9,0.95,1,1.1,1.3
21-25,0.75,0.8,0.85,0.9,1
26-30,0.65,0.7,0.75,0.8
31-35,0.55,0.6,0.65
36-40,0.5,0.55
41-50,0.4
`

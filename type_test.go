package cardrank

import (
	"reflect"
	"testing"
)

func TestRazz(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r HandRank
	}{
		{"Kh Qh Jh Th 9h Ks Qs", "Kh Qh Jh Th 9h", "Ks Qs", 7936},
		{"Ah Kh Qh Jh Th Ks Qs", "Kh Qh Jh Th Ah", "Ks Qs", 7681},
		{"2h 2c 2d 2s As Ks Qs", "2h 2c As Ks Qs", "2d 2s", 59569},
		{"Ah Ac Ad Ks Kh Ks Qs", "Ah Ac Ks Kh Qs", "Ad Ks", 63067},
		{"Ah Ac Ad Ks Qh Ks Qs", "Ks Ks Qh Qs Ah", "Ac Ad", 62935},
		{"Kh Kd Qd Qs Jh Ks Js", "Qd Qs Jh Js Kh", "Kd Ks", 62813},
		{"3h 3c Kh Qd Jd Ks Qs", "3h 3c Kh Qd Jd", "Ks Qs", 59734},
		{"2h 2c Kh Qd Jd Ks Qs", "2h 2c Kh Qd Jd", "Ks Qs", 59514},
		{"3h 2c Kh Qd Jd Ks Qs", "Kh Qd Jd 3h 2c", "Ks Qs", 7174},
	}
	for i, test := range tests {
		best, unused := Must(test.b), Must(test.u)
		h := Razz.RankHand(Must(test.v), nil)
		if h.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, h.Pocket, test.r, h.HiRank)
		}
		if !reflect.DeepEqual(h.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, h.Pocket, best, h.HiBest)
		}
		if !reflect.DeepEqual(h.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, h.Pocket, unused, h.HiUnused)
		}
	}
}

func TestBadugi(t *testing.T) {
	tests := []struct {
		v string
		b string
		u string
		r HandRank
	}{
		{"Kh Qh Jh Th", "Th", "Kh Qh Jh", 25088},
		{"Kh Qh Jd Th", "Jd Th", "Kh Qh", 17920},
		{"Kh Qc Jd Th", "Qc Jd Th", "Kh", 11776},
		{"Ks Qc Jd Th", "Ks Qc Jd Th", "", 7680},
		{"2h 2c 2d 2s", "2s", "2h 2d 2c", 24578},
		{"Ah Kh Qh Jh", "Ah", "Kh Qh Jh", 24577},
		{"Kh Kd Qd Qs", "Kh Qs", "Kd Qd", 22528},
		{"Ah Ac Ad Ks", "Ks Ah", "Ad Ac", 20481},
		{"3h 3c Kh Qd", "Kh Qd 3c", "3h", 14340},
		{"2h 2c Kh Qd", "Kh Qd 2c", "2h", 14338},
		{"3h 2c Kh Ks", "Ks 3h 2c", "Kh", 12294},
		{"3h 2c Kh Qd", "Qd 3h 2c", "Kh", 10246},
		{"Ah 2c 4s 6d", "6d 4s 2c Ah", "", 43},
		{"Ac 2h 4d 6s", "6s 4d 2h Ac", "", 43},
		{"Ah 2c 3s 6d", "6d 3s 2c Ah", "", 39},
		{"Ah 2c 4s 5d", "5d 4s 2c Ah", "", 27},
		{"Ah 2c 3s 5d", "5d 3s 2c Ah", "", 23},
		{"Ah 2c 3s 4d", "4d 3s 2c Ah", "", 15},
		{"Ac 2h 3s 4d", "4d 3s 2h Ac", "", 15},
	}
	for i, test := range tests {
		best, unused := Must(test.b), Must(test.u)
		h := Badugi.RankHand(Must(test.v), nil)
		if h.HiRank != test.r {
			t.Errorf("test %d %v expected rank %d, got: %d", i, h.Pocket, test.r, h.HiRank)
		}
		if !reflect.DeepEqual(h.HiBest, best) {
			t.Errorf("test %d %v expected best %v, got: %v", i, h.Pocket, best, h.HiBest)
		}
		if !reflect.DeepEqual(h.HiUnused, unused) {
			t.Errorf("test %d %v expected unused %v, got: %v", i, h.Pocket, unused, h.HiUnused)
		}
	}
}

func TestNumberedStreets(t *testing.T) {
	exp := []string{"Ante", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "11th", "101st", "102nd", "River"}
	streets := NumberedStreets(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 90, 1, 1)
	v := make([]string, len(streets))
	for i := 0; i < len(streets); i++ {
		v[i] = streets[i].Name
	}
	if !reflect.DeepEqual(v, exp) {
		t.Errorf("expected items to be equal:\n%v\n%v", exp, v)
	}
}

func TestTypeUnmarshal(t *testing.T) {
	tests := []struct {
		s   string
		exp Type
	}{
		{"HOLDEM", Holdem},
		{"omaha", Omaha},
		{"studHiLo", StudHiLo},
		{"razz", Razz},
		{"BaDUGI", Badugi},
		{"fusIon", Fusion},
	}
	for i, test := range tests {
		typ := Type(^uint16(0))
		if err := typ.UnmarshalText([]byte(test.s)); err != nil {
			t.Fatalf("test %d expected no error, got: %v", i, err)
		}
		if typ != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, typ)
		}
	}
}

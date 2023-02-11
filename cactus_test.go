//go:build !portable

package cardrank

import (
	"testing"
)

func TestNextBitPermutation(t *testing.T) {
	n := uint32(31)
	for _, exp := range []uint32{47, 55, 59, 61, 62} {
		if n = nextBitPermutation(n); n != exp {
			t.Errorf("expected n == %d, got: %d", exp, n)
		}
	}
}

func TestPrimeProduct(t *testing.T) {
	tests := []struct {
		hand []Card
		exp  uint32
	}{
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x21103, 0x22103}, 0x2ccbb},
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x21103, 0x00001}, 0x0eee9},
		{[]Card{0x802713, 0x8004b25, 0x200291d, 0x00001, 0x00001}, 0x04fa3},
	}
	for i, test := range tests {
		if n := primeProduct(test.hand[0], test.hand[1], test.hand[2], test.hand[3], test.hand[4]); n != test.exp {
			t.Errorf("test %d %v expected %d == %d", i, test.hand, test.exp, n)
		}
	}
}

func TestPrimeProductBits(t *testing.T) {
	tests := []struct {
		bits uint32
		exp  uint32
	}{
		{0x079, 0x84f2},
		{0x158, 0x759b},
		{0x037, 0x10c2},
		{0x01f, 0x0906},
		{0x04e, 0x06f9},
		{0x063, 0x052e},
		{0x015, 0x006e},
		{0x00b, 0x002a},
		{0x001, 0x0002},
	}
	for i, test := range tests {
		if n := primeProductBits(test.bits); n != test.exp {
			t.Errorf("test %d expected primeProductBits(%d) == %d, got: %d", i, test.bits, test.exp, n)
		}
	}
}

func TestCactusMaps(t *testing.T) {
	flushes, unique5 := cactusMaps()
	if n, exp := len(flushes), 1287; n != exp {
		t.Fatalf("expected len(flush) == %d, got: %d", exp, n)
	}
	flushesTests := []struct {
		r   uint32
		exp HandRank
	}{
		{0x005ffe37, 0x0184},
		{0x003d1623, 0x0185},
		{0x003c6619, 0x0340},
		{0x003a5e5d, 0x0334},
		{0x00345631, 0x01b1},
		{0x0029f659, 0x047b},
		{0x0017ae13, 0x0167},
		{0x00166e15, 0x01b3},
		{0x0014b621, 0x03ba},
		{0x00104625, 0x026f},
		{0x000b2e27, 0x01aa},
		{0x0009ee29, 0x01d5},
		{0x00092e22, 0x0274},
		{0x00044e32, 0x016e},
		{0x0003f7af, 0x05cf},
		{0x0003a61e, 0x04a5},
		{0x00026e35, 0x01e3},
		{0x00021f1b, 0x04f7},
		{0x0001f617, 0x0307},
		{0x0001ce36, 0x01dc},
		{0x0000b3b2, 0x031d},
		{0x0000725a, 0x0320},
		{0x00006f4a, 0x0475},
		{0x00004e2a, 0x0580},
		{0x0000310e, 0x0605},
	}
	for i, test := range flushesTests {
		if n := flushes[test.r]; n != test.exp {
			t.Fatalf("test %d flush[%d] to be %d, got: %d", i, test.r, test.exp, n)
		}
	}
	if n, exp := len(unique5), 6175; n != exp {
		t.Fatalf("expected len(unique5) == %d, got: %d", exp, n)
	}
	unique5Tests := []struct {
		r   uint32
		exp HandRank
	}{
		{0x01c51151, 0x0a3e},
		{0x005f112f, 0x0695},
		{0x00529143, 0x0d0f},
		{0x0021110f, 0x0ee4},
		{0x001f912d, 0x1af7},
		{0x0017912b, 0x0826},
		{0x00171a57, 0x0c19},
		{0x0010914d, 0x0059},
		{0x000c1147, 0x0787},
		{0x000b1101, 0x0beb},
		{0x00049123, 0x13b1},
		{0x0003915a, 0x112b},
		{0x000382e3, 0x110f},
		{0x00029135, 0x1cc6},
		{0x0002112e, 0x1bc6},
		{0x0001913e, 0x1c60},
		{0x00019113, 0x0a89},
		{0x00011132, 0x0c2c},
		{0x00002d8e, 0x08d3},
		{0x0000115e, 0x1733},
		{0x00001144, 0x1818},
		{0x00001138, 0x0981},
		{0x00001117, 0x008c},
		{0x000010fe, 0x1639},
		{0x000010fb, 0x094a},
	}
	for i, test := range unique5Tests {
		if n := unique5[test.r]; n != test.exp {
			t.Fatalf("test %d unique5[%d] to be %d, got: %d", i, test.r, test.exp, n)
		}
	}
}

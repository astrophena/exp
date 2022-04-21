package binary

import "testing"

func TestDecimalAndBinary(t *testing.T) {
	const (
		n = 100
		// 1   1   0   0  1  0  0
		// 64  32  16  8  4  2  1
		nBin = 0b1100100
		// 1   1   0   0  1  0  0
		// 0   1   1   0  0  1  0
		// 64  32  16  8  4  2  1
	)

	if n != nBin {
		t.Fatal("n != nBin")
	}

	t.Logf("Type of n is %T", n)
	t.Logf("Type of nBin is %T", nBin)

	for i := 0; i < 10; i++ {
		t.Logf("n >> %d = %d", i, n>>i)
	}
	for i := 0; i < 10; i++ {
		t.Logf("n << %d = %d", i, n<<i)
	}
}

// See https://yourbasic.org/golang/bitmask-flag-set-clear/.

type bits uint8

const (
	f0 bits = 1 << iota
	f1
	f2
	f3
	f4
	f5
	f6
	f7
)

func set(b, flag bits) bits    { return b | flag }
func clear(b, flag bits) bits  { return b &^ flag }
func toggle(b, flag bits) bits { return b ^ flag }
func has(b, flag bits) bool    { return b&flag != 0 }

func TestBitMask(t *testing.T) {
	var n = bits(0b11111111)
	var n2 = bits(0b01010100)
	t.Logf("n = %d", n)
	t.Logf("f1 = %v", f1)

	t.Logf("has(n, f7) = %v", has(n, f7))
	t.Logf("has(n2, f7) = %v", has(n2, f7))

	var n3 bits
	t.Logf("n3 = %d (%b)", n3, n3)
	n3 = toggle(n3, f7)
	n3 = toggle(n3, f0)
	t.Logf("n3 = %d (%b)", n3, n3)

	n4 := f1 | f2 | f4 | f7
	t.Logf("n4 = %d (%b)", n4, n4)
}

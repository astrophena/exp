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
}

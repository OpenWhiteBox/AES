package number

import (
	"testing"
)

func TestBinaryMultiplicationArbitrary(t *testing.T) {
	var a, b ByteFieldElem = 0x57, 0x83

	c, d := a.longMul(b), b.longMul(a)

	if c != 0x2b79 || d != 0x2b79 {
		t.Fatalf("Multiplication is wrong, 0x57 * 0x83 != 0x83 * 0x57 != 0x2b79")
	}
}

func TestBinaryMultiplicationOne(t *testing.T) {
	var a, b ByteFieldElem = 0x57, 0x01

	c, d := a.longMul(b), b.longMul(a)

	if c != 0x57 || d != 0x57 {
		t.Fatalf("Multiplication by mulBytest. identity is wrong, 0x57 * 0x01 != 0x01 * 0x57 != 0x57")
	}
}

func TestBinaryMultiplicationZero(t *testing.T) {
	var a, b ByteFieldElem = 0x57, 0x00

	c, d := a.longMul(b), b.longMul(a)

	if c != 0x00 || d != 0x00 {
		t.Fatalf("Multiplication by add. identity is wrong, 0x57 * 0x00 != 0x00 * 0x57 != 0x00")
	}
}

func TestBinaryDivisionArbitrary(t *testing.T) {
	var numer, denom ByteFieldElem = 0x2b79, 0x11b

	quo, rem := numer.longDiv(denom)

	if quo != 0x28 {
		t.Fatalf("Quotient is wrong, (denom << 5) ^ (denom << 3) ^ 0xc1 != 0x2b79")
	}

	if rem != 0xc1 {
		t.Fatalf("Remainder is wrong, 0x57 * 0x83 mod 0x11b != 0xc1")
	}
}

func TestBinaryDivisionOne(t *testing.T) {
	var numer, denom ByteFieldElem = 0xc1, 0x01

	quo, rem := numer.longDiv(denom)

	if quo != 0xc1 {
		t.Fatalf("Quotient is wrong, 0xc1 / 0x01 != 0xc1")
	}

	if rem != 0x00 {
		t.Fatalf("Remainder is wrong, 0xc1 mod 0x01 != 0x00")
	}
}

func TestBinaryDivisionZero(t *testing.T) {
	var numer, denom ByteFieldElem = 0xc1, 0x00

	_, rem := numer.longDiv(denom) // Quotient doesn't matter

	if rem != 0xc1 {
		t.Fatalf("Remainder is wrong, 0xc1 mod 0x00 != 0xc1")
	}
}

func TestByteFieldElemMul(t *testing.T) {
	x, y := ByteFieldElem(0x57), ByteFieldElem(0x83)

	if x.Mul(y) != 0xc1 || y.Mul(x) != 0xc1 {
		t.Fatalf("0x57 * 0x83 != 0xc1")
	}
}

func TestByteFieldElemInvert(t *testing.T) {
	x := ByteFieldElem(0x02)
	y := x.Invert()

	if x.Mul(y) != 0x01 || y.Mul(x) != 0x01 {
		t.Fatalf("0x02 * (0x02)^-1 != 0x01")
	}
}

func TestArrayFieldElemTrim(t *testing.T) {
	w := ArrayFieldElem{ByteFieldElem(0), ByteFieldElem(0), ByteFieldElem(0)}
	x := ArrayFieldElem{ByteFieldElem(0), ByteFieldElem(0), ByteFieldElem(1)}
	y := ArrayFieldElem{ByteFieldElem(0), ByteFieldElem(1), ByteFieldElem(0)}
	z := ArrayFieldElem{ByteFieldElem(1), ByteFieldElem(0), ByteFieldElem(0)}

	if len(w.trim()) != 0 {
		t.Fatalf("Polynomial trimmed incorrectly (should be empty)")
	}

	if len(x.trim()) != 3 {
		t.Fatalf("Polynomial trimmed incorrectly (should be three)")
	}

	if len(y.trim()) != 2 {
		t.Fatalf("Polynomial trimmed incorrectly (should be two)")
	}

	if len(z.trim()) != 1 {
		t.Fatalf("Polynomial trimmed incorrectly (should be one)")
	}
}

func TestArrayFieldElemMultiplicationArbitrary(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{ByteFieldElem(0x0e), ByteFieldElem(0x09), ByteFieldElem(0x0d), ByteFieldElem(0x0b)}

	if !x.Mul(y).IsOne() || !y.Mul(x).IsOne() {
		t.Fatalf("Multiplication is wrong, element * inverse != 1")
	}
}

func TestArrayFieldElemMultiplicationOne(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{ByteFieldElem(0x01)}

	xy, yx := x.Mul(y), y.Mul(x)

	for i := 0; i < 4; i++ {
		if xy[i] != x[i] || yx[i] != x[i] {
			t.Fatalf("Multiplication is wrong, element * 1 != element")
		}
	}
}

func TestArrayFieldElemMultiplicationZero(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{ByteFieldElem(0x00)}

	if !x.Mul(y).IsZero() || !y.Mul(x).IsZero() {
		t.Fatalf("Multiplication is wrong, element * 0 != 0")
	}
}

func TestArrayFieldElemDivisionArbitrary(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{ByteFieldElem(0x0e), ByteFieldElem(0x09), ByteFieldElem(0x0d), ByteFieldElem(0x0b)}
	xy := x.longMul(y)

	quo, rem := xy.longDiv(arrayModulus)

	if !rem.IsOne() {
		t.Fatalf("Remainder is wrong, should be one.")
	}

	back := arrayModulus.longMul(quo).Add(rem)

	if len(xy) != len(back) {
		t.Fatalf("x * y isn't equal to (denom * quotient) + remainder")
	}

	for i := 0; i < len(xy); i++ {
		if xy[i] != back[i] {
			t.Fatalf("x * y isn't equal to (denom * quotient) + remainder")
		}
	}
}

func TestArrayFieldElemDivisionOne(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{ByteFieldElem(0x01)}

	quo, rem := x.longDiv(y)

	for i := 0; i < 4; i++ {
		if x[i] != quo[i] {
			t.Fatalf("Quotient is wrong, element / 1 != element")
		}
	}

	if !rem.IsZero() {
		t.Fatalf("Remainder is wrong, should be zero.")
	}
}

func TestArrayFieldElemDivisionZero(t *testing.T) {
	x := ArrayFieldElem{ByteFieldElem(0x02), ByteFieldElem(0x01), ByteFieldElem(0x01), ByteFieldElem(0x03)}
	y := ArrayFieldElem{}

	_, rem := x.longDiv(y)

	for i := 0; i < 4; i++ {
		if x[i] != rem[i] {
			t.Fatalf("Remainder is wrong, should be element")
		}
	}
}

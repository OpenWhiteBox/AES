package gfmatrix

import (
	"testing"
)

var (
	E2_4 = Row{0x00, 0x01, 0x00, 0x00}
	E3_4 = Row{0x00, 0x00, 0x01, 0x00}

	PERM     = Row{0x00, 0x03, 0x02, 0x01}
	NOT_PERM = Row{0x01, 0x02, 0x03, 0x04}
)

func TestLessThan(t *testing.T) {
	if !LessThan(E2_4, E3_4) {
		t.Fatal("LessThan returned false on true case.")
	}

	if LessThan(E3_4, E3_4) {
		t.Fatal("LessThan returned true on equivalent rows.")
	}

	if LessThan(E3_4, E2_4) {
		t.Fatal("LessThan returned true on false case.")
	}
}

func TestIsPermutation(t *testing.T) {
	if !PERM.IsPermutation() {
		t.Fatal("IsPermutation returned false on true case.")
	}

	if NOT_PERM.IsPermutation() {
		t.Fatal("IsPermutation returned true on false case.")
	}
}

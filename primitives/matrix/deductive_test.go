package matrix

import (
	"testing"

	"crypto/rand"
)

func testingDeductiveMatrix() DeductiveMatrix {
	dm := NewDeductiveMatrix(128)
	m := GenerateRandom(rand.Reader, 128)

	found := 0
	for found < 126 {
		in := GenerateRandomRow(rand.Reader, 128)
		out := m.Mul(in)

		learned := dm.Assert(in, out)
		if learned {
			found++
		}
	}

	return dm
}

func TestDeductiveMatrixNovelInput(t *testing.T) {
	dm := testingDeductiveMatrix()

	for i := 0; i < 100; i++ {
		x := dm.NovelInput()

		if dm.IsInDomain(x) {
			t.Fatal("NovelInput returned row in domain!")
		}
	}
}

func TestDeductiveMatrixSuccess(t *testing.T) {
	dm := NewDeductiveMatrix(128)

	m := GenerateRandom(rand.Reader, 128)
	mInv, _ := m.Invert()

	assertions := 0
	for !dm.FullyDefined() {
		in := GenerateRandomRow(rand.Reader, 128)
		out := m.Mul(in)

		dm.Assert(in, out)
		assertions += 1
	}

	t.Logf("Took %v assertions to deduce 128-by-128 matrix.", assertions)

	dmMatrix, dmInv := dm.Matrix(), dm.Inverse()

	for i, _ := range m {
		for j, _ := range m[i] {
			if m[i][j] != dmMatrix[i][j] {
				t.Fatal("Deduced matrix is different than real matrix!")
			}

			if mInv[i][j] != dmInv[i][j] {
				t.Fatal("Deduced inverse matrix is different than real inverse matrix!")
			}
		}
	}
}

package matrix

import (
	"testing"

	"crypto/rand"
)

func recoverTestFromPanic(t *testing.T) func() {
	return func() {
		if r := recover(); r != nil {
			t.Log("Recovered goroutine from panic--test passed.")
		}
	}
}

func testingDeductiveMatrix() *DeductiveMatrix {
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

func TestDeductiveMatrixFailure1(t *testing.T) {
	defer recoverTestFromPanic(t)()
	dm := testingDeductiveMatrix()

	// Generate a random un-novel input element.
	in := dm.Input.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	// Generate a random un-novel output element.
	out := dm.Output.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

func TestDeductiveMatrixFailure2(t *testing.T) {
	defer recoverTestFromPanic(t)()
	dm := testingDeductiveMatrix()

	// Generate a random novel input element.
	in := dm.Input.NovelRow(0)

	// Generate a random un-novel output element.
	out := dm.Output.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

func TestDeductiveMatrixFailure3(t *testing.T) {
	defer recoverTestFromPanic(t)()
	dm := testingDeductiveMatrix()

	// Generate a random un-novel input element.
	in := dm.Input.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	// Generate a random novel output element.
	out := dm.Output.NovelRow(0)

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

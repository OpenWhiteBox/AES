package gfmatrix

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

func testingDeductiveMatrix() DeductiveMatrix {
	dm := NewDeductiveMatrix(16)
	m, _ := GenerateRandom(rand.Reader, 16)

	found := 0
	for found < 14 {
		in := GenerateRandomRow(rand.Reader, 16)
		out := m.Mul(in)

		learned := dm.Assert(in, out)
		if learned {
			found++
		}
	}

	return dm
}

func TestDeductiveMatrixSuccess(t *testing.T) {
	dm := NewDeductiveMatrix(16)

	m, mInv := GenerateRandom(rand.Reader, 16)

	assertions := 0
	for !dm.FullyDefined() {
		in := GenerateRandomRow(rand.Reader, 16)
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
	in := dm.input.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	// Generate a random un-novel output element.
	out := dm.output.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

func TestDeductiveMatrixFailure2(t *testing.T) {
	defer recoverTestFromPanic(t)()
	dm := testingDeductiveMatrix()

	// Generate a random novel input element.
	in := dm.NovelInput()

	// Generate a random un-novel output element.
	out := dm.output.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

func TestDeductiveMatrixFailure3(t *testing.T) {
	defer recoverTestFromPanic(t)()
	dm := testingDeductiveMatrix()

	// Generate a random un-novel input element.
	in := dm.input.Matrix().Transpose().Mul(GenerateRandomRow(rand.Reader, 128))

	// Generate a random novel output element.
	out := dm.NovelOutput()

	dm.Assert(in, out)
	t.Fatal("Goroutine did not panic when it should've!")
}

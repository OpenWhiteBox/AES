package matrix

// DeductiveMatrix is a generalization of IncrementalMatrix that allows the incremental deduction of matrices.
//
// Like incremental matrices, its primary use-case is in cryptanalyses and search algorithms, where we can do some work
// to obtain an (input, output) pair that we believe defines a matrix. We don't want to do more work than necessary and
// we also can't just obtain n (input, output) pairs because of linear dependence, etc.
type DeductiveMatrix struct {
	Input, Output IncrementalMatrix
}

// NewDeductiveMatrix returns a new n-by-n deductive matrix.
func NewDeductiveMatrix(n int) *DeductiveMatrix {
	return &DeductiveMatrix{
		Input:  NewIncrementalMatrix(n),
		Output: NewIncrementalMatrix(n),
	}
}

// Assert represents an assertion that A(in) = out. The function will panic if this is inconsistent with previous
// assertions. It it's not, it returns whether or not the assertion contained new information about A.
func (dm *DeductiveMatrix) Assert(in, out Row) (learned bool) {
	inReduced, inInverse := dm.Input.reduce(in)
	outReduced, outInverse := dm.Output.reduce(out)

	if inReduced.IsZero() || outReduced.IsZero() {
		real := dm.Output.Matrix().Transpose().Mul(inInverse)

		if !real.Equals(out) {
			panic("Asserted input, output pair is inconsistent with previous assertions!")
		}
		return false
	}

	dm.Input.addRows(in, inReduced, inInverse)
	dm.Output.addRows(out, outReduced, outInverse)
	return true
}

// FullyDefined returns true if the assertions made give a fully defined matrix.
func (dm *DeductiveMatrix) FullyDefined() bool {
	return dm.Input.FullyDefined() && dm.Output.FullyDefined()
}

// Matrix returns the deduced matrix.
func (dm *DeductiveMatrix) Matrix() Matrix {
	if !dm.FullyDefined() {
		return nil
	}
	return dm.Input.Inverse().Compose(dm.Output.Matrix()).Transpose()
}

// Inverse returns the deduced matrix's inverse.
func (dm *DeductiveMatrix) Inverse() Matrix {
	if !dm.FullyDefined() {
		return nil
	}
	return dm.Output.Inverse().Compose(dm.Input.Matrix()).Transpose()
}

// Dup returns a duplicate of dm.
func (dm *DeductiveMatrix) Dup() *DeductiveMatrix {
	return &DeductiveMatrix{
		Input:  dm.Input.Dup(),
		Output: dm.Output.Dup(),
	}
}

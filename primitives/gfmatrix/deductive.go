package gfmatrix

// DeductiveMatrix is a generalization of IncrementalMatrix that allows the incremental deduction of matrices.
//
// Like incremental matrices, its primary use-case is in cryptanalyses and search algorithms, where we can do some work
// to obtain an (input, output) pair that we believe defines a matrix. We don't want to do more work than necessary and
// we also can't just obtain n (input, output) pairs because of linear dependence, etc.
type DeductiveMatrix struct {
	input, output IncrementalMatrix
}

// NewDeductiveMatrix returns a new n-by-n deductive matrix.
func NewDeductiveMatrix(n int) DeductiveMatrix {
	return DeductiveMatrix{
		input:  NewIncrementalMatrix(n),
		output: NewIncrementalMatrix(n),
	}
}

// Assert represents an assertion that A(in) = out. The function will panic if this is inconsistent with previous
// assertions. It it's not, it returns whether or not the assertion contained new information about A.
func (dm *DeductiveMatrix) Assert(in, out Row) (learned bool) {
	inReduced, inInverse := dm.input.reduce(in)
	outReduced, outInverse := dm.output.reduce(out)

	if inReduced.IsZero() || outReduced.IsZero() {
		real := dm.output.Matrix().Transpose().Mul(inInverse)

		if !real.Equals(out) {
			panic("Asserted input, output pair is inconsistent with previous assertions!")
		}
		return false
	}

	dm.input.addRows(in, inReduced, inInverse)
	dm.output.addRows(out, outReduced, outInverse)
	return true
}

// FullyDefined returns true if the assertions made give a fully defined matrix.
func (dm *DeductiveMatrix) FullyDefined() bool {
	return dm.input.FullyDefined() && dm.output.FullyDefined()
}

// NovelInput returns a random x not in the domain of A.
func (dm *DeductiveMatrix) NovelInput() Row {
	return dm.input.Novel()
}

// NovelOutput returns a random y not in the span of A.
func (dm *DeductiveMatrix) NovelOutput() Row {
	return dm.output.Novel()
}

// IsInDomain returns whether or not x is in the known span of A.
func (dm *DeductiveMatrix) IsInDomain(x Row) bool {
	return dm.input.IsInSpan(x)
}

// IsInSpan returns whether or not y is in the known span of A.
func (dm *DeductiveMatrix) IsInSpan(y Row) bool {
	return dm.output.IsInSpan(y)
}

// Matrix returns the deduced matrix.
func (dm *DeductiveMatrix) Matrix() Matrix {
	if !dm.FullyDefined() {
		return nil
	}
	return dm.input.Inverse().Compose(dm.output.Matrix()).Transpose()
}

// Inverse returns the deduced matrix's inverse.
func (dm *DeductiveMatrix) Inverse() Matrix {
	if !dm.FullyDefined() {
		return nil
	}
	return dm.output.Inverse().Compose(dm.input.Matrix()).Transpose()
}

// Dup returns a duplicate of dm.
func (dm *DeductiveMatrix) Dup() DeductiveMatrix {
	return DeductiveMatrix{
		input:  dm.input.Dup(),
		output: dm.output.Dup(),
	}
}

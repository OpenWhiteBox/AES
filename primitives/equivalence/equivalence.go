// Package equivalence implements the linear equivalence algorithm. TODO: The affine equivalence algorithm.
package equivalence

import (
	"fmt"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

// A (linear) equivalence is a pair of matrices, (A, B) such that f(A(x)) = B(g(x)) for all x.
type Equivalence struct {
	A, B matrix.Matrix
}

// LinearEquivalence finds linear equivalences between f and g. cap is the maximum number of equivalences to return.
func LinearEquivalence(f, g encoding.Byte, cap int) []Equivalence {
	return search(f, g, matrix.NewDeductiveMatrix(8), matrix.NewDeductiveMatrix(8), 0, 0, cap)
}

// search contains the search logic of our dynamic programming algorithm.
// f and g are the functions we're finding equivalences for. A and B are the parasites. posA and posB are our positions
// in the span of A and B, respectively. cap is the maximum number of equivalences to return.
func search(f, g encoding.Byte, A, B *matrix.DeductiveMatrix, posA, posB, cap int) (res []Equivalence) {
	x, novelOutputSize := A.Input.NovelRow(0), A.Output.NovelSize()

	// 1. Take a guess for A(x).
	// 2. Check if its possible for any matrix B to satisfy an equivalence relation with what we know about A.
	for i := 0; i < novelOutputSize; i++ {
		AT, BT := A.Dup(), B.Dup()
		AT.Assert(x, A.Output.NovelRow(i))

		posAT, posBT, consistent := learn(f, g, AT, BT, posA, posB)

		// Our guess for A(x) ...
		if !consistent { // ... isn't consistent with any equivalence relation.
			continue
		} else if AT.FullyDefined() { // ... uniquely specified an equivalence relation.
			res = append(res, Equivalence{
				A: AT.Matrix(),
				B: BT.Matrix(),
			})
		} else { // ... has neither led to a contradiction nor a full definition.
			res = append(res, search(f, g, AT, BT, posAT, posBT, cap-len(res))...)
		}

		if len(res) >= cap {
			return
		}
	}

	return
}

// learn finds the logical implications of what we've specified about A and B.
// f and g are the functions we're finding equivalences for. A and B are the parasites.
// learn returns whether or not A and B are consistent with any possible equivalence. A and B are mutated to contain the
// new information.
func learn(f, g encoding.Byte, A, B *matrix.DeductiveMatrix, posA, posB int) (posAT, posBT int, consistent bool) {
	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprint(r) == "Asserted input, output pair is inconsistent with previous assertions!" {
				consistent = false
			} else {
				panic(r)
			}
		}
	}()

	consistent = true
	learning := true

	size := 0

	// We have to loop because of the "Needlework Effect." A gives info on B, which may in turn give more info on A, ....
	for learning {
		learning = false

		// For every vector x in the domain of A, we have that f(Ax) = B * g(x) -- the span of A gives us input-output
		// behavior that B *has* to satsify, and for every A(x) = y that we guess correctly, we get twice as many
		// input-output pairs. This means that we end up with a complete definition of B (allowing us to derive A) much
		// faster than if we were to try to guess all of A first (and then derive B).
		size = A.Input.Size()
		for ; posA < size; posA++ {
			x, y := A.Input.Row(posA), A.Output.Row(posA)

			xT := matrix.Row{g.Encode(x[0])}
			yT := matrix.Row{f.Encode(y[0])}

			learned := B.Assert(xT, yT)
			learning = learned || learned
		}

		// Lets say that we know two input-output pairs of B: f(Ax) = B * g(x) and f(Ay) = B * g(y). Because B is linear,
		// we get the following statement about B for free: f(Ax) + f(Ay) = B * (g(x) + g(y)). But this means there has to
		// be some z such that f(Ax) + f(Ay) = f(Az) and g(x) + g(y) = g(z). Find z and Az by solving the two previous
		// equations.
		size = B.Input.Size()
		for ; posB < size; posB++ {
			x, y := B.Input.Row(posB), B.Output.Row(posB)

			z := matrix.Row{g.Decode(x[0])}
			Az := matrix.Row{f.Decode(y[0])}

			learned := A.Assert(z, Az)
			learning = learning || learned
		}
	}

	return posA, posB, consistent
}

package equivalence

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type Equivalence struct {
	A, B matrix.Matrix
}

func LinearEquivalence(f, g encoding.Byte) <-chan Equivalence {
	res := make(chan Equivalence)

	go func() {
		search(f, g, NewMatrix(), NewMatrix(), res)
		close(res)
	}()

	return res
}

// search contains the search logic of our dynamic programming algorithm.
// f and g are the functions we're finding equivalences for. A and B are the parasites. As we find equivalences, we send
// them back on the channel res.
func search(f, g encoding.Byte, A, B Matrix, res chan Equivalence) {
	x := A.NovelInput()

	// 1. Take a guess for A(x).
	// 2. Check if its possible for any matrix B to satisfy an equivalence relation with what we know about A.
	for _, guess := range Universe {
		if A.IsInSpan(guess) { // Our guess for A(x) can't result in A being singular.
			continue
		}

		AT, BT := A.Dup(), B.Dup()
		AT.Assert(x, guess)

		consistent := learn(f, g, AT, BT)

		// Our guess for A(x) ...
		if !consistent { // ... isn't consistent with any equivalence relation.
			continue
		} else if AT.FullyDefined() { // ... uniquely specified an equivalence relation.
			res <- Equivalence{
				A: AT.Matrix(),
				B: BT.Matrix(),
			}
		} else { // ... has neither led to a contradiction nor a full definition.
			search(f, g, AT, BT, res)
		}
	}

	return
}

// learn finds the logical implications of what we've specified about A and B.
// f and g are the functions we're finding equivalences for. A and B are the parasites.
// learn returns whether or not A and B are consistent with any possible equivalence. A and B are mutated to contain the
// new information.
func learn(f, g encoding.Byte, A, B Matrix) (consistent bool) {
	defer func() {
		if r := recover(); r != nil {
			consistent = false
		}
	}()

	consistent = true
	learning := true

	// We have to loop because of the "Needlework Effect." A gives info on B, which may in turn give more info on A, ....
	for learning {
		learning = false

		// For every vector x in the domain of A, we have that f(Ax) = B * g(x) -- the span of A gives us input-output
		// behavior that B *has* to satsify, and for every A(x) = y that we guess correctly, we get twice as many
		// input-output pairs. This means that we end up with a complete definition of B (allowing us to derive A) much
		// faster than if we were to try to guess all of A first (and then derive B).
		for elem := range A.Span() {
			x := matrix.Row{g.Encode(elem.In[0])}
			y := matrix.Row{f.Encode(elem.Out[0])}

			learning = learning || B.Assert(x, y)
		}

		// Lets say that we know two input-output pairs of B: f(Ax) = B * g(x) and f(Ay) = B * g(y). Because B is linear,
		// we get the following statement about B for free: f(Ax) + f(Ay) = B * (g(x) + g(y)). But this means there has to
		// be some z such that f(Ax) + f(Ay) = f(Az) and g(x) + g(y) = g(z). Find z and Az by solving the two previous
		// equations.
		for elem := range B.Span() {
			z := matrix.Row{g.Decode(elem.In[0])}
			Az := matrix.Row{f.Decode(elem.Out[0])}

			learning = learning || A.Assert(z, Az)
		}
	}

	return
}

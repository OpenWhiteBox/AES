package matrix2

// gaussJordan reduces the matrix according to the Gauss-Jordan Method.  Returns the augment matrix, the transformed
// matrix, at what column elimination failed (if ever), and whether it succeeded (meaning the augment matrix computes
// the transformation).
func (e Matrix) gaussJordan() (Matrix, Matrix, int, bool) {
	out, in := e.Size()
	aug := GenerateIdentity(out)

	f := make([]Row, out) // Duplicate e away so we don't mutate it.
	copy(f, e)

	for row, _ := range f {
		if row >= in { // The matrix is tall and thin--we've finished before exhausting all the rows.
			break
		}

		// Find a row with a non-zero entry in the (row)th position.
		candId := -1
		for j, f_j := range f[row:] {
			if !f_j[row].IsZero() {
				candId = j + row
				break
			}
		}

		if candId == -1 { // If we can't find one, fail and return our partial work.
			return aug, f, row, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		aug[row], aug[candId] = aug[candId], aug[row]

		// Normalize the entry in this rows (row)th position.
		correction := f[row][row].Invert()
		f[row] = f[row].ScalarMul(correction)
		aug[row] = aug[row].ScalarMul(correction)

		// Cancel out the (row)th position for every row above and below it.
		for i, _ := range f {
			if i != row && !f[i][row].IsZero() {
				aug[i] = aug[i].Add(aug[row].ScalarMul(f[i][row]))
				f[i] = f[i].Add(f[row].ScalarMul(f[i][row]))
			}
		}
	}

	return aug, f, -1, true
}

package matrix

// gaussJordan reduces the matrix according to the Gauss-Jordan Method.  Returns the transformed matrix, at what column
// elimination failed (if ever), and whether it succeeded (meaning the augment matrix computes the transformation).
func (e Matrix) gaussJordan(aug Matrix, lower, upper int, ignore RowIgnore) (Matrix, int, bool) {
	a, b := e.Size()

	f := make([]Row, a) // Duplicate e away so we don't mutate it.
	copy(f, e)

	for i, _ := range f[lower:upper] {
		row := i + lower

		if row >= b { // The matrix is tall and thin--we've finished before exhausting all the rows.
			break
		}

		// Find a row with a non-zero entry in the (col)th position
		candId := -1
		for j, f_j := range f[row:] {
			if !ignore(j+row) && f_j.GetBit(row) == 1 {
				candId = j + row
				break
			}
		}

		if candId == -1 && lower > 0 {
			for j, f_j := range f[:lower] {
				if !ignore(j) && f_j.GetBit(row) == 1 {
					candId = j
					break
				}
			}

			if candId == -1 {
				return f, row, false
			}
		} else if candId == -1 { // If we can't find one and there's nowhere else, fail and return our partial work.
			return f, row, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		aug[row], aug[candId] = aug[candId], aug[row]

		// Cancel out the (row)th position for every row above and below it.
		for i, _ := range f {
			if !ignore(i) && i != row && f[i].GetBit(row) == 1 {
				f[i] = f[i].Add(f[row])
				aug[i] = aug[i].Add(aug[row])
			}
		}
	}

	return f, -1, true
}

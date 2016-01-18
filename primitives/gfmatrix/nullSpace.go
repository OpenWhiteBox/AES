package gfmatrix

// NullSpace returns a basis for the matrix's nullspace.
func (e Matrix) NullSpace() (basis []Row) {
	out, in := e.Size()
	if out == 0 {
		return []Row{}
	}

	_, f, frees := e.gaussJordan()

	for _, free := range frees {
		input := NewRow(in)
		input[free] = 0x01

		for _, row := range f {
			if !row[free].IsZero() {
				input[row.Height()] = row[free]
			}
		}

		basis = append(basis, input)
	}

	return
}

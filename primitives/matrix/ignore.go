package matrix

// RowIgnore blacklists rows of a matrix from an operation; rows are handeled at the byte-level, not the bit-level. It's
// used by GeneratePartialIdentity and GeneratePartialRandom to leave empty rows in a matrix.
type RowIgnore func(int) bool

// IgnoreNoRows implements the RowIgnore interface. It sets no rows to be blacklisted.
func IgnoreNoRows(row int) bool {
	return false
}

// IgnoreRows returns an impementation of the RowIgnore interface which is true at all given positions and false at all
// others.
func IgnoreRows(positions ...int) RowIgnore {
	return func(row int) bool {
		pos := row / 8

		for _, cand := range positions {
			if pos == cand {
				return true
			}
		}

		return false
	}
}

// ByteIgnore blacklists blocks of a matrix from an operation; rows and columns are handeled at the byte-level, not the
// bit-level. It's used by GenerateRandomPartial to control where a matrix is random and fixed.
type ByteIgnore func(int, int) bool

// IgnoreNoBytes implements the ByteIgnore interface. It sets no blocks to be blacklisted.
func IgnoreNoBytes(row, col int) bool {
	return false
}

// IgnoreBytes returns an implementation of the ByteIgnore interface which is true if the row OR column equals a given
// position and false otherwise.
func IgnoreBytes(positions ...int) ByteIgnore {
	return func(row, col int) bool {
		for _, pos := range positions {
			if row == pos || col == pos {
				return true
			}
		}

		return false
	}
}

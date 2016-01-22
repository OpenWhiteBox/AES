package equivalence

import (
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func TestIMAddRow(t *testing.T) {
	e := NewIncrementalMatrix()

	ok1 := e.Add(matrix.Row{0x01})
	ok2 := e.Add(matrix.Row{0x02})
	ok3 := e.Add(matrix.Row{0x04})
	ok4 := e.Add(matrix.Row{0x06})
	ok5 := e.Add(matrix.Row{0x05})

	if !ok1 || !ok2 || !ok3 || ok4 || ok5 {
		t.Fatalf("IncrementalMatrix.AddRow behaved incorrectly! %v %v %v %v", ok1, ok2, ok3, ok4)
	}
}

func TestAssert(t *testing.T) {
	e, learned := NewMatrix(), true

	learned = e.Assert(matrix.Row{0x01}, matrix.Row{0x39})
	if learned != true {
		t.Fatalf("Assert returned false on first novel assertion.")
	}

	learned = e.Assert(matrix.Row{0x02}, matrix.Row{0x90})
	if learned != true {
		t.Fatalf("Assert returned false on second novel assertion.")
	}

	learned = e.Assert(matrix.Row{0x03}, matrix.Row{0x39 ^ 0x90})
	if learned != false {
		t.Fatalf("Assert returned true on redundant assertion.")
	}
}

func TestAssertFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Assert didn't panic when given contradicting assertions.")
		}
	}()

	e := NewMatrix()
	e.Assert(matrix.Row{0x01}, matrix.Row{0x39})
	e.Assert(matrix.Row{0x02}, matrix.Row{0x90})
	e.Assert(matrix.Row{0x03}, matrix.Row{0x39 ^ 0x90 ^ 0x01})
}

func TestNovelInput(t *testing.T) {
	e := NewMatrix()
	e.Assert(matrix.Row{0x01}, matrix.Row{0x01})
	e.Assert(matrix.Row{0x02}, matrix.Row{0x02})
	e.Assert(matrix.Row{0x04}, matrix.Row{0x04})

	x := e.NovelInput()
	if x[0] < 0x08 {
		t.Fatalf("NovelInput gave input in known domain.")
	}
}

func TestIsInSpan(t *testing.T) {
	e := NewMatrix()
	e.Assert(matrix.Row{0x01}, matrix.Row{0x39})
	e.Assert(matrix.Row{0x02}, matrix.Row{0x90})

	if !e.IsInSpan(matrix.Row{0x39 ^ 0x90}) {
		t.Fatalf("IsInSpan returned false for value in span.")
	}

	if e.IsInSpan(matrix.Row{0x39 ^ 0x90 ^ 0x01}) {
		t.Fatalf("IsInSpan returned true for value not in span.")
	}
}

func TestFullyDefined(t *testing.T) {
	e := NewMatrix()
	for i := uint(0); i < 7; i++ {
		e.Assert(matrix.Row{1 << i}, matrix.Row{1 << i})
	}

	if e.FullyDefined() {
		t.Fatalf("FullyDefined returned true for under-defined matrix.")
	}

	e.Assert(matrix.Row{0x80}, matrix.Row{0x80})

	if !e.FullyDefined() {
		t.Fatalf("FullyDefined returned false for fully-defined matrix.")
	}
}

func TestDup(t *testing.T) {
	e := NewMatrix()
	e.Assert(matrix.Row{0x01}, matrix.Row{0x01})

	f := e.Dup()

	e.Assert(matrix.Row{0x02}, matrix.Row{0x02})
	f.Assert(matrix.Row{0x02}, matrix.Row{0x04})

	if e.IsInSpan(matrix.Row{0x04}) {
		t.Fatalf("Original e has vector in span that it shouldn't.")
	} else if f.IsInSpan(matrix.Row{0x02}) {
		t.Fatalf("Copy f has vector in span that it shouldn't.")
	}
}

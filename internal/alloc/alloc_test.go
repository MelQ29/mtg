package alloc

import "testing"

func TestOtherBuiltDeckBlocksTheOnlyCopy(t *testing.T) {
	if !CanPlace(Line{Owned: 1}, 1) {
		t.Fatal("a free copy should fit")
	}
	if CanPlace(Line{Owned: 1, InOtherBuilt: 1}, 1) {
		t.Fatal("the other built deck already holds it")
	}
	if Free(Line{Owned: 1, InThis: 1}) != 1 {
		t.Fatal("the deck being edited does not lock its own copy")
	}
}

func TestRemoveCannotPassZero(t *testing.T) {
	if CanPlace(Line{Owned: 1, InThis: 0}, -1) {
		t.Fatal("cannot remove a copy that is not in the deck")
	}
	if !CanPlace(Line{Owned: 1, InThis: 2}, -1) {
		t.Fatal("removing one of two should be allowed")
	}
}

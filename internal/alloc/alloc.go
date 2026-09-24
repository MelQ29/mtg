package alloc

// Line is one printing while a deck is being edited.
// InThis is that deck's own count and does not reduce Free.
type Line struct {
	Owned        int
	InThis       int
	InOtherBuilt int
}

// Free is how many physical copies are not already in some other built deck.
func Free(l Line) int {
	n := l.Owned - l.InOtherBuilt
	if n < 0 {
		return 0
	}
	return n
}

// CanPlace reports whether this deck may change its count by add.
// A negative add removes copies. A positive add may not pass Free.
func CanPlace(l Line, add int) bool {
	if add < 0 {
		return l.InThis+add >= 0
	}
	return l.InThis+add <= Free(l)
}

package importmd

import "testing"

func TestParseDeckLines(t *testing.T) {
	rows := ParseDeck("- 1 Item Shopkeep **[R]** `{1}{R}` — note\n- 11 Swamp **[B]**\n- not a card\n")
	if len(rows) != 2 || rows[0].Name != "Item Shopkeep" || rows[0].Qty != 1 || rows[1].Name != "Swamp" || rows[1].Qty != 11 {
		t.Fatalf("%+v", rows)
	}
}

func TestParseTailedAndBareLines(t *testing.T) {
	rows, err := Parse("- Hexing Squelcher ×1 — rare · nonfoil · $29.59 · ecl #317 · 2026-09-23\n- Lightning Strike ×1\n- Basic Island ×25\n")
	if err != nil || len(rows) != 3 {
		t.Fatal(err, len(rows))
	}
	if rows[0].Set != "ecl" || rows[0].Number != "317" || rows[0].Foil || rows[0].Price != "29.59" {
		t.Fatalf("%+v", rows[0])
	}
	if rows[1].Name != "Lightning Strike" || rows[1].Set != "" {
		t.Fatalf("%+v", rows[1])
	}
	if rows[2].Set != "basic" || rows[2].Number != "island" || rows[2].Qty != 25 {
		t.Fatalf("%+v", rows[2])
	}
}

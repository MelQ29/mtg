package deckrules

import "testing"

func TestSingletonAllowsBasicsAndAnyNumber(t *testing.T) {
	have := []Card{{Name: "Swamp", TypeLine: "Basic Land — Swamp", Qty: 4}}
	if err := CheckAdd("commander", Card{}, false, have, Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"}, 1); err != nil {
		t.Fatal(err)
	}
	spell := Card{Name: "Relentless Rats", TypeLine: "Creature — Rat", Oracle: "A deck can have any number of cards named Relentless Rats."}
	if err := CheckAdd("commander", Card{}, false, []Card{spell}, spell, 1); err != nil {
		t.Fatal(err)
	}
	one := Card{Name: "Overkill", TypeLine: "Instant", Qty: 1}
	if err := CheckAdd("commander", Card{}, false, []Card{one}, one, 1); err == nil {
		t.Fatal("expected singleton")
	}
	if err := CheckAdd("kitchen", Card{}, false, []Card{one}, one, 1); err != nil {
		t.Fatal(err)
	}
}

func TestCommanderColorsAndForest(t *testing.T) {
	leader := Card{Name: "Abigale", TypeLine: "Legendary Creature — Bird Bard", Colors: "W"}
	if err := CheckAdd("commander", leader, true, nil, Card{Name: "Overkill", TypeLine: "Instant", Colors: "B"}, 1); err == nil {
		t.Fatal("black spell in a white deck")
	}
	forest := Card{Name: "Forest", TypeLine: "Basic Land — Forest", Colors: "G"}
	if err := CheckAdd("commander", leader, true, nil, forest, 1); err == nil {
		t.Fatal("forest in a white deck")
	}
	if err := CheckAdd("commander", leader, true, nil, Card{Name: "Plains", TypeLine: "Basic Land — Plains"}, 1); err != nil {
		t.Fatal(err)
	}
	if !CanBeCommander(leader.TypeLine, "") {
		t.Fatal("legendary creature")
	}
	if CanBeCommander("Instant", "") {
		t.Fatal("instant")
	}
	if !CanBeCommander("Legendary Planeswalker — Ajani", "Ajani can be your commander.") {
		t.Fatal("planeswalker ability")
	}
}

func TestSizeAndOutsideCommander(t *testing.T) {
	var many []Card
	for i := 0; i < 100; i++ {
		many = append(many, Card{Name: string(rune('A' + i%20)), Qty: 1})
	}
	// 100 already, names collide but we only check size after counting qty
	filled := []Card{{Name: "Soldier", TypeLine: "Creature — Soldier", Qty: 100}}
	if err := CheckAdd("commander", Card{}, false, filled, Card{Name: "Scout", TypeLine: "Creature — Scout"}, 1); err == nil {
		t.Fatal("expected 100-card cap")
	}
	leader := Card{Name: "Abigale", TypeLine: "Legendary Creature — Bird Bard", Colors: "W"}
	if err := CheckCommander(leader, []Card{{Name: "Overkill", Colors: "B", TypeLine: "Instant"}}); err == nil {
		t.Fatal("expected outside colors")
	}
}

package parser

import (
	"testing"
)

func TestParseStructuredSimple(t *testing.T) {
	p := Parse("look", "goblin", ClassStructured)
	if p.Verb != "look" {
		t.Errorf("expected verb 'look', got '%s'", p.Verb)
	}
	if p.Target.Noun != "goblin" {
		t.Errorf("expected target noun 'goblin', got '%s'", p.Target.Noun)
	}
	if !p.Instrument.IsEmpty() {
		t.Errorf("expected no instrument, got '%s'", p.Instrument.Noun)
	}
}

func TestParseStripFillers(t *testing.T) {
	p := Parse("look", "at the goblin", ClassStructured)
	if p.Target.Noun != "goblin" {
		t.Errorf("expected target 'goblin' after stripping fillers, got '%s'", p.Target.Noun)
	}
}

func TestParseAdjective(t *testing.T) {
	p := Parse("get", "red sword", ClassStructured)
	if p.Target.Noun != "sword" {
		t.Errorf("expected noun 'sword', got '%s'", p.Target.Noun)
	}
	if len(p.Target.Adjectives) != 1 || p.Target.Adjectives[0] != "red" {
		t.Errorf("expected adjectives ['red'], got %v", p.Target.Adjectives)
	}
}

func TestParsePreposition(t *testing.T) {
	p := Parse("unlock", "door with key", ClassStructured)
	if p.Target.Noun != "door" {
		t.Errorf("expected target 'door', got '%s'", p.Target.Noun)
	}
	if p.Preposition != "with" {
		t.Errorf("expected preposition 'with', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "key" {
		t.Errorf("expected instrument 'key', got '%s'", p.Instrument.Noun)
	}
}

func TestParseFullSentence(t *testing.T) {
	p := Parse("unlock", "the red door with the green key", ClassStructured)
	if p.Target.Noun != "door" {
		t.Errorf("expected target 'door', got '%s'", p.Target.Noun)
	}
	if len(p.Target.Adjectives) != 1 || p.Target.Adjectives[0] != "red" {
		t.Errorf("expected target adj ['red'], got %v", p.Target.Adjectives)
	}
	if p.Preposition != "with" {
		t.Errorf("expected preposition 'with', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "key" {
		t.Errorf("expected instrument 'key', got '%s'", p.Instrument.Noun)
	}
	if len(p.Instrument.Adjectives) != 1 || p.Instrument.Adjectives[0] != "green" {
		t.Errorf("expected instrument adj ['green'], got %v", p.Instrument.Adjectives)
	}
}

func TestParseDisambiguation(t *testing.T) {
	p := Parse("look", "sword#2", ClassStructured)
	if p.Target.Noun != "sword" {
		t.Errorf("expected noun 'sword', got '%s'", p.Target.Noun)
	}
	if p.Target.Index != 2 {
		t.Errorf("expected index 2, got %d", p.Target.Index)
	}
}

func TestParseSpecialPrefix(t *testing.T) {
	p := Parse("attack", "@123", ClassStructured)
	if p.Target.Noun != "@123" {
		t.Errorf("expected noun '@123', got '%s'", p.Target.Noun)
	}

	p = Parse("attack", "#456", ClassStructured)
	if p.Target.Noun != "#456" {
		t.Errorf("expected noun '#456', got '%s'", p.Target.Noun)
	}
}

func TestParseFreeform(t *testing.T) {
	p := Parse("say", "hello the world about things", ClassFreeform)
	if p.Rest != "hello the world about things" {
		t.Errorf("freeform rest should be unmodified, got '%s'", p.Rest)
	}
	if p.Extra != "hello the world about things" {
		t.Errorf("freeform extra should match, got '%s'", p.Extra)
	}
}

func TestParseEmptyRest(t *testing.T) {
	p := Parse("look", "", ClassStructured)
	if !p.Target.IsEmpty() {
		t.Errorf("expected empty target, got '%s'", p.Target.Noun)
	}
	if p.Rest != "" {
		t.Errorf("expected empty rest, got '%s'", p.Rest)
	}
}

func TestParseGetFromContainer(t *testing.T) {
	p := Parse("get", "sword from chest", ClassStructured)
	if p.Target.Noun != "sword" {
		t.Errorf("expected target 'sword', got '%s'", p.Target.Noun)
	}
	if p.Preposition != "from" {
		t.Errorf("expected preposition 'from', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "chest" {
		t.Errorf("expected instrument 'chest', got '%s'", p.Instrument.Noun)
	}
}

func TestParseGiveToPlayer(t *testing.T) {
	p := Parse("give", "gold to merchant", ClassStructured)
	if p.Target.Noun != "gold" {
		t.Errorf("expected target 'gold', got '%s'", p.Target.Noun)
	}
	if p.Preposition != "to" {
		t.Errorf("expected preposition 'to', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "merchant" {
		t.Errorf("expected instrument 'merchant', got '%s'", p.Instrument.Noun)
	}
}

func TestParseOnlyFillers(t *testing.T) {
	p := Parse("look", "at the", ClassStructured)
	if p.Target.Noun != "" {
		t.Errorf("expected empty target after stripping all fillers, got '%s'", p.Target.Noun)
	}
}

func TestMultiWordCollapse(t *testing.T) {
	verb, rest, ok := TryMultiWordCollapse("pick up sword")
	if !ok {
		t.Error("expected multi-word collapse")
	}
	if verb != "get" {
		t.Errorf("expected verb 'get', got '%s'", verb)
	}
	if rest != "sword" {
		t.Errorf("expected rest 'sword', got '%s'", rest)
	}
}

func TestMultiWordCollapseNoMatch(t *testing.T) {
	_, _, ok := TryMultiWordCollapse("attack goblin")
	if ok {
		t.Error("expected no multi-word collapse")
	}
}

func TestPrefixMatch(t *testing.T) {
	commands := []string{"teleport", "tell", "throw", "attack", "ask"}

	match := TryPrefixMatch("tel", commands)
	if match != "" {
		t.Errorf("'tel' should be ambiguous (teleport, tell), got '%s'", match)
	}

	match = TryPrefixMatch("tele", commands)
	if match != "teleport" {
		t.Errorf("expected 'teleport', got '%s'", match)
	}

	match = TryPrefixMatch("att", commands)
	if match != "attack" {
		t.Errorf("expected 'attack', got '%s'", match)
	}

	match = TryPrefixMatch("x", commands)
	if match != "" {
		t.Errorf("single char should not match, got '%s'", match)
	}
}

func TestParsePassthrough(t *testing.T) {
	p := Parse("server", "reload items", ClassPassthrough)
	if p.Rest != "reload items" {
		t.Errorf("passthrough rest should be unmodified, got '%s'", p.Rest)
	}
}

func TestParsePreservesOrigRest(t *testing.T) {
	p := Parse("look", "at the red goblin", ClassStructured)
	if p.OrigRest != "at the red goblin" {
		t.Errorf("expected OrigRest 'at the red goblin', got '%s'", p.OrigRest)
	}
	// But Rest should be cleaned
	if p.Rest == p.OrigRest {
		t.Error("Rest should differ from OrigRest after filler stripping")
	}
}

func TestNounPhraseString(t *testing.T) {
	np := NounPhrase{Adjectives: []string{"red"}, Noun: "sword", Index: 2}
	s := FormatNounPhrase(np)
	if s != "red sword#2" {
		t.Errorf("expected 'red sword#2', got '%s'", s)
	}
}

func TestParseAll(t *testing.T) {
	p := Parse("get", "all", ClassStructured)
	if !p.Target.All {
		t.Error("expected All=true")
	}
	if p.Target.Noun != "all" {
		t.Errorf("expected noun 'all', got '%s'", p.Target.Noun)
	}
}

func TestParseQuantity(t *testing.T) {
	p := Parse("get", "5 gold", ClassStructured)
	if p.Target.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", p.Target.Quantity)
	}
	if p.Target.Noun != "gold" {
		t.Errorf("expected noun 'gold', got '%s'", p.Target.Noun)
	}
}

func TestParseQuantityFromContainer(t *testing.T) {
	p := Parse("get", "3 arrows from quiver", ClassStructured)
	if p.Target.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", p.Target.Quantity)
	}
	if p.Target.Noun != "arrows" {
		t.Errorf("expected noun 'arrows', got '%s'", p.Target.Noun)
	}
	if p.Preposition != "from" {
		t.Errorf("expected preposition 'from', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "quiver" {
		t.Errorf("expected instrument 'quiver', got '%s'", p.Instrument.Noun)
	}
}

func TestParseWildcard(t *testing.T) {
	p := Parse("attack", "*goblin", ClassStructured)
	if p.Target.Noun != "*goblin" {
		t.Errorf("expected noun '*goblin', got '%s'", p.Target.Noun)
	}
}

func TestParseAllFromContainer(t *testing.T) {
	p := Parse("get", "all from chest", ClassStructured)
	if !p.Target.All {
		t.Error("expected All=true")
	}
	if p.Preposition != "from" {
		t.Errorf("expected preposition 'from', got '%s'", p.Preposition)
	}
	if p.Instrument.Noun != "chest" {
		t.Errorf("expected instrument 'chest', got '%s'", p.Instrument.Noun)
	}
}

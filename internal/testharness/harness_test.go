package testharness

import (
	"testing"
)

func TestHarnessCreation(t *testing.T) {
	h := New(t)

	if h.User == nil {
		t.Fatal("expected user to be created")
	}
	if h.Room == nil {
		t.Fatal("expected room to be created")
	}
	if h.User.Character.Gold != 1000 {
		t.Errorf("expected 1000 gold, got %d", h.User.Character.Gold)
	}
	if h.User.Character.Level != 10 {
		t.Errorf("expected level 10, got %d", h.User.Character.Level)
	}
}

func TestLookCommand(t *testing.T) {
	h := New(t)

	handled, err := h.Run("look")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected look to be handled")
	}
	// Look should produce output mentioning the room title or description
	if len(h.Output()) == 0 {
		t.Error("expected some output from look")
	}
}

func TestLookWithFillers(t *testing.T) {
	h := New(t)

	h.Room.Nouns["sign"] = "A test sign."

	handled, _ := h.Run("look at the sign")
	if !handled {
		t.Error("expected look to be handled")
	}
	h.ExpectOutput("sign")
}

func TestPrefixMatching(t *testing.T) {
	h := New(t)

	// "loo" should prefix-match to "look"
	handled, _ := h.Run("loo")
	if !handled {
		t.Error("expected 'loo' to prefix-match 'look'")
	}
	if len(h.Output()) == 0 {
		t.Error("expected output from look via prefix match")
	}
}

func TestMultiWordCollapse(t *testing.T) {
	h := New(t)

	// "pick up" should collapse to "get"
	handled, _ := h.Run("pick up")
	if !handled {
		t.Error("expected 'pick up' to be handled as 'get'")
	}
	h.ExpectOutput("Get what")
}

func TestGoldTracking(t *testing.T) {
	h := New(t)

	startGold := h.UserGold()
	if startGold != 1000 {
		t.Errorf("expected 1000 starting gold, got %d", startGold)
	}
}

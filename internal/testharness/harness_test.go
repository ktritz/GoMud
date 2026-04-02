package testharness

import (
	"testing"
)

// === Harness setup tests ===

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
	if h.User.Character.Health != 100 {
		t.Errorf("expected 100 health, got %d", h.User.Character.Health)
	}
}

func TestGoldTracking(t *testing.T) {
	h := New(t)
	if h.UserGold() != 1000 {
		t.Errorf("expected 1000 starting gold, got %d", h.UserGold())
	}
}

// === Parser integration tests ===

func TestPrefixMatching(t *testing.T) {
	h := New(t)
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
	handled, _ := h.Run("pick up")
	if !handled {
		t.Error("expected 'pick up' to be handled as 'get'")
	}
	h.ExpectOutput("Get what")
}

func TestAliasResolution(t *testing.T) {
	h := New(t)
	// "l" is aliased to "look"
	handled, _ := h.Run("l")
	if !handled {
		t.Error("expected 'l' alias to resolve to 'look'")
	}
	if len(h.Output()) == 0 {
		t.Error("expected output from look via alias")
	}
}

// === Look handler tests ===

func TestLookRoom(t *testing.T) {
	h := New(t)
	handled, err := h.Run("look")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected look to be handled")
	}
	if len(h.Output()) == 0 {
		t.Error("expected room description output")
	}
}

func TestLookAtNoun(t *testing.T) {
	h := New(t)
	h.Room.Nouns["sign"] = "A weathered wooden sign."

	handled, _ := h.Run("look sign")
	if !handled {
		t.Error("expected look to be handled")
	}
	h.ExpectOutput("sign")
}

func TestLookAtNounWithFillers(t *testing.T) {
	h := New(t)
	h.Room.Nouns["fountain"] = "A bubbling fountain of clear water."

	handled, _ := h.Run("look at the fountain")
	if !handled {
		t.Error("expected look to be handled")
	}
	h.ExpectOutput("fountain")
}

func TestLookAtNounExamine(t *testing.T) {
	h := New(t)
	h.Room.Nouns["sign"] = "A test sign."

	// "examine" is aliased to "look"
	handled, _ := h.Run("examine sign")
	if !handled {
		t.Error("expected examine alias to work")
	}
	h.ExpectOutput("sign")
}

func TestLookDirection(t *testing.T) {
	h := New(t)
	// Looking in a direction that has an exit
	handled, _ := h.Run("look north")
	if !handled {
		t.Error("expected look north to be handled")
	}
}

// === Get handler tests ===

func TestGetNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("get")
	if !handled {
		t.Error("expected get to be handled")
	}
	h.ExpectOutput("Get what")
}

func TestGetGoldFromFloor(t *testing.T) {
	h := New(t)
	h.Room.Gold = 50
	startGold := h.UserGold()

	handled, _ := h.Run("get gold")
	if !handled {
		t.Error("expected get gold to be handled")
	}

	if h.UserGold() <= startGold {
		t.Error("expected user gold to increase after getting gold from floor")
	}
}

func TestGetAll(t *testing.T) {
	h := New(t)
	h.Room.Gold = 25
	startGold := h.UserGold()

	handled, _ := h.Run("get all")
	if !handled {
		t.Error("expected get all to be handled")
	}

	// Should have picked up floor gold at minimum
	if h.UserGold() <= startGold {
		t.Error("expected gold increase from get all")
	}
}

func TestGrabAlias(t *testing.T) {
	h := New(t)
	// "grab" is aliased to "get"
	handled, _ := h.Run("grab")
	if !handled {
		t.Error("expected 'grab' alias to resolve to 'get'")
	}
	h.ExpectOutput("Get what")
}

func TestTakeAlias(t *testing.T) {
	h := New(t)
	// "take" is aliased to "get"
	handled, _ := h.Run("take")
	if !handled {
		t.Error("expected 'take' alias to resolve to 'get'")
	}
	h.ExpectOutput("Get what")
}

// === Give handler tests ===

func TestGiveNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("give")
	if !handled {
		t.Error("expected give to be handled")
	}
	h.ExpectOutput("Give what")
}

// === Attack handler tests ===

func TestAttackNoTarget(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("attack")
	if !handled {
		t.Error("expected attack to be handled")
	}
	h.ExpectOutput("darkness")
}

func TestAttackWithFillers(t *testing.T) {
	h := New(t)
	// "attack the goblin" — fillers should be stripped, but since there's
	// no goblin in the room it should still say "attack the darkness"
	handled, _ := h.Run("attack the goblin")
	if !handled {
		t.Error("expected attack to be handled")
	}
	h.ExpectOutput("darkness")
}

func TestKillAlias(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("kill")
	if !handled {
		t.Error("expected 'kill' alias to resolve to 'attack'")
	}
	h.ExpectOutput("darkness")
}

// === Movement tests ===

func TestMovementHandled(t *testing.T) {
	h := New(t)

	// Movement commands should be handled (recognized as exits)
	handled, _ := h.Run("north")
	if !handled {
		t.Error("expected north to be handled as an exit")
	}
	// Note: actual room change requires MoveToRoom + event processing,
	// which needs deeper harness support. This test verifies the command
	// is recognized and doesn't error.
}

// === Drop handler tests ===

func TestDropNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("drop")
	if !handled {
		t.Error("expected drop to be handled")
	}
}

// === Say handler tests (parser freeform) ===

func TestSayCommand(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("say hello world")
	if !handled {
		t.Error("expected say to be handled")
	}
	// Say should echo back what was said
	h.ExpectOutput("hello world")
}

// === Eat/Drink handler tests ===

func TestEatNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("eat")
	if !handled {
		t.Error("expected eat to be handled")
	}
}

func TestEatWithFillers(t *testing.T) {
	h := New(t)
	// "eat the potion" should strip "the" and try to find "potion"
	handled, _ := h.Run("eat the potion")
	if !handled {
		t.Error("expected eat to be handled")
	}
	// Should say something about not having it (not "eat what")
}

func TestDrinkNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("drink")
	if !handled {
		t.Error("expected drink to be handled")
	}
}

// === Equip handler tests ===

func TestEquipNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("equip")
	if !handled {
		t.Error("expected equip to be handled")
	}
}

func TestEquipWithFillers(t *testing.T) {
	h := New(t)
	// "wear the sword" should strip filler
	handled, _ := h.Run("wear the sword")
	if !handled {
		t.Error("expected wear alias to be handled")
	}
}

// === Buy handler tests ===

func TestBuyNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("buy")
	if !handled {
		t.Error("expected buy to be handled")
	}
}

// === Sell handler tests ===

func TestSellNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("sell")
	if !handled {
		t.Error("expected sell to be handled")
	}
}

// === Use handler tests ===

func TestUseNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("use")
	if !handled {
		t.Error("expected use to be handled")
	}
}

// === Stash handler tests ===

func TestStashNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("stash")
	if !handled {
		t.Error("expected stash to be handled")
	}
}

// === Read handler tests ===

func TestReadNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("read")
	if !handled {
		t.Error("expected read to be handled")
	}
}

// === Appraise handler tests ===

func TestAppraiseNoArgs(t *testing.T) {
	h := New(t)
	handled, _ := h.Run("appraise")
	if !handled {
		t.Error("expected appraise to be handled")
	}
}

func TestSayPreservesFillers(t *testing.T) {
	h := New(t)
	// Freeform commands should NOT strip fillers
	handled, _ := h.Run("say the quick brown fox")
	if !handled {
		t.Error("expected say to be handled")
	}
	h.ExpectOutput("the quick brown fox")
}

// Package testharness provides a test environment for running command handlers
// against real game objects without requiring a running game server.
package testharness

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// Harness provides a test environment with a pre-built user, room, and
// utilities for running commands and checking results.
type Harness struct {
	t    *testing.T
	User *users.UserRecord
	Room *rooms.Room

	// Adjacent rooms for movement tests
	Rooms map[int]*rooms.Room

	// Captured output from SendText calls (collected from event queue)
	output []string
}

// New creates a test harness with a fully equipped user in a fully stocked room.
func New(t *testing.T) *Harness {
	t.Helper()

	// Ensure configs are loaded (uses defaults if no file)
	configs.AddOverlayOverrides(map[string]any{
		"FilePaths.DataFiles": t.TempDir(),
	})

	// Initialize keywords for command alias resolution
	keywords.InitTestKeywords()

	h := &Harness{
		t:     t,
		Rooms: make(map[int]*rooms.Room),
	}

	// Initialize biomes so rooms can compute visibility
	rooms.InitTestBiomes()

	// Create the god-room: exits, items, nouns, gold
	h.Room = rooms.NewTestRoom(9999, "TestZone", "The Test Chamber")
	h.Room.Description = "A featureless room that contains everything needed for testing."
	h.Room.Biome = "city"
	h.Room.Gold = 500
	h.Room.Nouns = map[string]string{
		"sign":     "A test sign on the wall.",
		"fountain": "A bubbling fountain.",
	}

	// Add exits to adjacent rooms
	north := rooms.NewTestRoom(9998, "TestZone", "North Room")
	south := rooms.NewTestRoom(9997, "TestZone", "South Room")
	east := rooms.NewTestRoom(9996, "TestZone", "East Room")

	h.Room.Exits["north"] = exit.RoomExit{RoomId: 9998}
	h.Room.Exits["south"] = exit.RoomExit{RoomId: 9997}
	h.Room.Exits["east"] = exit.RoomExit{RoomId: 9996}
	north.Exits["south"] = exit.RoomExit{RoomId: 9999}
	south.Exits["north"] = exit.RoomExit{RoomId: 9999}
	east.Exits["west"] = exit.RoomExit{RoomId: 9999}

	h.Rooms[9999] = h.Room
	h.Rooms[9998] = north
	h.Rooms[9997] = south
	h.Rooms[9996] = east

	// Register all rooms
	for _, r := range h.Rooms {
		rooms.RegisterTestRoom(r)
	}

	// Create the god-user: level 10, all stats, gold, items
	h.User = users.NewUserRecord(99999, 0)
	h.User.Username = "testuser"
	h.User.Role = users.RoleAdmin
	h.User.Character.Name = "TestPlayer"
	h.User.Character.Level = 10
	h.User.Character.Health = 100
	h.User.Character.HealthMax.SetValue(100)
	h.User.Character.Mana = 50
	h.User.Character.ManaMax.SetValue(50)
	h.User.Character.Gold = 1000
	h.User.Character.Bank = 5000
	h.User.Character.RoomId = 9999
	h.User.Character.Zone = "TestZone"
	h.User.Character.ExtraLives = 3

	// Train all stats
	h.User.Character.Stats.Strength.Training = 10
	h.User.Character.Stats.Speed.Training = 10
	h.User.Character.Stats.Smarts.Training = 10
	h.User.Character.Stats.Vitality.Training = 10
	h.User.Character.Stats.Mysticism.Training = 10
	h.User.Character.Stats.Perception.Training = 10

	// Register user
	users.RegisterTestUser(h.User)

	// Cleanup on test end
	t.Cleanup(func() {
		users.DeregisterTestUser(h.User.UserId)
		for id := range h.Rooms {
			rooms.DeregisterTestRoom(id)
		}
		h.DrainEvents()
	})

	return h
}

// Run executes a command string as if the user typed it.
// Returns (handled, error).
func (h *Harness) Run(input string) (bool, error) {
	h.t.Helper()
	h.output = nil // Clear previous output

	// Split command and rest
	command := input
	rest := ""
	if idx := strings.Index(input, " "); idx != -1 {
		command = strings.ToLower(input[:idx])
		rest = input[idx+1:]
	}

	// Multi-word collapse
	if verb, r, ok := parser.TryMultiWordCollapse(input); ok {
		command = verb
		rest = r
	}

	// Run through TryCommand
	handled, err := usercommands.TryCommand(command, rest, h.User.UserId, 0)

	// Collect output from event queue
	h.collectOutput()

	return handled, err
}

// Output returns all text sent to the user via SendText during the last Run().
func (h *Harness) Output() string {
	return strings.Join(h.output, "")
}

// OutputLines returns output split by newlines, empty lines removed.
func (h *Harness) OutputLines() []string {
	var lines []string
	for _, line := range strings.Split(h.Output(), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// OutputContains checks if the output contains the given substring.
func (h *Harness) OutputContains(substr string) bool {
	return strings.Contains(strings.ToLower(h.Output()), strings.ToLower(substr))
}

// ExpectOutput asserts the output contains the given substring.
func (h *Harness) ExpectOutput(substr string) {
	h.t.Helper()
	if !h.OutputContains(substr) {
		h.t.Errorf("expected output to contain %q, got: %s", substr, h.Output())
	}
}

// ExpectNoOutput asserts the output does NOT contain the given substring.
func (h *Harness) ExpectNoOutput(substr string) {
	h.t.Helper()
	if h.OutputContains(substr) {
		h.t.Errorf("expected output NOT to contain %q, got: %s", substr, h.Output())
	}
}

// GiveItem adds an item to the user's backpack.
func (h *Harness) GiveItem(itemId int) {
	item := items.New(itemId)
	h.User.Character.StoreItem(item)
}

// PlaceItemInRoom adds an item to the room floor.
func (h *Harness) PlaceItemInRoom(itemId int) {
	item := items.New(itemId)
	h.Room.AddItem(item, false)
}

// DrainEvents pulls all events from the queue, discarding them.
// Call between tests to prevent event leakage.
func (h *Harness) DrainEvents() {
	events.ProcessEvents()
}

// collectOutput drains Message events from the queue and captures text
// directed at our test user.
func (h *Harness) collectOutput() {
	// Process any pending events to trigger message delivery
	// For now, just drain the queue and capture Message events
	for events.ProcessSingle(func(e events.Event) events.ListenerReturn {
		if msg, ok := e.(events.Message); ok {
			if msg.UserId == h.User.UserId {
				h.output = append(h.output, msg.Text)
			}
		}
		// Don't run listeners — just capture messages
		return events.Continue
	}) {
	}
}

// UserGold returns the user's current gold.
func (h *Harness) UserGold() int {
	return h.User.Character.Gold
}

// UserHealth returns the user's current health.
func (h *Harness) UserHealth() int {
	return h.User.Character.Health
}

// BackpackContains checks if the user has an item with the given name in their backpack.
func (h *Harness) BackpackContains(name string) bool {
	_, found := h.User.Character.FindInBackpack(name)
	return found
}

// RoomHasItem checks if the room floor has an item with the given name.
func (h *Harness) RoomHasItem(name string) bool {
	for _, item := range h.Room.Items {
		if strings.Contains(strings.ToLower(item.Name()), strings.ToLower(name)) {
			return true
		}
	}
	return false
}

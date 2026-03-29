package actions

import (
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// ActionContext provides a unified interface for game actions that work
// for both players and mobs. The caller constructs this from either a
// UserRecord or Mob, then passes it to shared action functions.
type ActionContext struct {
	// Identity
	UserId        int // 0 if mob
	MobInstanceId int // 0 if user

	// Core references
	Character *characters.Character
	Room      *rooms.Room

	// Messaging callbacks — set by the caller to handle user vs mob differences
	SendToActor func(string)           // Send text to the acting entity (user gets it, mobs ignore)
	SendToRoom  func(string, ...int)   // Send text to the room (with optional exclude IDs)

	// Display
	ActorName string // The character's display name
	ActorTag  string // "username" or "mobname" — used in ANSI tags
}

// IsUser returns true if this context represents a player.
func (ctx *ActionContext) IsUser() bool {
	return ctx.UserId > 0
}

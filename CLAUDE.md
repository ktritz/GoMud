# GoMud - AI Agent Guide

## Build & Test
```bash
go build ./...    # Build everything
go test ./...     # Run all tests
go vet ./...      # Static analysis (some pre-existing warnings in memory.go files are expected)
```

## Architecture Overview

**Foundation** (no/few internal deps): `configs`, `util`, `mudlog`, `events`, `stats`, `statmods`, `skills`
**Domain models**: `characters`, `items`, `buffs`, `races`, `spells`, `quests`, `pets`, `mobs`
**Infrastructure**: `rooms` (entity container + manager), `users` (session + persistence)
**Shared logic**: `actions` (unified mob/player actions via `ActionContext`), `targeting` (target resolution), `locksmith` (lock/key operations), `combat/resolution.go` (combat helpers)
**Orchestration**: `hooks` (event listeners), `usercommands`, `mobcommands`, `scripting`
**Extensions**: `modules/` (pluggable features: gmcp, auctions, follow, leaderboards, time, cleanup, webhelp)

### Key patterns
- **Event-driven**: `events.AddToQueue()` emits events, `hooks/` registers listeners via `events.RegisterListener()`
- **Command signature**: `func CmdName(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error)`
- **Stat accessors**: Use `Stats.GetValueAdj("StatName")` for direct stat reads. Use `Stats.ActionValueAdj("ActionName")` for game-logic stat reads (resolves action→stat via config). See `config.statistics.go` for the action mapping.
- **Template functions**: Registered in `internal/templates/templatesfunctions.go`, used in `_datafiles/**/templates/`
- **Modules**: Register via `plugins.RegisterPlugin()` in `init()`. Can add commands, listen for events, serve web pages. See `modules/README.md`.

### Config
- Single config file: `_datafiles/config.yaml` → parsed into `internal/configs/` structs
- Game data: `_datafiles/world/default/` (rooms, items, mobs, buffs, spells, races, templates)
- Config types use `ConfigInt`/`ConfigFloat` wrappers (see `config_types.go`)

## Known Technical Debt

### Mob/Player duplication
Many operations are implemented twice — once for mobs, once for players. When modifying combat, equip, spell casting, or skill logic, check both `usercommands/` and `mobcommands/` for parallel implementations that need the same change. Key duplicated files:
- `attack.go`, `equip.go`, `cast.go`, `backstab.go`, `aid.go`, `get.go`, `drop.go`
- `hooks/NewRound_DoCombat.go` handles player combat (~lines 41-650) and mob combat (~650+) with similar logic

### Oversized files
These files have multiple responsibilities and are candidates for splitting:
- `characters/character.go` (2040 lines) — combat calcs, presentation, key management, shops, cooldowns
- `rooms/rooms.go` (2311 lines) — data container, mob/item lifecycle, buff application, search/find methods
- `hooks/NewRound_DoCombat.go` (1107 lines) — all combat round processing
- `usercommands/admin.room.go` (1574 lines) — all room admin operations
- `scripting/actor_func.go` (857 lines) — 90 script API functions

### Global state
`mobs`, `rooms`, `users` packages use package-level maps as registries. These are not synchronized and prevent parallel testing.

### Business logic in command handlers
`usercommands/` files often contain orchestration logic (targeting, combat resolution, room movement + scripting events) that should be in domain packages.

## Conventions

- **Stat access in game logic**: Always use `Stats.ActionValueAdj("ActionName")` — never hardcode stat names like `"Strength"` in game mechanics. Action→stat mappings live in `config.yaml` under `Statistics.StatActions`.
- **StatInfo fields**: Use accessor methods (`GetValue()`, `SetValue()`, `GetMods()`, etc.) — don't access `StatInfo` struct fields directly.
- **Templates**: `.template` files use Go `text/template` syntax. Use `charStatValue`, `charStatValueAdj` etc. helper functions rather than direct field access.
- **Shared actions**: When adding a new command that works for both mobs and players, put the shared logic in `internal/actions/` using `ActionContext`, then write thin wrappers in `usercommands/` and `mobcommands/`. See `equip.go` and `drop.go` as examples.
- **Target resolution**: Use `targeting.FindAutoTarget()` and `targeting.FindRandomTarget()` instead of inline target-finding loops. See `attack.go` and `backstab.go`.
- **Lock/key operations**: Use `locksmith.FindKey()` and `locksmith.ConsumeBackpackKey()` instead of inline key-checking. See `lock.go` and `go.go`.
- **Mob commands** mirror user commands — changes to shared logic often need updates in both. Prefer extracting to `actions/` package when possible.
- **Events**: Prefer `events.AddToQueue()` over direct function calls for cross-package communication.

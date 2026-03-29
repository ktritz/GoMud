# GoMud Refactoring Plan

## Problem Statement

The codebase (77,583 lines across 398 Go files) has grown organically and suffers from:
- **Oversized files**: 5 files over 1,000 lines, with `rooms.go` at 2,311 and `character.go` at 2,040
- **Mob/player duplication**: Nearly identical logic implemented twice across `usercommands/` and `mobcommands/` (~15 command pairs)
- **Mixed concerns**: Domain models (Character, Room) contain presentation, persistence, and orchestration logic
- **Business logic in command handlers**: `usercommands/` imports 27 packages and directly orchestrates combat, spells, items
- **Global mutable state**: Package-level maps in `mobs`, `rooms`, `users` with no synchronization

## Goals

- Soft limit of **400-500 lines per file**
- Eliminate mob/player code duplication via shared logic
- Separate concerns: data, behavior, presentation, orchestration
- Improve maintainability for AI agent workflows (stop/resume, parallel work)
- No behavioral changes — all refactoring is structural

## Progress Tracker

### Phase 4: Split `rooms/rooms.go` (2,311 lines -> ~5 files)
Split within same package. No import changes anywhere.

- [x] Create `room_messaging.go` (104 lines) — SendText, SendTextCommunication, PlaySound, SendTextToExits
- [x] Create `room_queries.go` (939 lines) — FindByName, findPlayerByName, findMobByName, FindExitByName, FindExitTo, FindContainerByName, FindNoun, FindCorpse, FindOnFloor, GetMobs, GetPlayers, MobCt, PlayerCt, GetAllFloorItems, GetRandomExit, IsCalm, ArePlayersAttacking, AreMobsAttacking, isInRoom, findMobExit, findUserExit, FindTemporaryExitByUserId, GetExitInfo
- [x] Create `room_visitors.go` (105 lines) — MarkVisited, Visitors, HasVisited, HasRecentVisitors, PruneVisitors
- [x] Create `room_mutations.go` (526 lines) — AddMob, RemoveMob, AddPlayer, RemovePlayer, AddItem, RemoveItem, SetExitLock, AddCorpse, RemoveCorpse, UpdateCorpses, AddSign, GetPublicSigns, GetPrivateSigns, AddTemporaryExit, RemoveTemporaryExit, PruneTemporaryExits, PruneSigns, RepeatSpawnItem, SpawnTempContainer, ApplyBuffIdToPlayers, ApplyBuffIdToMobs, ApplyBuffIdToNativeMobs
- [x] Clean up `rooms.go` (678 lines) — struct def, constants, Prepare, CleanupMobSpawns, RoundTick, Validate, Id, Filename, Filepath, GetBiome, ActiveMutators, IsPvp, CanPvp, data store methods, script methods, description methods
- [x] Verify: `go build ./... && go test ./...`

### Phase 5: Split `characters/character.go` (2,040 lines -> ~6 files)
Split within same package. Existing sibling files (aggro.go, cooldowns.go, worn.go) show the pattern.

- [x] Create `character_inventory.go` (157 lines) — StoreItem, RemoveItem, UpdateItem, UseItem, FindInBackpack, FindOnBody, GetRandomItem, GetAllBackpackItems, HandsRequired, CarryCapacity
- [x] Create `character_skills.go` (246 lines) — GetSkills, SetSkill, TrainSkill, GetSkillLevel, GetSkillLevelCost, HasSpell, LearnSpell, DisableSpell, EnableSpell, TrackSpellCast, GetSpells, GetBaseCastSuccessChance, AutoTrain, CanDualWield, GetMaxCharmedCreatures, GetMemoryCapacity, GetMapSprawlCapacity
- [x] Create `character_equipment.go` (429 lines) — Wear, RemoveFromBody, GetAllWornItems, GetGearValue, Uncurse, reapplyPermabuffs, SetPermaBuffs, GetDefaultDiceRoll, GetDefense
- [x] Create `character_combat.go` (151 lines) — ApplyHealthChange, ApplyManaChange, TrackPlayerDamage, Heal, HealthPerRound, ManaPerRound, GetHealthAppearance, MovementCost, BarterPrice
- [x] Create `character_quests.go` (83 lines) — IsQuestDone, HasQuest, GetQuestProgress, GiveQuestToken, ClearQuestToken, RememberRoom
- [x] Clean up `character.go` (1013 lines) — struct def, New(), Validate, RecalculateStats, LevelUp, GrantXP, XPTNL/XPTL/XPTNLActual, StatMod, buff methods, charm methods, misc data/settings/timers/keys, description/name methods
- [x] Verify: `go build ./... && go test ./...`

### Phase 2: Extract combat resolution helpers from `hooks/NewRound_DoCombat.go` (1,107 lines)
Move repeated combat blocks into `internal/combat/` package.

- [x] Create `combat/resolution.go` (159 lines) — SendRoundMessages, HandleEquipmentBreak, TriggerCharmedMobRetaliation
- [ ] Create `combat/flee.go` — extract flee logic (player flee + mob flee share structure) [deferred — flee is player-only currently]
- [ ] Create `combat/spellresolution.go` — extract spell combat handling (deferred — moderate risk, spell logic differs between player/mob)
- [x] Refactor DoCombat to use helpers — 1,107 -> 908 lines, eliminated 3x equipment break duplication, 2x charmed mob retaliation, 4x message dispatch
- [x] Verify: `go build ./... && go test ./...`

### Phase 6: Split `usercommands/admin.room.go` (1,574 lines)
Split by subcommand. Main function becomes thin router.

- [x] Create `admin.room.containers.go` (615 lines) — container editing + helper
- [x] Create `admin.room.exits.go` (441 lines) — exit editing
- [x] Create `admin.room.mutators.go` (141 lines) — mutator editing
- [x] Keep `admin.room.go` (420 lines) — dispatcher + inline noun/copy/info/set/exit subcommands
- [x] Verify: `go build ./... && go test ./...`

### Phase 7: Split `scripting/actor_func.go` (857 lines)
Split by domain. 90+ methods grouped into focused files.

- [x] Create `actor_func_inventory.go` (105 lines) — HasItemId, GetBackpackItems, GiveItem, TakeItem, etc.
- [x] Create `actor_func_combat.go` (131 lines) — AddHealth, SetHealth, GetHealth*, AddMana, GetMana*, IsAggro, buffs, etc.
- [x] Create `actor_func_progression.go` (164 lines) — GrantXP, TrainSkill, GetSkillLevel, LearnSpell, etc.
- [x] Create `actor_func_social.go` (288 lines) — SendText, Command, GetPartyMembers, charms, timers, etc.
- [x] Keep `actor_func.go` (199 lines) — struct def, constructors, GetStat, core identity methods
- [x] Verify: `go build ./... && go test ./...`

### Phase 1: Extract shared targeting logic (new package)
Deduplicate target-finding across user and mob commands.

- [x] Create `internal/targeting/targeting.go` (114 lines) — FindAutoTarget, FindRandomTarget
- [x] Update `usercommands/attack.go` (273->199) and `mobcommands/attack.go` (157->99)
- [x] Update `usercommands/skill.skulduggery.backstab.go` (150->123) and `mobcommands/backstab.go` (80->56)
- [ ] Update spell targeting in `usercommands/skill.cast.go` and `mobcommands/cast.go` [deferred — spell targeting is more specialized]
- [x] Verify: `go build ./... && go test ./...`

### Phase 3: Extract lock/key operations (new package)
Deduplicate lock/unlock/key logic across go.go, lock.go, unlock.go.

- [x] Create `internal/locksmith/locksmith.go` (55 lines) — BuildLockId, FindKey, ConsumeBackpackKey
- [x] Unified `lock.go` + `unlock.go` into single `lock.go` (156 lines) via shared `lockOrUnlock`
- [x] Updated `go.go` (400->366) to use locksmith package
- [x] Verify: `go build ./... && go test ./...`

### Phase 8: Unified action commands (future — depends on Phases 1-3)
Create `internal/actions/` package with shared mob/player command logic.

- [x] Design ActionContext struct — `internal/actions/context.go` (32 lines)
- [x] Proof of concept: unify equip.go — `internal/actions/equip.go` (77 lines), thin wrappers in user (54) and mob (59)
- [x] Unify drop.go — `internal/actions/drop.go` (56 lines), user (76) and mob (60) wrappers
- [ ] Unify attack.go — complex due to party logic differences
- [ ] Unify remaining command pairs: give, shoot are good candidates; get, cast are too divergent for simple unification
- [x] Verify: `go build ./... && go test ./...`

## Execution Notes

- Phases 4, 5, 6, 7 are **same-package file splits** — safest, no external import changes
- Phases 1, 3 create **new packages** — moderate risk, changes import graphs
- Phase 2 **adds to existing combat package** — moderate risk, touches hot code path
- Phase 8 is **highest risk, highest reward** — do last, one command pair at a time
- After each phase, run `go build ./... && go test ./...` before proceeding
- Each phase should be a separate commit

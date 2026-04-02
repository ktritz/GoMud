# Production Audit

## Scope

This file records the ongoing production-readiness audit in segments.

## Segment 1: Runtime, Persistence, Config, Connections, Web, Plugins

Date: 2026-03-29

### Findings

1. `Plugin.ReadIntoStruct` silently drops deserialization failures.
   - File: `internal/plugins/plugins.go`
   - Problem: `ReadIntoStruct()` returns `nil` when `yaml.Unmarshal()` fails.
   - Impact: Corrupt plugin state can be treated as a successful load.
   - Example call sites: `modules/auctions/auctions.go`, `modules/leaderboards/leaderboards.go`.

2. Connection lifecycle management has correctness bugs.
   - File: `internal/connections/connections.go`
   - Problems:
   - `Kick()` closes a connection but does not delete it from `netConnections`.
   - `GetAllConnectionIds()` allocates `len(netConnections)` and then appends, producing zero-value IDs in the returned slice.
   - Impact: Stale connections, incorrect cleanup, confusing operational behavior.

3. Global connection lock is held during network writes.
   - File: `internal/connections/connections.go`
   - Problem: `Broadcast()` and `SendTo()` call `cd.Write()` while the registry mutex is held.
   - Impact: A slow client can stall unrelated sends and registry mutations.

4. Config overlay mutation is not concurrency-safe.
   - File: `internal/configs/configs.go`
   - Problem: `AddOverlayOverrides()` takes an `RLock` while mutating shared maps and config state.
   - Impact: Race conditions during plugin load and config reload flows.

5. `SaveAllFlatFiles()` is panic-driven and not resilient.
   - File: `internal/fileloader/fileloader.go`
   - Problems:
   - Worker goroutines `panic()` on marshal/write/rename failures.
   - Parent directories are not created in the bulk save path.
   - Impact: Recoverable save-path failures can crash the server.

6. Websocket and template-serving boundaries are not production-safe.
   - File: `internal/web/web.go`
   - Problems:
   - Websocket `CheckOrigin` returns `true` for all requests.
   - Template parse errors write `http.Error()` but do not return immediately.
   - Impact: Cross-origin websocket abuse risk and malformed/double HTTP responses.

7. Event dispatch runs while holding the listener registry lock.
   - File: `internal/events/listeners.go`
   - Problem: `DoListeners()` executes callbacks under the listener lock.
   - Impact: Dispatch is coupled to registration, and listener mutation patterns become deadlock-prone.

8. Debug-only and panic-driven bootstrap code remains in production paths.
   - Files: `modules/gmcp/gmcp.go`, multiple module `init()` functions
   - Problems:
   - GMCP dispatch still contains `TODO: REMOVE` debug output branches.
   - Multiple modules call `panic(err)` during `init()` on setup failures.
   - Impact: Optional-module faults can crash the whole process.

### Structural Concerns

1. `main.go` and `world.go` are oversized orchestration files.
   - `main.go` owns startup, server boot, plugin integration, shutdown, and telnet session bootstrap.
   - `world.go` owns broad event/input orchestration.
   - Recommendation: Extract session runtime, bootstrap, and worker coordination into smaller units.

2. `internal/web/web.go` mixes too many concerns.
   - Routing, template composition, TLS/HTTP lifecycle, plugin integration, and MUD locking all live together.
   - Recommendation: Split router construction, page rendering, websocket serving, and server lifecycle.

### Immediate Refactor Priority

1. Fix `ReadIntoStruct()` and make all persistence loads check errors.
2. Repair connection lifecycle bugs and stop holding the global connection lock during writes.
3. Replace panic-based bulk save behavior with error propagation and structured failure handling.
4. Tighten websocket origin policy and clean up HTTP error-return paths.
5. Remove debug-only branches from hot paths and replace module `panic(err)` bootstrap with explicit load failures.

## Segment 2: Command Layer, Room Admin, Entity Managers, User Persistence

Date: 2026-03-29

### Findings

1. Several `room` admin mutations do not persist to disk.
   - File: `internal/usercommands/admin.room.go`
   - Problems:
   - Deleting a noun mutates `room.Nouns` but does not call `rooms.SaveRoomTemplate()`.
   - Deleting or renaming an exit mutates `room.Exits` but does not save.
   - Toggling mutators via `room set mutator` returns without saving.
   - Setting `biome` mutates the room in memory but does not save.
   - Impact: Admin tooling appears to succeed but changes are lost on reload.

2. User persistence still swallows load/search failures.
   - Files: `internal/users/users.go`
   - Problems:
   - `LoadUser()` logs YAML unmarshal errors and still returns a partially loaded user.
   - `SearchOfflineUsers()` ignores the return value from `filepath.Walk()`, so unreadable/corrupt files and early-stop conditions are discarded.
   - Impact: Corrupt user records can silently degrade behavior, and offline searches are not reliable under filesystem or data errors.

3. Role permission checks emit info-level logs on the hot path.
   - File: `internal/users/userrecord.go`
   - Problem: `HasRolePermission()` logs each permission check and each comparison through `mudlog.Info()`.
   - Impact: Unnecessary log volume and avoidable work on a path hit by virtually every command.

4. Command dispatch is centralized in a single stringly-typed runtime hub.
   - File: `internal/usercommands/usercommands.go`
   - Problems:
   - The package owns a large static command registry, alias expansion, room-script interception, item-script interception, self-target rewriting, permission gating, profiling, emote fallback, spell fallback, and movement fallback in one function.
   - Cross-package hooks use `GetExportedFunction(string) any` plus runtime type assertions.
   - Impact: This is difficult to reason about, difficult to test in isolation, and easy to regress when adding new command behaviors.

5. Admin item and mob tooling is heavily copy-pasted.
   - Files: `internal/usercommands/admin.item.go`, `internal/usercommands/admin.mob.go`
   - Problem: The list/spawn/create command structure is near-duplicate across large files, with only domain-specific details swapped in.
   - Impact: Behavior will drift over time, fixes must be applied twice, and new admin tooling is incentivized to repeat the same pattern.

6. Reusable transient-storage helpers are duplicated across core entity types.
   - Files: `internal/items/items.go`, `internal/mobs/mobs.go`, `internal/rooms/rooms.go`, `internal/users/userrecord.go`
   - Problem: `SetTempData()` / `GetTempData()` are reimplemented separately for items, mobs, rooms, and users.
   - Impact: The codebase pays repeated maintenance cost for identical behavior and invites divergence when one implementation changes.

7. Zone/path sanitization logic is duplicated across packages.
   - Files: `internal/rooms/roommanager.go`, `internal/mobs/mobs.go`
   - Problem: `ZoneNameSanitize()` exists in both packages with the same behavior.
   - Impact: Small today, but this is exactly the kind of duplication that creates file-path inconsistencies later.

8. Room management still combines indexing, file lookup, movement, zone mutation, and unload policy in one package-global manager.
   - File: `internal/rooms/roommanager.go`
   - Problems:
   - `searchForRoomFile()` does a `filepath.Walk()` over the rooms tree to resolve a room path on cache misses.
   - The same file also owns player movement, ephemeral-room spawning, zone creation, room connection, maintenance, and load/bootstrap.
   - Impact: Slow path discovery is mixed with live runtime behavior, and the room subsystem remains harder to test or optimize than it should be.

### Immediate Refactor Priority

1. Fix the non-persisted `room` admin mutations before relying on the tool for live world editing.
2. Make `LoadUser()` and `SearchOfflineUsers()` fail explicitly on bad data instead of logging and continuing.
3. Remove info-level permission logging from the command hot path or gate it behind a debug flag.
4. Split command dispatch into smaller stages with typed extension points instead of string-based exported-function lookup.
5. Extract shared admin prompt/list helpers and shared temp-data helpers to reduce structural duplication.

## Segment 3: World Loop, Event Delivery, Combat, Scripting, Module Boundaries

Date: 2026-03-29

### Findings

1. The main world loop busy-polls the event system every millisecond under the global mud lock.
   - File: `world.go`
   - Problems:
   - `MainWorker()` creates an `eventLoopTimer` at 1ms and resets it on every tick.
   - The event loop runs inside `util.LockMud()` / `util.UnlockMud()`.
   - Impact: The server pays constant polling overhead and expands lock hold time around all event processing.

2. Event listeners still perform outbound connection writes inline during world-event processing.
   - Files: `internal/hooks/Message_SendMessages.go`, `internal/hooks/Broadcast_SendToAll.go`, `internal/hooks/WebClientCommand_SendWebClientCommand.go`, `world.go`
   - Problem: The main worker calls `w.EventLoop()` while the mud lock is held, and message/broadcast/web-command listeners call `connections.SendTo()` directly in that path.
   - Impact: Slow telnet/websocket clients can directly increase world-loop latency and stall unrelated game work.

3. Hook sequencing is implicit and order-dependent rather than modeled explicitly.
   - File: `internal/hooks/hooks.go`
   - Problem: Listener registration order is being used as a runtime sequencing mechanism, especially around `NewRound` listeners and combat.
   - Impact: Behavior depends on registration order in one central file, which is fragile and hard to validate when new hooks are added.

4. The scripting subsystem is duplicated enough that it already contains copy-paste configuration bugs.
   - Files: `internal/scripting/buff.go`, `internal/scripting/mob.go`, `internal/scripting/spell.go`
   - Problems:
   - Buff script execution uses `scriptRoomTimeout` instead of `scriptBuffTimeout`.
   - Mob script execution uses `scriptRoomTimeout` instead of `scriptMobTimeout`.
   - Spell script execution uses `scriptItemTimeout` instead of `scriptSpellTimeout`.
   - Impact: Timeout configuration is not applied as intended, and this strongly suggests the current five-way VM implementation is too duplicated to maintain safely.

5. Scripting uses package-global mutable output wrappers and a stringly-typed module export boundary.
   - Files: `internal/scripting/scripting.go`, `internal/scripting/room.go`, `internal/scripting/module_func.go`
   - Problems:
   - `userTextWrap` and `roomTextWrap` are package-global mutable state changed around each script invocation.
   - The module bridge is a global `map[string]map[string]any` injected directly into JS.
   - The public registration function is misspelled as `AddModlueFunction`.
   - Impact: The scripting boundary is hard to reason about, easy to misuse, and not structured for future concurrency or stronger typing.

6. Combat orchestration is still split across an oversized round hook and a smaller calculation package.
   - Files: `internal/hooks/NewRound_DoCombat.go`, `internal/combat/combat.go`, `internal/combat/resolution.go`
   - Problems:
   - `NewRound_DoCombat.go` is a 900+ line hook file that owns combat flow, flee logic, spell resolution, retaliation, death handling, and post-combat cleanup.
   - The `combat` package handles calculations and message resolution, but the main side-effect-heavy orchestration remains in the hook layer.
   - Impact: Combat behavior is difficult to test at the scenario level, and combat-specific changes are spread across multiple layers with no clear application boundary.

7. `world.go` still mixes unrelated runtime concerns into one orchestration file.
   - File: `world.go`
   - Problems:
   - Input handling, prompt replay, macro expansion, autocomplete, tick scheduling, event-loop driving, stats updates, zombie transitions, and kick/logout flows all live together.
   - `GetAutoComplete()` alone embeds a large command-aware branching tree that belongs closer to the command layer.
   - Impact: The core runtime is harder to isolate, profile, and evolve than it needs to be.

8. Debug-only and crash-oriented code still exists in module runtime paths.
   - Files: `modules/gmcp/gmcp.go`, `modules/gmcp/gmcp.Mudlet.go`, `modules/webhelp/webhelp.go`
   - Problems:
   - GMCP outbound handling still has `TODO: REMOVE` debug branches in the live send path.
   - Module initialization still uses `panic(err)` on attach/setup failures in additional modules.
   - Impact: Optional capability faults and leftover diagnostics still leak into production execution paths.

### Immediate Refactor Priority

1. Replace the 1ms polling event loop with a queue-driven or condition-driven event runner, and stop holding the global mud lock across outbound writes.
2. Separate event production from connection delivery so message hooks enqueue transport work instead of performing it inline.
3. Collapse the five script-runner implementations into one shared execution harness with per-domain adapters and correct timeout usage.
4. Extract combat orchestration into a dedicated application layer instead of leaving it inside a `NewRound` hook file.
5. Break `world.go` into focused runtime units: input processing, autocomplete, scheduling, and connection/session transitions.

## Segment 4: Dead Code, Hardcoded Behavior, and Test Gaps

Date: 2026-03-29

### Findings

1. Several symbols and methods are effectively dead or placeholder code.
   - Files: `internal/usercommands/usercommands.go`, `internal/users/userrecord.go`
   - Problems:
   - `CommandHelpItem` is defined but has no non-definition references.
   - `(*UserRecord).RoundTick()` is an empty method.
   - Impact: These make the codebase look more extensible than it really is and increase maintenance noise.

2. There are still explicit TODO placeholders in live runtime paths.
   - Files: `internal/hooks/PlayerSpawn_HandleJoin.go`, `internal/usercommands/experience.go`, `internal/scripting/actor_func_progression.go`
   - Problems:
   - `HandleJoin()` contains a `TODO HERE` directly in the player-spawn path.
   - Stat handling is still hardcoded in both user command code and scripting code, and one block still carries a note that `newG` is never used.
   - Impact: Behavior that should be finalized or centralized is still partially provisional in production code.

3. Hardcoded stat-name logic is duplicated across command and scripting surfaces.
   - Files: `internal/usercommands/experience.go`, `internal/scripting/actor_func_progression.go`
   - Problem: The six core stats are manually enumerated in multiple places instead of being driven by a shared stat registry/model.
   - Impact: Any stat-system evolution will require synchronized edits across unrelated packages and is likely to drift.

4. High-risk operational packages still have little or no direct test coverage.
   - Evidence from repository test inventory:
   - `internal/connections`: 0 tests
   - `internal/web`: 0 tests
   - `internal/hooks`: 0 tests
   - `internal/usercommands`: 0 tests
   - `internal/plugins`: 0 tests
   - `internal/mobs`: 0 tests
   - `internal/items`: 0 tests
   - Impact: The packages with the most stateful behavior, I/O boundaries, and orchestration logic are the least protected against regression.

5. Existing tests are concentrated in lower-level utility/model areas rather than the current failure-prone runtime boundaries.
   - Evidence from repository test inventory:
   - `internal/rooms`: 2 test files
   - `internal/scripting`: 1 test file
   - `internal/combat`: 1 test file
   - Problem: There is some coverage, but it is thin relative to the complexity of the systems now carrying world orchestration, scripting execution, admin mutation, and event delivery.
   - Impact: Production bugs are more likely to emerge from integration seams that currently have no dedicated tests at all.

### Immediate Refactor Priority

1. Remove or implement dead placeholders such as `RoundTick()` and delete unused types that no longer participate in the design.
2. Centralize stat metadata so commands and scripts stop maintaining their own hardcoded stat lists.
3. Add targeted tests for the highest-risk packages first: `connections`, `web`, `hooks`, `usercommands`, and `plugins`.
4. Add regression tests for admin room edits, user/plugin persistence failure handling, scripting timeout selection, and event-to-transport delivery behavior.

## Segment 5: Feature Packages, Optional Modules, and Integrations

Date: 2026-03-29

### Findings

1. The Discord webhook client has multiple production-safety bugs.
   - File: `internal/integrations/discord/client.go`
   - Problems:
   - `http.NewRequest()` errors are ignored before `request.Header.Set(...)`, so a failed request build can still dereference `request`.
   - `response.Body` is never closed after `client.Do(...)`.
   - Every send spawns its own goroutine with no queue or worker limit.
   - Impact: Error paths can panic, successful paths leak HTTP resources, and bursts of Discord events can create unbounded goroutine fan-out.

2. The Discord event listener path can panic on missing connection state.
   - Files: `internal/integrations/discord/listeners.go`, `internal/connections/connections.go`
   - Problem: `HandlePlayerSpawn()` calls `connections.Get(user.ConnectionId())` and immediately dereferences `connDetails.IsWebSocket()` without a nil check.
   - Impact: Racey login/disconnect edges can crash the Discord integration path.

3. The leaderboard module load path appears broken and likely discards persisted state.
   - File: `modules/leaderboards/leaderboards.go`
   - Problems:
   - `loadLBs()` calls `l.plug.ReadIntoStruct(..., &l)`, passing a pointer-to-pointer receiver variable rather than the module struct.
   - The same method then reinitializes `LB_Gold`, `LB_Experience`, and `LB_Kills`, which would overwrite loaded leaderboard content even if deserialization succeeded.
   - Impact: Persisted leaderboard data is unlikely to be restored correctly.

4. Optional modules still use inconsistent and sometimes unsafe config access patterns.
   - Files: `modules/auctions/auctions.go`, other modules
   - Problem: Most module config reads use checked type assertions, but `auctions.go` directly does `mod.plug.Config.Get("UpdateSeconds").(int)` in the live update path.
   - Impact: A malformed or missing config value can panic a running module instead of falling back safely.

5. Party state is a package-global mutable registry with no persistence, no locks, and at least one dead/misleading method.
   - File: `internal/parties/parties.go`
   - Problems:
   - `partyMap` is a global mutable map.
   - There is a receiver method `(*Party).New()` that does not create a party and is semantically misleading.
   - The package has no tests.
   - Impact: Party logic remains fragile, implicit, and hard to evolve or validate.

6. Pet and quest data loading still uses panic-driven bootstrap, and the packages have no direct tests.
   - Files: `internal/pets/pets.go`, `internal/quests/quests.go`
   - Problem: `LoadDataFiles()` in both packages still `panic(err)` on load failure, and neither package has direct `_test.go` coverage.
   - Impact: Non-core content/data issues can still crash server startup, with little automated protection.

7. Optional module initialization is still highly repetitive and crash-oriented.
   - Files: multiple `modules/*/*.go`
   - Pattern:
   - Construct plugin in `init()`
   - `AttachFileSystem(files)` and `panic(err)` on failure
   - Register commands/listeners/callbacks inline
   - Impact: Module boot remains copy-pasted and brittle, and optional features are not isolated from process-wide startup failure.

8. Feature and integration packages have effectively no direct automated coverage.
   - Evidence:
   - `internal/parties`: 0 tests
   - `internal/pets`: 0 tests
   - `internal/quests`: 0 tests
   - `internal/actions`: 0 tests
   - `internal/integrations/discord`: 0 tests
   - Impact: The non-core subsystems most likely to regress via integration behavior have no package-level safety net.

### Immediate Refactor Priority

1. Fix the Discord client first: handle `NewRequest` errors, close response bodies, and route outbound webhooks through a bounded worker/queue.
2. Repair `leaderboards.loadLBs()` so it deserializes into the real module instance and does not overwrite restored data.
3. Standardize module config access so optional modules never use unchecked assertions in live paths.
4. Replace repeated module `init()` bootstrap patterns with a shared registration/bootstrap helper that reports module-load failure without panicking the whole process.
5. Add focused tests for party behavior, leaderboard persistence, Discord webhook delivery/error handling, and auction config fallback behavior.

## Production Roadmap

### Remediation Progress

Completed:
1. Fixed plugin persistence error handling and module load checks.
2. Fixed admin room mutations that changed memory without saving.
3. Fixed scripting timeout variable selection in buff, mob, and spell runners.
4. Fixed connection lifecycle bugs around stale entries, zero-padded ID lists, and write-under-registry-lock behavior.
5. Fixed leaderboard restore so persisted data is actually loaded.
6. Hardened the Discord client and listener against request, body, queue, and nil-dereference failures.
7. Moved prompt/message/broadcast/GMCP/log/web-command/MSP delivery onto bounded async connection workers so outbound I/O is no longer done inline on the main event path.
8. Changed listener dispatch to snapshot handlers before invocation, so listeners can register/unregister safely without holding the listener registry lock during handler execution.
9. Replaced the 1ms polled world event loop with event-driven wakeups from the event queue and requeue paths.
10. Fixed the config overlay write path to use an exclusive lock instead of mutating shared config state under an `RLock`.
11. Reworked bulk flat-file saves to create parent directories and return save errors instead of panicking worker goroutines.
12. Hardened the web template/origin boundary by rejecting cross-host websocket origins and returning immediately on template parse failures.
13. Fixed user load/search paths so malformed user YAML and offline-user scan failures are surfaced instead of silently ignored.
14. Replaced hard-fail optional module init paths with degraded startup logging for embedded plugin modules (`auctions`, `follow`, `leaderboards`, `webhelp`, `cleanup`, `time`, `gmcp.Mudlet`) instead of panicking the process.
15. Hardened `mudlog` so init-time error logging can fall back to stderr before the main logger is configured.
16. Introduced an explicit transport event boundary with `TransportBatch` events and single-event world dispatch, so socket/webclient/GMCP/log/MSP prompt delivery now runs outside the mud lock after simulation listeners build immutable delivery intents.
17. Moved Discord/webhook delivery onto the same event-boundary model by queuing `DiscordWebhook` transport events from simulation listeners and processing the actual webhook queueing outside the mud lock.
18. Added per-listener simulation/transport phasing so selected listeners on normal gameplay events can run after the simulation phase, outside the mud lock, without moving the entire event type.
19. Moved Discord’s gameplay listeners onto the transport phase and extended `PlayerSpawn` with connection snapshot data so those notifications no longer need live user/connection lookups.
20. Added online-player snapshot data to spawn/despawn events and moved `gmcp.Game` join/leave payload generation onto the transport phase without live active-user reads.
21. Moved `WebClientCommand` dispatch to the transport phase and split log following into simulation-phase subscription management plus transport-phase output delivery with its own lock.

Still open:
1. Simulation/event mutation still runs under the mud lock; more listeners need explicit categorization, snapshots, or local synchronization before they can safely leave it.
2. High-risk runtime packages still need broader concurrency and integration coverage.
3. Optional subsystems still need more consistent degraded-mode behavior instead of ad hoc failure handling.

Note:
The remaining mud-lock reduction now depends on listener-by-listener categorization. The mechanism exists, but additional listeners need either event snapshot data or local synchronization before they can be moved to the transport/read-only phase safely.

### Tier 0: Correctness Bugs To Fix Before Trusting Live Edits Or Runtime Safety

1. Fix `Plugin.ReadIntoStruct()` error handling and make all plugin persistence loads check failures.
2. Repair connection lifecycle bugs in `internal/connections`, especially stale entries and zero-padded ID lists.
3. Fix all admin room mutations that currently modify in memory without saving.
4. Correct scripting timeout variable selection in buff, mob, and spell script execution.
5. Fix `leaderboards.loadLBs()` so persisted state actually restores.
6. Fix the Discord client request/error/body-handling bugs.

### Tier 1: Runtime Model Changes Needed For Production Scale

1. Stop holding the global mud lock across world event processing that performs connection output.
2. Replace the 1ms polling event loop with a queue-driven event runner.
3. Separate event generation from connection delivery so outbound I/O is isolated from the simulation loop.
4. Remove hot-path info logging such as permission-check logging.
5. Tighten concurrency boundaries:
   - keep world-state mutation serialized
   - move socket/webhook delivery onto bounded async workers
   - shorten lock scope in connection and event systems
   - eliminate or guard package-global mutable registries and caches
6. Replace ad hoc goroutine usage with explicit queue/worker designs where backpressure matters, especially for outbound integrations.

### Tier 2: Crash-Oriented Boot And Save Paths To Replace

1. Replace `panic(err)` startup/data-load behavior in core packages and optional modules with explicit load failures.
2. Make bulk save paths return and aggregate errors instead of panicking worker goroutines.
3. Standardize module config access and fallback handling so bad config cannot panic a live subsystem.

### Tier 3: Structural Refactors With High Long-Term Payoff

1. Split `main.go` and `world.go` into focused runtime/session/bootstrap units.
2. Break command dispatch into explicit stages with typed extension points.
3. Extract combat orchestration from hook files into a dedicated combat application layer.
4. Collapse duplicated script-runner implementations into one shared execution harness.
5. Split `internal/web/web.go` into router, rendering, websocket, and server-lifecycle components.

### Tier 4: Cleanup And Consistency Work

1. Remove dead or misleading code such as `CommandHelpItem`, `(*Party).New()`, and empty placeholder methods.
2. Centralize shared concepts that are currently duplicated:
   - temp-data helpers
   - zone-name sanitization
   - stat metadata
   - admin prompt/list scaffolding
3. Remove debug-only branches still present in production paths.

### Tier 5: Testing Strategy

1. Add direct tests for:
   - `internal/connections`
   - `internal/web`
   - `internal/hooks`
   - `internal/usercommands`
   - `internal/plugins`
2. Add regression coverage for:
   - admin room persistence
   - plugin/user persistence corruption handling
   - scripting timeout selection
   - event-to-transport delivery behavior
   - module config fallback behavior
   - Discord webhook failures and rate bursts
3. Add concurrency-focused tests for:
   - slow-client behavior not stalling simulation
   - connection removal during delivery
   - listener registration/unregistration safety during dispatch
   - bounded integration worker behavior under burst load

## Concurrency Guidance

1. Treat the world simulation as a mostly single-writer domain.
2. Push network and webhook delivery out to dedicated worker queues with bounded capacity.
3. Never hold global locks while doing socket writes, websocket writes, disk I/O, or webhook calls.
4. Prefer immutable snapshots or copied slices/maps when broadcasting over shared registries.
5. If a package owns mutable global state, either move it behind the world/event thread or add explicit synchronization and tests.

## Additional Recommendations

1. Write down the target runtime model explicitly.
   - Define which subsystems are single-writer, which are async, and which boundaries are queue-based.
2. Introduce a transport layer between world/events and outbound delivery.
   - Telnet, websocket, GMCP, and Discord should consume delivery intents rather than being called directly from simulation logic.
3. Add degraded-mode behavior for optional subsystems.
   - Discord, GMCP, web, and plugin persistence should fail closed or fail soft without taking the process down.
4. Add a short architecture note for maintainers.
   - Document event flow, lock ownership, queue ownership, and package responsibilities.
5. Use race detection as part of verification for high-risk packages.
   - Prioritize `connections`, `events`, `hooks`, `scripting`, and any transport worker code once concurrency fixes start landing.

## Recent Remediation Updates

1. `GMCPOut` now snapshots connection target and wrapper choice at enqueue time and dispatches in the transport phase.
2. GMCP payload delivery no longer depends on a live `users.GetConnectionId()` lookup during send.
3. Added GMCP regression coverage for snapshotted telnet/web delivery wrappers.
4. `Message` and `Broadcast` now split recipient selection from render/send work so ANSI parsing and payload assembly run in the transport phase instead of under the simulation lock.
5. `RedrawPrompt` and `MSP` now use simulation-phase delivery planning with transport-phase payload emission, keeping prompt rendering and telnet/websocket wrapper selection out of the simulation hot path.
6. Mudlet-originated GMCP sends now go through a dedicated transport-phase outbound plan, and Discord/Mudlet producer events now carry `UserId` directly instead of rescanning active users by connection.
7. Mudlet Discord status requests now carry prepared GMCP messages into a transport-phase handler, and Mudlet’s local client cache now has its own lock so cleanup can leave the mud lock safely.

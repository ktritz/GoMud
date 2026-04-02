# Tinymap ANSI Tag Rendering Bug

## Problem

The inline tinymap (minimap shown alongside room descriptions) renders broken ANSI tags in the web client. Instead of colored map symbols, players see raw tag fragments:

```
║  ="map-yard" bg="mapbg-yard">Y─G║
║  │ │║
║  ="map-you" bg="mapbg-you">@─="map-kitchen" bg="mapbg-kitchen">K║
```

Should render as just `Y`, `G`, `@`, `K` etc. with colors.

## How to Reproduce

```bash
# Connect with the gomud-agent
/opt/gomud/bin/gomud-agent connect --url ws://localhost:80/ws --user admin --pass password &
sleep 4

# Make sure tinymap is on (toggle if needed)
/opt/gomud/bin/gomud-agent send "set tinymap"

# Go to a room with map legend symbols
/opt/gomud/bin/gomud-agent send "teleport 1"
sleep 2
/opt/gomud/bin/gomud-agent read

# Also visible in web browser at http://192.168.1.25/webclient
```

The `@` (player position) and plain box-drawing chars (`│`, `─`, `╔`) render fine. Only symbols with `maplegend` entries break (like `S` for Street, `Y` for Yard, etc.).

## Root Cause Analysis

### Tag generation (FIXED)
The tinymap code in `internal/usercommands/look.go:506` generates ANSI tags for map legend symbols. Originally it produced nested tags:
```go
// OLD (fixed in this branch):
fmt.Sprintf(`<ansi fg="map-room"><ansi fg="map-%s" bg="mapbg-%s">%c</ansi></ansi>`, ...)

// NEW:
fmt.Sprintf(`<ansi fg="map-%s" bg="mapbg-%s">%c</ansi>`, ...)
```
The same pattern existed in `internal/scripting/room_func.go:472` and `internal/usercommands/skill.map.go:178` - both also fixed.

### Color aliases (FIXED)
The `ansi-aliases.yaml` was missing `map-*` and `mapbg-*` entries for all 176 maplegend values. Added them all with fg=15 (white) and bg=0 (black).

### Maplegend spaces (FIXED)
Some maplegend values like "Bus Stop" had spaces which would produce invalid tag attributes `fg="map-bus stop"`. Fixed by replacing spaces with hyphens.

### The actual unsolved bug: ANSI conversion pipeline

The tags ARE valid. Standalone test proves the `ansitags` library converts them correctly:

```go
// This works perfectly:
ansitags.LoadAliases("/opt/gomud/data/world/insideout/ansi-aliases.yaml")
result := ansitags.Parse(`<ansi fg="map-yard" bg="mapbg-yard">Y</ansi>`)
// result = "\x1b[38;5;15m\x1b[48;5;0mY\x1b[0m"  ← correct escape sequence
```

But in the actual server output, the tags arrive at the web client as raw text with only the `<ansi fg` portion consumed:
```
="map-yard" bg="mapbg-yard">Y
```

The `<ansi fg` part was converted to an ANSI escape sequence, but the `="map-yard" bg="mapbg-yard">` part leaked as plain text.

### What I investigated

1. **`templates.AnsiParse()`** in `internal/templates/templates.go:427` - This is called in `internal/hooks/Message_SendMessages.go:33` on every message before sending to the client. When `forceAnsiFlags == AnsiTagsParse` (the default, set at line 57), it calls `ansitags.Parse(input)` which should convert all tags.

2. **`ProcessText()`** in `internal/templates/templates.go:270` - Template processing does NOT apply ANSI parsing (lines 276-277 have the force flag check commented out). So templates output raw `<ansi>` tags.

3. **The rendering chain for `look` command:**
   - `look.go` builds tinymap lines with `<ansi fg="map-..." bg="mapbg-...">` tags
   - `rooms.GetDetails()` in `internal/rooms/roomdetails.go:42` merges tinymap into description text
   - The merged description goes through `templates.Process("descriptions/room", ...)` which does NOT parse ANSI
   - The processed template text is sent via `user.SendText()` which creates a `Message` event
   - `Message_SendMessages.go` calls `templates.AnsiParse()` which SHOULD convert the tags
   - But the tags in the tinymap portion are NOT being converted

4. **Hex dump of gomud-agent output** confirms no `\x1b` bytes at all for the map portion - the escape sequences are not present. The `<ansi fg` opening tag IS being partially consumed (not visible in output) but the attributes and closing `>` leak as text.

### Possible causes not yet explored

- The `<ansi fg=` portion might match a DIFFERENT tag pattern earlier in the ansitags parser pipeline (maybe the template engine or color pattern system intercepts it)
- The room description template `descriptions/room.template` wraps the description in `<ansi fg="room-description">...</ansi>` - this outer tag might interfere with inner map tags
- The `SplitString` function (which we patched for width calculation) might be altering the tag text when joining the tinymap with the description
- The `GetDetails()` function at `roomdetails.go:125` does `description[i] += spaces + tinymap[i]` using `len()` instead of visible width, which may miscalculate padding and corrupt tag boundaries

### Files involved

- `internal/usercommands/look.go:480-510` - tinymap generation and tag insertion
- `internal/rooms/roomdetails.go:115-136` - merging tinymap into description
- `internal/templates/templates.go:427-448` - AnsiParse function
- `internal/hooks/Message_SendMessages.go:33` - where AnsiParse is called
- `internal/util/util.go` - SplitString (patched for ANSI-aware width)
- `_datafiles/world/insideout/ansi-aliases.yaml` - color alias definitions
- `_datafiles/world/insideout/templates/descriptions/room.template` - wraps description in outer ansi tag

### Quick test to verify the issue

```go
// Add debug logging in Message_SendMessages.go before AnsiParse:
// Check if the tinymap tags are present in message.Text
if strings.Contains(message.Text, "map-yard") {
    mudlog.Debug("MAP TAG PRESENT", "text", message.Text[:200])
}
textOut := templates.AnsiParse(message.Text)
if strings.Contains(textOut, "map-yard") {
    mudlog.Error("MAP TAG NOT CONVERTED", "text", textOut[:200])
}
```

This would confirm whether the tags arrive at AnsiParse and whether they survive the conversion.

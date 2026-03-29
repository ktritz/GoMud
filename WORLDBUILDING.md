# GoMud World Building Reference

Guide for creating world content via YAML data files. All paths relative to the `DataFiles` directory (e.g., `/opt/gomud/data/world/default/`).

## Quick Start: New World

```bash
cp -r _datafiles/world/empty _datafiles/world/myworld
# Edit config.yaml: DataFiles: _datafiles/world/myworld
```

## Directory Structure

```
world/myworld/
  rooms/{zone-name}/          # Room files: {roomid}.yaml + zone-config.yaml
  mobs/{zone-name}/           # Mob files: {mobid}-{name}.yaml
  items/
    weapons-10000/            # Weapon items: {itemid}-{name}.yaml
    armor-20000/{slot}/       # Armor by slot: head/, body/, legs/, etc.
    consumables-30000/        # Potions, food, etc.
    other-0/                  # Keys, quest items, misc
  spells/                     # {spellid}.yaml
  buffs/                      # {buffid}-{name}.yaml
  races/                      # {raceid}-{name}.yaml
  quests/                     # {questid}-{name}.yaml
  templates/                  # Display templates (status, help, etc.)
  users/                      # Player save files (auto-generated)
```

## Rooms

**File:** `rooms/{zone}/{roomid}.yaml`

```yaml
roomid: 100                        # Unique integer ID
zone: MyZone                       # Must match folder name (case-sensitive)
title: A Dark Cave
description: >-
  The cave walls are damp and covered in moss. A faint light
  glimmers from deeper within.
mapsymbol: C                       # Single char for ASCII map (optional)
maplegend: Cave                    # Legend label (optional)
biome: cave                        # Affects lighting/weather. Options: city, forest, cave, dungeon, road, etc.
pvp: false                         # Allow PvP in this room (when server PvP is "limited")
exits:
  north:
    roomid: 101                    # Target room ID
  south:
    roomid: 99
  cave entrance:                   # Custom exit names work
    roomid: 102
    secret: true                   # Hidden from "exits" list until discovered
    lock:
      difficulty: 20               # 0 = no lock. Higher = harder to pick
nouns:                             # Examinable nouns in the description
  moss: The moss is thick and spongy, releasing a musty scent when disturbed.
  light: A phosphorescent fungus clings to the far wall.
containers:                        # Searchable containers
  chest:
    lock:
      difficulty: 10
    items:
      - itemid: 10004
    gold: 50
spawninfo:                         # Mobs/items that spawn here
  - mobid: 5                       # Mob to spawn
    message: A rat scurries out from the shadows.
    respawnrate: 5 real minutes    # How often. Options: "X real minutes", "X game hours"
    maxwander: 3                   # How far mob can wander from this room
    levelmod: 0                    # Level adjustment (added to mob base level)
  - itemid: 10001                  # Item to spawn on floor
    respawnrate: 10 real minutes
    containername: chest           # Spawn inside a container instead of floor
idlemessages:                      # Random ambient text shown to players
  - Water drips from the ceiling.
  - You hear a distant rumbling.
skilltraining:                     # Skills trainable in this room
  cast:
    min: 1
    max: 2
```

## Zone Config

**File:** `rooms/{zone}/zone-config.yaml`

```yaml
name: MyZone
roomid: 100                        # Starting/default room for this zone
autoscale:                         # Auto-scale mob levels to player level
  minimum: 1
  maximum: 10
idlemessages:                      # Zone-wide ambient messages
  - A cool breeze blows through the area.
musicfile: static/audio/music/myzone.mp3
defaultbiome: forest
```

## Mobs

**File:** `mobs/{zone}/{mobid}-{name}.yaml`

```yaml
mobid: 10                          # Unique integer ID
zone: MyZone
hostile: true                      # Attacks players on sight
itemdropchance: 25                 # % chance to drop inventory on death
maxwander: 5                       # Max rooms from home room
activitylevel: 20                  # 1-100, chance per round to do idle/combat action
groups:                            # Group affiliations (for callforhelp, hostility)
  - undead
  - cave-dwellers
hates:                             # Auto-attack these groups/races on sight
  - frostfang-law
idlecommands:                      # What they do when not fighting
  - emote growls menacingly
  - wander
  - ""                             # Empty = do nothing (reduces activity frequency)
combatcommands:                    # Special actions during combat
  - backstab                       # Use skills
  - cast mm                        # Cast spells
  - "callforhelp 3:yells for backup."  # Summon allies (range:message)
angrycommands:                     # Said/done when entering combat
  - shout You dare challenge me?
scripttag: myboss                  # Links to script file: mobs/{zone}/scripts/{mobid}-{name}-{tag}.js
buffids:                           # Permanent buffs (always active)
  - 9                              # e.g., hidden buff
questflags:                        # Quest flags this mob carries
  - quest-1-start
character:
  name: skeleton warrior
  description: >-
    A reanimated skeleton clad in rusted armor, wielding a
    chipped sword with eerie precision.
  level: 5
  raceid: 0                        # Race ID (affects stats, size, etc.)
  alignment: -50                   # -128 to 127 (-evil to +good)
  gold: 15                         # Gold dropped on death
  stats:
    strength:
      training: 5                  # Extra training points in this stat
    speed:
      training: 3
  equipment:
    weapon:
      itemid: 10002
    body:
      itemid: 20008
  items:                           # Backpack items (can be dropped on death)
    - itemid: 10001
  shop:                            # Makes this mob a merchant
    - itemid: 10004
      quantitymax: 3               # Max stock (restocks over time)
    - itemid: 20003
      quantitymax: 1
```

## Items

### Weapons
**File:** `items/weapons-10000/{itemid}-{name}.yaml`

```yaml
itemid: 10010
name: iron sword
namesimple: sword                  # Short name for "the sword breaks" messages
description: A sturdy iron sword with a leather-wrapped grip.
type: weapon
subtype: slashing                  # slashing, stabbing, cleaving, bludgeoning, shooting, whipping
hands: 1                           # 1 or 2 handed
damage:
  diceroll: 1d6+1                  # XdY+Z format
waitrounds: 0                     # Extra rounds between attacks (0 = normal speed)
breakchance: 5                    # % chance to break on crit received (offhand only)
statmods:                         # Stat bonuses when equipped
  speed: -1
  strength: 2
value: 50                         # Gold value for shops
```

### Armor
**File:** `items/armor-20000/{slot}/{itemid}-{name}.yaml`

Slots: `head/`, `neck/`, `body/`, `belt/`, `gloves/`, `ring/`, `legs/`, `feet/`

```yaml
itemid: 20010
name: leather helmet
namesimple: helmet
description: A hardened leather helmet providing basic protection.
type: head                         # Must match the slot folder
subtype: wearable
damagereduction: 3                # Flat damage reduction
statmods:
  perception: -1                   # Stat modifiers when worn
value: 25
wornbuffids:                      # Buffs applied while wearing
  - 3
```

### Consumables
**File:** `items/consumables-30000/{itemid}-{name}.yaml`

```yaml
itemid: 30010
name: health potion
namesimple: potion
description: A vial of red liquid that restores health.
type: potion
subtype: drinkable
uses: 1                           # How many times it can be used (destroyed at 0)
value: 30
```

### Other Items (keys, quest items)
**File:** `items/other-0/{itemid}-{name}.yaml`

```yaml
itemid: 50
name: rusty key
namesimple: key
description: An old rusty key. It might fit something.
type: key
keylockid: "100-chest"            # Matches room lock format: "{roomid}-{exit/container}"
value: 0
```

## Spells

**File:** `spells/{spellid}.yaml`

```yaml
spellid: fireball
name: Fireball
description: Hurls a ball of fire for 2d6 damage
type: harmsingle                   # helpsingle, helpmulti, helparea, harmsingle, harmmulti, harmarea
school: destruction                # School for casting skill checks
cost: 8                            # Mana cost
waitrounds: 3                     # Casting time in rounds
difficulty: 40                    # 0-100, affects success chance
```

Spell types:
- `helpsingle` — heals/buffs one target (defaults to self)
- `helpmulti` — heals/buffs party
- `helparea` — heals/buffs everyone in room
- `harmsingle` — damages one target
- `harmmulti` — damages all enemies
- `harmarea` — damages everyone in room

## Buffs

**File:** `buffs/{buffid}-{name}.yaml`

```yaml
buffid: 50
name: Burning
description: You are on fire!
triggerrate: 1 round               # How often the effect triggers
triggercount: 5                    # How many times (-1 = permanent until removed)
statmods:
  strength: -2                     # Stat modifications while active
flags:                             # Special behavior flags
  - cancel-on-combat               # Removed when entering combat
  - cancel-on-action               # Removed on any action
  - hidden                         # Character is hidden/sneaking
  - emits-light                    # Character emits light
  - no-combat                      # Prevents combat
  - poison                         # Poison DOT
  - perma-gear                     # Can't change equipment
  - revive-on-death                # Auto-revive once
```

## Races

**File:** `races/{raceid}-{name}.yaml`

```yaml
raceid: 10
name: ratkin
description: Small, agile rat-people with keen senses.
defaultalignment: -10
size: small                        # small, medium, large (affects weapon requirements)
unarmedname: claws
tnlscale: 0.9                     # XP to next level multiplier (1.0 = normal)
selectable: true                  # Can players choose this race?
knowsfirstaid: false
tameable: true                    # Can be tamed as pet
stats:
  speed:
    base: 3                        # Higher base = grows faster per level
  perception:
    base: 2
  strength:
    base: 1
  vitality:
    base: 1
damage:
  diceroll: 1d3                    # Unarmed damage
  attacks: 1
disabledslots:                    # Equipment slots this race can't use
  - ring
```

## ID Ranges

| Type | Range | Example |
|------|-------|---------|
| Weapons | 10000-19999 | `10001-sharp_stick.yaml` |
| Armor | 20000-29999 | `20001-rat_pelt.yaml` |
| Consumables | 30000-39999 | `30001-health_potion.yaml` |
| Other items | 0-9999 | `50-rusty_key.yaml` |
| Rooms | Any positive int | `100.yaml` |
| Mobs | Any positive int | `1-rat.yaml` |
| Races | 0+ | `0-ghostly_spirit.yaml` |
| Buffs | 0+ | `0-meditating.yaml` |

## Room ID Assignment

Room IDs must be globally unique across all zones. Check existing max:
```bash
find rooms/ -name "*.yaml" ! -name "zone-*" | sed 's/.*\///' | sed 's/\.yaml//' | sort -n | tail -1
```

## Connecting Rooms

Exits are one-directional. To make a two-way connection, both rooms need exits pointing to each other:

```yaml
# Room 100
exits:
  north:
    roomid: 101

# Room 101
exits:
  south:
    roomid: 100
```

## After Creating Content

Restart the server to load new files:
```bash
sudo systemctl restart gomud
```

Or use the in-game admin command to reload without restart:
```
reload
```

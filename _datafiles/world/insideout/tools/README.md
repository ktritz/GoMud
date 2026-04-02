# GoMud World Content Tools

Content generation and validation tools for GoMud world data. Designed to be used with AI coding assistants to minimize token usage by handling YAML boilerplate automatically.

## Structure

```
tools/
  worldgen.py          CLI entry point for generation and validation
  validate.py          Standalone validation/fix script
  gomudlib/
    __init__.py
    yamlutil.py        YAML dumping, file slugification, string sanitization
    generators.py      Content generators (rooms, mobs, items, quests, etc.)
    validate.py        World validation and auto-fix logic
  batch-*.yaml         Batch definition files (input to worldgen)
```

## Quick Start

### Generate content from a batch file

```bash
python3 tools/worldgen.py rooms tools/batch-rooms-headquarters.yaml
python3 tools/worldgen.py mobs tools/batch-mobs-headquarters.yaml
python3 tools/worldgen.py items tools/batch-items-all.yaml
python3 tools/worldgen.py quests tools/batch-quests.yaml
python3 tools/worldgen.py races tools/batch-races.yaml
python3 tools/worldgen.py buffs tools/batch-buffs.yaml
python3 tools/worldgen.py spells tools/batch-spells.yaml
python3 tools/worldgen.py conversations tools/batch-convos.yaml
```

### Validate a world

```bash
# Check for errors
python3 tools/validate.py

# Check and auto-fix
python3 tools/validate.py --fix

# Validate a different directory (e.g., deployed server)
python3 tools/validate.py --base-dir /opt/gomud/data/world/insideout --fix
```

## Batch File Formats

### Rooms

```yaml
zone: rileys_neighborhood
zone_config:
  name: Rileys Neighborhood
  roomid: 1
  autoscale: {minimum: 1, maximum: 3}
  defaultbiome: city
rooms:
  - id: 1
    title: "Riley's Street"
    desc: >-
      Room description here. 3-5 sentences.
    biome: city
    symbol: S
    legend: Street
    exits: {north: 2, east: 6}
    nouns: {sign: "A neighborhood map is posted here."}
    idle: ["A car drives past.", "The fog rolls in."]
    spawns:
      - {mob: 1, rate: "3 real minutes", msg: "A thought worm appears."}
```

Exit variants:
- Simple: `north: 2`
- Locked: `north: {roomid: 2, lock: 15}`
- Secret: `north: {roomid: 2, secret: true}`

### Mobs

```yaml
zone: headquarters
zone_config:
  name: Headquarters
mobs:
  - id: 20
    name: joy
    desc: >-
      Mob description here.
    level: 99
    race: 2
    hostile: false
    wander: 0
    activity: 30
    drop: 10
    groups: [headquarters-emotion]
    idle: ["emote bounces", "say Hello!", ""]
    combat: ["backstab", "cast mm"]
    angry: ["shout Let's go!"]
    script: questgiver
    questflags: ["3-start", "3-end"]
    alignment: 10
    gold: 50
    equip: {weapon: 10001, body: 20010}
    items: [30001]
    shop:
      - {item: 30001, qty: 10}
      - {item: 10001, qty: 5}
```

### Items

```yaml
weapons:
  - id: 10001
    name: pencil
    simple: pencil
    desc: "A sharp pencil."
    damage: 1d2
    hands: 1
    subtype: stabbing
    value: 5
    statmods: {speed: 1}

armor:
  - id: 20001
    name: baseball cap
    simple: cap
    desc: "A worn cap."
    slot: head       # head, neck, body, belt, gloves, ring, legs, feet, offhand
    dr: 1
    value: 10
    statmods: {perception: 1}

consumables:
  - id: 30001
    name: juice box
    simple: juice
    desc: "Apple juice."
    uses: 1
    value: 10
    buffids: [40]

other:
  - id: 1
    name: rusty key
    simple: key
    desc: "An old key."
    type: key
    keylock: "100-north"
```

### Quests

```yaml
quests:
  - id: 1
    name: The Missing Homework
    desc: "Find your homework!"
    steps:
      - id: start
        desc: Find the homework.
        hint: Check the park.
      - id: found
        desc: Return it to the teacher.
        hint: Classroom 101.
      - id: end
        desc: "Turned in!"
    rewards:
      xp: 1000
      gold: 50
      item: 20001
      msg: "Great job!"
      roommsg: "Homework delivered!"
```

### Races

```yaml
races:
  - id: 1
    name: human
    desc: A regular human.
    size: medium          # small, medium, large
    unarmed: fists
    tnl: 1                # XP multiplier (1.0 = normal)
    selectable: true
    firstaid: true
    alignment: 0
    stats: {strength: 1, smarts: 1, vitality: 1, perception: 1}
    damage: 1d3
    disabled: []          # disabled equipment slots
    angry: ["shout Let's go!"]
```

### Buffs

```yaml
buffs:
  - id: 40
    name: Juice Box Healing
    desc: "Refreshing juice heals you."
    rate: 2 rounds        # trigger interval
    count: 3              # must be >= 1
    statmods: {vitality: 5}
    flags: [cancel-on-combat]
```

### Spells

```yaml
spells:
  - id: heal
    name: Minor Heal
    desc: "Heals for 2d3"
    type: helpsingle      # helpsingle, helpmulti, harmsingle, harmmulti
    school: restoration
    cost: 3
    wait: 2
    difficulty: 0
```

### Conversations

```yaml
zone: headquarters
conversations:
  - mobid: 20
    variants:
      - supported:
          "joy": ["sadness", "anger"]
        lines:
          - ["#1 say Remember when we thought only happy memories mattered?"]
          - ["#2 say I'm glad we learned better."]
```

## Validation Checks

The validator catches issues that cause GoMud server panics:

| Check | Auto-fixable | Description |
|-------|-------------|-------------|
| YAML parse errors | No | Invalid YAML syntax |
| Colon in idle messages | Yes | Strings parsed as key:value dicts |
| Zone name mismatch | Yes | Zone name doesn't derive to folder name |
| Filename hyphens | Yes | Hyphens in mob/item names (GoMud uses underscores) |
| Filename parentheses | Yes | Parens in item names (GoMud uses underscores) |
| Buff triggercount < 1 | Yes | GoMud requires triggercount >= 1 |

## Naming Rules

GoMud derives filesystem paths from content names. The `slugify()` function in `gomudlib/yamlutil.py` matches this behavior:

- Lowercase everything
- Replace spaces, apostrophes, hyphens, parentheses with underscores
- Collapse multiple underscores
- Strip trailing underscores

**Avoid** these in zone names, item names, and mob names:
- Apostrophes (`Riley's` -> use `Rileys`)
- Parentheses (`core memory (gold)` -> use `core memory gold`)
- Hyphens in the name portion (`void-touched` -> use `void touched`)

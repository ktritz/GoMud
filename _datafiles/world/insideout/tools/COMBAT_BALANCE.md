# Combat & Balance Guide -- Inside Out GoMud World

Reference document for understanding GoMud combat mechanics and tuning
mob/weapon balance in the Inside Out world.

---

## Stat -> Combat Action Mappings

From `_datafiles/config.yaml` under `Statistics.StatActions`. Each stat
governs the following game actions:

### Strength
| Action | Effect |
|---|---|
| DiceSides | Additional weapon damage dice sides |
| CritPrimary | Critical hit chance (primary component) |
| CarryCapacity | How many items can be carried |
| DisarmDefense1 | Resist disarm (primary) |
| BumpOffense | Bump/push success |

### Speed
| Action | Effect |
|---|---|
| DiceCount | Additional weapon damage dice count |
| AttackCount | Extra attacks per combat round |
| HitChance | Chance to land a hit |
| CritSecondary | Critical hit chance (secondary component) |
| FleeChance | Chance to flee from combat |
| MovementSpeed | Extra movements per round |
| TackleOffense | Tackle success chance |
| DisarmOffense1 | Disarm success (primary) |
| Pickpocket1 | Pickpocket success (primary) |
| MobFollowSpeed | Speed component of mob follow |

### Smarts
| Action | Effect |
|---|---|
| DisarmOffense2 | Disarm success (secondary) |
| Pickpocket2 | Pickpocket success (secondary) |
| MapMemory | Room memory capacity |
| MapSprawl | Map generation range |
| TameGrowth2 | Taming skill growth (secondary) |

### Vitality
| Action | Effect |
|---|---|
| MaxHealth | Contribution to maximum HP |

### Mysticism
| Action | Effect |
|---|---|
| MaxMana | Contribution to maximum MP |
| Casting | Spell casting success chance |
| EnchantPower | Enchantment damage/defense/stat bonus |
| EnchantProtection | Reduction in enchant item destruction |
| PortalDuration | How long portals last |
| PrayerPower | Number of prayer buffs received |

### Perception
| Action | Effect |
|---|---|
| DamageBonus | Flat bonus damage per hit |
| TackleDefense | Resist being tackled |
| DisarmDefense2 | Resist disarm (secondary) |
| Pickpocket3 | Pickpocket success (tertiary) |
| PickpocketDefense | Resist pickpocket |
| SearchSuccess | Search for hidden things |
| BarterDiscount | Negotiate better prices |
| MobFollowDefense | Mob chase awareness |
| HostilityReduction | Reduce hostile mob timer |
| MobWeaponSearch | Mob AI weapon pickup chance |
| TameGrowth1 | Taming skill growth (primary) |

---

## How Combat Resolution Works

Source: `internal/combat/combat.go`, `internal/characters/character_equipment.go`,
`internal/stats/stats.go`, `internal/characters/character.go`

### Stat Value Calculation

Every stat goes through `StatInfo.Recalculate(level)`:

```
Racial = GainsForLevel(level) = floor((level - 1) * BaseModFactor * RacialBase) + floor(level * NaturalGainsModFactor)
Value  = Racial + Training + Mods
```

Where:
- `BaseModFactor` = 0.3333 (from config)
- `NaturalGainsModFactor` = 0.5 (from config)
- `RacialBase` = the `base` value from the race YAML (e.g., human has base 1 for most stats)
- `Training` = stat points spent by the player via the `train` command
- `Mods` = sum of equipment statmods + buff statmods + pet statmods

Soft cap at 105: if `ValueAdj >= 105`, overage is compressed:
```
ValueAdj = 100 + round(sqrt(overage) * 2)
```

### HP Calculation

From `Character.RecalculateStats()`:

```
HealthMax.Mod = 5
             + StatMod("healthmax")       // from buffs/equipment
             + Level                       // +1 HP per level
             + ActionValueAdj("MaxHealth") * 4   // Vitality stat * 4

HealthMax.Value = HealthMax.Racial + HealthMax.Training + HealthMax.Mod
```

Since HealthMax has Base=1, Racial = floor((level-1) * 0.3333 * 1) + floor(level * 0.5).

In practice for a level 1 human with no gear:
- Vitality stat value = floor(0 * 0.3333 * 1) + floor(1 * 0.5) = 0 (both floor to 0)
- Wait, NaturalGainsModFactor: floor(1 * 0.5) = 0
- So Vitality ValueAdj = 0 + 0 + 0 = 0
- HealthMax.Mod = 5 + 0 + 1 + 0*4 = 6
- HealthMax.Racial for HealthMax (Base=1): floor(0 * 0.3333) + floor(0.5) = 0
- HealthMax.Value = 0 + 0 + 6 = 6
- Starting Health = 10 (capped at HealthMax), so actual = 6

**Important**: Starting health is set to 10 in code, but then capped by Validate() to
HealthMax, so a level 1 human with 0 training starts with **6 HP**.

### Mana Calculation

```
ManaMax.Mod = 4
            + StatMod("manamax")
            + Level
            + ActionValueAdj("MaxMana") * 3   // Mysticism stat * 3
```

### Hit Chance

From `combat.Hits()` and `combat.hitChance()`:

```
hitChance = 30 + (attackerHitChance / (attackerHitChance + defenderHitChance)) * 70
```

Where `attackerHitChance` and `defenderHitChance` are the Speed stat's
`ActionValueAdj("HitChance")` values.

Then apply any penalty (e.g., -35 for dual wield without max skill).

Clamped to range [5, 95]. Roll `Rand(100)` and hit if roll < hitChance.

**Key insight**: If attacker and defender have equal Speed, hitChance = 30 + 35 = 65%.
If the defender has 0 Speed (most insideout mobs), hitChance = 30 + 70 = 100% (capped at 95%).

### Damage Calculation

1. Get weapon dice: `attacks, dCount, dSides, bonus, critBuffs`
   - If no weapon equipped, use racial defaults (e.g., human = 1d3)
   - Then add stat bonuses:
     ```
     dCount += floor(ActionValueAdj("DiceCount") / 50)    // Speed stat / 50
     dSides += floor(ActionValueAdj("DiceSides") / 12)    // Strength stat / 12
     bonus  += floor(ActionValueAdj("DamageBonus") / 25)  // Perception stat / 25
     ```

2. For each attack swing, roll: `damage = RollDice(dCount, dSides) + bonus`

3. If crit: `damage += dCount * dSides + bonus` (adds max possible roll as bonus)

4. Apply defense reduction:
   ```
   defenseRoll = Rand(targetDefense)     // random 0 to defense-1
   reduction = round((defenseRoll / 100) * damage)
   finalDamage = damage - reduction
   ```
   Defense is a percentage (0-100) representing damage reduction chance.

### Attack Count (Multi-attacks)

```
attackCount = ceil((attackerAttackCount - defenderAttackCount) / 25)
```

Minimum 1. Additional attacks from statmods (buffs) are added on top.
AttackCount comes from the Speed stat.

### Critical Hits

From `combat.Crits()`:

```
levelDiff = max(attackerLevel - defenderLevel, 1)
critChance = 5 + round((CritPrimary + CritSecondary) / levelDiff)
```

Where CritPrimary = Strength stat, CritSecondary = Speed stat.
Doubled if attacker has Accuracy buff, halved if defender has Blink buff.
Minimum 5%.

On crit: full max damage added (dCount * dSides + bonus), plus any
critBuff effects applied to target.

Backstab = guaranteed crit on first attack.

### Defense (Armor)

`Character.GetDefense()` sums `DamageReduction` from all 10 equipment
slots (weapon, offhand, head, neck, body, belt, gloves, ring, legs, feet).

If offhand is a shield (non-weapon with DamageReduction > 0), total
defense gets a 1.5x multiplier.

Capped at 100.

Defense is used as: roll random 0 to defense, that percentage of damage
is blocked.

---

## Damage Formulas Summary

### Weapon Damage Per Hit (no crits, no stat bonuses)

```
damage = RollDice(dCount, dSides) + bonus
```

Where dCount, dSides, bonus come from the weapon's `diceroll` field.

### Full Damage Per Round

```
totalDamage = 0
for each attackCount:
    for each weapon (1 or 2 if dual wield):
        for each weapon.attacks (from dice format "2@1d6" = 2 attacks):
            if Hits():
                dmg = Roll(dCount, dSides) + bonus + statDamageBonus
                if Crits():
                    dmg += maxRoll + bonus
                dmg -= armorReduction
                totalDamage += dmg
```

### Average Damage Formula (simplified)

```
avgHit = dCount * (dSides + 1) / 2 + bonus
avgDamage = avgHit * hitChance * attackCount * weaponAttacks
```

---

## Player Progression Reference

Assumptions: Human race (base 1 for Str/Spd/Sma/Vit/Per, base 1 for Spd),
no stat training, no buffs, using level-appropriate weapon from the
Inside Out world.

### Stat Value by Level (Human, base 1, no training)

```
StatValue = floor((level - 1) * 0.3333 * 1) + floor(level * 0.5) + mods
```

| Level | Racial Calc | Stat Value (no mods) |
|-------|------------|---------------------|
| 1 | floor(0) + floor(0.5) = 0 | 0 |
| 2 | floor(0.333) + floor(1) = 0 + 1 = 1 | 1 |
| 3 | floor(0.666) + floor(1.5) = 0 + 1 = 1 | 1 |
| 4 | floor(1.0) + floor(2.0) = 1 + 2 = 3 | 3 |
| 5 | floor(1.333) + floor(2.5) = 1 + 2 = 3 | 3 |
| 7 | floor(2.0) + floor(3.5) = 2 + 3 = 5 | 5 |
| 10 | floor(3.0) + floor(5.0) = 3 + 5 = 8 | 8 |
| 15 | floor(4.666) + floor(7.5) = 4 + 7 = 11 | 11 |
| 20 | floor(6.333) + floor(10) = 6 + 10 = 16 | 16 |

### HP by Level (Human, Vitality base 1, no training/gear)

```
VitalityStat = see table above
HealthMax.Mod = 5 + Level + VitalityStat * 4
HealthMax.Racial = floor((level-1) * 0.3333) + floor(level * 0.5)  [HealthMax Base=1]
HealthMax = HealthMax.Racial + HealthMax.Mod
```

| Level | Vit Stat | HM.Mod | HM.Racial | HP |
|-------|----------|--------|-----------|-----|
| 1 | 0 | 5+1+0 = 6 | 0 | **6** |
| 2 | 1 | 5+2+4 = 11 | 1 | **12** |
| 3 | 1 | 5+3+4 = 12 | 1 | **13** |
| 5 | 3 | 5+5+12 = 22 | 3 | **25** |
| 7 | 5 | 5+7+20 = 32 | 5 | **37** |
| 10 | 8 | 5+10+32 = 47 | 8 | **55** |
| 15 | 11 | 5+15+44 = 64 | 11 | **75** |
| 20 | 16 | 5+20+64 = 89 | 16 | **105** |

With gear adding Vitality, these numbers climb significantly. A level 10
player with +5 Vitality from gear would have Vit 13 -> HP ~67.

### Damage Output by Level (Human, level-appropriate weapon, no stat training)

At level 1 unarmed (1d3, no stat bonuses):
- Avg damage per hit: 2
- Hit chance vs 0-speed mob: 95%
- Avg damage per round: ~1.9

With Pencil (1d2):
- Avg damage per hit: 1.5
- Avg per round: ~1.4

With Hockey Stick (1d4) at level 3-5:
- Player Speed stat ~3, Str ~3, Per ~3
- dCount bonus: floor(3/50) = 0, dSides bonus: floor(3/12) = 0, dmgBonus: floor(3/25) = 0
- Avg damage per hit: 2.5
- Avg per round: ~2.4

With Memory Orb Shard (1d4+1) at level 5-7:
- Avg damage per hit: 3.5
- Avg per round: ~3.3

With Imagination Sword (1d6) at level 7-10:
- Player Speed ~5, Str ~5, Per ~5
- dCount bonus: 0, dSides bonus: 0, dmgBonus: 0 (need 12/25/50 to gain +1)
- Avg damage per hit: 3.5
- Avg per round: ~3.3

With Dream Prop Blade (1d6+1) at level 10-12:
- Player Speed ~8, Str ~8
- dSides bonus: 0, dmgBonus: 0
- Avg damage per hit: 4.5
- Avg per round: ~4.3

With Abstract Edge (1d8) at level 12-15:
- Player Speed ~11, Str ~11, Per ~11
- dSides bonus: 0, dmgBonus: 0 (need 12 for +1 side)
- Avg damage per hit: 4.5
- Avg per round: ~4.3

With Fear Fang Dagger (1d8+1) at level 14-17:
- Avg damage per hit: 5.5

With Void Shard Blade (2d6, +2 Str) at level 18-20:
- Player Speed ~16, Str ~18 (with +2 from weapon)
- dSides bonus: floor(18/12) = 1 -> 2d7
- Avg damage per hit: 8
- Avg per round: ~7.6

---

## Mob Balance Guidelines

### Key Balance Principles

1. Mobs have NO equipment unless explicitly given in YAML (most insideout
   hostile mobs have none).
2. Mob damage comes from their **racial unarmed dice** plus stat bonuses
   from their level.
3. Mob HP uses the same formula as players.
4. Most insideout hostile mobs use races with limited or no equipment slots
   (fear_construct, forgotten_memory, abstract_being), meaning they rely
   purely on racial stats and level.

### The Critical Problem: Mob Stats Scale with Level via Race Base Stats

A mob's effective stats are calculated the same as a player's:
```
StatValue = floor((level - 1) * 0.3333 * RaceBase) + floor(level * 0.5)
```

For a level 5 mob using **fear_construct** (race 5):
- Strength base 1: floor(4 * 0.333 * 1) + floor(2.5) = 1 + 2 = 3
- Speed base 2: floor(4 * 0.333 * 2) + floor(2.5) = 2 + 2 = 4
- Perception base 1: floor(4 * 0.333) + floor(2.5) = 1 + 2 = 3
- Vitality base 0: floor(0) + floor(2.5) = 0 + 2 = 2
- Unarmed: 1d5 (racial) + stat bonuses

For a level 5 mob using **forgotten_memory** (race 7):
- Mysticism base 1: 1 + 2 = 3
- All other stats base 0: 0 + 2 = 2
- Unarmed: 1d4 + stat bonuses
- Vitality base 0: stat = 2 -> HP = 5 + 5 + 2*4 + (floor(4*0.333)+floor(2.5)) = 5+5+8+3 = 21

### Recommended Mob Stats by Level Tier

Target: Player should need 3-6 hits to kill a normal mob, 10-15 for a boss.
Player should survive 5-8 hits from a normal mob, 4-6 from a boss.

| Tier | Mob Level | Target HP | Target Dmg/Hit | Player Hits to Kill | Mob Hits to Kill Player |
|------|-----------|-----------|----------------|--------------------|-----------------------|
| Starter (Riley's Neighborhood) | 1-3 | 6-15 | 1-2 | 3-5 | 4-8 |
| Early (School, LTM) | 4-7 | 20-35 | 2-4 | 5-8 | 6-10 |
| Mid (Imagination, Dream) | 8-12 | 40-60 | 4-6 | 6-10 | 6-8 |
| Upper (Abstract, Subconscious) | 12-16 | 55-80 | 5-8 | 8-12 | 5-7 |
| High (Memory Dump, Back of Mind) | 16-20 | 75-110 | 7-10 | 10-14 | 5-7 |
| Boss (Jangles, Corrupted Core) | 20-30 | 120-200 | 8-12 | 15-25 | 6-8 |

---

## Inside Out World Mob Audit

### Estimated Mob HP Calculation

Using the formula for each mob, based on race Vitality base and mob level.
No equipment on most hostile mobs means no gear HP bonuses.

**Formula reminder**:
```
VitStat = floor((level-1) * 0.3333 * vit_base) + floor(level * 0.5)
HP = floor((level-1)*0.3333*1) + floor(level*0.5)      // HealthMax.Racial (base=1)
   + 5 + level + VitStat * 4                            // HealthMax.Mod
```

### Riley's Neighborhood (Intended: Level 1-3)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| thought worm | 1 | forgotten_memory (vit 0) | Yes | 0 | 6 | 1d4 (avg 2.5) | OK for starter |
| neighbor kid | 3 | human (vit 1) | No | 1 | 13 | 1d3 | Non-hostile, fine |
| rileys_dad | 20 | human | No | 1 | 105 | 1d3 | Non-hostile NPC |
| rileys_mom | 20 | human | No | 1 | 105 | 1d3 | Non-hostile NPC |
| shopkeeper | 10 | human | No | 1 | 55 | 1d3 | Non-hostile NPC |
| hockey_coach | 15 | human | No | 1 | 75 | 1d3 | Non-hostile NPC |

**Assessment**: Thought worm at level 1 is correct for the starter zone. Only
one hostile mob in the zone, which is appropriate.

### School (Intended: Level 4-7)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| doubt shadow | 5 | fear_construct (vit 0) | Yes | 0 | 18 | 1d5 (avg 3) | **See analysis below** |
| anxiety spiral | 6 | fear_construct (vit 0) | Yes | 0 | 20 | 1d5 (avg 3) | **See analysis below** |
| teacher | 15 | human | No | 1 | 75 | 1d3 | Non-hostile NPC |
| principal | 20 | human | No | 1 | 105 | 1d3 | Non-hostile NPC |
| classmate | 5 | human | No | 1 | 25 | 1d3 | Non-hostile |
| lunch_lady | 10 | human | No | 1 | 55 | 1d3 | Non-hostile NPC |
| janitor | 10 | human | No | 1 | 55 | 1d3 | Non-hostile NPC |

**Doubt Shadow Analysis (level 5, fear_construct)**:
- Fear construct: Str base 1, Speed base 2, Per base 1, Vit base 0
- Speed stat at level 5: floor(4 * 0.333 * 2) + floor(2.5) = 2 + 2 = 4
- Str stat: floor(4*0.333) + floor(2.5) = 1 + 2 = 3
- Per stat: 1 + 2 = 3
- Unarmed: 1d5 + dCount(floor(4/50)=0) + dSides(floor(3/12)=0) + bonus(floor(3/25)=0)
- Avg damage: 3 per hit
- HitChance vs level 1 player (Speed 0): 95% (asymmetric advantage)
- Has `backstab` combat command = **guaranteed crit on first hit = 3 + 5 + 0 = 8 damage**
- Level 1 player HP: **6**

**CRITICAL ISSUE**: A doubt shadow's backstab crit can deal 8 damage to a
player with 6 HP. That is an instant kill. Even a normal hit of 3-5 takes
more than half the player's health. A level 1 player entering School will
be killed in 1-2 hits.

### Long Term Memory (Intended: Level 5-8)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| memory leech | 5 | forgotten_memory (vit 0) | Yes | 0 | 18 | 1d4 (avg 2.5) | **Has backstab** |
| fading memory | 6 | forgotten_memory (vit 0) | Yes | 0 | 20 | 1d4 (avg 2.5) | Has backstab |
| gloom raider | 7 | fear_construct (vit 0) | Yes | 0 | 22 | 1d5 (avg 3) | Has backstab |
| mind worker | 8 | memory_worker (vit 0) | No | 0 | 24 | 1d3 | Shop NPC |
| bing bong | 99 | imaginary_friend | No | 1 | ~537 | 1d4 | Quest NPC |

**Memory Leech Analysis (level 5, forgotten_memory)**:
- forgotten_memory: Mys base 1, all others base 0
- Speed stat: floor(0) + floor(2.5) = 2
- Str stat: 0 + 2 = 2
- Unarmed: 1d4
- Avg damage: 2.5 per normal hit
- Has `backstab`: crit = 2.5 + 4 + 0 = ~6.5 damage
- **This will one-shot a level 1 player (6 HP) on a backstab crit**
- Even a level 3 player (~13 HP) would lose half their health to one backstab

**This matches the reported bug**: Player being one-shotted by a memory leech.

### Imagination Land (Intended: Level 7-10)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| corrupted imaginary friend | 8 | imaginary_friend (vit 1) | Yes | 1 | 32 | 1d4 (avg 2.5) | Has backstab |
| gloom raider scout | 7 | fear_construct (vit 0) | Yes | 0 | 22 | 1d5 (avg 3) | Has backstab |
| imaginary boyfriend | 5 | imaginary_friend | No | 1 | 25 | 1d4 | Non-hostile |
| cloud builder | 12 | imaginary_friend | No | 1 | 50 | 1d4 | Shop NPC |

### Dream Productions (Intended: Level 9-12)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| dream nightmare | 10 | fear_construct (vit 0) | Yes | 0 | 30 | 1d5 (avg 3) | Has backstab |
| corrupted dream | 11 | forgotten_memory (vit 0) | Yes | 0 | 32 | 1d4 (avg 2.5) | Has backstab |
| dream director | 20 | memory_worker (vit 0) | No | 0 | 56 | 1d3 | Quest NPC |
| dream actor | 8 | memory_worker (vit 0) | No | 0 | 24 | 1d3 | Non-hostile |
| stage hand | 10 | memory_worker (vit 0) | No | 0 | 30 | 1d3 | Shop NPC |

### Abstract Thought (Intended: Level 10-14)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| geometric fragment | 10 | abstract_being (vit 0) | Yes | 0 | 30 | 1d6 (avg 3.5) | Has backstab |
| abstract form | 12 | abstract_being (vit 0) | Yes | 0 | 35 | 1d6 (avg 3.5) | Has backstab |
| perspective shifter | 11 | abstract_being (vit 0) | Yes | 0 | 32 | 1d6 (avg 3.5) | Has backstab |
| non-figurative horror | 14 | abstract_being (vit 0) | Yes | 0 | 41 | 1d6 (avg 3.5) | Boss-like, backstab + cast mm |

### The Subconscious (Intended: Level 12-16)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| broccoli horror | 12 | fear_construct (vit 0) | Yes | 0 | 35 | 1d5 (avg 3) | |
| darkness creeper | 13 | fear_construct (vit 0) | Yes | 0 | 38 | 1d5 (avg 3) | |
| fear spider | 14 | fear_construct (vit 0) | Yes | 0 | 41 | 1d5 (avg 3) | Has backstab |
| doubt shadow (sub) | 15 | fear_construct (vit 0) | Yes | 0 | 44 | 1d5 (avg 3) | backstab + cast mm |
| **jangles the clown** | **30** | fear_construct (vit 0) | Yes | 0 | **86** | **2d6 weapon** (avg 7) | **BOSS**: backstab + cast mm + callforhelp, has joy buzzer, 500g |

**Jangles Analysis**:
- Level 30, fear_construct race
- Speed base 2: floor(29 * 0.333 * 2) + floor(15) = 19 + 15 = 34
- Str base 1: floor(29*0.333) + floor(15) = 9 + 15 = 24
- Per base 1: 9 + 15 = 24
- Weapon: Jangles' Joy Buzzer (2d6, avg 7)
- Stat bonuses: dCount+0 (34/50=0), dSides+2 (24/12=2), bonus+0 (24/25=0)
- Effective weapon: 2d8, avg 9 per hit
- Attack count: ceil((34 - playerAttackCount) / 25) = at least 1, probably 2 vs low-level
- Backstab crit: 9 + 16 = 25 damage
- This is a proper boss. A level 15 player with ~75 HP should survive 3-4 hits.

### Memory Dump (Intended: Level 16-20)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| dissolving shade | 16 | forgotten_memory (vit 0) | Yes | 0 | 46 | 1d4 (avg 2.5) | |
| forgotten memory | 17 | forgotten_memory (vit 0) | Yes | 0 | 49 | 1d4 (avg 2.5) | |
| memory wraith | 19 | forgotten_memory (vit 0) | Yes | 0 | 55 | 1d4 (avg 2.5) | backstab + cast mm |
| **corrupted core memory** | **20** | forgotten_memory (vit 0) | Yes | 0 | **58** | 1d4 (avg 2.5) | **BOSS**: backstab + cast mm + cast sparks + callforhelp, 1000g |

**Corrupted Core Memory Issue**: At level 20 with forgotten_memory race
(1d4 unarmed), this boss does very low base damage. Its danger comes from
spells and calling for help, but a player at level 15+ may find it underwhelming
as a final boss in terms of raw physical threat.

### The Back of the Mind (Intended: Level 15-18)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| anxiety spiral (BotM) | 16 | fear_construct (vit 0) | Yes | 0 | 46 | 1d5 (avg 3) | backstab + cast mm |
| doubt shadow elder | 17 | fear_construct (vit 0) | Yes | 0 | 49 | 1d5 (avg 3) | backstab + cast mm |
| suppression guard | 18 | abstract_being (vit 0) | Yes | 0 | 52 | 1d6 (avg 3.5) | backstab + cast mm |

### Belief System (Intended: Level 11-14)

| Mob | Level | Race | Hostile | Vit Base | Est. HP | Est. Dmg | Notes |
|-----|-------|------|---------|----------|---------|----------|-------|
| crumbling belief | 11 | abstract_being (vit 0) | Yes | 0 | 32 | 1d6 (avg 3.5) | |
| toxic thought | 12 | fear_construct (vit 0) | Yes | 0 | 35 | 1d5 (avg 3) | Has backstab |
| identity guardian | 14 | emotion (vit 0) | No | 0 | 41 | 1d6 (avg 3.5) | Non-hostile |

### Headquarters (Level 99 NPCs, non-hostile)

All emotions (Joy, Sadness, Anger, Fear, Disgust, Anxiety, Envy, Ennui,
Embarrassment, Nostalgia) are level 99, emotion race, non-hostile. They
have `callforhelp` combat commands as a defensive measure. These are
functioning as intended as unkillable quest NPCs.

### Tutorial

| Mob | Level | Race | Hostile | Notes |
|-----|-------|------|---------|-------|
| orb of knowledge | 1 | orb (all slots disabled) | No | Tutorial guide |
| training dummy | 1 | dummy (all slots disabled, -5 vit training) | No | Very low HP on purpose |

---

## Weapon Progression

Listed by intended progression order, with damage dice and estimated
average damage.

| Weapon | Dice | Avg Dmg | Value | Intended Level | Zone |
|--------|------|---------|-------|---------------|------|
| Pencil | 1d2 | 1.5 | 5 | 1-2 | Riley's Neighborhood |
| Rubber Band Slingshot | 1d3 | 2 | 10 | 1-3 | Riley's Neighborhood |
| Textbook | 1d3 | 2 | 15 | 2-4 | School |
| Hockey Stick | 1d4 (2H) | 2.5 | 25 | 3-5 | Riley's Neighborhood |
| Memory Orb Shard | 1d4+1 | 3.5 | 40 | 5-7 | Long Term Memory |
| Imagination Sword | 1d6 | 3.5 | 75 | 7-9 | Imagination Land |
| Broccoli Flail | 1d6 | 3.5 | 80 | 7-9 | The Subconscious |
| Dream Prop Blade | 1d6+1 | 4.5 | 100 | 9-11 | Dream Productions |
| Bing Bong's Candy Cane | 2d4 (2H) | 5 | 120 | 8-10 | Long Term Memory (quest) |
| Abstract Edge | 1d8 | 4.5 | 150 | 11-13 | Abstract Thought |
| Emotion Amplifier Wand | 1d6+2 (+5 Mys) | 5.5 | 200 | 10-14 | Dream Productions |
| Fear Fang Dagger | 1d8+1 (+1 Spd) | 5.5 | 200 | 13-16 | The Subconscious |
| Jangles' Joy Buzzer | 2d6 | 7 | 350 | 16+ | The Subconscious (boss drop) |
| Core Memory Staff | 2d5 (2H, +3 Sma) | 6 | 400 | 15+ | Quest/Rare |
| Void Shard Blade | 2d6 (+2 Str) | 7 | 500 | 18+ | Memory Dump |

### Progression Curve Issues

The weapon curve is reasonable overall. Notable gaps:

1. **Level 1-3 gap**: Only the pencil (1d2) and slingshot (1d3) are
   available. Both are very weak. The hockey stick (1d4) requires 2 hands
   and costs 25g. There is nothing between 1d2 and 1d4 for one-handed
   weapons.

2. **Level 13-16 gap**: Between the Abstract Edge (1d8) and Fear Fang
   Dagger (1d8+1), there is almost no progression. The fear fang is only
   marginally better.

3. **Endgame weapons are close together**: Joy Buzzer (2d6), Core Memory
   Staff (2d5), and Void Shard Blade (2d6) are all very similar in damage.

---

## Recommended Adjustments

### CRITICAL: Backstab on Low-Level Mobs

**Problem**: Many low-level hostile mobs (memory leech level 5, doubt shadow
level 5, fading memory level 6) have `backstab` as a combat command. This
gives them a free guaranteed critical hit, which can one-shot or nearly
one-shot low-level players.

A memory leech backstab crit = ~6-8 damage vs a level 1 player with 6 HP.

**Fix Options** (pick one or combine):

1. **Remove backstab from mobs below level 8**: Edit the following mob files
   to remove `backstab` from their `combatcommands`:
   - `mobs/long_term_memory/31-memory_leech.yaml`
   - `mobs/long_term_memory/35-fading_memory.yaml`
   - `mobs/school/13-doubt_shadow.yaml`
   - `mobs/school/14-anxiety_spiral.yaml`

2. **Lower these mobs' levels to 2-3**: This reduces their stats and HP,
   making them appropriate for the players who encounter them first.

3. **Add a starter zone buffer**: Ensure there is a clear level 1-3 zone
   (Riley's Neighborhood) with only level 1-3 hostile mobs before the
   player ever encounters level 5+ mobs.

### Specific Mob Adjustments

#### Memory Leech (mob 31) -- REPORTED BUG
- **Current**: Level 5, forgotten_memory race, hostile, backstab
- **Problem**: One-shots level 1 players with backstab crit
- **Recommendation**: Lower to level 3, remove backstab, OR gate Long Term
  Memory behind a level 3+ quest requirement. The zone feels intended for
  levels 5-8 but there is nothing stopping a level 1 from wandering in.

#### Doubt Shadow (mob 13, School)
- **Current**: Level 5, fear_construct race, hostile, backstab
- **Problem**: Fear construct has Speed base 2, making it fast AND it backstabs.
  A level 1 player will be destroyed.
- **Recommendation**: Lower to level 3 and remove backstab, OR keep at level 5
  but remove backstab and add a quest gate to the School hallways.

#### Anxiety Spiral (mob 14, School)
- **Current**: Level 6, fear_construct, hostile, backstab
- **Problem**: Same as doubt shadow but slightly worse.
- **Recommendation**: Lower to level 4, remove backstab.

#### Gloom Raider (mob 34, LTM)
- **Current**: Level 7, fear_construct, hostile, backstab
- **Problem**: At level 7 with Speed base 2, this mob has aggressive stats
  for the zone it shares with level 5 memory leeches.
- **Recommendation**: Appropriate for level 7 if players are gated properly.
  Consider removing backstab to reduce spike damage.

#### Corrupted Core Memory (mob 56, Memory Dump) -- BOSS
- **Current**: Level 20, forgotten_memory race, 1d4 unarmed damage
- **Problem**: As the "final boss" of the Memory Dump (level 16-20 zone with
  1000g drop and 50% item drop chance), it does pathetically low physical
  damage. The forgotten_memory race has 1d4 unarmed and almost no combat
  stats (only Mysticism base 1).
- **Recommendation**: Either:
  - Give it a weapon (e.g., Void Shard Blade or a custom boss weapon)
  - Increase level to 25
  - Change race to abstract_being (1d6 base) or fear_construct (1d5 base)
  - Give it stat training in Strength and Speed

#### Jangles the Clown (mob 49, Subconscious) -- BOSS
- **Current**: Level 30, fear_construct, has Joy Buzzer weapon (2d6), backstab + cast mm + callforhelp
- **Assessment**: Well-designed boss. Level 30 is significantly above the zone
  (level 12-16). With 2d6 weapon and fear_construct Speed base 2, Jangles
  hits hard and has multi-attack potential. The callforhelp bringing in fear
  mobs adds danger. This is the strongest boss in the world and feels
  appropriate as a major challenge.

#### Non-Figurative Horror (mob 47, Abstract Thought)
- **Current**: Level 14, abstract_being, backstab + cast mm
- **Assessment**: Functions as a mini-boss for Abstract Thought. Stats are
  reasonable. The combination of backstab + magic makes it dangerous.
  Appropriate for its zone.

### Zone Flow Recommendations

The world currently has no hard gates between zones. A fresh level 1
player can walk from Riley's Neighborhood directly to Long Term Memory or
School and encounter level 5-7 mobs with backstab. This is the root cause
of the reported one-shot issue.

**Recommended zone level ranges and flow**:

```
Tutorial (Level 1) -> Riley's Neighborhood (Level 1-3)
  -> School (Level 3-6)
  -> Headquarters (safe hub)
     -> Long Term Memory (Level 5-8)
     -> Imagination Land (Level 7-10)
     -> Dream Productions (Level 9-12)
     -> Belief System (Level 11-14)
     -> Abstract Thought (Level 12-15)
     -> The Subconscious (Level 13-16, boss: Jangles level 30)
     -> The Back of the Mind (Level 15-18)
     -> Memory Dump (Level 16-20, boss: Corrupted Core level 20)
```

Consider adding level-check scripts on zone transitions that warn low-level
players, or adding quest prerequisites to reach deeper zones.

### Global Backstab Assessment

Backstab on hostile mobs is extremely dangerous because:
1. It triggers as a combat command during normal rounds
2. It is a guaranteed critical (doubles damage)
3. Mobs use it unpredictably during combat, not just on opener

**Recommended backstab distribution**:
- Levels 1-4: NO backstab on any mob
- Levels 5-8: Backstab on max 1-2 mobs per zone, as a zone "elite" variant
- Levels 9-14: Backstab on most hostile mobs (current state is fine)
- Levels 15+: Backstab + cast combinations on dangerous mobs

### Summary of Highest Priority Changes

1. **Remove backstab from memory leech** (`mobs/long_term_memory/31-memory_leech.yaml`) -- this is the reported bug
2. **Remove backstab from doubt shadow** (`mobs/school/13-doubt_shadow.yaml`)
3. **Remove backstab from anxiety spiral** (`mobs/school/14-anxiety_spiral.yaml`)
4. **Remove backstab from fading memory** (`mobs/long_term_memory/35-fading_memory.yaml`)
5. **Give the corrupted core memory a weapon or stat training** to make the endgame boss feel threatening
6. **Add 1-2 more level 2-3 hostile mobs to Riley's Neighborhood** so players have something to fight before heading to School/LTM

---

## Appendix: Race Reference

| ID | Name | Vit Base | Notable Stats | Unarmed | Size |
|----|------|----------|---------------|---------|------|
| 0 | ghostly spirit | 0 | none | 1d1 | small |
| 1 | human | 1 | all base 1 | 1d3 | medium |
| 2 | emotion | 0 | Sma 2, Per 2 | 1d6 | medium |
| 3 | memory worker | 0 | Sma 1 | 1d3 | medium |
| 4 | abstract being | 0 | Sma 2, Mys 1 | 1d6 | large |
| 5 | fear construct | 0 | Str 1, **Spd 2**, Per 1 | 1d5 | large |
| 6 | imaginary friend | 0 | Vit 1, Per 1 | 1d4 | medium |
| 7 | forgotten memory | 0 | Mys 1 | 1d4 | medium |
| 19 | dummy | 0 | none | 1d1 | medium |
| 20 | orb | 0 | none | 1d1 | small |

Note: Fear construct (race 5) is the most combat-capable race due to Speed
base 2, which maps to HitChance, DiceCount, AttackCount, and CritSecondary.
This makes all fear_construct mobs disproportionately dangerous compared to
their level.

---

*Last updated: 2026-03-29*

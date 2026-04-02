#!/usr/bin/env python3
"""
GoMud Mob Balance Checker

Flags mobs that are out of bounds for their zone's autoscale range,
missing stat training, using overpowered combat commands at low levels,
or otherwise unbalanced.

Usage:
    python3 check_mobs.py [--base-dir <dir>]
"""

import argparse
import glob
import os
import sys
import yaml


BACKSTAB_MIN_LEVEL = 15
BOSS_MIN_LEVEL = 20

OPPOSITE_RACES_SPEED = {
    'fear construct': 2,  # Speed base 2 - already dangerous
}


def load_zones(base_dir):
    zones = {}
    for f in glob.glob(os.path.join(base_dir, 'rooms', '*', 'zone-config.yaml')):
        with open(f) as fh:
            data = yaml.safe_load(fh)
        if data and 'name' in data:
            zones[data['name']] = {
                'min': data.get('autoscale', {}).get('minimum'),
                'max': data.get('autoscale', {}).get('maximum'),
            }
    return zones


def load_races(base_dir):
    races = {}
    for f in glob.glob(os.path.join(base_dir, 'races', '*.yaml')):
        with open(f) as fh:
            data = yaml.safe_load(fh)
        if data:
            races[data['raceid']] = {
                'name': data['name'],
                'stats': {k: v.get('base', 0) for k, v in data.get('stats', {}).items()},
                'damage': data.get('damage', {}).get('diceroll', '1d1'),
            }
    return races


def check_mobs(base_dir):
    zones = load_zones(base_dir)
    races = load_races(base_dir)
    errors = []
    warnings = []

    for f in sorted(glob.glob(os.path.join(base_dir, 'mobs', '**', '*.yaml'), recursive=True)):
        with open(f) as fh:
            data = yaml.safe_load(fh)
        if not data or 'character' not in data:
            continue

        char = data['character']
        mobid = data['mobid']
        name = char.get('name', '?')
        level = char.get('level', 1)
        zone = data.get('zone', '?')
        hostile = data.get('hostile', False)
        race_id = char.get('raceid', 0)
        race = races.get(race_id, {'name': '?', 'stats': {}, 'damage': '1d1'})
        training = {k: v.get('training', 0) for k, v in char.get('stats', {}).items()}
        combat = data.get('combatcommands', [])
        equipment = char.get('equipment', {})

        tag = f"#{mobid} {name} (lvl {level}, {zone})"

        if not hostile:
            continue

        # Check 1: Level vs zone autoscale
        zinfo = zones.get(zone, {})
        zmin = zinfo.get('min')
        zmax = zinfo.get('max')
        if zmin is not None and level < zmin:
            warnings.append(f"{tag}: level {level} below zone minimum {zmin}")
        if zmax is not None and level > zmax + 10:
            errors.append(f"{tag}: level {level} WAY above zone maximum {zmax}")
        elif zmax is not None and level > zmax:
            warnings.append(f"{tag}: level {level} above zone maximum {zmax}")

        # Check 2: No stat training on hostile mob
        if not training and level > 1:
            errors.append(f"{tag}: no stat training (will be weak for its level)")

        # Check 3: Backstab on low-level non-boss mobs
        if 'backstab' in combat and level < BACKSTAB_MIN_LEVEL:
            errors.append(f"{tag}: has backstab at level {level} (min recommended: {BACKSTAB_MIN_LEVEL})")

        # Check 4: Boss mob without proper stats
        if level >= BOSS_MIN_LEVEL and not equipment and hostile:
            warnings.append(f"{tag}: high-level mob with no equipment")

        # Check 5: Vitality training check - too low means instant death
        vit_train = training.get('vitality', 0)
        if level >= 10 and vit_train < 5 and hostile:
            warnings.append(f"{tag}: low vitality training ({vit_train}) for level {level}")

        # Check 6: Fear construct speed warning
        race_speed = race['stats'].get('speed', 0)
        if race_speed >= 2 and level <= 8:
            warnings.append(f"{tag}: race '{race['name']}' has speed base {race_speed} - very dangerous at low levels")

    return errors, warnings


def main():
    parser = argparse.ArgumentParser(description='GoMud Mob Balance Checker')
    parser.add_argument('--base-dir', default=os.path.join(os.path.dirname(__file__), '..'))
    args = parser.parse_args()
    base_dir = os.path.abspath(args.base_dir)

    print(f"Checking mobs in: {base_dir}\n")

    errors, warnings = check_mobs(base_dir)

    if errors:
        print(f"ERRORS ({len(errors)}):")
        for e in errors:
            print(f"  {e}")
        print()

    if warnings:
        print(f"WARNINGS ({len(warnings)}):")
        for w in warnings:
            print(f"  {w}")
        print()

    total = len(errors) + len(warnings)
    if total == 0:
        print("All mobs balanced.")
    else:
        print(f"Total: {len(errors)} errors, {len(warnings)} warnings")

    sys.exit(1 if errors else 0)


if __name__ == '__main__':
    main()

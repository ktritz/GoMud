#!/usr/bin/env python3
"""
Room Grid Tool

Manages spatial grid definitions for GoMud zones.

Commands:
    generate    Walk exits from room YAML files and generate grid.yaml per zone
    validate    Check room exits against grid.yaml for consistency
    sync        Update room YAML exit blocks to match grid.yaml
    show        Print ASCII map of a zone's grid

Usage:
    python3 tools/grid_tool.py generate [--base-dir <dir>] [--zone <name>]
    python3 tools/grid_tool.py validate [--base-dir <dir>] [--zone <name>]
    python3 tools/grid_tool.py sync [--base-dir <dir>] [--zone <name>] [--dry-run]
    python3 tools/grid_tool.py show [--base-dir <dir>] --zone <name>
"""

import sys
import os
import argparse
import glob
from collections import deque, defaultdict

import yaml


DIRECTION_VECTORS = {
    "north": (0, 1),
    "south": (0, -1),
    "east": (1, 0),
    "west": (-1, 0),
    "northeast": (1, 1),
    "southwest": (-1, -1),
    "northwest": (-1, 1),
    "southeast": (1, -1),
}

DIRECTION_OPPOSITES = {
    "north": "south", "south": "north",
    "east": "west", "west": "east",
    "northeast": "southwest", "southwest": "northeast",
    "northwest": "southeast", "southeast": "northwest",
}

CARDINAL_DIRS = set(DIRECTION_VECTORS.keys())


def load_rooms(base_dir):
    rooms = {}
    rooms_dir = os.path.join(base_dir, "rooms")
    for zone_dir in sorted(os.listdir(rooms_dir)):
        zone_path = os.path.join(rooms_dir, zone_dir)
        if not os.path.isdir(zone_path):
            continue
        for filepath in sorted(glob.glob(os.path.join(zone_path, "*.yaml"))):
            fn = os.path.basename(filepath)
            if fn in ("zone-config.yaml", "grid.yaml"):
                continue
            with open(filepath) as f:
                data = yaml.safe_load(f)
            if not data or "roomid" not in data:
                continue
            rid = data["roomid"]
            rooms[rid] = {
                "roomid": rid,
                "zone": data.get("zone", zone_dir),
                "zone_dir": zone_dir,
                "title": data.get("title", "(untitled)"),
                "exits": data.get("exits") or {},
                "file": filepath,
            }
    return rooms


def load_grid(grid_path):
    """Load a grid.yaml file. Returns (coords, special_exits)."""
    if not os.path.exists(grid_path):
        return None, None
    with open(grid_path) as f:
        data = yaml.safe_load(f)
    if not data:
        return {}, {}

    coords = {}
    for key, rid in (data.get("grid") or {}).items():
        parts = str(key).split(",")
        x, y = int(parts[0]), int(parts[1])
        coords[rid] = (x, y)

    special = data.get("special_exits") or {}
    return coords, special


def derive_grid_from_exits(rooms, zone_rooms):
    """BFS through cardinal exits to assign coordinates. Returns {roomid: (x,y)} and list of conflicts."""
    if not zone_rooms:
        return {}, []

    start = min(zone_rooms)
    coords = {start: (0, 0)}
    queue = deque([start])
    conflicts = []

    while queue:
        current = queue.popleft()
        cx, cy = coords[current]
        room = rooms.get(current)
        if not room:
            continue

        for direction, exit_data in room["exits"].items():
            if direction not in CARDINAL_DIRS:
                continue
            target = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target is None or target not in zone_rooms:
                continue

            dx, dy = DIRECTION_VECTORS[direction]
            expected = (cx + dx, cy + dy)

            if target in coords:
                if coords[target] != expected:
                    conflicts.append({
                        "room": target,
                        "title": rooms[target]["title"],
                        "existing": coords[target],
                        "expected": expected,
                        "from_room": current,
                        "from_title": rooms[current]["title"],
                        "direction": direction,
                    })
            else:
                coords[target] = expected
                queue.append(target)

    return coords, conflicts


def get_special_exits(rooms, zone_rooms, coords):
    """Identify non-cardinal exits for the zone."""
    special = defaultdict(dict)
    for rid in zone_rooms:
        room = rooms.get(rid)
        if not room:
            continue
        for direction, exit_data in room["exits"].items():
            if direction in CARDINAL_DIRS:
                continue
            if isinstance(exit_data, dict):
                special[rid][direction] = exit_data
            else:
                special[rid][direction] = {"roomid": exit_data}
    return dict(special)


def expected_exits_from_grid(coords, special_exits):
    """Given grid coordinates, compute what exits each room should have."""
    # Build reverse lookup: (x,y) -> roomid
    pos_to_room = {pos: rid for rid, pos in coords.items()}

    expected = defaultdict(dict)
    for rid, (x, y) in coords.items():
        for direction, (dx, dy) in DIRECTION_VECTORS.items():
            neighbor_pos = (x + dx, y + dy)
            if neighbor_pos in pos_to_room:
                target = pos_to_room[neighbor_pos]
                expected[rid][direction] = {"roomid": target}

        # Add special exits
        if rid in special_exits:
            for direction, exit_data in special_exits[rid].items():
                expected[rid][direction] = exit_data

    return dict(expected)


def write_grid_yaml(grid_path, coords, special_exits, rooms):
    """Write a grid.yaml file."""
    # Build grid dict sorted by roomid
    grid = {}
    for rid in sorted(coords.keys()):
        x, y = coords[rid]
        title = rooms[rid]["title"] if rid in rooms else "?"
        grid[f"{x},{y}"] = rid

    # Build comment map
    comments = {}
    for rid in sorted(coords.keys()):
        x, y = coords[rid]
        title = rooms[rid]["title"] if rid in rooms else "?"
        comments[f"{x},{y}"] = title

    lines = ["# Auto-generated grid layout", "#"]

    # ASCII map
    if coords:
        min_x = min(x for x, y in coords.values())
        max_x = max(x for x, y in coords.values())
        min_y = min(y for x, y in coords.values())
        max_y = max(y for x, y in coords.values())

        pos_to_room = {pos: rid for rid, pos in coords.items()}
        col_width = 8

        for y in range(max_y, min_y - 1, -1):
            row = f"# y{y:+d}: "
            for x in range(min_x, max_x + 1):
                if (x, y) in pos_to_room:
                    rid = pos_to_room[(x, y)]
                    cell = f"[{rid}]"
                else:
                    cell = ""
                row += cell.center(col_width)
            lines.append(row.rstrip())
        lines.append("#")

    lines.append("")
    lines.append("grid:")
    for rid in sorted(coords.keys()):
        x, y = coords[rid]
        key = f"{x},{y}"
        title = comments.get(key, "")
        lines.append(f'  "{key}": {rid}    # {title}')

    if special_exits:
        lines.append("")
        lines.append("special_exits:")
        for rid in sorted(special_exits.keys()):
            lines.append(f"  {rid}:")
            for direction, exit_data in sorted(special_exits[rid].items()):
                lines.append(f"    {direction}:")
                for k, v in exit_data.items():
                    if isinstance(v, bool):
                        lines.append(f"      {k}: {'true' if v else 'false'}")
                    else:
                        lines.append(f"      {k}: {v}")

    lines.append("")

    with open(grid_path, "w") as f:
        f.write("\n".join(lines))


def cmd_generate(base_dir, zone_filter):
    rooms = load_rooms(base_dir)

    # Group rooms by zone_dir
    zones = defaultdict(set)
    for rid, room in rooms.items():
        zones[room["zone_dir"]].add(rid)

    for zone_dir in sorted(zones.keys()):
        if zone_filter and zone_dir != zone_filter:
            continue

        zone_rooms = zones[zone_dir]
        coords, conflicts = derive_grid_from_exits(rooms, zone_rooms)
        special = get_special_exits(rooms, zone_rooms, coords)

        grid_path = os.path.join(base_dir, "rooms", zone_dir, "grid.yaml")

        if conflicts:
            print(f"\n{zone_dir}: {len(conflicts)} spatial conflicts!")
            for c in conflicts:
                print(f"  #{c['room']} \"{c['title']}\" at {c['existing']}, "
                      f"but #{c['from_room']} \"{c['from_title']}\" "
                      f"--{c['direction']}--> places it at {c['expected']}")

        # Check for rooms not reached by BFS
        unreached = zone_rooms - set(coords.keys())
        if unreached:
            print(f"\n{zone_dir}: {len(unreached)} rooms not reachable via cardinal exits:")
            for rid in sorted(unreached):
                print(f"  #{rid} \"{rooms[rid]['title']}\"")

        write_grid_yaml(grid_path, coords, special, rooms)
        zone_name = rooms[next(iter(zone_rooms))]["zone"] if zone_rooms else zone_dir
        print(f"{zone_dir}: wrote grid.yaml ({len(coords)} rooms, {len(special)} special exits)")


def cmd_validate(base_dir, zone_filter):
    rooms = load_rooms(base_dir)

    zones = defaultdict(set)
    for rid, room in rooms.items():
        zones[room["zone_dir"]].add(rid)

    total_errors = 0

    for zone_dir in sorted(zones.keys()):
        if zone_filter and zone_dir != zone_filter:
            continue

        grid_path = os.path.join(base_dir, "rooms", zone_dir, "grid.yaml")
        coords, special = load_grid(grid_path)
        if coords is None:
            print(f"{zone_dir}: no grid.yaml found, skipping")
            continue

        expected = expected_exits_from_grid(coords, special or {})
        zone_errors = 0

        for rid in sorted(zones[zone_dir]):
            room = rooms[rid]
            actual_exits = room["exits"]
            exp_exits = expected.get(rid, {})

            # Check for missing exits (in grid but not in room YAML)
            for direction, exit_data in exp_exits.items():
                target = exit_data.get("roomid")
                actual_target = None
                if direction in actual_exits:
                    ae = actual_exits[direction]
                    actual_target = ae.get("roomid") if isinstance(ae, dict) else ae

                if actual_target is None:
                    print(f"  MISSING: #{rid} \"{room['title']}\" should have {direction} -> #{target}")
                    zone_errors += 1
                elif actual_target != target:
                    print(f"  WRONG:   #{rid} \"{room['title']}\" {direction} -> #{actual_target}, expected #{target}")
                    zone_errors += 1

            # Check for extra cardinal exits not in grid
            for direction, exit_data in actual_exits.items():
                if direction not in CARDINAL_DIRS:
                    continue
                if direction not in exp_exits:
                    target = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
                    print(f"  EXTRA:   #{rid} \"{room['title']}\" has {direction} -> #{target} (not in grid)")
                    zone_errors += 1

        if zone_errors:
            print(f"{zone_dir}: {zone_errors} issues")
        else:
            print(f"{zone_dir}: OK")
        total_errors += zone_errors

    if total_errors:
        print(f"\n{total_errors} total issues found.")
        sys.exit(1)
    else:
        print("\nAll zones consistent.")


def cmd_sync(base_dir, zone_filter, dry_run):
    rooms = load_rooms(base_dir)

    zones = defaultdict(set)
    for rid, room in rooms.items():
        zones[room["zone_dir"]].add(rid)

    changes = 0

    for zone_dir in sorted(zones.keys()):
        if zone_filter and zone_dir != zone_filter:
            continue

        grid_path = os.path.join(base_dir, "rooms", zone_dir, "grid.yaml")
        coords, special = load_grid(grid_path)
        if coords is None:
            continue

        expected = expected_exits_from_grid(coords, special or {})

        for rid in sorted(zones[zone_dir]):
            room = rooms[rid]
            exp_exits = expected.get(rid, {})
            if not exp_exits:
                continue

            # Read the raw YAML
            with open(room["file"]) as f:
                raw = f.read()
            data = yaml.safe_load(raw)
            current_exits = data.get("exits") or {}

            # Build new exits: start with expected, preserve non-cardinal properties (lock, secret, etc.)
            new_exits = {}
            for direction, exit_data in exp_exits.items():
                new_exit = dict(exit_data)
                # Preserve extra properties from current YAML (lock, secret, etc.)
                if direction in current_exits:
                    curr = current_exits[direction]
                    if isinstance(curr, dict):
                        for k, v in curr.items():
                            if k != "roomid":
                                new_exit[k] = v
                new_exits[direction] = new_exit

            if new_exits == current_exits:
                continue

            changes += 1
            action = "WOULD UPDATE" if dry_run else "UPDATED"
            print(f"  {action}: #{rid} \"{room['title']}\"")

            # Show diff
            for d in sorted(set(list(new_exits.keys()) + list(current_exits.keys()))):
                old_target = None
                new_target = None
                if d in current_exits:
                    ce = current_exits[d]
                    old_target = ce.get("roomid") if isinstance(ce, dict) else ce
                if d in new_exits:
                    new_target = new_exits[d].get("roomid")
                if old_target != new_target:
                    if old_target is None:
                        print(f"    + {d} -> #{new_target}")
                    elif new_target is None:
                        print(f"    - {d} -> #{old_target}")
                    else:
                        print(f"    ~ {d}: #{old_target} -> #{new_target}")

            if not dry_run:
                data["exits"] = new_exits
                with open(room["file"], "w") as f:
                    yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

    if changes:
        print(f"\n{changes} rooms {'would be ' if dry_run else ''}updated.")
    else:
        print("All rooms already match their grids.")


def cmd_show(base_dir, zone_dir):
    rooms = load_rooms(base_dir)
    grid_path = os.path.join(base_dir, "rooms", zone_dir, "grid.yaml")
    coords, special = load_grid(grid_path)

    if coords is None:
        print(f"No grid.yaml found for {zone_dir}")
        sys.exit(1)

    if not coords:
        print("Grid is empty.")
        return

    pos_to_room = {pos: rid for rid, pos in coords.items()}

    min_x = min(x for x, y in coords.values())
    max_x = max(x for x, y in coords.values())
    min_y = min(y for x, y in coords.values())
    max_y = max(y for x, y in coords.values())

    # Find max label width
    labels = {}
    for rid, pos in coords.items():
        title = rooms[rid]["title"] if rid in rooms else f"#{rid}"
        # Truncate long titles
        if len(title) > 14:
            title = title[:13] + "…"
        labels[pos] = f"{rid}:{title}"

    col_width = max(len(v) for v in labels.values()) + 2

    print(f"\n  {zone_dir}")
    print()
    for y in range(max_y, min_y - 1, -1):
        # Room row
        row = ""
        for x in range(min_x, max_x + 1):
            pos = (x, y)
            if pos in labels:
                cell = f"[{labels[pos]}]"
            else:
                cell = ""
            row += cell.ljust(col_width)
        print(f"  {row.rstrip()}")

        # Connection row (vertical lines)
        if y > min_y:
            conn_row = ""
            for x in range(min_x, max_x + 1):
                pos = (x, y)
                below = (x, y - 1)
                if pos in pos_to_room and below in pos_to_room:
                    conn_row += "|".center(col_width)
                else:
                    conn_row += " " * col_width
            if conn_row.strip():
                print(f"  {conn_row.rstrip()}")

    # Show special exits
    if special:
        print()
        print("  Special exits:")
        for rid in sorted(special.keys()):
            for direction, exit_data in special[rid].items():
                target = exit_data.get("roomid", "?")
                title = rooms[rid]["title"] if rid in rooms else f"#{rid}"
                target_title = rooms[target]["title"] if target in rooms else f"#{target}"
                extra = ""
                if exit_data.get("secret"):
                    extra = " (secret)"
                print(f"    #{rid} \"{title}\" --{direction}--> #{target} \"{target_title}\"{extra}")
    print()


def main():
    parser = argparse.ArgumentParser(description="Room Grid Tool")
    parser.add_argument("command", choices=["generate", "validate", "sync", "show"])
    parser.add_argument("--base-dir", default=os.path.join(os.path.dirname(__file__), ".."))
    parser.add_argument("--zone", default=None, help="Only process this zone directory name")
    parser.add_argument("--dry-run", action="store_true", help="Show changes without writing (sync only)")
    args = parser.parse_args()

    base_dir = os.path.abspath(args.base_dir)

    if args.command == "generate":
        cmd_generate(base_dir, args.zone)
    elif args.command == "validate":
        cmd_validate(base_dir, args.zone)
    elif args.command == "sync":
        cmd_sync(base_dir, args.zone, args.dry_run)
    elif args.command == "show":
        if not args.zone:
            print("--zone is required for show command")
            sys.exit(1)
        cmd_show(base_dir, args.zone)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Room Exit Consistency Checker

Validates spatial consistency of room exits across all zones:
- One-way exits (A->B exists but B->A does not)
- Direction mismatches (A goes north to B, but B does not go south back)
- Spatial grid conflicts (room assigned two different coordinates)
- Cross-zone connections
- Orphan rooms (no incoming exits)
- Dead ends (only one exit)

Usage:
    python3 tools/check_exits.py [--base-dir <dir>]
"""

import sys
import os
import argparse
import glob
from collections import defaultdict

import yaml


DIRECTION_OPPOSITES = {
    "north": "south",
    "south": "north",
    "east": "west",
    "west": "east",
    "up": "down",
    "down": "up",
    "northeast": "southwest",
    "southwest": "northeast",
    "northwest": "southeast",
    "southeast": "northwest",
}

DIRECTION_VECTORS = {
    "north": (0, 1, 0),
    "south": (0, -1, 0),
    "east": (1, 0, 0),
    "west": (-1, 0, 0),
    "northeast": (1, 1, 0),
    "southwest": (-1, -1, 0),
    "northwest": (-1, 1, 0),
    "southeast": (1, -1, 0),
    "up": (0, 0, 1),
    "down": (0, 0, -1),
}


def load_rooms(base_dir):
    """Load all room YAML files from rooms/{zone}/*.yaml, excluding zone-config.yaml."""
    rooms = {}
    rooms_dir = os.path.join(base_dir, "rooms")
    if not os.path.isdir(rooms_dir):
        print(f"Error: rooms directory not found at {rooms_dir}", file=sys.stderr)
        sys.exit(1)

    for zone_dir in sorted(os.listdir(rooms_dir)):
        zone_path = os.path.join(rooms_dir, zone_dir)
        if not os.path.isdir(zone_path):
            continue
        for filepath in sorted(glob.glob(os.path.join(zone_path, "*.yaml"))):
            filename = os.path.basename(filepath)
            if filename == "zone-config.yaml":
                continue
            with open(filepath, "r") as f:
                data = yaml.safe_load(f)
            if data is None:
                continue
            roomid = data.get("roomid")
            if roomid is None:
                continue
            zone = data.get("zone", zone_dir)
            exits = data.get("exits") or {}
            rooms[roomid] = {
                "roomid": roomid,
                "zone": zone,
                "title": data.get("title", "(untitled)"),
                "exits": exits,
                "file": filepath,
            }
    return rooms


def is_cardinal_direction(direction):
    """Return True if the direction is a standard cardinal/vertical direction."""
    return direction in DIRECTION_VECTORS


def check_one_way_exits(rooms):
    """Find exits where A->B exists but B has no exit back to A."""
    issues = []
    for roomid, room in rooms.items():
        for direction, exit_data in room["exits"].items():
            target_id = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target_id is None:
                continue
            if target_id not in rooms:
                continue  # target room doesn't exist in loaded data, skip
            target_room = rooms[target_id]
            # Check if target has ANY exit back to this room
            has_return = False
            for _, back_exit in target_room["exits"].items():
                back_target = back_exit.get("roomid") if isinstance(back_exit, dict) else back_exit
                if back_target == roomid:
                    has_return = True
                    break
            if not has_return:
                issues.append({
                    "from_room": roomid,
                    "from_zone": room["zone"],
                    "from_title": room["title"],
                    "direction": direction,
                    "to_room": target_id,
                    "to_zone": target_room["zone"],
                    "to_title": target_room["title"],
                })
    return issues


def check_direction_mismatches(rooms):
    """Find exits where A goes direction X to B, but B's return is not the opposite direction."""
    issues = []
    seen = set()
    for roomid, room in rooms.items():
        for direction, exit_data in room["exits"].items():
            if not is_cardinal_direction(direction):
                continue
            target_id = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target_id is None or target_id not in rooms:
                continue
            opposite = DIRECTION_OPPOSITES[direction]
            target_room = rooms[target_id]
            # Find which exit(s) in target lead back to this room
            return_dirs = []
            for back_dir, back_exit in target_room["exits"].items():
                back_target = back_exit.get("roomid") if isinstance(back_exit, dict) else back_exit
                if back_target == roomid:
                    return_dirs.append(back_dir)
            # If there's a return but not via the expected opposite direction
            for ret_dir in return_dirs:
                if is_cardinal_direction(ret_dir) and ret_dir != opposite:
                    pair_key = tuple(sorted([(roomid, direction), (target_id, ret_dir)]))
                    if pair_key not in seen:
                        seen.add(pair_key)
                        issues.append({
                            "room_a": roomid,
                            "zone_a": room["zone"],
                            "title_a": room["title"],
                            "dir_a": direction,
                            "room_b": target_id,
                            "zone_b": target_room["zone"],
                            "title_b": target_room["title"],
                            "dir_b": ret_dir,
                            "expected_b": opposite,
                        })
    return issues


def build_spatial_grid(rooms):
    """Walk the room graph assigning coordinates based on direction vectors.

    Returns (coords, conflicts) where coords maps roomid -> (x,y,z)
    and conflicts is a list of rooms that got assigned two different positions.
    """
    coords = {}
    conflicts = []

    # Start from room 1 if it exists, otherwise pick the lowest roomid
    start_room = 1 if 1 in rooms else min(rooms.keys())

    # BFS from start_room
    from collections import deque
    queue = deque()
    coords[start_room] = (0, 0, 0)
    queue.append(start_room)
    visited_edges = set()

    while queue:
        current = queue.popleft()
        cx, cy, cz = coords[current]
        room = rooms[current]
        for direction, exit_data in room["exits"].items():
            if not is_cardinal_direction(direction):
                continue
            target_id = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target_id is None or target_id not in rooms:
                continue
            edge = (current, target_id, direction)
            if edge in visited_edges:
                continue
            visited_edges.add(edge)

            dx, dy, dz = DIRECTION_VECTORS[direction]
            expected = (cx + dx, cy + dy, cz + dz)

            if target_id in coords:
                if coords[target_id] != expected:
                    conflicts.append({
                        "room": target_id,
                        "zone": rooms[target_id]["zone"],
                        "title": rooms[target_id]["title"],
                        "existing_coord": coords[target_id],
                        "new_coord": expected,
                        "from_room": current,
                        "from_zone": room["zone"],
                        "from_title": room["title"],
                        "direction": direction,
                    })
            else:
                coords[target_id] = expected
                queue.append(target_id)

    return coords, conflicts


def find_cross_zone_exits(rooms):
    """List all exits that connect rooms in different zones."""
    connections = []
    seen = set()
    for roomid, room in rooms.items():
        for direction, exit_data in room["exits"].items():
            target_id = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target_id is None or target_id not in rooms:
                continue
            target_room = rooms[target_id]
            if room["zone"] != target_room["zone"]:
                pair_key = tuple(sorted([roomid, target_id]))
                if pair_key not in seen:
                    seen.add(pair_key)
                    connections.append({
                        "from_room": roomid,
                        "from_zone": room["zone"],
                        "from_title": room["title"],
                        "direction": direction,
                        "to_room": target_id,
                        "to_zone": target_room["zone"],
                        "to_title": target_room["title"],
                    })
    return connections


def find_orphan_rooms(rooms):
    """Find rooms that no other room has an exit to."""
    has_incoming = set()
    for roomid, room in rooms.items():
        for direction, exit_data in room["exits"].items():
            target_id = exit_data.get("roomid") if isinstance(exit_data, dict) else exit_data
            if target_id is not None:
                has_incoming.add(target_id)
    orphans = []
    for roomid, room in rooms.items():
        if roomid not in has_incoming:
            orphans.append({
                "roomid": roomid,
                "zone": room["zone"],
                "title": room["title"],
            })
    return orphans


def find_dead_ends(rooms):
    """Find rooms with exactly one exit."""
    dead_ends = []
    for roomid, room in rooms.items():
        if len(room["exits"]) == 1:
            dead_ends.append({
                "roomid": roomid,
                "zone": room["zone"],
                "title": room["title"],
                "exit_dir": list(room["exits"].keys())[0],
            })
    return dead_ends


def zone_stats(rooms):
    """Compute room and exit counts per zone."""
    stats = defaultdict(lambda: {"rooms": 0, "exits": 0})
    for room in rooms.values():
        z = room["zone"]
        stats[z]["rooms"] += 1
        stats[z]["exits"] += len(room["exits"])
    return dict(stats)


def room_label(roomid, room_data):
    """Format a room reference for display."""
    if room_data:
        return f"#{roomid} \"{room_data['title']}\" [{room_data['zone']}]"
    return f"#{roomid}"


def main():
    parser = argparse.ArgumentParser(description="Room Exit Consistency Checker")
    parser.add_argument(
        "--base-dir",
        default=os.path.join(os.path.dirname(__file__), ".."),
        help="World directory (parent of rooms/)",
    )
    args = parser.parse_args()
    base_dir = os.path.abspath(args.base_dir)

    print(f"Scanning: {base_dir}")
    print()

    rooms = load_rooms(base_dir)
    if not rooms:
        print("No rooms found.")
        sys.exit(1)

    stats = zone_stats(rooms)
    one_way = check_one_way_exits(rooms)
    mismatches = check_direction_mismatches(rooms)
    _, grid_conflicts = build_spatial_grid(rooms)
    cross_zone = find_cross_zone_exits(rooms)
    orphans = find_orphan_rooms(rooms)
    dead_ends = find_dead_ends(rooms)

    # Group everything by zone for the report
    zones = sorted(stats.keys())
    total_issues = 0

    # --- Per-zone summary ---
    print("=" * 70)
    print("ZONE SUMMARY")
    print("=" * 70)
    for zone in zones:
        s = stats[zone]
        print(f"  {zone:30s}  {s['rooms']:3d} rooms, {s['exits']:3d} exits")
    total_rooms = sum(s["rooms"] for s in stats.values())
    total_exits = sum(s["exits"] for s in stats.values())
    print(f"  {'TOTAL':30s}  {total_rooms:3d} rooms, {total_exits:3d} exits")
    print()

    # --- One-way exits ---
    print("=" * 70)
    print("ONE-WAY EXITS (warning)")
    print("=" * 70)
    if one_way:
        by_zone = defaultdict(list)
        for issue in one_way:
            by_zone[issue["from_zone"]].append(issue)
        for zone in sorted(by_zone):
            print(f"\n  [{zone}]")
            for i in by_zone[zone]:
                print(f"    #{i['from_room']} \"{i['from_title']}\" --{i['direction']}--> "
                      f"#{i['to_room']} \"{i['to_title']}\" [{i['to_zone']}]  (no return)")
        total_issues += len(one_way)
    else:
        print("  None found.")
    print()

    # --- Direction mismatches ---
    print("=" * 70)
    print("DIRECTION MISMATCHES")
    print("=" * 70)
    if mismatches:
        by_zone = defaultdict(list)
        for issue in mismatches:
            by_zone[issue["zone_a"]].append(issue)
        for zone in sorted(by_zone):
            print(f"\n  [{zone}]")
            for i in by_zone[zone]:
                print(f"    #{i['room_a']} \"{i['title_a']}\" --{i['dir_a']}--> "
                      f"#{i['room_b']} \"{i['title_b']}\" --{i['dir_b']}--> back")
                print(f"      Expected return direction: {i['expected_b']}, got: {i['dir_b']}")
        total_issues += len(mismatches)
    else:
        print("  None found.")
    print()

    # --- Grid conflicts ---
    print("=" * 70)
    print("SPATIAL GRID CONFLICTS")
    print("=" * 70)
    if grid_conflicts:
        by_zone = defaultdict(list)
        for c in grid_conflicts:
            by_zone[c["zone"]].append(c)
        for zone in sorted(by_zone):
            print(f"\n  [{zone}]")
            for c in by_zone[zone]:
                print(f"    #{c['room']} \"{c['title']}\" has conflicting coordinates:")
                print(f"      Already at {c['existing_coord']}, "
                      f"but #{c['from_room']} \"{c['from_title']}\" "
                      f"--{c['direction']}--> places it at {c['new_coord']}")
        total_issues += len(grid_conflicts)
    else:
        print("  None found.")
    print()

    # --- Cross-zone connections ---
    print("=" * 70)
    print("CROSS-ZONE CONNECTIONS (verify manually)")
    print("=" * 70)
    if cross_zone:
        for c in cross_zone:
            print(f"  #{c['from_room']} \"{c['from_title']}\" [{c['from_zone']}] "
                  f"--{c['direction']}--> "
                  f"#{c['to_room']} \"{c['to_title']}\" [{c['to_zone']}]")
    else:
        print("  None found.")
    print()

    # --- Orphan rooms ---
    print("=" * 70)
    print("ORPHAN ROOMS (no incoming exits)")
    print("=" * 70)
    if orphans:
        by_zone = defaultdict(list)
        for o in orphans:
            by_zone[o["zone"]].append(o)
        for zone in sorted(by_zone):
            print(f"\n  [{zone}]")
            for o in by_zone[zone]:
                print(f"    #{o['roomid']} \"{o['title']}\"")
        total_issues += len(orphans)
    else:
        print("  None found.")
    print()

    # --- Dead ends ---
    print("=" * 70)
    print("DEAD ENDS (single exit)")
    print("=" * 70)
    if dead_ends:
        by_zone = defaultdict(list)
        for d in dead_ends:
            by_zone[d["zone"]].append(d)
        for zone in sorted(by_zone):
            print(f"\n  [{zone}]")
            for d in by_zone[zone]:
                print(f"    #{d['roomid']} \"{d['title']}\" (exit: {d['exit_dir']})")
    else:
        print("  None found.")
    print()

    # --- Final tally ---
    print("=" * 70)
    warning_count = len(one_way) + len(dead_ends)
    error_count = len(mismatches) + len(grid_conflicts)
    info_count = len(cross_zone) + len(orphans)
    print(f"Errors: {error_count}  Warnings: {warning_count}  Info: {info_count}")
    if error_count > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()

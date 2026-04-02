#!/usr/bin/env python3
"""
GoMud World Content Generator

Reads compact batch YAML definitions and generates individual GoMud YAML files.
Saves significant AI tokens by handling boilerplate structure automatically.

Usage:
    python3 worldgen.py <type> <batch_file> [--base-dir <dir>]
    python3 worldgen.py validate [--fix] [--base-dir <dir>]

Types: rooms, mobs, items, quests, races, buffs, spells, conversations, validate

Examples:
    python3 worldgen.py rooms batch-neighborhood.yaml
    python3 worldgen.py mobs batch-hq-mobs.yaml
    python3 worldgen.py validate --fix
"""

import sys
import os
import argparse
import yaml

# Add tools/ to path so gomudlib is importable
sys.path.insert(0, os.path.dirname(__file__))

from gomudlib.generators import GENERATORS
from gomudlib.validate import validate_world


def cmd_generate(args):
    base_dir = os.path.abspath(args.base_dir)
    print(f"Base directory: {base_dir}")

    with open(args.batch_file) as f:
        batch = yaml.safe_load(f)

    gen_func = GENERATORS[args.type]
    gen_func(batch, base_dir)
    print("Done!")


def cmd_validate(args):
    base_dir = os.path.abspath(args.base_dir)
    print(f"Validating: {base_dir}")
    print()

    errors, fixes = validate_world(base_dir, fix=args.fix)

    if fixes:
        print(f"\nAuto-fixed {len(fixes)} issues:")
        for fix in fixes:
            print(f"  {fix}")

    if errors:
        # Re-check after fixes
        if args.fix:
            errors, _ = validate_world(base_dir, fix=False)

        if errors:
            print(f"\n{len(errors)} errors found:")
            for err in errors:
                print(f"  {err}")
            sys.exit(1)
        else:
            print("\nAll issues fixed!")
    else:
        print("No errors found.")


def main():
    parser = argparse.ArgumentParser(description='GoMud World Content Generator')
    parser.add_argument('--base-dir', default=os.path.join(os.path.dirname(__file__), '..'),
                        help='Base world directory (default: parent of tools/)')

    sub = parser.add_subparsers(dest='command')

    # Generate subcommand
    gen_parser = sub.add_parser('generate', aliases=list(GENERATORS.keys()),
                                help='Generate content from batch file')
    gen_parser.add_argument('batch_file', nargs='?', help='YAML batch definition file')

    # Validate subcommand
    val_parser = sub.add_parser('validate', help='Validate world data')
    val_parser.add_argument('--fix', action='store_true', help='Attempt to auto-fix issues')

    args = parser.parse_args()

    if args.command == 'validate':
        cmd_validate(args)
    elif args.command in GENERATORS:
        # Shorthand: `worldgen.py rooms batch.yaml` instead of `worldgen.py generate rooms batch.yaml`
        args.type = args.command
        if not args.batch_file:
            parser.error(f"batch_file is required for '{args.command}'")
        cmd_generate(args)
    elif args.command == 'generate':
        parser.error("Use content type directly: worldgen.py rooms batch.yaml")
    else:
        parser.print_help()


if __name__ == '__main__':
    main()

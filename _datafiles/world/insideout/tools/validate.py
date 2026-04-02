#!/usr/bin/env python3
"""
GoMud World Validator

Scans a GoMud world directory for issues that would cause server panics:
- YAML parse errors
- Idle messages parsed as dicts (colons creating key:value)
- Zone name / folder name mismatches
- Filename special characters (hyphens, parentheses)
- Buff triggercount < 1

Usage:
    python3 validate.py [--fix] [--base-dir <dir>]
    python3 validate.py --fix --base-dir /opt/gomud/data/world/insideout
"""

import sys
import os
import argparse

sys.path.insert(0, os.path.dirname(__file__))

from gomudlib.validate import validate_world


def main():
    parser = argparse.ArgumentParser(description='GoMud World Validator')
    parser.add_argument('--base-dir', default=os.path.join(os.path.dirname(__file__), '..'),
                        help='World directory to validate')
    parser.add_argument('--fix', action='store_true',
                        help='Attempt to auto-fix issues')
    args = parser.parse_args()

    base_dir = os.path.abspath(args.base_dir)
    print(f"Validating: {base_dir}")
    print()

    errors, fixes = validate_world(base_dir, fix=args.fix)

    if fixes:
        print(f"\nAuto-fixed {len(fixes)} issues:")
        for fix in fixes:
            print(f"  {fix}")
        print()

    # Re-check after fixes
    if args.fix and fixes:
        errors, _ = validate_world(base_dir, fix=False)

    if errors:
        print(f"{len(errors)} errors remaining:")
        for err in errors:
            print(f"  {err}")
        sys.exit(1)
    else:
        print("No errors found." if not fixes else "All issues resolved.")


if __name__ == '__main__':
    main()

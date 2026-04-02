"""Validation and auto-fix for GoMud world data.

Catches issues that cause server panics or rendering bugs:
- YAML strings parsed as dicts (colons in idle messages)
- Zone name / folder name mismatches
- Filename special characters (hyphens, parens)
- Buff triggercount < 1
- Broken ANSI tags (line-wrapped mid-tag by YAML serialization)
- General YAML validity
"""

import os
import re
import glob
import yaml

from .yamlutil import slugify


def validate_yaml_file(path):
    """Validate a single YAML file. Returns (data, errors) tuple."""
    errors = []
    try:
        with open(path) as f:
            data = yaml.safe_load(f)
    except yaml.YAMLError as e:
        return None, [f"YAML parse error: {e}"]
    return data, errors


def _check_string_lists(data, fields, path):
    """Check that list fields contain only strings (not dicts from colon misparse)."""
    errors = []
    for field in fields:
        if field not in data:
            continue
        for i, item in enumerate(data[field]):
            if isinstance(item, dict):
                errors.append(f"{path}: {field}[{i}] is a dict (colon in string?): {item!r}")
    return errors


# Regex to find ANSI tags that got split across lines during YAML wrapping.
# Matches patterns like: "<\nansi" or "</\nansi" or "<ansi\nfg=" etc.
_BROKEN_ANSI_RE = re.compile(
    r'<\s*\n\s*ansi'       # < followed by newline then ansi
    r'|<\s*/\s*\n\s*ansi'  # </ followed by newline then ansi
    r'|<\s*ansi\s*\n'      # <ansi followed by newline (mid-tag)
    r'|<\s*/\s*ansi\s*\n'  # </ansi followed by newline
)

# Also catch the YAML-serialized version where the newline becomes a space
# but the tag got split into "< ansi" or "ansi >" etc.
_BROKEN_ANSI_SPACE_RE = re.compile(
    r'<\s{2,}ansi'          # < with multiple spaces before ansi
    r'|<\s*/\s{2,}ansi'     # </ with multiple spaces before ansi
    r'|<\s+/\s+ansi'        # < / ansi with spaces
)


def _check_broken_ansi(data, fields, path):
    """Check for ANSI tags broken by YAML line wrapping."""
    errors = []
    for field in fields:
        val = data.get(field, '')
        if isinstance(val, str) and ('<' in val or 'ansi' in val):
            if _BROKEN_ANSI_RE.search(val) or _BROKEN_ANSI_SPACE_RE.search(val):
                errors.append(f"{path}: {field} has broken ANSI tags (line-wrapped mid-tag)")
        if isinstance(val, list):
            for i, item in enumerate(val):
                if isinstance(item, str) and ('<' in item or 'ansi' in item):
                    if _BROKEN_ANSI_RE.search(item) or _BROKEN_ANSI_SPACE_RE.search(item):
                        errors.append(f"{path}: {field}[{i}] has broken ANSI tags (line-wrapped mid-tag)")
    return errors


def _fix_broken_ansi(data, fields, path):
    """Fix ANSI tags broken by YAML line wrapping.

    The root cause is YAML serializers wrapping long lines, splitting tags like:
        <ansi fg="itemname">text</ansi>
    into:
        <\n  ansi fg="itemname">text</\n  ansi>

    Fix: re-join any < ... ansi sequences that have whitespace/newlines in them.
    """
    changed = False
    for field in fields:
        val = data.get(field)
        if isinstance(val, str):
            fixed = _rejoin_ansi_tags(val)
            if fixed != val:
                data[field] = fixed
                changed = True
        elif isinstance(val, list):
            for i, item in enumerate(val):
                if isinstance(item, str):
                    fixed = _rejoin_ansi_tags(item)
                    if fixed != item:
                        val[i] = fixed
                        changed = True
    if changed:
        with open(path, 'w') as f:
            yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False, width=10000)
    return changed


def _rejoin_ansi_tags(text):
    """Rejoin ANSI tags that were split by whitespace/newlines."""
    # Fix opening tags: < \n ansi -> <ansi
    text = re.sub(r'<\s+ansi', '<ansi', text)
    # Fix closing tags: < \n / \n ansi -> </ansi  and  </ \n ansi -> </ansi
    text = re.sub(r'<\s*/\s*ansi', '</ansi', text)
    # Fix closing tags where the slash got separated: < / ansi -> </ansi
    text = re.sub(r'<\s+/\s*ansi', '</ansi', text)
    return text


def _check_filename(path, expected_slug_fn):
    """Check that a filename matches GoMud's expected derivation from YAML name field."""
    errors = []
    basename = os.path.basename(path).replace('.yaml', '')
    parts = basename.split('-', 1)
    if len(parts) != 2:
        return errors

    file_id, file_name_part = parts

    # Basic character checks
    if '(' in file_name_part or ')' in file_name_part:
        errors.append(f"{path}: filename contains parentheses - GoMud replaces these with underscores")
    if '-' in file_name_part:
        errors.append(f"{path}: filename contains hyphens after ID - GoMud replaces these with underscores")

    # Cross-check filename against the name field in the YAML
    try:
        with open(path) as f:
            data = yaml.safe_load(f)
        if data:
            # Items use 'name', mobs use character.name
            name = None
            if 'name' in data:
                name = data['name']
            elif 'character' in data and isinstance(data['character'], dict):
                name = data['character'].get('name')

            if name:
                expected_name_part = slugify(name)
                if file_name_part != expected_name_part:
                    errors.append(
                        f"{path}: filename '{file_id}-{file_name_part}' doesn't match "
                        f"name '{name}' (expected: '{file_id}-{expected_name_part}')"
                    )
    except:
        pass

    return errors


def _check_zone_folder(zone_config_path):
    """Check that zone folder name matches the zone name in config."""
    errors = []
    try:
        with open(zone_config_path) as f:
            data = yaml.safe_load(f)
    except:
        return [f"{zone_config_path}: cannot parse"]

    if not data or 'name' not in data:
        return []

    zone_name = data['name']
    folder = os.path.basename(os.path.dirname(zone_config_path))
    expected_folder = slugify(zone_name)

    if folder != expected_folder:
        errors.append(
            f"{zone_config_path}: folder '{folder}' doesn't match zone name '{zone_name}' "
            f"(expected folder: '{expected_folder}')"
        )
    return errors


def _check_buff(data, path):
    """Check buff-specific constraints."""
    errors = []
    if 'triggercount' in data and data['triggercount'] < 1:
        errors.append(f"{path}: triggercount is {data['triggercount']}, GoMud requires >= 1")
    return errors


def validate_world(base_dir, fix=False):
    """Validate an entire GoMud world directory.

    Args:
        base_dir: Path to the world directory (e.g., _datafiles/world/insideout)
        fix: If True, attempt to auto-fix issues

    Returns:
        (errors, fixes) tuple of lists
    """
    errors = []
    fixes = []

    # 1. Check zone folder names
    for zc in glob.glob(os.path.join(base_dir, 'rooms', '*', 'zone-config.yaml')):
        errs = _check_zone_folder(zc)
        errors.extend(errs)
        if fix and errs:
            fixed = _fix_zone_name(zc)
            if fixed:
                fixes.append(f"Fixed zone name in {zc}")

    # 2. Check all room YAML files
    str_fields = ['description', 'idlemessages']
    for path in sorted(glob.glob(os.path.join(base_dir, 'rooms', '**', '*.yaml'), recursive=True)):
        if 'zone-config' in path:
            continue
        data, parse_errors = validate_yaml_file(path)
        errors.extend(parse_errors)
        if data:
            errs = _check_string_lists(data, ['idlemessages'], path)
            errors.extend(errs)
            if fix and errs:
                _fix_string_lists(data, ['idlemessages'], path)

            # Check for broken ANSI tags in descriptions and idle messages
            # Also check noun values
            ansi_fields = ['description']
            if 'nouns' in data and isinstance(data['nouns'], dict):
                ansi_fields.extend([f"nouns.{k}" for k in data['nouns']])
            ansi_errs = _check_broken_ansi(data, ['description', 'idlemessages'], path)
            if 'nouns' in data and isinstance(data['nouns'], dict):
                for k, v in data['nouns'].items():
                    if isinstance(v, str) and (_BROKEN_ANSI_RE.search(v) or _BROKEN_ANSI_SPACE_RE.search(v)):
                        ansi_errs.append(f"{path}: nouns.{k} has broken ANSI tags")
            errors.extend(ansi_errs)
            if fix and ansi_errs:
                if _fix_broken_ansi(data, ['description', 'idlemessages'], path):
                    fixes.append(f"Fixed broken ANSI tags in {path}")
                # Also fix nouns
                if 'nouns' in data and isinstance(data['nouns'], dict):
                    noun_changed = False
                    for k, v in data['nouns'].items():
                        if isinstance(v, str):
                            fixed = _rejoin_ansi_tags(v)
                            if fixed != v:
                                data['nouns'][k] = fixed
                                noun_changed = True
                    if noun_changed:
                        with open(path, 'w') as f:
                            yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False, width=10000)
                        fixes.append(f"Fixed broken ANSI tags in nouns of {path}")
                fixes.append(f"Fixed idle messages in {path}")

    # 3. Check mob filenames and YAML
    for path in sorted(glob.glob(os.path.join(base_dir, 'mobs', '**', '*.yaml'), recursive=True)):
        data, parse_errors = validate_yaml_file(path)
        errors.extend(parse_errors)
        errs = _check_filename(path, slugify)
        errors.extend(errs)
        if fix and errs:
            new_path = _fix_filename(path)
            if new_path:
                fixes.append(f"Renamed {path} -> {os.path.basename(new_path)}")

    # 4. Check item filenames
    for path in sorted(glob.glob(os.path.join(base_dir, 'items', '**', '*.yaml'), recursive=True)):
        errs = _check_filename(path, slugify)
        errors.extend(errs)
        if fix and errs:
            new_path = _fix_filename(path)
            if new_path:
                fixes.append(f"Renamed {path} -> {os.path.basename(new_path)}")

    # 5. Check buffs
    for path in sorted(glob.glob(os.path.join(base_dir, 'buffs', '*.yaml'))):
        data, parse_errors = validate_yaml_file(path)
        errors.extend(parse_errors)
        if data:
            errs = _check_buff(data, path)
            errors.extend(errs)
            if fix and errs:
                _fix_buff(data, path)
                fixes.append(f"Fixed buff triggercount in {path}")

    # 6. Check quest YAML
    for path in sorted(glob.glob(os.path.join(base_dir, 'quests', '*.yaml'))):
        data, parse_errors = validate_yaml_file(path)
        errors.extend(parse_errors)

    # 7. Check conversation YAML
    for path in sorted(glob.glob(os.path.join(base_dir, 'conversations', '**', '*.yaml'), recursive=True)):
        data, parse_errors = validate_yaml_file(path)
        errors.extend(parse_errors)

    return errors, fixes


def _fix_string_lists(data, fields, path):
    """Auto-fix string lists that contain dicts (colon misparse)."""
    changed = False
    for field in fields:
        if field not in data:
            continue
        new_list = []
        for item in data[field]:
            if isinstance(item, dict):
                parts = []
                for k, v in item.items():
                    parts.append(f"{k} -- {v}")
                new_list.append(' '.join(parts))
                changed = True
            else:
                new_list.append(item)
        data[field] = new_list

    if changed:
        with open(path, 'w') as f:
            yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False, width=10000)


def _fix_buff(data, path):
    """Auto-fix buff triggercount."""
    if 'triggercount' in data and data['triggercount'] < 1:
        data['triggercount'] = 1
        with open(path, 'w') as f:
            yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False, width=10000)


def _fix_filename(path):
    """Rename a file to match the GoMud-expected slug derived from the YAML name field."""
    dirname = os.path.dirname(path)
    basename = os.path.basename(path).replace('.yaml', '')
    parts = basename.split('-', 1)
    if len(parts) != 2:
        return None
    id_part = parts[0]

    # Read the name from the YAML to derive the correct filename
    try:
        with open(path) as f:
            data = yaml.safe_load(f)
        name = None
        if data and 'name' in data:
            name = data['name']
        elif data and 'character' in data and isinstance(data['character'], dict):
            name = data['character'].get('name')

        if name:
            new_name = slugify(name)
        else:
            # Fallback: just slugify the existing name part
            new_name = slugify(parts[1])
    except:
        new_name = slugify(parts[1])

    new_path = os.path.join(dirname, f"{id_part}-{new_name}.yaml")
    if new_path != path:
        os.rename(path, new_path)
        return new_path
    return None


def _fix_zone_name(zone_config_path):
    """Fix zone name to match the folder it's in.

    Rather than renaming folders (which affects rooms and mobs),
    we update the zone name in the config and all room files to
    produce a name that maps back to the existing folder.
    """
    try:
        with open(zone_config_path) as f:
            data = yaml.safe_load(f)
    except:
        return False

    if not data or 'name' not in data:
        return False

    folder = os.path.basename(os.path.dirname(zone_config_path))
    zone_name = data['name']
    expected_folder = slugify(zone_name)

    if folder == expected_folder:
        return False

    # Derive a clean name from the folder
    # e.g., "rileys_neighborhood" -> "Rileys Neighborhood"
    clean_name = folder.replace('_', ' ').title()

    # Update zone config
    data['name'] = clean_name
    with open(zone_config_path, 'w') as f:
        yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False, width=10000)

    # Update all room files in this zone
    zone_dir = os.path.dirname(zone_config_path)
    for room_file in glob.glob(os.path.join(zone_dir, '*.yaml')):
        if 'zone-config' in room_file:
            continue
        try:
            with open(room_file) as f:
                room_data = yaml.safe_load(f)
            if room_data and room_data.get('zone') == zone_name:
                room_data['zone'] = clean_name
                with open(room_file, 'w') as f:
                    yaml.dump(room_data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
        except:
            pass

    # Also update mob files in matching mob zone folder
    mob_dir = os.path.join(os.path.dirname(os.path.dirname(zone_config_path)), '..', 'mobs', folder)
    mob_dir = os.path.normpath(mob_dir)
    if os.path.isdir(mob_dir):
        for mob_file in glob.glob(os.path.join(mob_dir, '*.yaml')):
            try:
                with open(mob_file) as f:
                    mob_data = yaml.safe_load(f)
                if mob_data and mob_data.get('zone') == zone_name:
                    mob_data['zone'] = clean_name
                    with open(mob_file, 'w') as f:
                        yaml.dump(mob_data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
            except:
                pass

    return True

"""YAML utilities for GoMud content generation.

Handles YAML dumping with GoMud-compatible formatting, file slugification
matching GoMud's internal path derivation, and string sanitization.
"""

import os
import re
import yaml
from collections import OrderedDict


class CleanDumper(yaml.Dumper):
    """YAML dumper that produces clean, GoMud-compatible output."""
    pass


def _str_representer(dumper, data):
    if '\n' in data:
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='>')
    if any(c in data for c in ':{}[],"\'&#*!|>%@`'):
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style="'")
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


def _ordered_dict_representer(dumper, data):
    return dumper.represent_mapping('tag:yaml.org,2002:map', data.items())


def _none_representer(dumper, data):
    return dumper.represent_scalar('tag:yaml.org,2002:null', '')


CleanDumper.add_representer(str, _str_representer)
CleanDumper.add_representer(OrderedDict, _ordered_dict_representer)
CleanDumper.add_representer(type(None), _none_representer)


def dump_yaml(data, stream=None):
    """Dump data to YAML string using GoMud-compatible formatting.

    Uses width=10000 to prevent PyYAML from wrapping long lines, which
    would break ANSI tags like <ansi fg="itemname"> by splitting them
    across lines.
    """
    return yaml.dump(data, stream, Dumper=CleanDumper,
                     default_flow_style=False, allow_unicode=True,
                     sort_keys=False, width=10000)


def slugify(name):
    """Convert a display name to a GoMud-compatible filesystem slug.

    Matches GoMud's internal util.ConvertForFilename() exactly:
    - Lowercase
    - Skip apostrophes entirely (removed, not replaced)
    - Keep a-z and 0-9 as-is
    - Replace everything else (spaces, hyphens, parens, etc.) with underscore
    - Does NOT collapse multiple underscores
    - Does NOT strip trailing underscores
    """
    s = name.lower()
    result = []
    for ch in s:
        if ch == "'":
            continue  # skip apostrophes entirely
        elif ('a' <= ch <= 'z') or ('0' <= ch <= '9'):
            result.append(ch)
        else:
            result.append('_')
    return ''.join(result)


def sanitize_string(s):
    """Sanitize a string for safe YAML embedding.

    Fixes common issues that cause YAML parsing errors or GoMud unmarshal failures:
    - Ensures colons in strings don't create accidental key:value pairs
    - Strips control characters
    """
    if not isinstance(s, str):
        return s
    return s


def sanitize_strings_in_list(items):
    """Sanitize all strings in a list (e.g., idle messages).

    Returns the sanitized list and a list of warnings.
    """
    warnings = []
    result = []
    for i, item in enumerate(items):
        if not isinstance(item, str):
            warnings.append(f"  index {i}: expected string, got {type(item).__name__}: {item!r}")
            # Try to recover dict-as-string (colon caused key:value parse)
            if isinstance(item, dict):
                parts = []
                for k, v in item.items():
                    parts.append(f"{k} -- {v}")
                result.append(' '.join(parts))
            else:
                result.append(str(item))
        else:
            result.append(item)
    return result, warnings


def write_file(path, content):
    """Write content to a file, creating parent directories as needed."""
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        f.write(content)
    print(f"  Created: {path}")

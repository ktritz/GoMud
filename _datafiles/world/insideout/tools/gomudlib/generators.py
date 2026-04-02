"""Content generators for GoMud world data.

Each generator reads a compact batch YAML format and produces
individual GoMud-compatible YAML files.
"""

import os
from collections import OrderedDict
from .yamlutil import dump_yaml, write_file, slugify, sanitize_strings_in_list


def _get_zone_name(batch):
    return batch.get('zone_config', {}).get('name', batch.get('zone', 'Unknown'))


def _get_zone_folder(zone_name):
    return slugify(zone_name)


def _validate_zone_name(zone_name):
    """Warn if zone name will produce a folder name that differs from the batch 'zone' key."""
    folder = _get_zone_folder(zone_name)
    if "'" in zone_name:
        print(f"  WARNING: Zone name '{zone_name}' contains apostrophe -> folder '{folder}'")
        print(f"           GoMud expects folder to match. Consider removing the apostrophe.")


def _sanitize_idle(items, context=""):
    """Sanitize idle messages list, warn on issues."""
    clean, warnings = sanitize_strings_in_list(items)
    for w in warnings:
        print(f"  WARNING ({context}): idle message issue: {w}")
    return clean


# ── Rooms ─────────────────────────────────────────────────────────────

def gen_rooms(batch, base_dir):
    zone_name = _get_zone_name(batch)
    _validate_zone_name(zone_name)
    zone_folder = _get_zone_folder(zone_name)

    # Zone config
    zc = batch.get('zone_config', {})
    if zc:
        config = OrderedDict()
        config['name'] = zc.get('name', batch.get('zone', zone_name))
        config['roomid'] = zc.get('roomid', batch.get('rooms', [{}])[0].get('id', 1))
        if 'autoscale' in zc:
            config['autoscale'] = zc['autoscale']
        if 'idlemessages' in zc:
            config['idlemessages'] = _sanitize_idle(zc['idlemessages'], f"zone {zone_name}")
        if 'musicfile' in zc:
            config['musicfile'] = zc['musicfile']
        if 'defaultbiome' in zc:
            config['defaultbiome'] = zc['defaultbiome']
        path = os.path.join(base_dir, 'rooms', zone_folder, 'zone-config.yaml')
        write_file(path, dump_yaml(dict(config)))

    # Rooms
    for room in batch.get('rooms', []):
        r = OrderedDict()
        r['roomid'] = room['id']
        r['zone'] = zone_name
        r['title'] = room['title']
        r['description'] = room.get('desc', room.get('description', ''))

        if 'symbol' in room:
            r['mapsymbol'] = room['symbol']
        if 'legend' in room:
            r['maplegend'] = room['legend']
        if 'biome' in room:
            r['biome'] = room['biome']
        if 'pvp' in room:
            r['pvp'] = room['pvp']

        if 'exits' in room:
            exits = OrderedDict()
            for direction, target in room['exits'].items():
                if isinstance(target, dict):
                    exits[direction] = OrderedDict()
                    exits[direction]['roomid'] = target['roomid']
                    if 'secret' in target:
                        exits[direction]['secret'] = target['secret']
                    if 'lock' in target:
                        exits[direction]['lock'] = OrderedDict()
                        exits[direction]['lock']['difficulty'] = target['lock']
                else:
                    exits[direction] = OrderedDict([('roomid', target)])
            r['exits'] = exits

        if 'nouns' in room:
            r['nouns'] = room['nouns']
        if 'containers' in room:
            r['containers'] = room['containers']

        if 'spawns' in room:
            spawns = []
            for s in room['spawns']:
                si = OrderedDict()
                for src, dst in [('mob', 'mobid'), ('item', 'itemid'), ('msg', 'message'),
                                 ('rate', 'respawnrate'), ('wander', 'maxwander'),
                                 ('levelmod', 'levelmod'), ('container', 'containername')]:
                    if src in s:
                        si[dst] = s[src]
                if 'idle' in s:
                    si['idlecommands'] = s['idle']
                spawns.append(si)
            r['spawninfo'] = spawns

        if 'idle' in room:
            r['idlemessages'] = _sanitize_idle(room['idle'], f"room {room['id']}")
        if 'skills' in room:
            r['skilltraining'] = room['skills']

        path = os.path.join(base_dir, 'rooms', zone_folder, f"{room['id']}.yaml")
        write_file(path, dump_yaml(dict(r)))

    print(f"  Generated {len(batch.get('rooms', []))} rooms for zone '{zone_name}'")


# ── Mobs ──────────────────────────────────────────────────────────────

def gen_mobs(batch, base_dir):
    zone_name = _get_zone_name(batch)
    _validate_zone_name(zone_name)
    zone_folder = _get_zone_folder(zone_name)

    for mob in batch.get('mobs', []):
        m = OrderedDict()
        m['mobid'] = mob['id']
        m['zone'] = zone_name

        for src, dst in [('drop', 'itemdropchance'), ('hostile', 'hostile'),
                         ('wander', 'maxwander'), ('activity', 'activitylevel')]:
            if src in mob:
                m[dst] = mob[src]

        for field in ['groups', 'hates', 'idlecommands', 'combatcommands', 'angrycommands']:
            short = {'idlecommands': 'idle', 'combatcommands': 'combat',
                     'angrycommands': 'angry'}.get(field, field)
            if short in mob:
                m[field] = mob[short]

        if 'script' in mob:
            m['scripttag'] = mob['script']
        if 'buffs' in mob:
            m['buffids'] = mob['buffs']
        if 'questflags' in mob:
            m['questflags'] = mob['questflags']

        char = OrderedDict()
        char['name'] = mob['name']
        char['description'] = mob.get('desc', mob.get('description', ''))
        char['level'] = mob.get('level', 1)
        char['raceid'] = mob.get('race', 1)

        if 'alignment' in mob:
            char['alignment'] = mob['alignment']
        if 'gold' in mob:
            char['gold'] = mob['gold']

        if 'stats' in mob:
            stats = OrderedDict()
            for stat_name, val in mob['stats'].items():
                stats[stat_name] = OrderedDict([('training', val)])
            char['stats'] = stats

        if 'equip' in mob:
            equipment = OrderedDict()
            for slot, itemid in mob['equip'].items():
                equipment[slot] = OrderedDict([('itemid', itemid)])
            char['equipment'] = equipment

        if 'items' in mob:
            char['items'] = [
                item if isinstance(item, dict) else OrderedDict([('itemid', item)])
                for item in mob['items']
            ]

        if 'shop' in mob:
            shop = []
            for si in mob['shop']:
                entry = OrderedDict()
                if 'item' in si:
                    entry['itemid'] = si['item']
                if 'pet' in si:
                    entry['pettype'] = si['pet']
                if 'qty' in si:
                    entry['quantitymax'] = si['qty']
                if 'price' in si:
                    entry['price'] = si['price']
                shop.append(entry)
            char['shop'] = shop

        m['character'] = char

        name_slug = slugify(mob['name'])
        path = os.path.join(base_dir, 'mobs', zone_folder, f"{mob['id']}-{name_slug}.yaml")
        write_file(path, dump_yaml(dict(m)))

    print(f"  Generated {len(batch.get('mobs', []))} mobs for zone '{zone_name}'")


# ── Items ─────────────────────────────────────────────────────────────

def gen_items(batch, base_dir):
    count = 0
    type_map = {'weapons': 'weapon', 'armor': 'armor', 'consumables': 'consumable', 'other': 'other'}

    for batch_key, item_type in type_map.items():
        for item in batch.get(batch_key, []):
            i = OrderedDict()
            i['itemid'] = item['id']
            i['name'] = item['name']
            if 'simple' in item:
                i['namesimple'] = item['simple']
            i['description'] = item.get('desc', item.get('description', ''))

            if item_type == 'weapon':
                i['type'] = 'weapon'
                i['hands'] = item.get('hands', 1)
                if 'subtype' in item:
                    i['subtype'] = item['subtype']
                i['uses'] = item.get('uses', 0)
                i['damage'] = OrderedDict([('diceroll', item.get('damage', '1d2'))])
                if 'wait' in item:
                    i['waitrounds'] = item['wait']
                if 'breakchance' in item:
                    i['breakchance'] = item['breakchance']
            elif item_type == 'armor':
                slot = item.get('slot', item.get('type', 'body'))
                i['type'] = slot
                i['subtype'] = 'wearable'
                if 'dr' in item:
                    i['damagereduction'] = item['dr']
            elif item_type == 'consumable':
                i['type'] = item.get('type', 'potion')
                i['subtype'] = item.get('subtype', 'drinkable')
                i['uses'] = item.get('uses', 1)
            elif item_type == 'other':
                i['type'] = item.get('type', 'object')
                if 'keylock' in item:
                    i['keylockid'] = item['keylock']

            for src, dst in [('statmods', 'statmods'), ('value', 'value'),
                             ('buffids', 'buffids'), ('wornbuffs', 'wornbuffids'),
                             ('cursed', 'cursed')]:
                if src in item:
                    i[dst] = item[src]

            name_slug = slugify(item['name'])
            if item_type == 'weapon':
                path = os.path.join(base_dir, 'items', 'weapons-10000', f"{item['id']}-{name_slug}.yaml")
            elif item_type == 'armor':
                slot = item.get('slot', item.get('type', 'body'))
                path = os.path.join(base_dir, 'items', 'armor-20000', slot, f"{item['id']}-{name_slug}.yaml")
            elif item_type == 'consumable':
                path = os.path.join(base_dir, 'items', 'consumables-30000', f"{item['id']}-{name_slug}.yaml")
            else:
                path = os.path.join(base_dir, 'items', 'other-0', f"{item['id']}-{name_slug}.yaml")

            write_file(path, dump_yaml(dict(i)))
            count += 1

    print(f"  Generated {count} items")


# ── Quests ────────────────────────────────────────────────────────────

def gen_quests(batch, base_dir):
    for quest in batch.get('quests', []):
        q = OrderedDict()
        q['questid'] = quest['id']
        q['name'] = quest['name']
        q['description'] = quest.get('desc', quest.get('description', ''))
        if quest.get('secret'):
            q['secret'] = True

        steps = []
        for step in quest.get('steps', []):
            s = OrderedDict()
            s['id'] = step['id']
            s['description'] = step.get('desc', step.get('description', ''))
            if 'hint' in step:
                s['hint'] = step['hint']
            steps.append(s)
        q['steps'] = steps

        if 'rewards' in quest:
            rewards = OrderedDict()
            rw = quest['rewards']
            for src, dst in [('msg', 'playermessage'), ('roommsg', 'roommessage'),
                             ('xp', 'experience'), ('gold', 'gold'), ('item', 'itemid'),
                             ('buff', 'buffid'), ('skill', 'skillinfo'),
                             ('quest', 'questid'), ('room', 'roomid')]:
                if src in rw:
                    rewards[dst] = rw[src]
            q['rewards'] = rewards

        name_slug = slugify(quest['name'])
        path = os.path.join(base_dir, 'quests', f"{quest['id']}-{name_slug}.yaml")
        write_file(path, dump_yaml(dict(q)))

    print(f"  Generated {len(batch.get('quests', []))} quests")


# ── Races ─────────────────────────────────────────────────────────────

def gen_races(batch, base_dir):
    for race in batch.get('races', []):
        r = OrderedDict()
        r['raceid'] = race['id']
        r['name'] = race['name']
        r['description'] = race.get('desc', race.get('description', ''))
        r['defaultalignment'] = race.get('alignment', 0)
        r['size'] = race.get('size', 'medium')
        r['unarmedname'] = race.get('unarmed', 'fists')
        r['tnlscale'] = race.get('tnl', 1)
        r['selectable'] = race.get('selectable', False)
        r['knowsfirstaid'] = race.get('firstaid', False)
        r['tameable'] = race.get('tameable', False)

        if 'angry' in race:
            r['angrycommands'] = race['angry']

        stats = OrderedDict()
        for stat_name in ['strength', 'speed', 'smarts', 'vitality', 'perception', 'mysticism']:
            base_val = race.get('stats', {}).get(stat_name, 0)
            if base_val > 0:
                stats[stat_name] = OrderedDict([('base', base_val)])
        if stats:
            r['stats'] = stats

        r['damage'] = OrderedDict([('diceroll', race.get('damage', '1d3'))])
        if 'attacks' in race:
            r['damage']['attacks'] = race['attacks']

        r['disabledslots'] = race.get('disabled', [])

        name_slug = slugify(race['name'])
        path = os.path.join(base_dir, 'races', f"{race['id']}-{name_slug}.yaml")
        write_file(path, dump_yaml(dict(r)))

    print(f"  Generated {len(batch.get('races', []))} races")


# ── Buffs ─────────────────────────────────────────────────────────────

def gen_buffs(batch, base_dir):
    for buff in batch.get('buffs', []):
        b = OrderedDict()
        b['buffid'] = buff['id']
        b['name'] = buff['name']
        b['description'] = buff.get('desc', buff.get('description', ''))

        if 'rate' in buff:
            b['triggerrate'] = buff['rate']
        if 'count' in buff:
            count = buff['count']
            if count < 1:
                print(f"  WARNING: Buff {buff['id']} has triggercount {count}, clamping to 1 (GoMud requires >= 1)")
                count = 1
            b['triggercount'] = count
        if 'statmods' in buff:
            b['statmods'] = buff['statmods']
        if 'flags' in buff:
            b['flags'] = buff['flags']
        if buff.get('secret'):
            b['secret'] = True

        name_slug = slugify(buff['name'])
        path = os.path.join(base_dir, 'buffs', f"{buff['id']}-{name_slug}.yaml")
        write_file(path, dump_yaml(dict(b)))

    print(f"  Generated {len(batch.get('buffs', []))} buffs")


# ── Spells ────────────────────────────────────────────────────────────

def gen_spells(batch, base_dir):
    for spell in batch.get('spells', []):
        s = OrderedDict()
        s['spellid'] = spell['id']
        s['name'] = spell['name']
        s['description'] = spell.get('desc', spell.get('description', ''))
        s['type'] = spell.get('type', 'harmsingle')
        s['school'] = spell.get('school', 'conjuration')
        s['cost'] = spell.get('cost', 5)
        s['waitrounds'] = spell.get('wait', 1)
        s['difficulty'] = spell.get('difficulty', 0)

        path = os.path.join(base_dir, 'spells', f"{spell['id']}.yaml")
        write_file(path, dump_yaml(dict(s)))

    print(f"  Generated {len(batch.get('spells', []))} spells")


# ── Conversations ─────────────────────────────────────────────────────

def gen_conversations(batch, base_dir):
    zone = batch.get('zone', 'unknown')
    zone_folder = slugify(zone)

    for conv in batch.get('conversations', []):
        mobid = conv['mobid']
        entries = []
        for variant in conv.get('variants', []):
            entry = OrderedDict()
            entry['Supported'] = variant['supported']
            entry['Conversation'] = variant['lines']
            entries.append(entry)

        path = os.path.join(base_dir, 'conversations', zone_folder, f"{mobid}.yaml")
        write_file(path, dump_yaml(entries))

    print(f"  Generated {len(batch.get('conversations', []))} conversation files for zone '{zone}'")


# ── Registry ──────────────────────────────────────────────────────────

GENERATORS = {
    'rooms': gen_rooms,
    'mobs': gen_mobs,
    'items': gen_items,
    'quests': gen_quests,
    'races': gen_races,
    'buffs': gen_buffs,
    'spells': gen_spells,
    'conversations': gen_conversations,
}

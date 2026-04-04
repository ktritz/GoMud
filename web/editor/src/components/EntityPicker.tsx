import { useState, useRef, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { mobs, items, rooms, type MobSummary, type ItemSummary, type RoomSummary } from '../api/client';

type EntityType = 'mob' | 'item' | 'room';

interface EntityOption {
  id: number;
  label: string;
  sublabel?: string;
}

interface EntityPickerProps {
  type: EntityType;
  value: number;
  onChange: (id: number) => void;
  className?: string;
}

function useEntityOptions(type: EntityType): EntityOption[] {
  const mobQuery = useQuery({
    queryKey: ['mobs'],
    queryFn: () => mobs.list(),
    enabled: type === 'mob',
    staleTime: 60_000,
  });

  const itemQuery = useQuery({
    queryKey: ['items'],
    queryFn: () => items.list(),
    enabled: type === 'item',
    staleTime: 60_000,
  });

  const roomQuery = useQuery({
    queryKey: ['rooms'],
    queryFn: () => rooms.list(),
    enabled: type === 'room',
    staleTime: 60_000,
  });

  if (type === 'mob' && mobQuery.data?.mobs) {
    return mobQuery.data.mobs.map((m: MobSummary) => ({
      id: m.mobId,
      label: m.name,
      sublabel: `L${m.level} ${m.zone}${m.hostile ? ' (hostile)' : ''}`,
    }));
  }

  if (type === 'item' && itemQuery.data?.items) {
    return itemQuery.data.items.map((it: ItemSummary) => ({
      id: it.itemId,
      label: it.name,
      sublabel: `${it.type}${it.subtype ? '/' + it.subtype : ''}`,
    }));
  }

  if (type === 'room' && roomQuery.data?.rooms) {
    return roomQuery.data.rooms.map((r: RoomSummary) => ({
      id: r.roomId,
      label: r.title,
      sublabel: r.zone,
    }));
  }

  return [];
}

export function EntityPicker({ type, value, onChange, className }: EntityPickerProps) {
  const options = useEntityOptions(type);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Close on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const filtered = search
    ? options.filter(
        (o) =>
          o.label.toLowerCase().includes(search.toLowerCase()) ||
          String(o.id).includes(search) ||
          (o.sublabel && o.sublabel.toLowerCase().includes(search.toLowerCase()))
      )
    : options;

  const selected = options.find((o) => o.id === value);
  const displayLabel = selected ? `#${selected.id} ${selected.label}` : value ? `#${value}` : '';

  const colors: Record<EntityType, { text: string; border: string; highlight: string }> = {
    mob: { text: 'text-blue-400', border: 'border-blue-600', highlight: 'bg-blue-900/40' },
    item: { text: 'text-green-400', border: 'border-green-600', highlight: 'bg-green-900/40' },
    room: { text: 'text-purple-400', border: 'border-purple-600', highlight: 'bg-purple-900/40' },
  };
  const color = colors[type];

  return (
    <div ref={containerRef} className={`relative ${className || ''}`}>
      <button
        type="button"
        onClick={() => {
          setOpen(!open);
          setSearch('');
          setTimeout(() => inputRef.current?.focus(), 0);
        }}
        className={`w-full text-left px-2 py-1 bg-gray-700 border ${
          open ? color.border : 'border-gray-600'
        } rounded text-sm ${value ? color.text : 'text-gray-500'} hover:border-gray-500 transition-colors`}
      >
        {displayLabel || `Select ${type}...`}
      </button>

      {open && (
        <div className="absolute z-50 mt-1 w-72 bg-gray-800 border border-gray-600 rounded shadow-xl max-h-64 flex flex-col">
          <div className="p-1.5 border-b border-gray-700">
            <input
              ref={inputRef}
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={`Search ${type}s...`}
              className="w-full px-2 py-1 bg-gray-900 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-gray-400"
            />
          </div>
          <div className="overflow-y-auto flex-1">
            {value > 0 && (
              <button
                type="button"
                onClick={() => {
                  onChange(0);
                  setOpen(false);
                }}
                className="w-full text-left px-3 py-1.5 text-gray-500 text-sm hover:bg-gray-700 border-b border-gray-700"
              >
                Clear selection
              </button>
            )}
            {filtered.length === 0 && (
              <div className="px-3 py-2 text-gray-500 text-sm">No matches</div>
            )}
            {filtered.slice(0, 50).map((opt) => (
              <button
                key={opt.id}
                type="button"
                onClick={() => {
                  onChange(opt.id);
                  setOpen(false);
                }}
                className={`w-full text-left px-3 py-1.5 text-sm hover:bg-gray-700 flex items-center justify-between ${
                  opt.id === value ? color.highlight : ''
                }`}
              >
                <span>
                  <span className="text-gray-500 font-mono">#{opt.id}</span>{' '}
                  <span className={color.text}>{opt.label}</span>
                </span>
                {opt.sublabel && (
                  <span className="text-gray-500 text-xs ml-2">{opt.sublabel}</span>
                )}
              </button>
            ))}
            {filtered.length > 50 && (
              <div className="px-3 py-1.5 text-gray-500 text-xs text-center">
                {filtered.length - 50} more — refine search
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

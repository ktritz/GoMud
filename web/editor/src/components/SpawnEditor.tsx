import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mobs, items, type MobSummary, type ItemSummary } from '../api/client';
import { EntityPicker } from './EntityPicker';

interface SpawnInfo {
  MobId?: number;
  ItemId?: number;
  Message?: string;
  RespawnRate?: string;
  Container?: string;
}

interface SpawnEditorProps {
  spawns: SpawnInfo[];
  onSave: (spawns: SpawnInfo[]) => void;
}

function stripAnsi(text: string): string {
  return text.replace(/<\/?ansi[^>]*>/g, '');
}

export function SpawnEditor({ spawns, onSave }: SpawnEditorProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<SpawnInfo[]>([]);

  // Fetch names for read-only display
  const { data: mobData } = useQuery({ queryKey: ['mobs'], queryFn: () => mobs.list(), staleTime: 60_000 });
  const { data: itemData } = useQuery({ queryKey: ['items'], queryFn: () => items.list(), staleTime: 60_000 });
  const mobNames = new Map((mobData?.mobs || []).map((m: MobSummary) => [m.mobId, m.name]));
  const itemNames = new Map((itemData?.items || []).map((it: ItemSummary) => [it.itemId, it.name]));

  const startEdit = () => {
    setDraft(JSON.parse(JSON.stringify(spawns || [])));
    setEditing(true);
  };

  const handleSave = () => {
    onSave(draft.filter((s) => s.MobId || s.ItemId));
    setEditing(false);
  };

  const addSpawn = () => {
    setDraft([...draft, { MobId: 0, Message: '', RespawnRate: '5 real minutes' }]);
  };

  const removeSpawn = (index: number) => {
    setDraft(draft.filter((_, i) => i !== index));
  };

  const updateSpawn = (index: number, field: string, value: any) => {
    const updated = [...draft];
    updated[index] = { ...updated[index], [field]: value };
    setDraft(updated);
  };

  if (!editing) {
    return (
      <div>
        {spawns && spawns.length > 0 ? (
          <div className="space-y-2">
            {spawns.map((spawn, i) => (
              <div key={i} className="bg-gray-800 rounded px-3 py-2">
                <div className="flex items-center gap-2">
                  {spawn.MobId ? (
                    <Link to={`/mobs/${spawn.MobId}`} className="text-blue-400 hover:text-blue-300">
                      #{spawn.MobId} {mobNames.get(spawn.MobId) || 'Mob'}
                    </Link>
                  ) : null}
                  {spawn.ItemId ? (
                    <Link to={`/items/${spawn.ItemId}`} className="text-green-400 hover:text-green-300">
                      #{spawn.ItemId} {itemNames.get(spawn.ItemId) || 'Item'}
                    </Link>
                  ) : null}
                  <span className="text-gray-500 text-sm">{spawn.RespawnRate}</span>
                </div>
                {spawn.Message && (
                  <p className="text-gray-400 text-sm mt-1">{stripAnsi(spawn.Message)}</p>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No spawns</p>
        )}
        <button
          onClick={startEdit}
          className="mt-3 px-3 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 text-sm rounded"
        >
          Edit Spawns
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {draft.map((spawn, i) => (
        <div key={i} className="bg-gray-800 rounded px-3 py-3 space-y-2">
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-400 w-16">Mob</label>
            <EntityPicker
              type="mob"
              value={spawn.MobId || 0}
              onChange={(id) => updateSpawn(i, 'MobId', id)}
              className="w-56"
            />
            <label className="text-sm text-gray-400 w-16 ml-2">Item</label>
            <EntityPicker
              type="item"
              value={spawn.ItemId || 0}
              onChange={(id) => updateSpawn(i, 'ItemId', id)}
              className="w-56"
            />
            <button
              onClick={() => removeSpawn(i)}
              className="ml-auto text-red-400 hover:text-red-300 text-sm"
            >
              ✕
            </button>
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-400 w-16">Rate</label>
            <input
              type="text"
              value={spawn.RespawnRate || ''}
              onChange={(e) => updateSpawn(i, 'RespawnRate', e.target.value)}
              className="w-40 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
              placeholder="5 real minutes"
            />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-400 w-16">Message</label>
            <input
              type="text"
              value={spawn.Message || ''}
              onChange={(e) => updateSpawn(i, 'Message', e.target.value)}
              className="flex-1 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
              placeholder="A mob appears..."
            />
          </div>
        </div>
      ))}

      <button
        onClick={addSpawn}
        className="px-3 py-1 bg-green-700 hover:bg-green-600 text-white text-sm rounded"
      >
        + Add Spawn
      </button>

      <div className="flex gap-2">
        <button
          onClick={handleSave}
          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded"
        >
          Save Spawns
        </button>
        <button
          onClick={() => setEditing(false)}
          className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}

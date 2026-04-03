import { useState } from 'react';
import { Link } from 'react-router-dom';

interface Exit {
  RoomId: number;
  Secret?: boolean;
  Lock?: { Difficulty: number };
  MapDirection?: string;
}

interface ExitEditorProps {
  exits: Record<string, Exit>;
  onSave: (exits: Record<string, Exit>) => void;
}

export function ExitEditor({ exits, onSave }: ExitEditorProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<Record<string, Exit>>({});
  const [newDir, setNewDir] = useState('');
  const [newRoomId, setNewRoomId] = useState('');

  const startEdit = () => {
    setDraft(JSON.parse(JSON.stringify(exits || {})));
    setEditing(true);
  };

  const handleSave = () => {
    onSave(draft);
    setEditing(false);
  };

  const addExit = () => {
    if (!newDir || !newRoomId) return;
    setDraft({ ...draft, [newDir]: { RoomId: parseInt(newRoomId) } });
    setNewDir('');
    setNewRoomId('');
  };

  const removeExit = (dir: string) => {
    const updated = { ...draft };
    delete updated[dir];
    setDraft(updated);
  };

  const toggleSecret = (dir: string) => {
    const updated = { ...draft };
    updated[dir] = { ...updated[dir], Secret: !updated[dir].Secret };
    setDraft(updated);
  };

  if (!editing) {
    return (
      <div>
        {exits && Object.keys(exits).length > 0 ? (
          <div className="grid grid-cols-2 gap-2">
            {Object.entries(exits).map(([dir, exit]) => (
              <Link
                key={dir}
                to={`/rooms/${exit.RoomId}`}
                className="flex items-center justify-between bg-gray-800 rounded px-3 py-2 hover:bg-gray-700 transition-colors"
              >
                <span className="font-mono text-blue-400">{dir}</span>
                <span className="text-gray-400">→ #{exit.RoomId}</span>
                {exit.Secret && <span className="text-yellow-500 text-xs ml-2">secret</span>}
                {exit.Lock?.Difficulty ? (
                  <span className="text-red-400 text-xs ml-2">locked ({exit.Lock.Difficulty})</span>
                ) : null}
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No exits</p>
        )}
        <button
          onClick={startEdit}
          className="mt-3 px-3 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 text-sm rounded"
        >
          Edit Exits
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {Object.entries(draft).map(([dir, exit]) => (
        <div key={dir} className="flex items-center gap-2 bg-gray-800 rounded px-3 py-2">
          <span className="font-mono text-blue-400 w-24">{dir}</span>
          <span className="text-gray-400">→</span>
          <input
            type="number"
            value={exit.RoomId}
            onChange={(e) => {
              const updated = { ...draft };
              updated[dir] = { ...updated[dir], RoomId: parseInt(e.target.value) || 0 };
              setDraft(updated);
            }}
            className="w-20 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          />
          <label className="flex items-center gap-1 text-sm text-gray-400">
            <input
              type="checkbox"
              checked={exit.Secret || false}
              onChange={() => toggleSecret(dir)}
              className="rounded"
            />
            Secret
          </label>
          <button
            onClick={() => removeExit(dir)}
            className="ml-auto text-red-400 hover:text-red-300 text-sm"
          >
            ✕
          </button>
        </div>
      ))}

      <div className="flex items-center gap-2">
        <input
          type="text"
          placeholder="direction"
          value={newDir}
          onChange={(e) => setNewDir(e.target.value)}
          className="w-32 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
        />
        <span className="text-gray-400">→</span>
        <input
          type="number"
          placeholder="room ID"
          value={newRoomId}
          onChange={(e) => setNewRoomId(e.target.value)}
          className="w-20 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
        />
        <button
          onClick={addExit}
          className="px-3 py-1 bg-green-700 hover:bg-green-600 text-white text-sm rounded"
        >
          + Add
        </button>
      </div>

      <div className="flex gap-2">
        <button
          onClick={handleSave}
          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded"
        >
          Save Exits
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

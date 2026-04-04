import { useState } from 'react';

// --- EditableText ---

interface EditableTextProps {
  value: string;
  onSave: (value: string) => void;
  label?: string;
  multiline?: boolean;
}

export function EditableText({ value, onSave, multiline }: EditableTextProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);

  const handleSave = () => {
    onSave(draft);
    setEditing(false);
  };

  const handleCancel = () => {
    setDraft(value);
    setEditing(false);
  };

  if (!editing) {
    return (
      <div
        className="group cursor-pointer hover:bg-gray-800/50 rounded px-2 py-1 -mx-2 transition-colors"
        onClick={() => { setDraft(value); setEditing(true); }}
      >
        {multiline ? (
          <p className="text-gray-300 whitespace-pre-wrap">{value || <span className="text-gray-600 italic">Click to edit</span>}</p>
        ) : (
          <span className="text-gray-300">{value || <span className="text-gray-600 italic">Click to edit</span>}</span>
        )}
        <span className="text-gray-600 text-xs ml-2 opacity-0 group-hover:opacity-100 transition-opacity">&#9998;</span>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {multiline ? (
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          className="w-full px-3 py-2 bg-gray-700 border border-blue-500 rounded text-white text-sm focus:outline-none min-h-24"
          autoFocus
          rows={4}
        />
      ) : (
        <input
          type="text"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          className="w-full px-3 py-2 bg-gray-700 border border-blue-500 rounded text-white text-sm focus:outline-none"
          autoFocus
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSave();
            if (e.key === 'Escape') handleCancel();
          }}
        />
      )}
      <div className="flex gap-2">
        <button onClick={handleSave} className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded">Save</button>
        <button onClick={handleCancel} className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded">Cancel</button>
      </div>
    </div>
  );
}

// --- EditableNumber ---

interface EditableNumberProps {
  value: number;
  onSave: (value: number) => void;
  label?: string;
  min?: number;
  max?: number;
  float?: boolean;
}

export function EditableNumber({ value, onSave, min, max, float }: EditableNumberProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(String(value));

  const handleSave = () => {
    const parsed = float ? parseFloat(draft) : parseInt(draft, 10);
    if (!isNaN(parsed)) {
      onSave(parsed);
    }
    setEditing(false);
  };

  const handleCancel = () => {
    setDraft(String(value));
    setEditing(false);
  };

  if (!editing) {
    return (
      <div
        className="group cursor-pointer hover:bg-gray-800/50 rounded px-2 py-1 -mx-2 transition-colors inline-block"
        onClick={() => { setDraft(String(value)); setEditing(true); }}
      >
        <span className="text-gray-300">{value}</span>
        <span className="text-gray-600 text-xs ml-2 opacity-0 group-hover:opacity-100 transition-opacity">&#9998;</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <input
        type="number"
        value={draft}
        min={min}
        max={max}
        step={float ? 'any' : 1}
        onChange={(e) => setDraft(e.target.value)}
        className="w-24 px-2 py-1 bg-gray-700 border border-blue-500 rounded text-white text-sm focus:outline-none"
        autoFocus
        onKeyDown={(e) => {
          if (e.key === 'Enter') handleSave();
          if (e.key === 'Escape') handleCancel();
        }}
      />
      <button onClick={handleSave} className="px-2 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded">&#10003;</button>
      <button onClick={handleCancel} className="px-2 py-1 bg-gray-600 hover:bg-gray-500 text-white text-xs rounded">&#10005;</button>
    </div>
  );
}

// --- EditableBoolean ---

interface EditableBooleanProps {
  value: boolean;
  onSave: (value: boolean) => void;
  label?: string;
}

export function EditableBoolean({ value, onSave }: EditableBooleanProps) {
  return (
    <button
      onClick={() => onSave(!value)}
      className={`px-2 py-0.5 rounded text-sm cursor-pointer transition-colors ${
        value
          ? 'bg-green-900/50 text-green-300 hover:bg-green-800/50'
          : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
      }`}
    >
      {value ? 'Yes' : 'No'}
    </button>
  );
}

// --- EditableSelect ---

interface EditableSelectProps {
  value: string;
  options: string[];
  onSave: (value: string) => void;
  label?: string;
}

export function EditableSelect({ value, options, onSave }: EditableSelectProps) {
  return (
    <select
      value={value}
      onChange={(e) => onSave(e.target.value)}
      className="px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500 cursor-pointer"
    >
      {options.map((opt) => (
        <option key={opt} value={opt}>{opt}</option>
      ))}
    </select>
  );
}

// --- EditableList ---

interface EditableListProps {
  value: string[];
  onSave: (value: string[]) => void;
  label?: string;
  placeholder?: string;
}

export function EditableList({ value, onSave, placeholder }: EditableListProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<string[]>([]);
  const [newItem, setNewItem] = useState('');

  const startEdit = () => {
    setDraft([...(value || [])]);
    setEditing(true);
  };

  const handleSave = () => {
    onSave(draft.filter(Boolean));
    setEditing(false);
  };

  const addItem = () => {
    if (newItem.trim()) {
      setDraft([...draft, newItem.trim()]);
      setNewItem('');
    }
  };

  if (!editing) {
    return (
      <div
        className="group cursor-pointer hover:bg-gray-800/50 rounded px-2 py-1 -mx-2 transition-colors"
        onClick={startEdit}
      >
        {value && value.length > 0 ? (
          <div className="flex flex-wrap gap-1">
            {value.map((item, i) => (
              <span key={i} className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{item}</span>
            ))}
          </div>
        ) : (
          <span className="text-gray-600 italic">Click to edit</span>
        )}
        <span className="text-gray-600 text-xs ml-2 opacity-0 group-hover:opacity-100 transition-opacity">&#9998;</span>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {draft.map((item, i) => (
        <div key={i} className="flex items-center gap-2">
          <input
            type="text"
            value={item}
            onChange={(e) => {
              const updated = [...draft];
              updated[i] = e.target.value;
              setDraft(updated);
            }}
            className="flex-1 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          />
          <button
            onClick={() => setDraft(draft.filter((_, j) => j !== i))}
            className="text-red-400 hover:text-red-300 text-sm"
          >&#10005;</button>
        </div>
      ))}
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={newItem}
          onChange={(e) => setNewItem(e.target.value)}
          placeholder={placeholder || 'Add item...'}
          className="flex-1 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          onKeyDown={(e) => { if (e.key === 'Enter') addItem(); }}
        />
        <button onClick={addItem} className="px-2 py-1 bg-green-700 hover:bg-green-600 text-white text-xs rounded">+</button>
      </div>
      <div className="flex gap-2">
        <button onClick={handleSave} className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded">Save</button>
        <button onClick={() => setEditing(false)} className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded">Cancel</button>
      </div>
    </div>
  );
}

// --- EditableMap ---

interface EditableMapProps {
  value: Record<string, any>;
  onSave: (value: Record<string, any>) => void;
  label?: string;
  valueType?: 'string' | 'number';
}

export function EditableMap({ value, onSave, valueType = 'string' }: EditableMapProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<[string, string][]>([]);
  const [newKey, setNewKey] = useState('');
  const [newVal, setNewVal] = useState('');

  const startEdit = () => {
    setDraft(Object.entries(value || {}).map(([k, v]) => [k, String(v)]));
    setEditing(true);
  };

  const handleSave = () => {
    const result: Record<string, any> = {};
    for (const [k, v] of draft) {
      if (k.trim()) {
        result[k.trim()] = valueType === 'number' ? Number(v) || 0 : v;
      }
    }
    onSave(result);
    setEditing(false);
  };

  const addEntry = () => {
    if (newKey.trim()) {
      setDraft([...draft, [newKey.trim(), newVal]]);
      setNewKey('');
      setNewVal('');
    }
  };

  if (!editing) {
    const entries = Object.entries(value || {});
    return (
      <div
        className="group cursor-pointer hover:bg-gray-800/50 rounded px-2 py-1 -mx-2 transition-colors"
        onClick={startEdit}
      >
        {entries.length > 0 ? (
          <div className="grid grid-cols-3 gap-2 text-sm">
            {entries.map(([k, v]) => (
              <div key={k}>
                <span className="text-gray-500">{k}:</span>{' '}
                <span className={valueType === 'number' && Number(v) > 0 ? 'text-green-400' : valueType === 'number' && Number(v) < 0 ? 'text-red-400' : 'text-gray-300'}>
                  {valueType === 'number' && Number(v) > 0 ? `+${v}` : String(v)}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <span className="text-gray-600 italic">Click to edit</span>
        )}
        <span className="text-gray-600 text-xs ml-2 opacity-0 group-hover:opacity-100 transition-opacity">&#9998;</span>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {draft.map(([k, v], i) => (
        <div key={i} className="flex items-center gap-2">
          <input
            type="text"
            value={k}
            onChange={(e) => {
              const updated = [...draft] as [string, string][];
              updated[i] = [e.target.value, v];
              setDraft(updated);
            }}
            className="w-32 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          />
          <input
            type={valueType === 'number' ? 'number' : 'text'}
            value={v}
            onChange={(e) => {
              const updated = [...draft] as [string, string][];
              updated[i] = [k, e.target.value];
              setDraft(updated);
            }}
            className="w-24 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          />
          <button
            onClick={() => setDraft(draft.filter((_, j) => j !== i))}
            className="text-red-400 hover:text-red-300 text-sm"
          >&#10005;</button>
        </div>
      ))}
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          placeholder="key"
          className="w-32 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
        />
        <input
          type={valueType === 'number' ? 'number' : 'text'}
          value={newVal}
          onChange={(e) => setNewVal(e.target.value)}
          placeholder="value"
          className="w-24 px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm"
          onKeyDown={(e) => { if (e.key === 'Enter') addEntry(); }}
        />
        <button onClick={addEntry} className="px-2 py-1 bg-green-700 hover:bg-green-600 text-white text-xs rounded">+</button>
      </div>
      <div className="flex gap-2">
        <button onClick={handleSave} className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded">Save</button>
        <button onClick={() => setEditing(false)} className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded">Cancel</button>
      </div>
    </div>
  );
}

// --- Shared Section + Prop (used across detail pages) ---

export function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-6">
      <h2 className="text-lg font-semibold text-gray-200 mb-2 border-b border-gray-700 pb-1">
        {title}
      </h2>
      {children}
    </section>
  );
}

// --- SaveStatus banner ---

interface SaveStatusProps {
  isPending: boolean;
  isError: boolean;
  isSuccess: boolean;
  error?: Error | null;
}

export function SaveStatus({ isPending, isError, isSuccess, error }: SaveStatusProps) {
  return (
    <>
      {isPending && (
        <div className="bg-blue-900/30 border border-blue-700 text-blue-300 px-3 py-1 rounded mb-4 text-sm">Saving...</div>
      )}
      {isError && (
        <div className="bg-red-900/30 border border-red-700 text-red-300 px-3 py-1 rounded mb-4 text-sm">
          Error: {error?.message || 'Save failed'}
        </div>
      )}
      {isSuccess && (
        <div className="bg-green-900/30 border border-green-700 text-green-300 px-3 py-1 rounded mb-4 text-sm">Saved!</div>
      )}
    </>
  );
}

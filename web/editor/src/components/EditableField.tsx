import { useState } from 'react';

interface EditableTextProps {
  value: string;
  onSave: (value: string) => void;
  label: string;
  multiline?: boolean;
}

export function EditableText({ value, onSave, label, multiline }: EditableTextProps) {
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
        <span className="text-gray-600 text-xs ml-2 opacity-0 group-hover:opacity-100 transition-opacity">✎</span>
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
        <button
          onClick={handleSave}
          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded"
        >
          Save
        </button>
        <button
          onClick={handleCancel}
          className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}

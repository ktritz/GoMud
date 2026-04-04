import { useState } from 'react';

interface FieldDef {
  key: string;
  label: string;
  type: 'text' | 'number' | 'select';
  required?: boolean;
  options?: string[];
  placeholder?: string;
}

interface CreateEntityModalProps {
  title: string;
  fields: FieldDef[];
  onSubmit: (data: Record<string, any>) => Promise<any>;
  onCreated: (result: any) => void;
  onClose: () => void;
}

export function CreateEntityModal({ title, fields, onSubmit, onCreated, onClose }: CreateEntityModalProps) {
  const [values, setValues] = useState<Record<string, any>>(() => {
    const initial: Record<string, any> = {};
    for (const f of fields) {
      initial[f.key] = f.type === 'number' ? 0 : f.options?.[0] || '';
    }
    return initial;
  });
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    for (const f of fields) {
      if (f.required && !values[f.key]) {
        setError(`${f.label} is required`);
        return;
      }
    }

    setSaving(true);
    try {
      const result = await onSubmit(values);
      onCreated(result);
    } catch (err: any) {
      setError(err.message || 'Failed to create');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-gray-800 border border-gray-600 rounded-lg p-6 w-96 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-white mb-4">{title}</h2>

        {error && (
          <div className="bg-red-900/30 border border-red-700 text-red-300 px-3 py-1 rounded mb-3 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-3">
          {fields.map((f) => (
            <div key={f.key}>
              <label className="block text-sm text-gray-400 mb-1">
                {f.label}{f.required && <span className="text-red-400"> *</span>}
              </label>
              {f.type === 'select' ? (
                <select
                  value={values[f.key]}
                  onChange={(e) => setValues({ ...values, [f.key]: e.target.value })}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                >
                  {f.options?.map((opt) => (
                    <option key={opt} value={opt}>{opt}</option>
                  ))}
                </select>
              ) : (
                <input
                  type={f.type}
                  value={values[f.key]}
                  onChange={(e) =>
                    setValues({
                      ...values,
                      [f.key]: f.type === 'number' ? Number(e.target.value) : e.target.value,
                    })
                  }
                  placeholder={f.placeholder}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
                />
              )}
            </div>
          ))}

          <div className="flex gap-2 pt-2">
            <button
              type="submit"
              disabled={saving}
              className="flex-1 px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 text-white text-sm rounded font-medium"
            >
              {saving ? 'Creating...' : 'Create'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-2 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

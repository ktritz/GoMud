import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { Section } from './EditableField';

interface ScriptEditorProps {
  entityType: string;  // 'mob' | 'room' | 'item' | 'buff'
  entityId: number;
}

export function ScriptEditor({ entityType, entityId }: ScriptEditorProps) {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState('');

  const { data: scriptData } = useQuery({
    queryKey: ['script', entityType, entityId],
    queryFn: () => api.get<{ content: string; exists: boolean; path: string }>(`/admin/scripts/${entityType}/${entityId}`),
    enabled: entityId > 0,
  });

  const saveMut = useMutation({
    mutationFn: (content: string) =>
      api.put(`/admin/scripts/${entityType}/${entityId}`, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['script', entityType, entityId] });
      setEditing(false);
    },
  });

  useEffect(() => {
    if (scriptData) {
      setDraft(scriptData.content);
    }
  }, [scriptData]);

  const hasScript = scriptData?.exists && scriptData.content.length > 0;

  return (
    <Section title="Script">
      {!editing ? (
        <div>
          {hasScript ? (
            <pre className="text-xs font-mono text-gray-300 bg-gray-800 rounded p-3 overflow-auto max-h-64 whitespace-pre">
              {scriptData!.content}
            </pre>
          ) : (
            <p className="text-gray-500">No script attached</p>
          )}
          <button
            onClick={() => {
              setDraft(scriptData?.content || '');
              setEditing(true);
            }}
            className="mt-2 px-3 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 text-sm rounded"
          >
            {hasScript ? 'Edit Script' : 'Add Script'}
          </button>
        </div>
      ) : (
        <div className="space-y-2">
          <div className="relative">
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              className="w-full h-80 px-3 py-2 bg-gray-900 border border-blue-500 rounded text-gray-200 text-xs font-mono focus:outline-none resize-y"
              spellCheck={false}
              autoFocus
              onKeyDown={(e) => {
                // Tab support
                if (e.key === 'Tab') {
                  e.preventDefault();
                  const start = e.currentTarget.selectionStart;
                  const end = e.currentTarget.selectionEnd;
                  setDraft(draft.substring(0, start) + '  ' + draft.substring(end));
                  setTimeout(() => {
                    e.currentTarget.selectionStart = e.currentTarget.selectionEnd = start + 2;
                  }, 0);
                }
              }}
            />
            <div className="absolute top-1 right-2 text-xs text-gray-600">
              {draft.split('\n').length} lines
            </div>
          </div>

          {saveMut.isError && (
            <div className="bg-red-900/30 border border-red-700 text-red-300 px-3 py-1 rounded text-sm">
              {(saveMut.error as Error).message}
            </div>
          )}

          <div className="flex gap-2">
            <button
              onClick={() => saveMut.mutate(draft)}
              disabled={saveMut.isPending}
              className="px-3 py-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 text-white text-sm rounded"
            >
              {saveMut.isPending ? 'Saving...' : 'Save Script'}
            </button>
            <button
              onClick={() => setEditing(false)}
              className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded"
            >
              Cancel
            </button>
            {hasScript && (
              <button
                onClick={() => {
                  if (confirm('Delete this script?')) {
                    saveMut.mutate('');
                  }
                }}
                className="px-3 py-1 bg-red-700 hover:bg-red-600 text-white text-sm rounded ml-auto"
              >
                Delete Script
              </button>
            )}
          </div>

          {scriptData?.path && (
            <div className="text-xs text-gray-600 font-mono">{scriptData.path}</div>
          )}
        </div>
      )}
    </Section>
  );
}

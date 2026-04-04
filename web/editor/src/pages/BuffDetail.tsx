import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { EditableText, EditableNumber, EditableBoolean, EditableMap, Section, SaveStatus } from '../components/EditableField';
import { useBuffReferences } from '../hooks/useReferences';
import { ReferencesPanel } from '../components/ReferencesPanel';
import { DetailActions } from '../components/DetailActions';

export function BuffDetail() {
  const { buffId } = useParams();
  const id = Number(buffId);
  const queryClient = useQueryClient();

  const { data: buff, isLoading, error } = useQuery({
    queryKey: ['buffs', id],
    queryFn: () => api.get<any>(`/admin/buffs/${id}`),
    enabled: !isNaN(id),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/buffs/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buffs', id] });
      queryClient.invalidateQueries({ queryKey: ['buffs'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  const references = useBuffReferences(id);

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!buff) return <div className="text-gray-400">Buff not found</div>;

  const flags = buff.Flags || [];

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/buffs" className="text-gray-400 hover:text-white">&larr; Buffs</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{buff.BuffId}</span>
      </div>

      <DetailActions
        entityType="buffs"
        entityId={id}
        entityName={buff.Name || 'buff'}
        listPath="/buffs"
        detailPath="/buffs"
        duplicateData={{ Name: `${buff.Name} (Copy)`, Description: buff.Description }}
      />
      <SaveStatus
        isPending={updateMutation.isPending}
        isError={updateMutation.isError}
        isSuccess={updateMutation.isSuccess}
        error={updateMutation.error as Error}
      />

      <Section title="Identity">
        <div className="space-y-3">
          <div>
            <label className="text-gray-500 text-sm block mb-1">Name</label>
            <EditableText value={buff.Name || ''} onSave={(v) => saveField('Name', v)} />
          </div>
        </div>
      </Section>

      <Section title="Description">
        <EditableText
          value={buff.Description || ''}
          onSave={(v) => saveField('Description', v)}
          multiline
        />
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <label className="text-gray-500 block mb-1">Trigger Rate</label>
            <EditableText value={buff.TriggerRate || ''} onSave={(v) => saveField('TriggerRate', v)} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Trigger Count</label>
            <EditableNumber value={buff.TriggerCount || 0} onSave={(v) => saveField('TriggerCount', v)} />
          </div>
        </div>
      </Section>

      <Section title="Flags">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Secret: </span>
            <EditableBoolean value={!!buff.Secret} onSave={(v) => saveField('Secret', v)} />
          </div>
          <div>
            <span className="text-gray-500">Trigger Now: </span>
            <EditableBoolean value={!!buff.TriggerNow} onSave={(v) => saveField('TriggerNow', v)} />
          </div>
        </div>
        {flags.length > 0 && (
          <div className="flex flex-wrap gap-2 mt-3">
            {flags.map((flag: string) => (
              <span key={flag} className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">
                {flag}
              </span>
            ))}
          </div>
        )}
      </Section>

      <Section title="Stat Modifiers">
        <EditableMap
          value={buff.StatMods || {}}
          onSave={(v) => saveField('StatMods', v)}
          valueType="number"
        />
      </Section>

      <ReferencesPanel references={references} />

      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(buff, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

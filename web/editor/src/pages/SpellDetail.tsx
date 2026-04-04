import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { EditableText, EditableNumber, EditableSelect, Section, SaveStatus } from '../components/EditableField';
import { DetailActions } from '../components/DetailActions';

const typeOptions = [
  'neutral', 'harmsingle', 'harmmulti', 'helpsingle', 'helpmulti', 'harmarea', 'helparea',
];

const schoolOptions = [
  'restoration', 'illusion', 'conjuration',
];

export function SpellDetail() {
  const { spellId } = useParams();
  const queryClient = useQueryClient();

  const { data: spell, isLoading, error } = useQuery({
    queryKey: ['spells', spellId],
    queryFn: () => api.get<any>(`/admin/spells/${spellId}`),
    enabled: !!spellId,
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/spells/${spellId}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['spells', spellId] });
      queryClient.invalidateQueries({ queryKey: ['spells'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!spell) return <div className="text-gray-400">Spell not found</div>;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/spells" className="text-gray-400 hover:text-white">&larr; Spells</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">{spell.SpellId}</span>
      </div>

      <DetailActions
        entityType="spells"
        entityId={spellId!}
        entityName={spell.Name || 'spell'}
        listPath="/spells"
        detailPath="/spells"
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
            <EditableText value={spell.Name || ''} onSave={(v) => saveField('Name', v)} />
          </div>
        </div>
      </Section>

      <Section title="Description">
        <EditableText
          value={spell.Description || ''}
          onSave={(v) => saveField('Description', v)}
          multiline
        />
      </Section>

      <Section title="Classification">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-gray-500 text-sm block mb-1">Type</label>
            <EditableSelect
              value={spell.Type || ''}
              options={typeOptions}
              onSave={(v) => saveField('Type', v)}
            />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">School</label>
            <EditableSelect
              value={spell.School || ''}
              options={schoolOptions}
              onSave={(v) => saveField('School', v)}
            />
          </div>
        </div>
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <label className="text-gray-500 block mb-1">Cost</label>
            <EditableNumber value={spell.Cost || 0} onSave={(v) => saveField('Cost', v)} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Wait Rounds</label>
            <EditableNumber value={spell.WaitRounds || 0} onSave={(v) => saveField('WaitRounds', v)} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Difficulty</label>
            <EditableNumber value={spell.Difficulty || 0} onSave={(v) => saveField('Difficulty', v)} min={0} max={100} />
          </div>
        </div>
      </Section>

      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(spell, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

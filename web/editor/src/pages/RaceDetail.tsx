import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { EditableText, EditableNumber, EditableBoolean, EditableSelect, Section, SaveStatus } from '../components/EditableField';
import { DetailActions } from '../components/DetailActions';

const sizeOptions = ['small', 'medium', 'large'];

export function RaceDetail() {
  const { raceId } = useParams();
  const id = Number(raceId);
  const queryClient = useQueryClient();

  const { data: race, isLoading, error } = useQuery({
    queryKey: ['races', id],
    queryFn: () => api.get<any>(`/admin/races/${id}`),
    enabled: !isNaN(id),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/races/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['races', id] });
      queryClient.invalidateQueries({ queryKey: ['races'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!race) return <div className="text-gray-400">Race not found</div>;

  const buffIds = race.BuffIds || [];
  const angryCommands = race.AngryCommands || [];
  const disabledSlots = race.DisabledSlots || [];
  const damage = race.Damage || {};
  const stats = race.Stats || {};

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/races" className="text-gray-400 hover:text-white">&larr; Races</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{race.RaceId}</span>
      </div>

      <DetailActions
        entityType="races"
        entityId={id}
        entityName={race.Name || 'race'}
        listPath="/races"
        detailPath="/races"
        duplicateData={{ Name: `${race.Name} (Copy)`, Description: race.Description, Size: race.Size }}
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
            <EditableText value={race.Name || ''} onSave={(v) => saveField('Name', v)} />
          </div>
        </div>
      </Section>

      <Section title="Description">
        <EditableText
          value={race.Description || ''}
          onSave={(v) => saveField('Description', v)}
          multiline
        />
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <label className="text-gray-500 block mb-1">Size</label>
            <EditableSelect
              value={race.Size || ''}
              options={sizeOptions}
              onSave={(v) => saveField('Size', v)}
            />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Default Alignment</label>
            <EditableNumber value={race.DefaultAlignment || 0} onSave={(v) => saveField('DefaultAlignment', v)} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">TNL Scale</label>
            <EditableNumber value={race.TNLScale || 0} onSave={(v) => saveField('TNLScale', v)} float />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Unarmed Name</label>
            <EditableText value={race.UnarmedName || ''} onSave={(v) => saveField('UnarmedName', v)} />
          </div>
        </div>
      </Section>

      <Section title="Flags">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Tameable: </span>
            <EditableBoolean value={!!race.Tameable} onSave={(v) => saveField('Tameable', v)} />
          </div>
          <div>
            <span className="text-gray-500">Selectable: </span>
            <EditableBoolean value={!!race.Selectable} onSave={(v) => saveField('Selectable', v)} />
          </div>
          <div>
            <span className="text-gray-500">Knows First Aid: </span>
            <EditableBoolean value={!!race.KnowsFirstAid} onSave={(v) => saveField('KnowsFirstAid', v)} />
          </div>
        </div>
      </Section>

      <Section title="Buff IDs">
        {buffIds.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {buffIds.map((id: number) => (
              <Link
                key={id}
                to={`/buffs/${id}`}
                className="px-2 py-0.5 bg-gray-700 rounded text-sm text-blue-300 hover:text-blue-200"
              >
                #{id}
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">None</p>
        )}
      </Section>

      <Section title="Angry Commands">
        {angryCommands.length > 0 ? (
          <ul className="space-y-1">
            {angryCommands.map((cmd: string, i: number) => (
              <li key={i} className="text-gray-400 text-sm font-mono">{cmd}</li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">None</p>
        )}
      </Section>

      <Section title="Disabled Slots">
        {disabledSlots.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {disabledSlots.map((slot: string) => (
              <span key={slot} className="px-2 py-0.5 bg-red-900/50 rounded text-sm text-red-300">
                {slot}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">None</p>
        )}
      </Section>

      <Section title="Damage">
        {Object.keys(damage).length > 0 ? (
          <div className="grid grid-cols-3 gap-4 text-sm">
            {Object.entries(damage).map(([key, val]) => (
              <div key={key}>
                <span className="text-gray-500">{key}:</span>{' '}
                <span className="text-gray-300">{String(val)}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No damage data</p>
        )}
      </Section>

      <Section title="Stats">
        {Object.keys(stats).length > 0 ? (
          <div className="grid grid-cols-3 gap-2 text-sm">
            {Object.entries(stats).map(([stat, val]) => (
              <div key={stat}>
                <span className="text-gray-500">{stat}:</span>{' '}
                <span
                  className={
                    Number(val) > 0
                      ? 'text-green-400'
                      : Number(val) < 0
                      ? 'text-red-400'
                      : 'text-gray-300'
                  }
                >
                  {Number(val) > 0 ? `+${val}` : String(val)}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No stats</p>
        )}
      </Section>

      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(race, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

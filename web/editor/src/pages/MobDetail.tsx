import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mobs, api } from '../api/client';
import { EditableText, EditableNumber, EditableBoolean, EditableList, Section, SaveStatus } from '../components/EditableField';
import { useMobValidation, ValidationPanel } from '../hooks/useValidation';
import { useMobReferences } from '../hooks/useReferences';
import { ReferencesPanel } from '../components/ReferencesPanel';
import { DetailActions } from '../components/DetailActions';

export function MobDetail() {
  const { mobId } = useParams();
  const id = Number(mobId);
  const queryClient = useQueryClient();

  const { data: mob, isLoading, error } = useQuery({
    queryKey: ['mobs', id],
    queryFn: () => mobs.get(id),
    enabled: !isNaN(id),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/mobs/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mobs', id] });
      queryClient.invalidateQueries({ queryKey: ['mobs'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  const warnings = useMobValidation(mob);
  const references = useMobReferences(id);

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!mob) return <div className="text-gray-400">Mob not found</div>;

  const char = mob.Character || {};

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/mobs" className="text-gray-400 hover:text-white">&larr; Mobs</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{mob.MobId}</span>
      </div>

      <DetailActions
        entityType="mobs"
        entityId={id}
        entityName={char.Name || 'mob'}
        listPath="/mobs"
        detailPath="/mobs"
        duplicateData={{ Name: `${char.Name} (Copy)`, Zone: mob.Zone, Level: char.Level }}
      />
      <ValidationPanel warnings={warnings} />
      <SaveStatus
        isPending={updateMutation.isPending}
        isError={updateMutation.isError}
        isSuccess={updateMutation.isSuccess}
        error={updateMutation.error as Error}
      />

      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold">
          <EditableText value={char.Name || ''} onSave={(v) => saveField('CharName', v)} label="Name" />
        </h1>
        <span className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{mob.Zone}</span>
        {mob.Hostile && (
          <span className="px-2 py-0.5 bg-red-900/50 rounded text-sm text-red-300">Hostile</span>
        )}
        <span className="px-2 py-0.5 bg-blue-900/50 rounded text-sm text-blue-300">
          Level {char.Level}
        </span>
      </div>

      <Section title="Description">
        <EditableText
          value={char.Description || ''}
          onSave={(v) => saveField('CharDescription', v)}
          label="Description"
          multiline
        />
      </Section>

      <Section title="Character Stats">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Level: </span>
            <EditableNumber value={char.Level || 0} onSave={(v) => saveField('CharLevel', v)} label="Level" min={1} />
          </div>
          <div>
            <span className="text-gray-500">Gold: </span>
            <EditableNumber value={char.Gold || 0} onSave={(v) => saveField('CharGold', v)} label="Gold" min={0} />
          </div>
          <div>
            <span className="text-gray-500">Race ID: </span>
            <span className="text-gray-300">{char.RaceId || ''}</span>
          </div>
          <div>
            <span className="text-gray-500">Alignment: </span>
            <span className="text-gray-300">{char.Alignment || 0}</span>
          </div>
        </div>
      </Section>

      <Section title="Behavior">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Hostile: </span>
            <EditableBoolean value={!!mob.Hostile} onSave={(v) => saveField('Hostile', v)} label="Hostile" />
          </div>
          <div>
            <span className="text-gray-500">Activity Level: </span>
            <EditableNumber value={mob.ActivityLevel || 0} onSave={(v) => saveField('ActivityLevel', v)} label="Activity Level" min={0} />
          </div>
          <div>
            <span className="text-gray-500">Max Wander: </span>
            <EditableNumber value={mob.MaxWander || 0} onSave={(v) => saveField('MaxWander', v)} label="Max Wander" min={0} />
          </div>
          <div>
            <span className="text-gray-500">No Corpse: </span>
            <EditableBoolean value={!!mob.NoCorpse} onSave={(v) => saveField('NoCorpse', v)} label="No Corpse" />
          </div>
          <div>
            <span className="text-gray-500">Item Drop Chance: </span>
            <EditableNumber value={mob.ItemDropChance || 0} onSave={(v) => saveField('ItemDropChance', v)} min={0} max={100} />
          </div>
          <div>
            <span className="text-gray-500">Script Tag: </span>
            <EditableText value={mob.ScriptTag || ''} onSave={(v) => saveField('ScriptTag', v)} />
          </div>
        </div>
      </Section>

      <Section title="Death Message">
        <EditableText
          value={mob.DeathMessage ? stripAnsi(mob.DeathMessage) : ''}
          onSave={(v) => saveField('DeathMessage', v)}
          label="Death Message"
        />
      </Section>

      <Section title="Groups">
        <EditableList
          value={mob.Groups || []}
          onSave={(v) => saveField('Groups', v)}
          label="Groups"
          placeholder="Add group..."
        />
      </Section>

      <Section title="Hates">
        <EditableList
          value={mob.Hates || []}
          onSave={(v) => saveField('Hates', v)}
          placeholder="Add hate group..."
        />
      </Section>

      <Section title="Angry Commands">
        <EditableList
          value={mob.AngryCommands || []}
          onSave={(v) => saveField('AngryCommands', v)}
          placeholder="Add angry command..."
        />
      </Section>

      <Section title="Combat Commands">
        <EditableList
          value={mob.CombatCommands || []}
          onSave={(v) => saveField('CombatCommands', v)}
          placeholder="Add combat command..."
        />
      </Section>

      <Section title="Quest Flags">
        <EditableList
          value={mob.QuestFlags || []}
          onSave={(v) => saveField('QuestFlags', v)}
          placeholder="Add quest flag..."
        />
      </Section>

      <Section title="Buff IDs">
        <EditableList
          value={(mob.BuffIds || []).map(String)}
          onSave={(v) => saveField('BuffIds', v.map(Number).filter(n => !isNaN(n)))}
          placeholder="Add buff ID..."
        />
      </Section>

      <Section title="Idle Commands">
        <EditableList
          value={(mob.IdleCommands || []).filter((c: string) => c)}
          onSave={(v) => saveField('IdleCommands', v)}
          label="Idle Commands"
          placeholder="Add command..."
        />
      </Section>

      {/* Shop */}
      {char.Shop && char.Shop.length > 0 && (
        <Section title="Shop Stock">
          <div className="space-y-1">
            {char.Shop.map((entry: any, i: number) => (
              <div key={i} className="bg-gray-800 rounded px-3 py-2 flex items-center gap-3">
                {entry.ItemId > 0 && (
                  <Link to={`/items/${entry.ItemId}`} className="text-green-400 hover:text-green-300">
                    Item #{entry.ItemId}
                  </Link>
                )}
                <span className="text-gray-500 text-sm">qty: {entry.QuantityMax}</span>
              </div>
            ))}
          </div>
        </Section>
      )}

      <ReferencesPanel references={references} />

      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(mob, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

function stripAnsi(text: string): string {
  return text.replace(/<\/?ansi[^>]*>/g, '');
}

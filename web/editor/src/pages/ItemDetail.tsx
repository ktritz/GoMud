import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { items, api } from '../api/client';
import { EditableText, EditableNumber, EditableBoolean, EditableSelect, EditableList, EditableMap, Section, SaveStatus } from '../components/EditableField';
import { useItemValidation, ValidationPanel } from '../hooks/useValidation';
import { useItemReferences } from '../hooks/useReferences';
import { ReferencesPanel } from '../components/ReferencesPanel';
import { DetailActions } from '../components/DetailActions';

const typeOptions = [
  'weapon', 'offhand', 'head', 'neck', 'body', 'belt', 'gloves', 'ring',
  'legs', 'feet', 'potion', 'food', 'drink', 'scroll', 'grenade', 'junk',
  'readable', 'key', 'object', 'gemstone', 'lockpicks', 'botanical', 'service',
];

const subtypeOptions = [
  '', 'wearable', 'drinkable', 'edible', 'usable', 'throwable', 'mundane',
  'generic', 'bludgeoning', 'cleaving', 'stabbing',
];

export function ItemDetail() {
  const { itemId } = useParams();
  const id = Number(itemId);
  const queryClient = useQueryClient();

  const { data: item, isLoading, error } = useQuery({
    queryKey: ['items', id],
    queryFn: () => items.get(id),
    enabled: !isNaN(id),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/items/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['items', id] });
      queryClient.invalidateQueries({ queryKey: ['items'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  const warnings = useItemValidation(item);
  const references = useItemReferences(id);

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!item) return <div className="text-gray-400">Item not found</div>;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/items" className="text-gray-400 hover:text-white">&larr; Items</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{item.ItemId}</span>
      </div>

      <DetailActions
        entityType="items"
        entityId={id}
        entityName={item.Name}
        listPath="/items"
        detailPath="/items"
        duplicateData={{ Name: `${item.Name} (Copy)`, Type: item.Type }}
      />
      <ValidationPanel warnings={warnings} />
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
            <EditableText value={item.Name || ''} onSave={(v) => saveField('Name', v)} label="Name" />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Display Name</label>
            <EditableText value={item.DisplayName || ''} onSave={(v) => saveField('DisplayName', v)} label="Display Name" />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Simple Name</label>
            <EditableText value={item.NameSimple || ''} onSave={(v) => saveField('NameSimple', v)} label="Simple Name" />
          </div>
        </div>
      </Section>

      <Section title="Description">
        <EditableText
          value={item.Description || ''}
          onSave={(v) => saveField('Description', v)}
          label="Description"
          multiline
        />
      </Section>

      <Section title="Classification">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-gray-500 text-sm block mb-1">Type</label>
            <EditableSelect
              value={item.Type || ''}
              options={typeOptions}
              onSave={(v) => saveField('Type', v)}
              label="Type"
            />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Subtype</label>
            <EditableSelect
              value={item.Subtype || ''}
              options={subtypeOptions}
              onSave={(v) => saveField('Subtype', v)}
              label="Subtype"
            />
          </div>
        </div>
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <label className="text-gray-500 block mb-1">Value</label>
            <EditableNumber value={item.Value || 0} onSave={(v) => saveField('Value', v)} label="Value" />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Uses</label>
            <EditableNumber value={item.Uses || 0} onSave={(v) => saveField('Uses', v)} label="Uses" />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Hands</label>
            <EditableNumber value={item.Hands || 0} onSave={(v) => saveField('Hands', v)} label="Hands" min={0} max={2} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Damage Reduction</label>
            <EditableNumber value={item.DamageReduction || 0} onSave={(v) => saveField('DamageReduction', v)} label="Damage Reduction" />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Wait Rounds</label>
            <EditableNumber value={item.WaitRounds || 0} onSave={(v) => saveField('WaitRounds', v)} label="Wait Rounds" />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Break Chance (%)</label>
            <EditableNumber value={item.BreakChance || 0} onSave={(v) => saveField('BreakChance', v)} label="Break Chance" min={0} max={100} />
          </div>
          <div>
            <label className="text-gray-500 block mb-1">Cursed</label>
            <EditableBoolean value={!!item.Cursed} onSave={(v) => saveField('Cursed', v)} label="Cursed" />
          </div>
        </div>
      </Section>

      <Section title="Damage">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Dice Roll: </span>
            <EditableText
              value={item.Damage?.DiceRoll || ''}
              onSave={(v) => saveField('Damage', { ...item.Damage, DiceRoll: v })}
              label="Dice Roll"
            />
          </div>
          <div>
            <span className="text-gray-500">Attacks: </span>
            <EditableNumber value={item.Damage?.Attacks || 1} onSave={(v) => saveField('Damage', { ...item.Damage, Attacks: v })} min={1} />
          </div>
          <div>
            <span className="text-gray-500">Bonus Damage: </span>
            <EditableNumber value={item.Damage?.BonusDamage || 0} onSave={(v) => saveField('Damage', { ...item.Damage, BonusDamage: v })} />
          </div>
        </div>
        <div className="mt-3">
          <label className="text-gray-500 text-sm block mb-1">Crit Buff IDs</label>
          <EditableList
            value={(item.Damage?.CritBuffIds || []).map(String)}
            onSave={(v) => saveField('Damage', { ...item.Damage, CritBuffIds: v.map(Number).filter(n => !isNaN(n)) })}
            placeholder="Add crit buff ID..."
          />
        </div>
      </Section>

      <Section title="References">
        <div className="space-y-3">
          <div>
            <label className="text-gray-500 text-sm block mb-1">Key / Lock ID</label>
            <EditableText value={item.KeyLockId || ''} onSave={(v) => saveField('KeyLockId', v)} label="Key Lock ID" />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Quest Token</label>
            <EditableText value={item.QuestToken || ''} onSave={(v) => saveField('QuestToken', v)} label="Quest Token" />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Element</label>
            <EditableText value={item.Element || ''} onSave={(v) => saveField('Element', v)} label="Element" />
          </div>
        </div>
      </Section>

      <Section title="Buffs">
        <div className="space-y-3">
          <div>
            <label className="text-gray-500 text-sm block mb-1">Buff IDs (on use)</label>
            <EditableList value={item.BuffIds || []} onSave={(v) => saveField('BuffIds', v)} label="Buff IDs" placeholder="Add buff ID..." />
          </div>
          <div>
            <label className="text-gray-500 text-sm block mb-1">Worn Buff IDs (while equipped)</label>
            <EditableList value={item.WornBuffIds || []} onSave={(v) => saveField('WornBuffIds', v)} label="Worn Buff IDs" placeholder="Add worn buff ID..." />
          </div>
        </div>
      </Section>

      <Section title="Stat Modifiers">
        <EditableMap
          value={item.StatMods || {}}
          onSave={(v) => saveField('StatMods', v)}
          label="Stat Modifiers"
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
            {JSON.stringify(item, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

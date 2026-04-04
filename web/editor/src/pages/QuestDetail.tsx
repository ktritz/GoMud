import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { quests, api } from '../api/client';
import { EditableText, EditableBoolean, EditableNumber, Section, SaveStatus } from '../components/EditableField';
import { useQuestValidation, ValidationPanel } from '../hooks/useValidation';
import { useQuestReferences } from '../hooks/useReferences';
import { ReferencesPanel } from '../components/ReferencesPanel';
import { DetailActions } from '../components/DetailActions';

interface Step {
  Id: string;
  Description: string;
  Hint: string;
}

interface Rewards {
  QuestId: string;
  Gold: number;
  ItemId: number;
  BuffId: number;
  Experience: number;
  SkillInfo: string;
  PlayerMessage: string;
  RoomMessage: string;
  RoomId: number;
}

export function QuestDetail() {
  const { questId } = useParams();
  const id = String(questId);
  const queryClient = useQueryClient();

  const { data: quest, isLoading, error } = useQuery({
    queryKey: ['quests', id],
    queryFn: () => quests.get(String(id)),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/quests/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quests', id] });
      queryClient.invalidateQueries({ queryKey: ['quests'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  const warnings = useQuestValidation(quest);
  const references = useQuestReferences(Number(questId));

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!quest) return <div className="text-gray-400">Quest not found</div>;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/quests" className="text-gray-400 hover:text-white">&larr; Quests</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{quest.QuestId}</span>
      </div>

      <DetailActions
        entityType="quests"
        entityId={questId!}
        entityName={quest.Name}
        listPath="/quests"
        detailPath="/quests"
        duplicateData={{ Name: `${quest.Name} (Copy)`, Description: quest.Description }}
      />
      <ValidationPanel warnings={warnings} />
      <SaveStatus
        isPending={updateMutation.isPending}
        isError={updateMutation.isError}
        isSuccess={updateMutation.isSuccess}
        error={updateMutation.error as Error}
      />

      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold">{quest.Name || 'Unnamed Quest'}</h1>
        {quest.Secret && (
          <span className="px-2 py-0.5 bg-yellow-900/50 rounded text-sm text-yellow-300">Secret</span>
        )}
      </div>

      <Section title="Name">
        <EditableText value={quest.Name || ''} onSave={(v) => saveField('Name', v)} label="Name" />
      </Section>

      <Section title="Description">
        <EditableText
          value={quest.Description || ''}
          onSave={(v) => saveField('Description', v)}
          label="Description"
          multiline
        />
      </Section>

      <Section title="Secret">
        <div className="flex items-center gap-2">
          <span className="text-gray-400 text-sm">Hidden from quest lists:</span>
          <EditableBoolean value={!!quest.Secret} onSave={(v) => saveField('Secret', v)} label="Secret" />
        </div>
      </Section>

      <Section title="Steps">
        <StepEditor
          steps={quest.Steps || []}
          onSave={(steps) => saveField('Steps', steps)}
        />
      </Section>

      <Section title="Rewards">
        <RewardEditor
          rewards={quest.Rewards || {}}
          onSave={(rewards) => saveField('Rewards', rewards)}
        />
      </Section>

      <ReferencesPanel references={references} />

      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(quest, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

// --- Step Editor ---

function StepEditor({ steps, onSave }: { steps: Step[]; onSave: (steps: Step[]) => void }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<Step[]>([]);

  const startEdit = () => {
    setDraft(steps.map((s) => ({ ...s })));
    setEditing(true);
  };

  const handleSave = () => {
    onSave(draft);
    setEditing(false);
  };

  const handleCancel = () => {
    setDraft([]);
    setEditing(false);
  };

  const updateStep = (index: number, field: keyof Step, value: string) => {
    const updated = draft.map((s, i) => (i === index ? { ...s, [field]: value } : s));
    setDraft(updated);
  };

  const addStep = () => {
    setDraft([...draft, { Id: '', Description: '', Hint: '' }]);
  };

  const removeStep = (index: number) => {
    setDraft(draft.filter((_, i) => i !== index));
  };

  if (!editing) {
    return (
      <div>
        {steps.length > 0 ? (
          <div className="space-y-2 mb-3">
            {steps.map((step, i) => (
              <div key={i} className="bg-gray-800 rounded px-3 py-2">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-gray-500">#{i + 1}</span>
                  <span className="font-mono text-yellow-400 text-sm">{step.Id}</span>
                </div>
                <p className="text-gray-300 text-sm mt-1">{step.Description}</p>
                {step.Hint && (
                  <p className="text-gray-500 text-xs mt-1 italic">Hint: {step.Hint}</p>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500 mb-3">No steps defined</p>
        )}
        <button
          onClick={startEdit}
          className="px-3 py-1 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded"
        >
          Edit Steps
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {draft.map((step, i) => (
        <div key={i} className="bg-gray-800 rounded p-3 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-500">Step #{i + 1}</span>
            <button
              onClick={() => removeStep(i)}
              className="text-red-400 hover:text-red-300 text-sm"
            >
              &#10005; Remove
            </button>
          </div>
          <div>
            <label className="text-gray-500 text-xs">Id</label>
            <input
              type="text"
              value={step.Id}
              onChange={(e) => updateStep(i, 'Id', e.target.value)}
              className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="step-id"
            />
          </div>
          <div>
            <label className="text-gray-500 text-xs">Description</label>
            <input
              type="text"
              value={step.Description}
              onChange={(e) => updateStep(i, 'Description', e.target.value)}
              className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="What the player must do"
            />
          </div>
          <div>
            <label className="text-gray-500 text-xs">Hint</label>
            <input
              type="text"
              value={step.Hint}
              onChange={(e) => updateStep(i, 'Hint', e.target.value)}
              className="w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="Optional hint for the player"
            />
          </div>
        </div>
      ))}
      <button
        onClick={addStep}
        className="px-3 py-1 bg-green-700 hover:bg-green-600 text-white text-sm rounded"
      >
        + Add Step
      </button>
      <div className="flex gap-2">
        <button onClick={handleSave} className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded">Save</button>
        <button onClick={handleCancel} className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded">Cancel</button>
      </div>
    </div>
  );
}

// --- Reward Editor ---

function RewardEditor({ rewards, onSave }: { rewards: Rewards; onSave: (rewards: Rewards) => void }) {
  const current: Rewards = {
    QuestId: rewards.QuestId || '',
    Gold: rewards.Gold || 0,
    ItemId: rewards.ItemId || 0,
    BuffId: rewards.BuffId || 0,
    Experience: rewards.Experience || 0,
    SkillInfo: rewards.SkillInfo || '',
    PlayerMessage: rewards.PlayerMessage || '',
    RoomMessage: rewards.RoomMessage || '',
    RoomId: rewards.RoomId || 0,
  };

  const saveReward = (field: keyof Rewards, value: any) => {
    onSave({ ...current, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <span className="text-gray-500 text-sm">Gold:</span>
          <EditableNumber value={current.Gold} onSave={(v) => saveReward('Gold', v)} label="Gold" min={0} />
        </div>
        <div>
          <span className="text-gray-500 text-sm">Experience:</span>
          <EditableNumber value={current.Experience} onSave={(v) => saveReward('Experience', v)} label="Experience" min={0} />
        </div>
        <div>
          <span className="text-gray-500 text-sm">ItemId:</span>
          <EditableNumber value={current.ItemId} onSave={(v) => saveReward('ItemId', v)} label="ItemId" min={0} />
        </div>
        <div>
          <span className="text-gray-500 text-sm">BuffId:</span>
          <EditableNumber value={current.BuffId} onSave={(v) => saveReward('BuffId', v)} label="BuffId" min={0} />
        </div>
        <div>
          <span className="text-gray-500 text-sm">RoomId:</span>
          <EditableNumber value={current.RoomId} onSave={(v) => saveReward('RoomId', v)} label="RoomId" min={0} />
        </div>
      </div>
      <div className="space-y-2">
        <div>
          <span className="text-gray-500 text-sm">Next Quest (QuestId):</span>
          <EditableText value={current.QuestId} onSave={(v) => saveReward('QuestId', v)} label="QuestId" />
        </div>
        <div>
          <span className="text-gray-500 text-sm">SkillInfo:</span>
          <EditableText value={current.SkillInfo} onSave={(v) => saveReward('SkillInfo', v)} label="SkillInfo" />
        </div>
        <div>
          <span className="text-gray-500 text-sm">Player Message:</span>
          <EditableText value={current.PlayerMessage} onSave={(v) => saveReward('PlayerMessage', v)} label="PlayerMessage" />
        </div>
        <div>
          <span className="text-gray-500 text-sm">Room Message:</span>
          <EditableText value={current.RoomMessage} onSave={(v) => saveReward('RoomMessage', v)} label="RoomMessage" />
        </div>
      </div>
    </div>
  );
}

import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rooms, api } from '../api/client';
import { EditableText, EditableBoolean, EditableList, EditableMap, Section, SaveStatus } from '../components/EditableField';
import { ExitEditor } from '../components/ExitEditor';
import { SpawnEditor } from '../components/SpawnEditor';
import { useRoomValidation, ValidationPanel } from '../hooks/useValidation';
import { useRoomReferences } from '../hooks/useReferences';
import { ReferencesPanel } from '../components/ReferencesPanel';
import { DetailActions } from '../components/DetailActions';

export function RoomDetail() {
  const { roomId } = useParams();
  const id = Number(roomId);
  const queryClient = useQueryClient();

  const { data: room, isLoading, error } = useQuery({
    queryKey: ['rooms', id],
    queryFn: () => rooms.get(id),
    enabled: !isNaN(id),
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Record<string, any>) =>
      api.put(`/admin/rooms/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rooms', id] });
      queryClient.invalidateQueries({ queryKey: ['rooms'] });
    },
  });

  const saveField = (field: string, value: any) => {
    updateMutation.mutate({ [field]: value });
  };

  const warnings = useRoomValidation(room);
  const references = useRoomReferences(id);

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!room) return <div className="text-gray-400">Room not found</div>;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/rooms" className="text-gray-400 hover:text-white">&larr; Rooms</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{room.RoomId}</span>
      </div>

      <DetailActions
        entityType="rooms"
        entityId={id}
        entityName={room.Title}
        listPath="/rooms"
        detailPath="/rooms"
        duplicateData={{ Zone: room.Zone, Title: `${room.Title} (Copy)`, Description: room.Description }}
      />
      <ValidationPanel warnings={warnings} />
      <SaveStatus
        isPending={updateMutation.isPending}
        isError={updateMutation.isError}
        isSuccess={updateMutation.isSuccess}
        error={updateMutation.error as Error}
      />

      <Section title="Title">
        <EditableText value={room.Title || ''} onSave={(v) => saveField('Title', v)} label="Title" />
      </Section>

      <Section title="Zone / Biome">
        <div className="flex gap-4">
          <span className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{room.Zone}</span>
          <EditableText value={room.Biome || ''} onSave={(v) => saveField('Biome', v)} label="Biome" />
        </div>
      </Section>

      <Section title="Description">
        <EditableText
          value={room.Description || ''}
          onSave={(v) => saveField('Description', v)}
          label="Description"
          multiline
        />
      </Section>

      {/* Exits */}
      <Section title="Exits">
        <ExitEditor
          exits={room.Exits || {}}
          onSave={(exits) => saveField('Exits', exits)}
        />
      </Section>

      {/* Spawns */}
      <Section title="Spawns">
        <SpawnEditor
          spawns={room.SpawnInfo || []}
          onSave={(spawns) => saveField('SpawnInfo', spawns)}
        />
      </Section>

      {/* Nouns */}
      <Section title="Nouns">
        <EditableMap
          value={room.Nouns || {}}
          onSave={(v) => saveField('Nouns', v)}
        />
      </Section>

      {/* Flags */}
      <Section title="Flags">
        <div className="flex gap-4">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-500">Bank:</span>
            <EditableBoolean value={room.IsBank || false} onSave={(v) => saveField('IsBank', v)} />
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-500">Storage:</span>
            <EditableBoolean value={room.IsStorage || false} onSave={(v) => saveField('IsStorage', v)} />
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-500">PvP:</span>
            <EditableBoolean value={room.Pvp || false} onSave={(v) => saveField('Pvp', v)} />
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-500">Character Room:</span>
            <EditableBoolean value={room.IsCharacterRoom || false} onSave={(v) => saveField('IsCharacterRoom', v)} />
          </div>
        </div>
      </Section>

      {/* Map info */}
      <Section title="Map">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Symbol: </span>
            <EditableText value={room.MapSymbol || ''} onSave={(v) => saveField('MapSymbol', v)} label="Symbol" />
          </div>
          <div>
            <span className="text-gray-500">Legend: </span>
            <EditableText value={room.MapLegend || ''} onSave={(v) => saveField('MapLegend', v)} label="Legend" />
          </div>
        </div>
      </Section>

      {/* Idle Messages */}
      <Section title="Idle Messages">
        <EditableList
          value={room.IdleMessages || []}
          onSave={(v) => saveField('IdleMessages', v)}
          placeholder="Add idle message..."
        />
      </Section>

      {/* Signs */}
      <Section title="Signs">
        <EditableList
          value={(room.Signs || []).map((s: any) => s.DisplayText)}
          onSave={(texts) => saveField('Signs', texts.map((t: string) => ({ DisplayText: t })))}
          placeholder="Add sign text..."
        />
      </Section>

      {/* Skill Training */}
      <Section title="Skill Training">
        {room.SkillTraining && Object.keys(room.SkillTraining).length > 0 ? (
          <div className="space-y-1 text-sm">
            {Object.entries(room.SkillTraining).map(([skill, range]: [string, any]) => (
              <div key={skill} className="flex items-center gap-2">
                <span className="text-gray-400 font-mono w-24">{skill}</span>
                <span className="text-gray-300">Level {range.Min}-{range.Max}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No skill training</p>
        )}
      </Section>

      {/* Music */}
      <Section title="Music">
        <EditableText value={room.MusicFile || ''} onSave={(v) => saveField('MusicFile', v)} />
      </Section>

      <ReferencesPanel references={references} />

      {/* Raw JSON */}
      <Section title="Raw Data">
        <details>
          <summary className="text-gray-500 cursor-pointer hover:text-gray-300 text-sm">
            Show JSON
          </summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-800 rounded p-3 overflow-auto max-h-96">
            {JSON.stringify(room, null, 2)}
          </pre>
        </details>
      </Section>
    </div>
  );
}

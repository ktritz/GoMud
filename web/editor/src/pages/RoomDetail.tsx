import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rooms, api } from '../api/client';
import { EditableText } from '../components/EditableField';

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

      {updateMutation.isPending && (
        <div className="bg-blue-900/30 border border-blue-700 text-blue-300 px-3 py-1 rounded mb-4 text-sm">
          Saving...
        </div>
      )}
      {updateMutation.isError && (
        <div className="bg-red-900/30 border border-red-700 text-red-300 px-3 py-1 rounded mb-4 text-sm">
          Error: {(updateMutation.error as Error).message}
        </div>
      )}
      {updateMutation.isSuccess && (
        <div className="bg-green-900/30 border border-green-700 text-green-300 px-3 py-1 rounded mb-4 text-sm">
          Saved!
        </div>
      )}

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

      {/* Exits (read-only for now) */}
      <Section title="Exits">
        {room.Exits && Object.keys(room.Exits).length > 0 ? (
          <div className="grid grid-cols-2 gap-2">
            {Object.entries(room.Exits).map(([dir, exit]: [string, any]) => (
              <Link
                key={dir}
                to={`/rooms/${exit.RoomId}`}
                className="flex items-center justify-between bg-gray-800 rounded px-3 py-2 hover:bg-gray-700 transition-colors"
              >
                <span className="font-mono text-blue-400">{dir}</span>
                <span className="text-gray-400">→ Room #{exit.RoomId}</span>
                {exit.Secret && <span className="text-yellow-500 text-xs ml-2">secret</span>}
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No exits</p>
        )}
      </Section>

      {/* Spawn Info (read-only for now) */}
      <Section title="Spawns">
        {room.SpawnInfo && room.SpawnInfo.length > 0 ? (
          <div className="space-y-2">
            {room.SpawnInfo.map((spawn: any, i: number) => (
              <div key={i} className="bg-gray-800 rounded px-3 py-2">
                <div className="flex items-center gap-2">
                  {spawn.MobId > 0 && (
                    <Link to={`/mobs/${spawn.MobId}`} className="text-blue-400 hover:text-blue-300">
                      Mob #{spawn.MobId}
                    </Link>
                  )}
                  <span className="text-gray-500 text-sm">{spawn.RespawnRate}</span>
                </div>
                {spawn.Message && (
                  <p className="text-gray-400 text-sm mt-1">{stripAnsi(spawn.Message)}</p>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No spawns</p>
        )}
      </Section>

      {/* Nouns */}
      <Section title="Nouns">
        {room.Nouns && Object.keys(room.Nouns).length > 0 ? (
          <div className="space-y-2">
            {Object.entries(room.Nouns).map(([noun, desc]: [string, any]) => (
              <div key={noun} className="bg-gray-800 rounded px-3 py-2">
                <span className="font-mono text-yellow-400">{noun}</span>
                <p className="text-gray-400 text-sm mt-1">{stripAnsi(String(desc)).slice(0, 120)}...</p>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No nouns</p>
        )}
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

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-6">
      <h2 className="text-lg font-semibold text-gray-200 mb-2 border-b border-gray-700 pb-1">
        {title}
      </h2>
      {children}
    </section>
  );
}

function stripAnsi(text: string): string {
  return text.replace(/<\/?ansi[^>]*>/g, '');
}

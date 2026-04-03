import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { rooms } from '../api/client';

export function RoomDetail() {
  const { roomId } = useParams();
  const id = Number(roomId);

  const { data: room, isLoading, error } = useQuery({
    queryKey: ['rooms', id],
    queryFn: () => rooms.get(id),
    enabled: !isNaN(id),
  });

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

      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold">{room.Title}</h1>
        <span className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{room.Zone}</span>
        {room.Biome && (
          <span className="px-2 py-0.5 bg-blue-900/50 rounded text-sm text-blue-300">{room.Biome}</span>
        )}
      </div>

      {/* Description */}
      <Section title="Description">
        <p className="text-gray-300 whitespace-pre-wrap">{room.Description || 'No description'}</p>
      </Section>

      {/* Exits */}
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
                {exit.Lock?.Difficulty > 0 && (
                  <span className="text-red-400 text-xs ml-2">locked ({exit.Lock.Difficulty})</span>
                )}
              </Link>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No exits</p>
        )}
      </Section>

      {/* Spawn Info */}
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
                  {spawn.ItemId > 0 && (
                    <Link to={`/items/${spawn.ItemId}`} className="text-green-400 hover:text-green-300">
                      Item #{spawn.ItemId}
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

      {/* Items on floor */}
      <Section title="Floor Items">
        {room.Items && room.Items.length > 0 ? (
          <div className="space-y-1">
            {room.Items.map((item: any, i: number) => (
              <div key={i} className="text-gray-300">
                <Link to={`/items/${item.ItemId}`} className="text-green-400 hover:text-green-300">
                  Item #{item.ItemId}
                </Link>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No items</p>
        )}
        {room.Gold > 0 && <p className="text-yellow-400 mt-1">{room.Gold} gold on the floor</p>}
      </Section>

      {/* Idle Messages */}
      <Section title="Idle Messages">
        {room.IdleMessages && room.IdleMessages.length > 0 ? (
          <ul className="space-y-1">
            {room.IdleMessages.map((msg: string, i: number) => (
              <li key={i} className="text-gray-400 text-sm">{stripAnsi(msg)}</li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">None</p>
        )}
      </Section>

      {/* Map info */}
      <Section title="Map">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Symbol:</span>{' '}
            <span className="font-mono text-white">{room.MapSymbol || '?'}</span>
          </div>
          <div>
            <span className="text-gray-500">Legend:</span>{' '}
            <span className="text-gray-300">{room.MapLegend || 'none'}</span>
          </div>
          <div>
            <span className="text-gray-500">Biome:</span>{' '}
            <span className="text-gray-300">{room.Biome || 'default'}</span>
          </div>
        </div>
      </Section>

      {/* Raw JSON (dev helper) */}
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

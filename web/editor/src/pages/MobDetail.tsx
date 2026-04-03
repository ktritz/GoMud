import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mobs } from '../api/client';

export function MobDetail() {
  const { mobId } = useParams();
  const id = Number(mobId);

  const { data: mob, isLoading, error } = useQuery({
    queryKey: ['mobs', id],
    queryFn: () => mobs.get(id),
    enabled: !isNaN(id),
  });

  if (isLoading) return <div className="text-gray-400">Loading...</div>;
  if (error) return <div className="text-red-400">Error: {(error as Error).message}</div>;
  if (!mob) return <div className="text-gray-400">Mob not found</div>;

  const char = mob.Character || mob.character || {};

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-2 mb-6">
        <Link to="/mobs" className="text-gray-400 hover:text-white">&larr; Mobs</Link>
        <span className="text-gray-600">/</span>
        <span className="text-gray-300">#{mob.MobId}</span>
      </div>

      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold">{char.Name || char.name || 'Unknown'}</h1>
        <span className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{mob.Zone}</span>
        {mob.Hostile && (
          <span className="px-2 py-0.5 bg-red-900/50 rounded text-sm text-red-300">Hostile</span>
        )}
        <span className="px-2 py-0.5 bg-blue-900/50 rounded text-sm text-blue-300">
          Level {char.Level || char.level}
        </span>
      </div>

      <Section title="Description">
        <p className="text-gray-300 whitespace-pre-wrap">
          {char.Description || char.description || 'No description'}
        </p>
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <Prop label="Activity Level" value={mob.ActivityLevel} />
          <Prop label="Max Wander" value={mob.MaxWander} />
          <Prop label="Item Drop Chance" value={`${mob.ItemDropChance}%`} />
          <Prop label="Gold" value={char.Gold || char.gold || 0} />
          <Prop label="Race ID" value={char.RaceId || char.raceid} />
          <Prop label="Alignment" value={char.Alignment || char.alignment || 0} />
          <Prop label="No Corpse" value={mob.NoCorpse ? 'Yes' : 'No'} />
          <Prop label="Script Tag" value={mob.ScriptTag || 'none'} />
        </div>
      </Section>

      {mob.DeathMessage && (
        <Section title="Death Message">
          <p className="text-gray-300">{stripAnsi(mob.DeathMessage)}</p>
        </Section>
      )}

      <Section title="Groups">
        {mob.Groups && mob.Groups.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {mob.Groups.map((g: string) => (
              <span key={g} className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{g}</span>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">None</p>
        )}
      </Section>

      <Section title="Idle Commands">
        {mob.IdleCommands && mob.IdleCommands.length > 0 ? (
          <ul className="space-y-1">
            {mob.IdleCommands.filter((c: string) => c).map((cmd: string, i: number) => (
              <li key={i} className="text-gray-400 text-sm font-mono">{cmd}</li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">None</p>
        )}
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

function Prop({ label, value }: { label: string; value: any }) {
  return (
    <div>
      <span className="text-gray-500">{label}:</span>{' '}
      <span className="text-gray-300">{String(value ?? '')}</span>
    </div>
  );
}

function stripAnsi(text: string): string {
  return text.replace(/<\/?ansi[^>]*>/g, '');
}

import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { items } from '../api/client';

export function ItemDetail() {
  const { itemId } = useParams();
  const id = Number(itemId);

  const { data: item, isLoading, error } = useQuery({
    queryKey: ['items', id],
    queryFn: () => items.get(id),
    enabled: !isNaN(id),
  });

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

      <div className="flex items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold">{item.Name}</h1>
        <span className="px-2 py-0.5 bg-gray-700 rounded text-sm text-gray-300">{item.Type}</span>
        {item.Subtype && (
          <span className="px-2 py-0.5 bg-purple-900/50 rounded text-sm text-purple-300">{item.Subtype}</span>
        )}
      </div>

      <Section title="Description">
        <p className="text-gray-300 whitespace-pre-wrap">{item.Description || 'No description'}</p>
      </Section>

      <Section title="Properties">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <Prop label="Value" value={`${item.Value} gold`} />
          <Prop label="Hands" value={item.Hands || 1} />
          <Prop label="Uses" value={item.Uses || 'unlimited'} />
          <Prop label="Wait Rounds" value={item.WaitRounds || 0} />
          {item.DamageReduction > 0 && <Prop label="Defense" value={item.DamageReduction} />}
          {item.BreakChance > 0 && <Prop label="Break Chance" value={`${item.BreakChance}%`} />}
          {item.Cursed && <Prop label="Cursed" value="Yes" />}
        </div>
      </Section>

      {item.Damage && (item.Damage.DiceRoll || item.Damage.DiceCount > 0) && (
        <Section title="Damage">
          <div className="grid grid-cols-3 gap-4 text-sm">
            <Prop label="Dice" value={item.Damage.DiceRoll || `${item.Damage.DiceCount}d${item.Damage.SideCount}`} />
            <Prop label="Attacks" value={item.Damage.Attacks || 1} />
            {item.Damage.BonusDamage > 0 && <Prop label="Bonus" value={`+${item.Damage.BonusDamage}`} />}
          </div>
        </Section>
      )}

      {item.StatMods && Object.keys(item.StatMods).length > 0 && (
        <Section title="Stat Modifiers">
          <div className="grid grid-cols-3 gap-2 text-sm">
            {Object.entries(item.StatMods).map(([stat, val]: [string, any]) => (
              <div key={stat}>
                <span className="text-gray-500">{stat}:</span>{' '}
                <span className={val > 0 ? 'text-green-400' : 'text-red-400'}>
                  {val > 0 ? `+${val}` : val}
                </span>
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
            {JSON.stringify(item, null, 2)}
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

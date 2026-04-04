import { Link } from 'react-router-dom';
import type { Reference } from '../hooks/useReferences';

const linkPrefixes: Record<string, string> = {
  room: '/rooms',
  mob: '/mobs',
  item: '/items',
  quest: '/quests',
  buff: '/buffs',
};

const typeColors: Record<string, string> = {
  room: 'text-purple-400',
  mob: 'text-blue-400',
  item: 'text-green-400',
  quest: 'text-yellow-400',
  buff: 'text-orange-400',
};

interface ReferencesPanelProps {
  references: Reference[];
}

export function ReferencesPanel({ references }: ReferencesPanelProps) {
  if (references.length === 0) return null;

  return (
    <section className="mb-6">
      <h2 className="text-lg font-semibold text-gray-200 mb-2 border-b border-gray-700 pb-1">
        Referenced By
      </h2>
      <div className="space-y-1">
        {references.map((ref, i) => (
          <Link
            key={`${ref.type}-${ref.id}-${i}`}
            to={`${linkPrefixes[ref.type] || ''}/${ref.id}`}
            className="flex items-center gap-3 bg-gray-800 rounded px-3 py-2 hover:bg-gray-700 transition-colors"
          >
            <span className={`text-xs font-mono uppercase ${typeColors[ref.type] || 'text-gray-400'}`}>
              {ref.type}
            </span>
            <span className="text-gray-300">#{String(ref.id)} {ref.name}</span>
            <span className="text-gray-500 text-sm ml-auto">{ref.context}</span>
          </Link>
        ))}
      </div>
    </section>
  );
}

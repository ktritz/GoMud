import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { zones } from '../api/client';

export function Zones() {
  const { data, isLoading } = useQuery({
    queryKey: ['zones'],
    queryFn: () => zones.list(),
  });

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Zones</h1>

      {isLoading ? (
        <div className="text-gray-400">Loading...</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {data?.zones?.map((zone: any) => (
            <Link
              key={zone.name}
              to={`/rooms?zone=${encodeURIComponent(zone.name)}`}
              className="bg-gray-800 rounded-lg p-4 border border-gray-700 hover:border-blue-500 transition-colors"
            >
              <div className="font-semibold text-white">{zone.name}</div>
              <div className="text-sm text-gray-400 mt-1">{zone.roomCount} rooms</div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

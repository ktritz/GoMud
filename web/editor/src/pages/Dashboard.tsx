import { useQuery } from '@tanstack/react-query';
import { server, rooms, mobs, items } from '../api/client';

export function Dashboard() {
  const stats = useQuery({ queryKey: ['stats'], queryFn: server.stats });
  const roomList = useQuery({ queryKey: ['rooms'], queryFn: () => rooms.list() });
  const mobList = useQuery({ queryKey: ['mobs'], queryFn: () => mobs.list() });
  const itemList = useQuery({ queryKey: ['items'], queryFn: () => items.list() });

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <StatCard label="Online Users" value={stats.data?.onlineUsers ?? '...'} />
        <StatCard label="Rooms" value={roomList.data?.total ?? '...'} />
        <StatCard label="Mobs" value={mobList.data?.total ?? '...'} />
        <StatCard label="Items" value={itemList.data?.total ?? '...'} />
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="text-sm text-gray-400">{label}</div>
      <div className="text-2xl font-bold text-white mt-1">{value}</div>
    </div>
  );
}

import { useQuery } from '@tanstack/react-query';
import { rooms } from '../api/client';
import { EntityList } from '../components/EntityList';

export function Rooms() {
  const { data, isLoading } = useQuery({
    queryKey: ['rooms'],
    queryFn: () => rooms.list(),
  });

  return (
    <EntityList
      title="Rooms"
      data={data?.rooms}
      isLoading={isLoading}
      linkPrefix="/rooms"
      idField="roomId"
      columns={[
        { key: 'roomId', label: 'ID' },
        { key: 'title', label: 'Title' },
        { key: 'zone', label: 'Zone' },
        { key: 'biome', label: 'Biome' },
        { key: 'exitCount', label: 'Exits' },
      ]}
    />
  );
}

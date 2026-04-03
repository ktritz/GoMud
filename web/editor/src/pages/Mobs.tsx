import { useQuery } from '@tanstack/react-query';
import { mobs } from '../api/client';
import { EntityList } from '../components/EntityList';

export function Mobs() {
  const { data, isLoading } = useQuery({
    queryKey: ['mobs'],
    queryFn: () => mobs.list(),
  });

  return (
    <EntityList
      title="Mobs"
      data={data?.mobs}
      isLoading={isLoading}
      linkPrefix="/mobs"
      idField="mobId"
      columns={[
        { key: 'mobId', label: 'ID' },
        { key: 'name', label: 'Name' },
        { key: 'zone', label: 'Zone' },
        { key: 'level', label: 'Level' },
        {
          key: 'hostile',
          label: 'Hostile',
          render: (item) => (
            <span className={item.hostile ? 'text-red-400' : 'text-green-400'}>
              {item.hostile ? 'Yes' : 'No'}
            </span>
          ),
        },
      ]}
    />
  );
}

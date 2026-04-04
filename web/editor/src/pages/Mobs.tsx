import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { mobs, zones } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

export function Mobs() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['mobs'],
    queryFn: () => mobs.list(),
  });

  const { data: zoneData } = useQuery({
    queryKey: ['zones'],
    queryFn: () => zones.list(),
  });

  const zoneNames = (zoneData?.zones || []).map((z: any) => z.name);

  return (
    <>
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
        onCreate={() => setShowCreate(true)}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Mob"
          fields={[
            { key: 'Name', label: 'Name', type: 'text', required: true, placeholder: 'Mob name' },
            { key: 'Zone', label: 'Zone', type: 'select', required: true, options: zoneNames.length > 0 ? zoneNames : ['default'] },
            { key: 'Level', label: 'Level', type: 'number' },
          ]}
          onSubmit={(data) => mobs.create(data as any)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['mobs'] });
            setShowCreate(false);
            if (result?.MobId) navigate(`/mobs/${result.MobId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

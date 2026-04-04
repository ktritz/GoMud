import { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { rooms, zones } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

export function Rooms() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();
  const zoneFilter = searchParams.get('zone') || '';

  const { data, isLoading } = useQuery({
    queryKey: ['rooms'],
    queryFn: () => rooms.list(),
  });

  const { data: zoneData } = useQuery({
    queryKey: ['zones'],
    queryFn: () => zones.list(),
  });

  const zoneNames = (zoneData?.zones || []).map((z: any) => z.name);

  return (
    <>
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
        onCreate={() => setShowCreate(true)}
        initialSearch={zoneFilter}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Room"
          fields={[
            { key: 'Zone', label: 'Zone', type: 'select', required: true, options: zoneNames.length > 0 ? zoneNames : ['default'] },
            { key: 'Title', label: 'Title', type: 'text', placeholder: 'Room title' },
            { key: 'Description', label: 'Description', type: 'text', placeholder: 'Room description' },
          ]}
          onSubmit={(data) => rooms.create(data as any)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['rooms'] });
            setShowCreate(false);
            if (result?.RoomId) navigate(`/rooms/${result.RoomId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

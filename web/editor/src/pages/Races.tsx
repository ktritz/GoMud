import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { races, api } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

export function Races() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['races'],
    queryFn: () => races.list(),
  });

  return (
    <>
      <EntityList
        title="Races"
        data={data?.races}
        isLoading={isLoading}
        linkPrefix="/races"
        idField="RaceId"
        columns={[
          { key: 'RaceId', label: 'ID' },
          { key: 'Name', label: 'Name' },
          { key: 'Description', label: 'Description' },
          { key: 'Size', label: 'Size' },
        ]}
        onCreate={() => setShowCreate(true)}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Race"
          fields={[
            { key: 'Name', label: 'Name', type: 'text', required: true, placeholder: 'Race name' },
            { key: 'Description', label: 'Description', type: 'text', placeholder: 'Description' },
            { key: 'Size', label: 'Size', type: 'select', options: ['small', 'medium', 'large'] },
          ]}
          onSubmit={(data) => api.post('/admin/races', data)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['races'] });
            setShowCreate(false);
            if (result?.RaceId) navigate(`/races/${result.RaceId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

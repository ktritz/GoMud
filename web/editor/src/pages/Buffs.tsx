import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { buffs, api } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

export function Buffs() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['buffs'],
    queryFn: () => buffs.list(),
  });

  return (
    <>
      <EntityList
        title="Buffs"
        data={data?.buffs}
        isLoading={isLoading}
        linkPrefix="/buffs"
        idField="buffId"
        columns={[
          { key: 'buffId', label: 'ID' },
          { key: 'name', label: 'Name' },
          { key: 'description', label: 'Description' },
        ]}
        onCreate={() => setShowCreate(true)}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Buff"
          fields={[
            { key: 'Name', label: 'Name', type: 'text', required: true, placeholder: 'Buff name' },
            { key: 'Description', label: 'Description', type: 'text', placeholder: 'Description' },
          ]}
          onSubmit={(data) => api.post('/admin/buffs', data)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['buffs'] });
            setShowCreate(false);
            if (result?.BuffId) navigate(`/buffs/${result.BuffId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

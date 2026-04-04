import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { spells, api } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

const SPELL_TYPES = ['neutral', 'harmsingle', 'harmmulti', 'helpsingle', 'helpmulti', 'harmarea', 'helparea'];

export function Spells() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['spells'],
    queryFn: () => spells.list(),
  });

  return (
    <>
      <EntityList
        title="Spells"
        data={data?.spells}
        isLoading={isLoading}
        linkPrefix="/spells"
        idField="spellId"
        columns={[
          { key: 'spellId', label: 'ID' },
          { key: 'name', label: 'Name' },
          { key: 'type', label: 'Type' },
          { key: 'cost', label: 'Cost' },
        ]}
        onCreate={() => setShowCreate(true)}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Spell"
          fields={[
            { key: 'SpellId', label: 'Spell ID (lowercase)', type: 'text', required: true, placeholder: 'e.g. fireball' },
            { key: 'Name', label: 'Display Name', type: 'text', required: true, placeholder: 'Fireball' },
            { key: 'Type', label: 'Type', type: 'select', options: SPELL_TYPES },
          ]}
          onSubmit={(data) => api.post('/admin/spells', data)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['spells'] });
            setShowCreate(false);
            if (result?.SpellId) navigate(`/spells/${result.SpellId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

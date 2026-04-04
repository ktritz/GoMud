import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { items } from '../api/client';
import { EntityList } from '../components/EntityList';
import { CreateEntityModal } from '../components/CreateEntityModal';

const ITEM_TYPES = [
  'weapon', 'offhand', 'head', 'neck', 'body', 'belt', 'gloves', 'ring', 'legs', 'feet',
  'potion', 'food', 'drink', 'scroll', 'grenade', 'junk',
  'readable', 'key', 'object', 'gemstone', 'lockpicks', 'botanical', 'service',
];

export function Items() {
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['items'],
    queryFn: () => items.list(),
  });

  return (
    <>
      <EntityList
        title="Items"
        data={data?.items}
        isLoading={isLoading}
        linkPrefix="/items"
        idField="itemId"
        columns={[
          { key: 'itemId', label: 'ID' },
          { key: 'name', label: 'Name' },
          { key: 'type', label: 'Type' },
          { key: 'subtype', label: 'Subtype' },
          { key: 'value', label: 'Value' },
        ]}
        onCreate={() => setShowCreate(true)}
      />

      {showCreate && (
        <CreateEntityModal
          title="Create Item"
          fields={[
            { key: 'Name', label: 'Name', type: 'text', required: true, placeholder: 'Item name' },
            { key: 'Type', label: 'Type', type: 'select', required: true, options: ITEM_TYPES },
          ]}
          onSubmit={(data) => items.create(data as any)}
          onCreated={(result) => {
            queryClient.invalidateQueries({ queryKey: ['items'] });
            setShowCreate(false);
            if (result?.ItemId) navigate(`/items/${result.ItemId}`);
          }}
          onClose={() => setShowCreate(false)}
        />
      )}
    </>
  );
}

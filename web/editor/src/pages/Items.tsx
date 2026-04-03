import { useQuery } from '@tanstack/react-query';
import { items } from '../api/client';
import { EntityList } from '../components/EntityList';

export function Items() {
  const { data, isLoading } = useQuery({
    queryKey: ['items'],
    queryFn: () => items.list(),
  });

  return (
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
    />
  );
}

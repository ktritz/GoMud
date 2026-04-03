import { useQuery } from '@tanstack/react-query';
import { quests } from '../api/client';
import { EntityList } from '../components/EntityList';

export function Quests() {
  const { data, isLoading } = useQuery({
    queryKey: ['quests'],
    queryFn: () => quests.list(),
  });

  return (
    <EntityList
      title="Quests"
      data={data?.quests}
      isLoading={isLoading}
      linkPrefix="/quests"
      idField="QuestId"
      columns={[
        { key: 'QuestId', label: 'ID' },
        { key: 'Name', label: 'Name' },
        { key: 'Description', label: 'Description' },
      ]}
    />
  );
}

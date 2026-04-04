import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export interface Reference {
  type: 'room' | 'mob' | 'item' | 'quest' | 'buff';
  id: number | string;
  name: string;
  context: string;
}

function useReferences(entityType: string, entityId: number | string | undefined) {
  const { data } = useQuery({
    queryKey: ['references', entityType, entityId],
    queryFn: () => api.get<{ references: Reference[]; total: number }>(`/admin/references/${entityType}/${entityId}`),
    enabled: entityId !== undefined && entityId !== 0,
    staleTime: 60_000,
  });

  return data?.references || [];
}

export function useRoomReferences(roomId: number) { return useReferences('room', roomId); }
export function useMobReferences(mobId: number) { return useReferences('mob', mobId); }
export function useItemReferences(itemId: number) { return useReferences('item', itemId); }
export function useQuestReferences(questId: number) { return useReferences('quest', questId); }
export function useBuffReferences(buffId: number) { return useReferences('buff', buffId); }

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

interface DetailActionsProps {
  entityType: string;       // 'rooms' | 'mobs' | 'items' | 'quests'
  entityId: number | string;
  entityName: string;
  listPath: string;         // '/rooms' | '/mobs' etc
  detailPath: string;       // '/rooms' | '/mobs' etc (for duplicate navigation)
  duplicateData?: Record<string, any>; // data to POST for duplicate
}

export function DetailActions({ entityType, entityId, entityName, listPath, detailPath, duplicateData }: DetailActionsProps) {
  const [showConfirm, setShowConfirm] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const deleteMut = useMutation({
    mutationFn: () => api.del<any>(`/admin/${entityType}/${entityId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [entityType] });
      navigate(listPath);
    },
  });

  const dupMut = useMutation({
    mutationFn: () => api.post<any>(`/admin/${entityType}`, duplicateData),
    onSuccess: (result: any) => {
      queryClient.invalidateQueries({ queryKey: [entityType] });
      // Navigate to the new entity — try common ID field patterns
      const newId = result?.RoomId || result?.MobId || result?.ItemId || result?.QuestId || result?.BuffId || result?.RaceId || result?.SpellId;
      if (newId) navigate(`${detailPath}/${newId}`);
    },
  });

  return (
    <div className="flex gap-2 mb-4">
      {duplicateData && (
        <button
          onClick={() => dupMut.mutate()}
          disabled={dupMut.isPending}
          className="px-3 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 text-sm rounded"
        >
          {dupMut.isPending ? 'Duplicating...' : 'Duplicate'}
        </button>
      )}

      {!showConfirm ? (
        <button
          onClick={() => setShowConfirm(true)}
          className="px-3 py-1 bg-gray-700 hover:bg-red-700 text-gray-400 hover:text-white text-sm rounded transition-colors"
        >
          Delete
        </button>
      ) : (
        <div className="flex items-center gap-2">
          <span className="text-red-400 text-sm">Delete "{entityName}"?</span>
          <button
            onClick={() => deleteMut.mutate()}
            disabled={deleteMut.isPending}
            className="px-3 py-1 bg-red-700 hover:bg-red-600 text-white text-sm rounded"
          >
            {deleteMut.isPending ? 'Deleting...' : 'Yes, Delete'}
          </button>
          <button
            onClick={() => setShowConfirm(false)}
            className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded"
          >
            Cancel
          </button>
        </div>
      )}

      {deleteMut.isError && (
        <span className="text-red-400 text-sm">{(deleteMut.error as Error).message}</span>
      )}
      {dupMut.isError && (
        <span className="text-red-400 text-sm">{(dupMut.error as Error).message}</span>
      )}
    </div>
  );
}

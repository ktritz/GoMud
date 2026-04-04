import { useQuery } from '@tanstack/react-query';
import { rooms, mobs, items, buffs, quests } from '../api/client';

export interface ValidationWarning {
  field: string;
  message: string;
  severity: 'error' | 'warning';
}

// Cache all entity IDs for cross-reference validation
function useEntityIds() {
  const { data: roomData } = useQuery({ queryKey: ['rooms'], queryFn: () => rooms.list(), staleTime: 60_000 });
  const { data: mobData } = useQuery({ queryKey: ['mobs'], queryFn: () => mobs.list(), staleTime: 60_000 });
  const { data: itemData } = useQuery({ queryKey: ['items'], queryFn: () => items.list(), staleTime: 60_000 });
  const { data: buffData } = useQuery({ queryKey: ['buffs'], queryFn: () => buffs.list(), staleTime: 60_000 });
  const { data: questData } = useQuery({ queryKey: ['quests'], queryFn: () => quests.list(), staleTime: 60_000 });

  return {
    roomIds: new Set((roomData?.rooms || []).map((r: any) => r.roomId)),
    mobIds: new Set((mobData?.mobs || []).map((m: any) => m.mobId)),
    itemIds: new Set((itemData?.items || []).map((i: any) => i.itemId)),
    buffIds: new Set((buffData?.buffs || []).map((b: any) => b.buffId)),
    questIds: new Set((questData?.quests || []).map((q: any) => q.QuestId)),
  };
}

export function useRoomValidation(room: any): ValidationWarning[] {
  const ids = useEntityIds();
  if (!room) return [];

  const warnings: ValidationWarning[] = [];

  // Validate exits point to existing rooms
  if (room.Exits) {
    for (const [dir, exit] of Object.entries(room.Exits) as [string, any][]) {
      if (exit.RoomId && ids.roomIds.size > 0 && !ids.roomIds.has(exit.RoomId)) {
        warnings.push({ field: `Exits.${dir}`, message: `Exit "${dir}" points to non-existent room #${exit.RoomId}`, severity: 'error' });
      }
    }
  }

  // Validate spawns reference existing mobs/items
  if (room.SpawnInfo) {
    for (let i = 0; i < room.SpawnInfo.length; i++) {
      const spawn = room.SpawnInfo[i];
      if (spawn.MobId && ids.mobIds.size > 0 && !ids.mobIds.has(spawn.MobId)) {
        warnings.push({ field: `SpawnInfo[${i}].MobId`, message: `Spawn #${i + 1} references non-existent mob #${spawn.MobId}`, severity: 'error' });
      }
      if (spawn.ItemId && ids.itemIds.size > 0 && !ids.itemIds.has(spawn.ItemId)) {
        warnings.push({ field: `SpawnInfo[${i}].ItemId`, message: `Spawn #${i + 1} references non-existent item #${spawn.ItemId}`, severity: 'error' });
      }
    }
  }

  return warnings;
}

export function useMobValidation(mob: any): ValidationWarning[] {
  const ids = useEntityIds();
  if (!mob) return [];

  const warnings: ValidationWarning[] = [];
  const char = mob.Character || {};

  // Validate shop items exist
  if (char.Shop) {
    for (let i = 0; i < char.Shop.length; i++) {
      const entry = char.Shop[i];
      if (entry.ItemId && ids.itemIds.size > 0 && !ids.itemIds.has(entry.ItemId)) {
        warnings.push({ field: `Shop[${i}]`, message: `Shop entry #${i + 1} references non-existent item #${entry.ItemId}`, severity: 'error' });
      }
    }
  }

  // Validate buff IDs
  if (mob.BuffIds) {
    for (const buffId of mob.BuffIds) {
      if (ids.buffIds.size > 0 && !ids.buffIds.has(buffId)) {
        warnings.push({ field: 'BuffIds', message: `References non-existent buff #${buffId}`, severity: 'error' });
      }
    }
  }

  return warnings;
}

export function useItemValidation(item: any): ValidationWarning[] {
  const ids = useEntityIds();
  if (!item) return [];

  const warnings: ValidationWarning[] = [];

  // Validate buff IDs
  if (item.BuffIds) {
    for (const buffId of item.BuffIds) {
      if (ids.buffIds.size > 0 && !ids.buffIds.has(buffId)) {
        warnings.push({ field: 'BuffIds', message: `Use buff references non-existent buff #${buffId}`, severity: 'error' });
      }
    }
  }
  if (item.WornBuffIds) {
    for (const buffId of item.WornBuffIds) {
      if (ids.buffIds.size > 0 && !ids.buffIds.has(buffId)) {
        warnings.push({ field: 'WornBuffIds', message: `Worn buff references non-existent buff #${buffId}`, severity: 'error' });
      }
    }
  }

  // Validate quest token references a real quest
  if (item.QuestToken) {
    const questId = parseInt(item.QuestToken.split('-')[0], 10);
    if (!isNaN(questId) && ids.questIds.size > 0 && !ids.questIds.has(questId)) {
      warnings.push({ field: 'QuestToken', message: `Quest token references non-existent quest #${questId}`, severity: 'error' });
    }
  }

  return warnings;
}

export function useQuestValidation(quest: any): ValidationWarning[] {
  const ids = useEntityIds();
  if (!quest) return [];

  const warnings: ValidationWarning[] = [];
  const rewards = quest.Rewards || {};

  if (rewards.ItemId && ids.itemIds.size > 0 && !ids.itemIds.has(rewards.ItemId)) {
    warnings.push({ field: 'Rewards.ItemId', message: `Reward references non-existent item #${rewards.ItemId}`, severity: 'error' });
  }
  if (rewards.BuffId && ids.buffIds.size > 0 && !ids.buffIds.has(rewards.BuffId)) {
    warnings.push({ field: 'Rewards.BuffId', message: `Reward references non-existent buff #${rewards.BuffId}`, severity: 'error' });
  }
  if (rewards.RoomId && ids.roomIds.size > 0 && !ids.roomIds.has(rewards.RoomId)) {
    warnings.push({ field: 'Rewards.RoomId', message: `Reward teleports to non-existent room #${rewards.RoomId}`, severity: 'error' });
  }
  if (rewards.QuestId) {
    const nextQuestId = parseInt(rewards.QuestId.split('-')[0], 10);
    if (!isNaN(nextQuestId) && ids.questIds.size > 0 && !ids.questIds.has(nextQuestId)) {
      warnings.push({ field: 'Rewards.QuestId', message: `Reward chains to non-existent quest #${nextQuestId}`, severity: 'error' });
    }
  }

  return warnings;
}

// Display component for warnings
export { ValidationPanel } from '../components/ValidationPanel';

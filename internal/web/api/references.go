package api

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/quests"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

type reference struct {
	Type    string `json:"type"`
	Id      any    `json:"id"`
	Name    string `json:"name"`
	Context string `json:"context"`
}

func handleGetReferences(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityId := r.PathValue("id")

	var refs []reference

	switch entityType {
	case "room":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid room ID")
			return
		}
		refs = findRoomReferences(id)
	case "mob":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid mob ID")
			return
		}
		refs = findMobReferences(id)
	case "item":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid item ID")
			return
		}
		refs = findItemReferences(id)
	case "quest":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid quest ID")
			return
		}
		refs = findQuestReferences(id)
	case "buff":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid buff ID")
			return
		}
		refs = findBuffReferences(id)
	default:
		writeError(w, http.StatusBadRequest, "Unknown entity type")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"references": refs,
		"total":      len(refs),
	})
}

func findRoomReferences(roomId int) []reference {
	var refs []reference

	// Find rooms with exits pointing here
	allRoomIds := rooms.GetAllRoomIds()
	for _, rid := range allRoomIds {
		if rid == roomId {
			continue
		}
		room := rooms.LoadRoom(rid)
		if room == nil {
			continue
		}
		for dir, ex := range room.Exits {
			if ex.RoomId == roomId {
				refs = append(refs, reference{Type: "room", Id: rid, Name: room.Title, Context: "Exit '" + dir + "' leads here"})
			}
		}
	}

	// Find quests that teleport here
	for _, q := range quests.GetAllQuests() {
		if q.Rewards.RoomId == roomId {
			refs = append(refs, reference{Type: "quest", Id: q.QuestId, Name: q.Name, Context: "Reward teleports here"})
		}
	}

	return refs
}

func findMobReferences(mobId int) []reference {
	var refs []reference

	// Find rooms that spawn this mob
	allRoomIds := rooms.GetAllRoomIds()
	for _, rid := range allRoomIds {
		room := rooms.LoadRoom(rid)
		if room == nil {
			continue
		}
		for _, spawn := range room.SpawnInfo {
			if spawn.MobId == mobId {
				refs = append(refs, reference{Type: "room", Id: rid, Name: room.Title, Context: "Spawns this mob"})
				break
			}
		}
	}

	return refs
}

func findItemReferences(itemId int) []reference {
	var refs []reference

	// Find rooms that spawn this item
	allRoomIds := rooms.GetAllRoomIds()
	for _, rid := range allRoomIds {
		room := rooms.LoadRoom(rid)
		if room == nil {
			continue
		}
		for _, spawn := range room.SpawnInfo {
			if spawn.ItemId == itemId {
				refs = append(refs, reference{Type: "room", Id: rid, Name: room.Title, Context: "Spawns this item"})
				break
			}
		}
	}

	// Find mobs that sell this item (shop)
	allMobs := mobs.GetAllMobInfo()
	for _, mob := range allMobs {
		for _, shopEntry := range mob.Character.Shop {
			if shopEntry.ItemId == itemId {
				refs = append(refs, reference{Type: "mob", Id: int(mob.MobId), Name: mob.Character.Name, Context: "Sells this item"})
				break
			}
		}
	}

	// Find quests that reward this item
	for _, q := range quests.GetAllQuests() {
		if q.Rewards.ItemId == itemId {
			refs = append(refs, reference{Type: "quest", Id: q.QuestId, Name: q.Name, Context: "Rewards this item"})
		}
	}

	return refs
}

func findQuestReferences(questId int) []reference {
	var refs []reference

	// Find quests that chain to this quest
	for _, q := range quests.GetAllQuests() {
		if q.QuestId == questId {
			continue
		}
		nextId, _ := quests.TokenToParts(q.Rewards.QuestId)
		if nextId == questId {
			refs = append(refs, reference{Type: "quest", Id: q.QuestId, Name: q.Name, Context: "Chains to this quest"})
		}
	}

	// Find mobs with quest flags referencing this quest
	allMobs := mobs.GetAllMobInfo()
	for _, mob := range allMobs {
		for _, flag := range mob.QuestFlags {
			flagId, _ := quests.TokenToParts(flag)
			if flagId == questId {
				refs = append(refs, reference{Type: "mob", Id: int(mob.MobId), Name: mob.Character.Name, Context: "Has quest flag for this quest"})
				break
			}
		}
	}

	return refs
}

func findBuffReferences(buffId int) []reference {
	var refs []reference

	// Items that use this buff
	allItems := items.GetAllItemSpecs()
	itemIds := []int{}
	for id := range allItems {
		itemIds = append(itemIds, id)
	}
	sort.Ints(itemIds)

	for _, id := range itemIds {
		item := allItems[id]
		for _, bid := range item.BuffIds {
			if bid == buffId {
				refs = append(refs, reference{Type: "item", Id: item.ItemId, Name: item.Name, Context: "Uses this buff on use"})
				break
			}
		}
		for _, bid := range item.WornBuffIds {
			if bid == buffId {
				refs = append(refs, reference{Type: "item", Id: item.ItemId, Name: item.Name, Context: "Applies this buff when worn"})
				break
			}
		}
	}

	// Mobs with this buff
	allMobs := mobs.GetAllMobInfo()
	for _, mob := range allMobs {
		for _, bid := range mob.BuffIds {
			if bid == buffId {
				refs = append(refs, reference{Type: "mob", Id: int(mob.MobId), Name: mob.Character.Name, Context: "Has this buff"})
				break
			}
		}
	}

	return refs
}

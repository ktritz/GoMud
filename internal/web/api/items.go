package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/items"
)

type itemSummary struct {
	ItemId  int    `json:"itemId"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Value   int    `json:"value"`
}

func handleListItems(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	itemType := r.URL.Query().Get("type")

	allItems := items.GetAllItemSpecs()
	result := []itemSummary{}

	for _, item := range allItems {
		if itemType != "" && string(item.Type) != itemType {
			continue
		}
		if search != "" && !containsCI(item.Name, search) {
			continue
		}
		result = append(result, itemSummary{
			ItemId:  item.ItemId,
			Name:    item.Name,
			Type:    string(item.Type),
			Subtype: string(item.Subtype),
			Value:   item.Value,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ItemId < result[j].ItemId })

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func handleGetItem(w http.ResponseWriter, r *http.Request) {
	itemId, err := strconv.Atoi(r.PathValue("itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item := items.GetItemSpec(itemId)
	if item.ItemId == 0 {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

type itemCreateRequest struct {
	Name    string `json:"Name"`
	Type    string `json:"Type"`
	Subtype string `json:"Subtype,omitempty"`
}

func handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var req itemCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "Name and Type are required")
		return
	}

	spec := items.ItemSpec{
		Name:     req.Name,
		Type:     items.ItemType(req.Type),
		Subtype:  items.ItemSubType(req.Subtype),
	}

	newId, err := items.CreateNewItemFile(spec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create item: "+err.Error())
		return
	}

	created := items.GetItemSpec(newId)
	writeJSON(w, http.StatusCreated, created)
}

type itemUpdateRequest struct {
	Name            *string            `json:"Name,omitempty"`
	DisplayName     *string            `json:"DisplayName,omitempty"`
	NameSimple      *string            `json:"NameSimple,omitempty"`
	Description     *string            `json:"Description,omitempty"`
	Type            *string            `json:"Type,omitempty"`
	Subtype         *string            `json:"Subtype,omitempty"`
	Value           *int               `json:"Value,omitempty"`
	Uses            *int               `json:"Uses,omitempty"`
	Hands           *int               `json:"Hands,omitempty"`
	DamageReduction *int               `json:"DamageReduction,omitempty"`
	WaitRounds      *int               `json:"WaitRounds,omitempty"`
	BreakChance     *int               `json:"BreakChance,omitempty"`
	Cursed          *bool              `json:"Cursed,omitempty"`
	KeyLockId       *string            `json:"KeyLockId,omitempty"`
	QuestToken      *string            `json:"QuestToken,omitempty"`
	Element         *string            `json:"Element,omitempty"`
	BuffIds         *[]int             `json:"BuffIds,omitempty"`
	WornBuffIds     *[]int             `json:"WornBuffIds,omitempty"`
	StatMods        map[string]int     `json:"StatMods,omitempty"`
	DiceRoll        *string            `json:"DiceRoll,omitempty"`
	Attacks         *int               `json:"Attacks,omitempty"`
	BonusDamage     *int               `json:"BonusDamage,omitempty"`
	CritBuffIds     *[]int             `json:"CritBuffIds,omitempty"`
}

func handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	itemId, err := strconv.Atoi(r.PathValue("itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item := items.GetItemSpec(itemId)
	if item.ItemId == 0 {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	var req itemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name != nil { item.Name = *req.Name }
	if req.DisplayName != nil { item.DisplayName = *req.DisplayName }
	if req.NameSimple != nil { item.NameSimple = *req.NameSimple }
	if req.Description != nil { item.Description = *req.Description }
	if req.Type != nil { item.Type = items.ItemType(*req.Type) }
	if req.Subtype != nil { item.Subtype = items.ItemSubType(*req.Subtype) }
	if req.Value != nil { item.Value = *req.Value }
	if req.Uses != nil { item.Uses = *req.Uses }
	if req.Hands != nil { item.Hands = items.WeaponHands(*req.Hands) }
	if req.DamageReduction != nil { item.DamageReduction = *req.DamageReduction }
	if req.WaitRounds != nil { item.WaitRounds = *req.WaitRounds }
	if req.BreakChance != nil { item.BreakChance = uint8(*req.BreakChance) }
	if req.Cursed != nil { item.Cursed = *req.Cursed }
	if req.KeyLockId != nil { item.KeyLockId = *req.KeyLockId }
	if req.QuestToken != nil { item.QuestToken = *req.QuestToken }
	if req.Element != nil { item.Element = items.Element(*req.Element) }
	if req.BuffIds != nil { item.BuffIds = *req.BuffIds }
	if req.WornBuffIds != nil { item.WornBuffIds = *req.WornBuffIds }
	if req.StatMods != nil {
		mods := make(map[string]int)
		for k, v := range req.StatMods {
			mods[k] = v
		}
		item.StatMods = mods
	}
	if req.DiceRoll != nil { item.Damage.DiceRoll = *req.DiceRoll }
	if req.Attacks != nil { item.Damage.Attacks = *req.Attacks }
	if req.BonusDamage != nil { item.Damage.BonusDamage = *req.BonusDamage }
	if req.CritBuffIds != nil { item.Damage.CritBuffIds = *req.CritBuffIds }

	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*items.ItemSpec](configs.GetFilePathsConfig().DataFiles.String()+`/items`, item, saveModes...); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save item: "+err.Error())
		return
	}

	items.LoadDataFiles()

	updated := items.GetItemSpec(itemId)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	itemId, err := strconv.Atoi(r.PathValue("itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	item := items.GetItemSpec(itemId)
	if item == nil {
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	itemPath := dataDir + "/items/" + item.Filepath()
	if err := os.Remove(itemPath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete item file: "+err.Error())
		return
	}

	items.LoadDataFiles()

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

package api

import (
	"net/http"
	"sort"
	"strconv"

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

func handleCreateItem(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

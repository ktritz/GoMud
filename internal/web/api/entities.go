package api

import (
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/spells"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// Zones

func handleListZones(w http.ResponseWriter, r *http.Request) {
	zones := rooms.GetAllZoneNames()
	result := []map[string]any{}

	for _, zone := range zones {
		result = append(result, map[string]any{
			"name": zone,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"zones": result,
		"total": len(result),
	})
}

func handleGetZone(w http.ResponseWriter, r *http.Request) {
	zoneName := r.PathValue("zoneName")
	if zoneName == "" {
		writeError(w, http.StatusBadRequest, "Zone name required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name": zoneName,
	})
}

// Buffs

func handleListBuffs(w http.ResponseWriter, r *http.Request) {
	buffIds := buffs.GetAllBuffIds()
	result := []map[string]any{}
	for _, id := range buffIds {
		if spec := buffs.GetBuffSpec(id); spec != nil {
			result = append(result, map[string]any{
				"buffId":      spec.BuffId,
				"name":        spec.Name,
				"description": spec.Description,
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"buffs": result,
		"total": len(result),
	})
}

func handleGetBuff(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

// Spells

func handleListSpells(w http.ResponseWriter, r *http.Request) {
	allSpells := spells.GetAllSpells()
	result := []map[string]any{}
	for id, spell := range allSpells {
		result = append(result, map[string]any{
			"spellId": id,
			"name":    spell.Name,
			"type":    spell.Type,
			"cost":    spell.Cost,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"spells": result,
		"total":  len(result),
	})
}

func handleGetSpell(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

// Races

func handleListRaces(w http.ResponseWriter, r *http.Request) {
	allRaces := races.GetRaces()
	writeJSON(w, http.StatusOK, map[string]any{
		"races": allRaces,
		"total": len(allRaces),
	})
}

func handleGetRace(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

// Server

func handleReload(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	connCount, disconnCount := connections.Stats()
	onlineUsers := users.GetAllActiveUsers()

	writeJSON(w, http.StatusOK, map[string]any{
		"connections":    connCount,
		"disconnections": disconnCount,
		"onlineUsers":   len(onlineUsers),
	})
}

// Utility

func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

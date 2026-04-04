package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/spells"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// Zones

func handleListZones(w http.ResponseWriter, r *http.Request) {
	zones := rooms.GetAllZoneNames()
	sort.Strings(zones)
	result := []map[string]any{}

	for _, zone := range zones {
		roomIds := rooms.GetAllZoneRoomsIds(zone)
		result = append(result, map[string]any{
			"name":      zone,
			"roomCount": len(roomIds),
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

func handleGetZoneMap(w http.ResponseWriter, r *http.Request) {
	zoneName := r.PathValue("zoneName")
	if zoneName == "" {
		writeError(w, http.StatusBadRequest, "Zone name required")
		return
	}

	cfg := rooms.GetZoneConfig(zoneName)
	if cfg.RoomId == 0 {
		writeError(w, http.StatusNotFound, "Zone not found")
		return
	}

	m := mapper.GetMapper(cfg.RoomId)
	if m == nil {
		writeError(w, http.StatusNotFound, "No map data for zone")
		return
	}

	nodes := m.GetMapData()

	// Also fetch room titles
	type mapRoomInfo struct {
		mapper.MapNodeInfo
		Title string `json:"title"`
	}

	result := make([]mapRoomInfo, 0, len(nodes))
	for _, node := range nodes {
		title := ""
		if room := rooms.LoadRoom(node.RoomId); room != nil {
			title = room.Title
		}
		result = append(result, mapRoomInfo{
			MapNodeInfo: node,
			Title:       title,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"zone":     zoneName,
		"rootRoom": cfg.RoomId,
		"rooms":    result,
		"total":    len(result),
	})
}

// Buffs

func handleListBuffs(w http.ResponseWriter, r *http.Request) {
	buffIds := buffs.GetAllBuffIds()
	sort.Ints(buffIds)
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
	buffId, err := strconv.Atoi(r.PathValue("buffId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid buff ID")
		return
	}

	spec := buffs.GetBuffSpec(buffId)
	if spec == nil {
		writeError(w, http.StatusNotFound, "Buff not found")
		return
	}

	writeJSON(w, http.StatusOK, spec)
}

type buffCreateRequest struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
}

func handleCreateBuff(w http.ResponseWriter, r *http.Request) {
	var req buffCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	spec := buffs.BuffSpec{
		Name:         req.Name,
		Description:  req.Description,
		TriggerCount: 1,
		TriggerRate:  "1 round",
	}

	newId, err := buffs.CreateNewBuff(spec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create buff: "+err.Error())
		return
	}

	created := buffs.GetBuffSpec(newId)
	writeJSON(w, http.StatusCreated, created)
}

type buffUpdateRequest struct {
	Name         *string           `json:"Name,omitempty"`
	Description  *string           `json:"Description,omitempty"`
	Secret       *bool             `json:"Secret,omitempty"`
	TriggerNow   *bool             `json:"TriggerNow,omitempty"`
	TriggerRate  *string           `json:"TriggerRate,omitempty"`
	TriggerCount *int              `json:"TriggerCount,omitempty"`
	StatMods     map[string]int    `json:"StatMods,omitempty"`
	Flags        *[]buffs.Flag     `json:"Flags,omitempty"`
}

func handleUpdateBuff(w http.ResponseWriter, r *http.Request) {
	buffId, err := strconv.Atoi(r.PathValue("buffId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid buff ID")
		return
	}

	spec := buffs.GetBuffSpec(buffId)
	if spec == nil {
		writeError(w, http.StatusNotFound, "Buff not found")
		return
	}

	var req buffUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name != nil { spec.Name = *req.Name }
	if req.Description != nil { spec.Description = *req.Description }
	if req.Secret != nil { spec.Secret = *req.Secret }
	if req.TriggerNow != nil { spec.TriggerNow = *req.TriggerNow }
	if req.TriggerRate != nil { spec.TriggerRate = *req.TriggerRate }
	if req.TriggerCount != nil { spec.TriggerCount = *req.TriggerCount }
	if req.StatMods != nil { spec.StatMods = req.StatMods }
	if req.Flags != nil { spec.Flags = *req.Flags }

	if err := buffs.SaveBuff(spec); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save buff: "+err.Error())
		return
	}

	buffs.LoadDataFiles()
	updated := buffs.GetBuffSpec(buffId)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteBuff(w http.ResponseWriter, r *http.Request) {
	buffId, err := strconv.Atoi(r.PathValue("buffId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid buff ID")
		return
	}

	spec := buffs.GetBuffSpec(buffId)
	if spec == nil {
		writeError(w, http.StatusNotFound, "Buff not found")
		return
	}

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	filePath := dataDir + "/buffs/" + spec.Filepath()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete: "+err.Error())
		return
	}

	buffs.LoadDataFiles()
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
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
	sort.Slice(result, func(i, j int) bool {
		return result[i]["spellId"].(string) < result[j]["spellId"].(string)
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"spells": result,
		"total":  len(result),
	})
}

func handleGetSpell(w http.ResponseWriter, r *http.Request) {
	spellId := r.PathValue("spellId")
	if spellId == "" {
		writeError(w, http.StatusBadRequest, "Spell ID required")
		return
	}

	spell := spells.GetSpell(spellId)
	if spell == nil {
		writeError(w, http.StatusNotFound, "Spell not found")
		return
	}

	writeJSON(w, http.StatusOK, spell)
}

type spellCreateRequest struct {
	SpellId string `json:"SpellId"`
	Name    string `json:"Name"`
	Type    string `json:"Type,omitempty"`
}

func handleCreateSpell(w http.ResponseWriter, r *http.Request) {
	var req spellCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if req.SpellId == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "SpellId and Name are required")
		return
	}

	spell := spells.SpellData{
		SpellId: req.SpellId,
		Name:    req.Name,
		Type:    spells.SpellType(req.Type),
	}

	_, err := spells.CreateNewSpell(spell)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create spell: "+err.Error())
		return
	}

	created := spells.GetSpell(req.SpellId)
	writeJSON(w, http.StatusCreated, created)
}

type spellUpdateRequest struct {
	Name        *string `json:"Name,omitempty"`
	Description *string `json:"Description,omitempty"`
	Type        *string `json:"Type,omitempty"`
	School      *string `json:"School,omitempty"`
	Cost        *int    `json:"Cost,omitempty"`
	WaitRounds  *int    `json:"WaitRounds,omitempty"`
	Difficulty  *int    `json:"Difficulty,omitempty"`
}

func handleUpdateSpell(w http.ResponseWriter, r *http.Request) {
	spellId := r.PathValue("spellId")
	if spellId == "" {
		writeError(w, http.StatusBadRequest, "Spell ID required")
		return
	}

	spell := spells.GetSpell(spellId)
	if spell == nil {
		writeError(w, http.StatusNotFound, "Spell not found")
		return
	}

	var req spellUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name != nil { spell.Name = *req.Name }
	if req.Description != nil { spell.Description = *req.Description }
	if req.Type != nil { spell.Type = spells.SpellType(*req.Type) }
	if req.School != nil { spell.School = spells.SpellSchool(*req.School) }
	if req.Cost != nil { spell.Cost = *req.Cost }
	if req.WaitRounds != nil { spell.WaitRounds = *req.WaitRounds }
	if req.Difficulty != nil { spell.Difficulty = *req.Difficulty }

	if err := spells.SaveSpell(spell); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save spell: "+err.Error())
		return
	}

	spells.LoadSpellFiles()
	updated := spells.GetSpell(spellId)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteSpell(w http.ResponseWriter, r *http.Request) {
	spellId := r.PathValue("spellId")
	if spellId == "" {
		writeError(w, http.StatusBadRequest, "Spell ID required")
		return
	}

	spell := spells.GetSpell(spellId)
	if spell == nil {
		writeError(w, http.StatusNotFound, "Spell not found")
		return
	}

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	filePath := dataDir + "/spells/" + spell.Filepath()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete: "+err.Error())
		return
	}

	spells.LoadSpellFiles()
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// Races

func handleListRaces(w http.ResponseWriter, r *http.Request) {
	allRaces := races.GetRaces()
	sort.Slice(allRaces, func(i, j int) bool { return allRaces[i].RaceId < allRaces[j].RaceId })
	writeJSON(w, http.StatusOK, map[string]any{
		"races": allRaces,
		"total": len(allRaces),
	})
}

func handleGetRace(w http.ResponseWriter, r *http.Request) {
	raceId, err := strconv.Atoi(r.PathValue("raceId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	race := races.GetRace(raceId)
	if race == nil {
		writeError(w, http.StatusNotFound, "Race not found")
		return
	}

	writeJSON(w, http.StatusOK, race)
}

type raceCreateRequest struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
	Size        string `json:"Size,omitempty"`
}

func handleCreateRace(w http.ResponseWriter, r *http.Request) {
	var req raceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	race := races.Race{
		Name:        req.Name,
		Description: req.Description,
		Size:        races.Size(req.Size),
	}
	if race.Size == "" {
		race.Size = "medium"
	}

	newId, err := races.CreateNewRace(race)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create race: "+err.Error())
		return
	}

	created := races.GetRace(newId)
	writeJSON(w, http.StatusCreated, created)
}

type raceUpdateRequest struct {
	Name             *string  `json:"Name,omitempty"`
	Description      *string  `json:"Description,omitempty"`
	Size             *string  `json:"Size,omitempty"`
	DefaultAlignment *int8    `json:"DefaultAlignment,omitempty"`
	TNLScale         *float32 `json:"TNLScale,omitempty"`
	UnarmedName      *string  `json:"UnarmedName,omitempty"`
	Tameable         *bool    `json:"Tameable,omitempty"`
	Selectable       *bool    `json:"Selectable,omitempty"`
	KnowsFirstAid   *bool    `json:"KnowsFirstAid,omitempty"`
}

func handleUpdateRace(w http.ResponseWriter, r *http.Request) {
	raceId, err := strconv.Atoi(r.PathValue("raceId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	race := races.GetRace(raceId)
	if race == nil {
		writeError(w, http.StatusNotFound, "Race not found")
		return
	}

	var req raceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name != nil { race.Name = *req.Name }
	if req.Description != nil { race.Description = *req.Description }
	if req.Size != nil { race.Size = races.Size(*req.Size) }
	if req.DefaultAlignment != nil { race.DefaultAlignment = *req.DefaultAlignment }
	if req.TNLScale != nil { race.TNLScale = *req.TNLScale }
	if req.UnarmedName != nil { race.UnarmedName = *req.UnarmedName }
	if req.Tameable != nil { race.Tameable = *req.Tameable }
	if req.Selectable != nil { race.Selectable = *req.Selectable }
	if req.KnowsFirstAid != nil { race.KnowsFirstAid = *req.KnowsFirstAid }

	if err := race.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save race: "+err.Error())
		return
	}

	races.LoadDataFiles()
	updated := races.GetRace(raceId)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteRace(w http.ResponseWriter, r *http.Request) {
	raceId, err := strconv.Atoi(r.PathValue("raceId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid race ID")
		return
	}

	race := races.GetRace(raceId)
	if race == nil {
		writeError(w, http.StatusNotFound, "Race not found")
		return
	}

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	filePath := dataDir + "/races/" + race.Filepath()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete: "+err.Error())
		return
	}

	races.LoadDataFiles()
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
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

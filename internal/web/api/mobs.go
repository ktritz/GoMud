package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/mobs"
)

type mobSummary struct {
	MobId    int    `json:"mobId"`
	Zone     string `json:"zone"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Hostile  bool   `json:"hostile"`
}

func handleListMobs(w http.ResponseWriter, r *http.Request) {
	zone := r.URL.Query().Get("zone")
	search := r.URL.Query().Get("search")

	allMobs := mobs.GetAllMobInfo()
	result := []mobSummary{}

	for _, mob := range allMobs {
		if zone != "" && mob.Zone != zone {
			continue
		}
		if search != "" && !containsCI(mob.Character.Name, search) {
			continue
		}
		result = append(result, mobSummary{
			MobId:   int(mob.MobId),
			Zone:    mob.Zone,
			Name:    mob.Character.Name,
			Level:   mob.Character.Level,
			Hostile: mob.Hostile,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].MobId < result[j].MobId })

	writeJSON(w, http.StatusOK, map[string]any{
		"mobs":  result,
		"total": len(result),
	})
}

func handleGetMob(w http.ResponseWriter, r *http.Request) {
	mobId, err := strconv.Atoi(r.PathValue("mobId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid mob ID")
		return
	}

	mob := mobs.GetMobSpec(mobs.MobId(mobId))
	if mob == nil {
		writeError(w, http.StatusNotFound, "Mob not found")
		return
	}

	// Resolve cached description hash to actual text
	mobCopy := *mob
	mobCopy.Character.Description = mob.Character.GetDescription()

	writeJSON(w, http.StatusOK, mobCopy)
}

func handleCreateMob(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

type mobUpdateRequest struct {
	Hostile       *bool     `json:"Hostile,omitempty"`
	DeathMessage  *string   `json:"DeathMessage,omitempty"`
	NoCorpse      *bool     `json:"NoCorpse,omitempty"`
	ActivityLevel *int      `json:"ActivityLevel,omitempty"`
	MaxWander     *int      `json:"MaxWander,omitempty"`
	Groups        *[]string `json:"Groups,omitempty"`
	IdleCommands  *[]string `json:"IdleCommands,omitempty"`
	CharName      *string   `json:"CharName,omitempty"`
	CharDesc      *string   `json:"CharDescription,omitempty"`
	CharLevel     *int      `json:"CharLevel,omitempty"`
	CharGold      *int      `json:"CharGold,omitempty"`
}

func handleUpdateMob(w http.ResponseWriter, r *http.Request) {
	mobId, err := strconv.Atoi(r.PathValue("mobId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid mob ID")
		return
	}

	mob := mobs.GetMobSpec(mobs.MobId(mobId))
	if mob == nil {
		writeError(w, http.StatusNotFound, "Mob not found")
		return
	}

	var req mobUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Hostile != nil { mob.Hostile = *req.Hostile }
	if req.DeathMessage != nil { mob.DeathMessage = *req.DeathMessage }
	if req.NoCorpse != nil { mob.NoCorpse = *req.NoCorpse }
	if req.ActivityLevel != nil { mob.ActivityLevel = *req.ActivityLevel }
	if req.MaxWander != nil { mob.MaxWander = *req.MaxWander }
	if req.Groups != nil { mob.Groups = *req.Groups }
	if req.IdleCommands != nil { mob.IdleCommands = *req.IdleCommands }
	if req.CharName != nil { mob.Character.Name = *req.CharName }
	if req.CharDesc != nil { mob.Character.Description = *req.CharDesc }
	if req.CharLevel != nil { mob.Character.Level = *req.CharLevel }
	if req.CharGold != nil { mob.Character.Gold = *req.CharGold }

	if err := mob.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save mob: "+err.Error())
		return
	}

	// Reload mob data from disk
	mobs.LoadDataFiles()

	updated := mobs.GetMobSpec(mobs.MobId(mobId))
	updated.Character.Description = updated.Character.GetDescription()
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteMob(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

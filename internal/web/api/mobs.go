package api

import (
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

func handleUpdateMob(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleDeleteMob(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

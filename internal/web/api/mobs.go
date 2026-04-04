package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
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

type mobCreateRequest struct {
	Name  string `json:"Name"`
	Zone  string `json:"Zone"`
	Level int    `json:"Level,omitempty"`
}

func handleCreateMob(w http.ResponseWriter, r *http.Request) {
	var req mobCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" || req.Zone == "" {
		writeError(w, http.StatusBadRequest, "Name and Zone are required")
		return
	}

	newMob := mobs.Mob{
		Zone: req.Zone,
		Character: characters.Character{
			Name:  req.Name,
			Level: req.Level,
		},
	}

	newId, err := mobs.CreateNewMobFile(newMob, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create mob: "+err.Error())
		return
	}

	mobs.LoadDataFiles()

	created := mobs.GetMobSpec(newId)
	if created != nil {
		created.Character.Description = created.Character.GetDescription()
	}
	writeJSON(w, http.StatusCreated, created)
}

type mobUpdateRequest struct {
	Hostile         *bool     `json:"Hostile,omitempty"`
	DeathMessage    *string   `json:"DeathMessage,omitempty"`
	NoCorpse        *bool     `json:"NoCorpse,omitempty"`
	ActivityLevel   *int      `json:"ActivityLevel,omitempty"`
	MaxWander       *int      `json:"MaxWander,omitempty"`
	ItemDropChance  *int      `json:"ItemDropChance,omitempty"`
	ScriptTag       *string   `json:"ScriptTag,omitempty"`
	Groups          *[]string `json:"Groups,omitempty"`
	Hates           *[]string `json:"Hates,omitempty"`
	IdleCommands    *[]string `json:"IdleCommands,omitempty"`
	AngryCommands   *[]string `json:"AngryCommands,omitempty"`
	CombatCommands  *[]string `json:"CombatCommands,omitempty"`
	QuestFlags      *[]string `json:"QuestFlags,omitempty"`
	BuffIds         *[]int    `json:"BuffIds,omitempty"`
	CharName        *string   `json:"CharName,omitempty"`
	CharDesc        *string   `json:"CharDescription,omitempty"`
	CharLevel       *int      `json:"CharLevel,omitempty"`
	CharGold        *int      `json:"CharGold,omitempty"`
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
	if req.ItemDropChance != nil { mob.ItemDropChance = *req.ItemDropChance }
	if req.ScriptTag != nil { mob.ScriptTag = *req.ScriptTag }
	if req.Groups != nil { mob.Groups = *req.Groups }
	if req.Hates != nil { mob.Hates = *req.Hates }
	if req.IdleCommands != nil { mob.IdleCommands = *req.IdleCommands }
	if req.AngryCommands != nil { mob.AngryCommands = *req.AngryCommands }
	if req.CombatCommands != nil { mob.CombatCommands = *req.CombatCommands }
	if req.QuestFlags != nil { mob.QuestFlags = *req.QuestFlags }
	if req.BuffIds != nil { mob.BuffIds = *req.BuffIds }
	if req.CharName != nil { mob.Character.Name = *req.CharName }
	if req.CharDesc != nil { mob.Character.Description = *req.CharDesc }
	if req.CharLevel != nil { mob.Character.Level = *req.CharLevel }
	if req.CharGold != nil { mob.Character.Gold = *req.CharGold }

	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}
	if err := fileloader.SaveFlatFile[*mobs.Mob](configs.GetFilePathsConfig().DataFiles.String()+`/mobs`, mob, saveModes...); err != nil {
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

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	mobPath := dataDir + "/mobs/" + mob.Filepath()
	if err := os.Remove(mobPath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete mob file: "+err.Error())
		return
	}

	mobs.LoadDataFiles()

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

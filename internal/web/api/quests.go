package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/quests"
)

func handleListQuests(w http.ResponseWriter, r *http.Request) {
	allQuests := quests.GetAllQuests()
	sort.Slice(allQuests, func(i, j int) bool { return allQuests[i].QuestId < allQuests[j].QuestId })
	writeJSON(w, http.StatusOK, map[string]any{
		"quests": allQuests,
		"total":  len(allQuests),
	})
}

func handleGetQuest(w http.ResponseWriter, r *http.Request) {
	questId := r.PathValue("questId")
	if questId == "" {
		writeError(w, http.StatusBadRequest, "Quest ID required")
		return
	}

	// Try as integer first for direct ID lookup
	if id, err := strconv.Atoi(questId); err == nil {
		token := quests.PartsToToken(id, "all+")
		quest := quests.GetQuest(token)
		if quest != nil {
			writeJSON(w, http.StatusOK, quest)
			return
		}
	}

	quest := quests.GetQuest(questId)
	if quest == nil {
		writeError(w, http.StatusNotFound, "Quest not found")
		return
	}

	writeJSON(w, http.StatusOK, quest)
}

type questCreateRequest struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
}

func handleCreateQuest(w http.ResponseWriter, r *http.Request) {
	var req questCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	quest := quests.Quest{
		Name:        req.Name,
		Description: req.Description,
		Steps: []quests.QuestStep{
			{Id: "start", Description: "Begin the quest"},
			{Id: "end", Description: "Complete the quest"},
		},
	}

	newId, err := quests.CreateNewQuest(quest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create quest: "+err.Error())
		return
	}

	token := quests.PartsToToken(newId, "all+")
	created := quests.GetQuest(token)
	writeJSON(w, http.StatusCreated, created)
}

type questUpdateRequest struct {
	Name        *string             `json:"Name,omitempty"`
	Description *string             `json:"Description,omitempty"`
	Secret      *bool               `json:"Secret,omitempty"`
	Steps       *[]quests.QuestStep `json:"Steps,omitempty"`
	Rewards     *quests.QuestReward `json:"Rewards,omitempty"`
}

func handleUpdateQuest(w http.ResponseWriter, r *http.Request) {
	questId, err := strconv.Atoi(r.PathValue("questId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid quest ID")
		return
	}

	token := quests.PartsToToken(questId, "all+")
	quest := quests.GetQuest(token)
	if quest == nil {
		writeError(w, http.StatusNotFound, "Quest not found")
		return
	}

	var req questUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name != nil { quest.Name = *req.Name }
	if req.Description != nil { quest.Description = *req.Description }
	if req.Secret != nil { quest.Secret = *req.Secret }
	if req.Steps != nil { quest.Steps = *req.Steps }
	if req.Rewards != nil { quest.Rewards = *req.Rewards }

	if err := quests.SaveQuest(quest); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save quest: "+err.Error())
		return
	}

	quests.LoadDataFiles()

	updated := quests.GetQuest(token)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
	questId, err := strconv.Atoi(r.PathValue("questId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid quest ID")
		return
	}

	token := quests.PartsToToken(questId, "all+")
	quest := quests.GetQuest(token)
	if quest == nil {
		writeError(w, http.StatusNotFound, "Quest not found")
		return
	}

	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	questPath := dataDir + "/quests/" + quest.Filepath()
	if err := os.Remove(questPath); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Failed to delete quest file: "+err.Error())
		return
	}

	quests.LoadDataFiles()

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

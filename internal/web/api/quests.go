package api

import (
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/quests"
)

func handleListQuests(w http.ResponseWriter, r *http.Request) {
	allQuests := quests.GetAllQuests()
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

	quest := quests.GetQuest(questId)
	if quest == nil {
		writeError(w, http.StatusNotFound, "Quest not found")
		return
	}

	writeJSON(w, http.StatusOK, quest)
}

func handleCreateQuest(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleUpdateQuest(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

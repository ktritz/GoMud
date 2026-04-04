package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func handleGetScript(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityId := r.PathValue("id")

	scriptPath, err := resolveScriptPath(entityType, entityId)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]any{
				"path":    scriptPath,
				"content": "",
				"exists":  false,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to read script: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":    scriptPath,
		"content": string(content),
		"exists":  true,
	})
}

type scriptSaveRequest struct {
	Content string `json:"content"`
}

func handleSaveScript(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityId := r.PathValue("id")

	scriptPath, err := resolveScriptPath(entityType, entityId)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req scriptSaveRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Content == "" {
		// Delete empty scripts
		os.Remove(scriptPath)
		writeJSON(w, http.StatusOK, map[string]any{"saved": true, "deleted": true})
		return
	}

	if err := os.WriteFile(scriptPath, []byte(req.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to write script: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

// Dialogue tree endpoints — stores JSON design data alongside scripts
func handleGetDialogue(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityId := r.PathValue("id")

	scriptPath, err := resolveScriptPath(entityType, entityId)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	dialoguePath := strings.Replace(scriptPath, ".js", ".dialogue.json", 1)
	content, err := os.ReadFile(dialoguePath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]any{"exists": false, "tree": nil})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to read dialogue: "+err.Error())
		return
	}

	var tree any
	json.Unmarshal(content, &tree)
	writeJSON(w, http.StatusOK, map[string]any{"exists": true, "tree": tree})
}

func handleSaveDialogue(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityId := r.PathValue("id")

	scriptPath, err := resolveScriptPath(entityType, entityId)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	dialoguePath := strings.Replace(scriptPath, ".js", ".dialogue.json", 1)

	body, _ := io.ReadAll(r.Body)

	// Ensure the directory exists
	dir := dialoguePath[:strings.LastIndex(dialoguePath, "/")]
	os.MkdirAll(dir, 0755)

	if err := os.WriteFile(dialoguePath, body, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save dialogue: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func resolveScriptPath(entityType, entityId string) (string, error) {
	switch entityType {
	case "mob":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			return "", fmt.Errorf("invalid mob ID")
		}
		mob := mobs.GetMobSpec(mobs.MobId(id))
		if mob == nil {
			return "", fmt.Errorf("mob not found")
		}
		return mob.GetScriptPath(), nil
	case "room":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			return "", fmt.Errorf("invalid room ID")
		}
		room := rooms.LoadRoom(id)
		if room == nil {
			return "", fmt.Errorf("room not found")
		}
		return room.GetScriptPath(), nil
	case "item":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			return "", fmt.Errorf("invalid item ID")
		}
		item := items.GetItemSpec(id)
		if item == nil {
			return "", fmt.Errorf("item not found")
		}
		return item.GetScriptPath(), nil
	case "buff":
		id, err := strconv.Atoi(entityId)
		if err != nil {
			return "", fmt.Errorf("invalid buff ID")
		}
		buff := buffs.GetBuffSpec(id)
		if buff == nil {
			return "", fmt.Errorf("buff not found")
		}
		return buff.GetScriptPath(), nil
	default:
		return "", fmt.Errorf("unknown entity type: %s", entityType)
	}
}

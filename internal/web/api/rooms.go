package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

type roomSummary struct {
	RoomId    int    `json:"roomId"`
	Zone      string `json:"zone"`
	Title     string `json:"title"`
	Biome     string `json:"biome"`
	ExitCount int    `json:"exitCount"`
}

func handleListRooms(w http.ResponseWriter, r *http.Request) {
	zone := r.URL.Query().Get("zone")
	search := r.URL.Query().Get("search")

	var roomIds []int
	if zone != "" {
		roomIds = rooms.GetAllZoneRoomsIds(zone)
	} else {
		roomIds = rooms.GetAllRoomIds()
	}
	result := []roomSummary{}

	for _, id := range roomIds {
		room := rooms.LoadRoom(id)
		if room == nil {
			continue
		}
		if search != "" {
			// Simple case-insensitive search
			if !containsCI(room.Title, search) && !containsCI(room.Description, search) {
				continue
			}
		}
		result = append(result, roomSummary{
			RoomId:    room.RoomId,
			Zone:      room.Zone,
			Title:     room.Title,
			Biome:     room.Biome,
			ExitCount: len(room.Exits),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].RoomId < result[j].RoomId })

	writeJSON(w, http.StatusOK, map[string]any{
		"rooms": result,
		"total": len(result),
	})
}

func handleGetRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(r.PathValue("roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		writeError(w, http.StatusNotFound, "Room not found")
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

// roomUpdateRequest contains the editable fields for a room.
// JSON field names match the Go struct's exported field names.
type roomUpdateRequest struct {
	Title        *string                       `json:"Title,omitempty"`
	Description  *string                       `json:"Description,omitempty"`
	MapSymbol    *string                       `json:"MapSymbol,omitempty"`
	MapLegend    *string                       `json:"MapLegend,omitempty"`
	Biome        *string                       `json:"Biome,omitempty"`
	IdleMessages *[]string                     `json:"IdleMessages,omitempty"`
	Nouns        *map[string]string            `json:"Nouns,omitempty"`
	Exits        *map[string]exit.RoomExit     `json:"Exits,omitempty"`
	SpawnInfo    *[]rooms.SpawnInfo            `json:"SpawnInfo,omitempty"`
}

func handleUpdateRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(r.PathValue("roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		writeError(w, http.StatusNotFound, "Room not found")
		return
	}

	var req roomUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Apply partial updates
	if req.Title != nil {
		room.Title = *req.Title
	}
	if req.Description != nil {
		room.Description = *req.Description
	}
	if req.MapSymbol != nil {
		room.MapSymbol = *req.MapSymbol
	}
	if req.MapLegend != nil {
		room.MapLegend = *req.MapLegend
	}
	if req.Biome != nil {
		room.Biome = *req.Biome
	}
	if req.IdleMessages != nil {
		room.IdleMessages = *req.IdleMessages
	}
	if req.Nouns != nil {
		room.Nouns = *req.Nouns
	}
	if req.Exits != nil {
		room.Exits = *req.Exits
	}
	if req.SpawnInfo != nil {
		room.SpawnInfo = *req.SpawnInfo
	}

	// Save the template (strips runtime state, writes YAML, rebuilds maps)
	if err := rooms.SaveRoomTemplate(*room); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save room: "+err.Error())
		return
	}

	// Reload to get the fresh state
	updated := rooms.LoadRoom(roomId)
	writeJSON(w, http.StatusOK, updated)
}

func handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

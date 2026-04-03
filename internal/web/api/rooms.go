package api

import (
	"net/http"
	"sort"
	"strconv"

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

func handleUpdateRoom(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

func handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not yet implemented")
}

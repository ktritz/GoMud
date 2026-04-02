package rooms

import (
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// Test helpers for registering rooms in the global room manager.
// Used by the test harness package for integration testing.

// RegisterTestRoom adds a room to the global room manager cache.
func RegisterTestRoom(room *Room) {
	roomManager.rooms[room.RoomId] = room
}

// DeregisterTestRoom removes a room from the global room manager cache.
func DeregisterTestRoom(roomId int) {
	delete(roomManager.rooms, roomId)
	delete(roomManager.roomsWithUsers, roomId)
	delete(roomManager.roomsWithMobs, roomId)
}

// ClearAllTestRooms removes all rooms from the global room manager cache.
func ClearAllTestRooms() {
	roomManager.rooms = make(map[int]*Room)
	roomManager.roomsWithUsers = make(map[int]int)
	roomManager.roomsWithMobs = make(map[int]int)
	roomManager.roomIdToFileCache = make(map[int]string)
}

// InitTestBiomes sets up a minimal biome registry for testing.
func InitTestBiomes() {
	biomes = map[string]*BiomeInfo{
		"default": {Name: "default"},
		"city":    {Name: "city"},
		"dungeon": {Name: "dungeon", DarkArea: true},
	}
}

// NewTestRoom creates a minimal Room struct for testing.
func NewTestRoom(roomId int, zone string, title string) *Room {
	r := &Room{
		RoomId:  roomId,
		Zone:    zone,
		Title:   title,
		players: []int{},
		mobs:    []int{},
		Items:   []items.Item{},
		Nouns:   map[string]string{},
	}
	r.Exits = make(map[string]exit.RoomExit)
	return r
}

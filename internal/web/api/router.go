package api

import (
	"net/http"
)

// RegisterRoutes mounts all API routes on the default HTTP mux.
// Auth endpoints are public; all /api/admin/* endpoints require JWT.
func RegisterRoutes() {

	// Auth endpoints (no JWT required)
	http.Handle("POST /api/auth/login", cors(http.HandlerFunc(handleLogin)))
	http.Handle("POST /api/auth/refresh", cors(jwtAuth(http.HandlerFunc(handleRefresh))))
	http.Handle("GET /api/auth/me", cors(jwtAuth(http.HandlerFunc(handleMe))))

	// Admin API endpoints (JWT required)
	// Rooms
	http.Handle("GET /api/admin/rooms", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListRooms)))))
	http.Handle("GET /api/admin/rooms/{roomId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetRoom)))))
	http.Handle("POST /api/admin/rooms", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleCreateRoom)))))
	http.Handle("PUT /api/admin/rooms/{roomId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleUpdateRoom)))))
	http.Handle("DELETE /api/admin/rooms/{roomId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleDeleteRoom)))))

	// Zones
	http.Handle("GET /api/admin/zones", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListZones)))))
	http.Handle("GET /api/admin/zones/{zoneName}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetZone)))))

	// Mobs
	http.Handle("GET /api/admin/mobs", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListMobs)))))
	http.Handle("GET /api/admin/mobs/{mobId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetMob)))))
	http.Handle("POST /api/admin/mobs", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleCreateMob)))))
	http.Handle("PUT /api/admin/mobs/{mobId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleUpdateMob)))))
	http.Handle("DELETE /api/admin/mobs/{mobId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleDeleteMob)))))

	// Items
	http.Handle("GET /api/admin/items", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListItems)))))
	http.Handle("GET /api/admin/items/{itemId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetItem)))))
	http.Handle("POST /api/admin/items", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleCreateItem)))))
	http.Handle("PUT /api/admin/items/{itemId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleUpdateItem)))))
	http.Handle("DELETE /api/admin/items/{itemId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleDeleteItem)))))

	// Quests
	http.Handle("GET /api/admin/quests", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListQuests)))))
	http.Handle("GET /api/admin/quests/{questId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetQuest)))))
	http.Handle("POST /api/admin/quests", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleCreateQuest)))))
	http.Handle("PUT /api/admin/quests/{questId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleUpdateQuest)))))
	http.Handle("DELETE /api/admin/quests/{questId}", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleDeleteQuest)))))

	// Buffs
	http.Handle("GET /api/admin/buffs", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListBuffs)))))
	http.Handle("GET /api/admin/buffs/{buffId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetBuff)))))

	// Spells
	http.Handle("GET /api/admin/spells", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListSpells)))))
	http.Handle("GET /api/admin/spells/{spellId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetSpell)))))

	// Races
	http.Handle("GET /api/admin/races", cors(jwtAuth(http.HandlerFunc(withReadLock(handleListRaces)))))
	http.Handle("GET /api/admin/races/{raceId}", cors(jwtAuth(http.HandlerFunc(withReadLock(handleGetRace)))))

	// Server
	http.Handle("POST /api/admin/reload", cors(jwtAuth(http.HandlerFunc(withWriteLock(handleReload)))))
	http.Handle("GET /api/admin/stats", cors(jwtAuth(http.HandlerFunc(withReadLock(handleStats)))))
}

package targeting

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// FindAutoTarget finds who is currently attacking the given actor.
// Provide either selfUserId (for players) or selfMobInstanceId (for mobs), set the other to 0.
func FindAutoTarget(room *rooms.Room, selfUserId int, selfMobInstanceId int) (playerId int, mobInstanceId int) {
	if selfUserId > 0 {
		return findAutoTargetForPlayer(room, selfUserId)
	}
	return findAutoTargetForMob(room, selfMobInstanceId)
}

func findAutoTargetForPlayer(room *rooms.Room, selfUserId int) (playerId int, mobInstanceId int) {
	// Check mobs fighting this player
	for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
		m := mobs.GetInstance(mId)
		if m.Character.Aggro != nil && m.Character.Aggro.UserId == selfUserId {
			return 0, m.InstanceId
		}
	}

	// Check players fighting this player
	for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
		u := users.GetByUserId(uId)
		if u.Character.Aggro != nil && u.Character.Aggro.UserId == selfUserId {
			return u.UserId, 0
		}
	}

	return 0, 0
}

func findAutoTargetForMob(room *rooms.Room, selfMobInstanceId int) (playerId int, mobInstanceId int) {
	// Check mobs fighting this mob
	for _, mId := range room.GetMobs(rooms.FindFightingMob) {
		m := mobs.GetInstance(mId)
		if m.Character.Aggro != nil && m.Character.Aggro.MobInstanceId == selfMobInstanceId {
			return 0, m.InstanceId
		}
	}

	// Check players fighting this mob
	for _, uId := range room.GetPlayers(rooms.FindFightingMob) {
		u := users.GetByUserId(uId)
		if u.Character.Aggro != nil && u.Character.Aggro.MobInstanceId == selfMobInstanceId {
			return u.UserId, 0
		}
	}

	return 0, 0
}

// FindRandomTarget selects a random target in the room.
// rest should be "*" (anyone), "*mob" (any mob), or "*user" / anything else (any player).
// excludeUserId and excludeMobInstanceId prevent self-targeting.
func FindRandomTarget(room *rooms.Room, rest string, excludeUserId int, excludeMobInstanceId int) (playerId int, mobInstanceId int) {
	if rest == `*` {
		// Anyone
		allMobs := []int{}
		for _, mId := range room.GetMobs() {
			if mId != excludeMobInstanceId {
				allMobs = append(allMobs, mId)
			}
		}
		allPlayers := []int{}
		for _, uId := range room.GetPlayers() {
			if uId != excludeUserId {
				allPlayers = append(allPlayers, uId)
			}
		}

		total := len(allMobs) + len(allPlayers)
		if total == 0 {
			return 0, 0
		}

		pick := util.Rand(total)
		if pick < len(allMobs) {
			return 0, allMobs[pick]
		}
		return allPlayers[pick-len(allMobs)], 0

	} else if rest == `*mob` {
		allMobs := []int{}
		for _, mId := range room.GetMobs() {
			if mId != excludeMobInstanceId {
				allMobs = append(allMobs, mId)
			}
		}
		if len(allMobs) > 0 {
			return 0, allMobs[util.Rand(len(allMobs))]
		}
		return 0, 0

	} else {
		// *user or anything else — random player
		allPlayers := []int{}
		for _, uId := range room.GetPlayers() {
			if uId != excludeUserId {
				allPlayers = append(allPlayers, uId)
			}
		}
		if len(allPlayers) > 0 {
			return allPlayers[util.Rand(len(allPlayers))], 0
		}
		return 0, 0
	}
}

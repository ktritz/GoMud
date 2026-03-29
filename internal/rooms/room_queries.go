package rooms

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func (r *Room) MobCt() int {

	return len(r.mobs)
}

func (r *Room) PlayerCt() int {

	return len(r.players)
}

func (r *Room) IsCalm() bool {
	return !r.ArePlayersAttacking(0) && !r.AreMobsAttacking(0)
}

func (r *Room) ArePlayersAttacking(userId int) bool {

	for _, playerId := range r.players {
		if playerId == userId {
			continue
		}
		if u := users.GetByUserId(playerId); u != nil {
			if u.Character.Aggro != nil && (userId == 0 || u.Character.Aggro.UserId == userId) {
				return true
			}
		}
	}

	return false
}

func (r *Room) AreMobsAttacking(userId int) bool {

	for _, mobId := range r.mobs {
		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}
		if mob.Character.Aggro != nil && (userId == 0 || mob.Character.Aggro.UserId == userId) {
			return true
		}
	}
	return false
}

func (r *Room) GetExitInfo(exitName string) (exitInfo exit.RoomExit, ok bool) {

	// Do mutators first to allow for ephemeral/temporary "taking over" of exits.
	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		if exitInfo, ok = spec.Exits[exitName]; ok {
			break
		}
	}

	if !ok {
		exitInfo, ok = r.Exits[exitName]
	}

	return exitInfo, ok
}

func (r *Room) GetRandomExit() (exitName string, roomId int) {

	allExits := map[string]int{}

	for exitName, exit := range r.Exits {
		if exit.Secret {
			continue
		}
		if exit.Lock.IsLocked() {
			continue
		}

		allExits[exitName] = exit.RoomId
	}

	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		for exitName, exit := range spec.Exits {
			if exit.Secret {
				continue
			}
			if exit.Lock.IsLocked() {
				continue
			}
			allExits[exitName] = exit.RoomId
		}
	}

	roomSelection := util.Rand(len(allExits))

	for exitName, roomId := range allExits {
		if roomSelection == 0 {
			return exitName, roomId
		}
		roomSelection--
	}

	return ``, 0
}

func (r *Room) GetAllFloorItems(stash bool) []items.Item {

	found := []items.Item{}

	if stash {
		found = append(found, r.Stash...)
	}

	found = append(found, r.Items...)

	return found
}

func (r *Room) FindCorpse(searchName string) (Corpse, bool) {

	// First search for player corpses that match

	playerCorpseLookup := map[string]int{}
	playerCorpses := []string{}

	mobCorpseLookup := map[string]int{}
	mobCorpses := []string{}

	for idx, c := range r.Corpses {

		if c.Prunable {
			continue
		}

		if c.UserId > 0 {
			name := c.Character.Name + ` corpse`
			if _, ok := playerCorpseLookup[name]; !ok {
				playerCorpseLookup[name] = idx
				playerCorpses = append(playerCorpses, name)
			}
		}

		if c.MobId > 0 {
			name := c.Character.Name + ` corpse`
			if _, ok := mobCorpseLookup[name]; !ok {
				mobCorpseLookup[name] = idx
				mobCorpses = append(mobCorpses, name)
			}
		}
	}

	userMatch, closeUserMatch := util.FindMatchIn(searchName, playerCorpses...)
	if userMatch != `` {
		return r.Corpses[playerCorpseLookup[userMatch]], true
	}

	mobMatch, closeMobMatch := util.FindMatchIn(searchName, mobCorpses...)
	if mobMatch != `` {
		return r.Corpses[mobCorpseLookup[mobMatch]], true
	}

	if closeUserMatch != `` {
		return r.Corpses[playerCorpseLookup[closeUserMatch]], true
	} else if closeMobMatch != `` {
		return r.Corpses[mobCorpseLookup[closeMobMatch]], true
	}

	return Corpse{}, false
}

func (r *Room) FindOnFloor(itemName string, stash bool) (items.Item, bool) {

	if stash {
		// search the stash
		closeMatchItem, matchItem := items.FindMatchIn(itemName, r.Stash...)

		if matchItem.ItemId != 0 {
			return matchItem, true
		}

		if closeMatchItem.ItemId != 0 {
			return closeMatchItem, true
		}

		return items.Item{}, false
	}

	// Search floor
	closeMatchItem, matchItem := items.FindMatchIn(itemName, r.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}

func (r *Room) FindByName(searchName string, findTypes ...FindFlag) (playerId int, mobInstanceId int) {
	if len(findTypes) < 1 {
		findTypes = []FindFlag{FindAll}
	}
	mobInstanceId, _ = r.findMobByName(searchName, findTypes...)
	playerId, _ = r.findPlayerByName(searchName, findTypes...)
	return playerId, mobInstanceId
}

func (r *Room) FindByPetName(searchName string) (playerId int) {
	// Map name to display name
	petOwners := map[string]int{}
	petNames := []string{}

	for _, uId := range r.GetPlayers(FindHasPet) {
		if u := users.GetByUserId(uId); u != nil {
			petOwners[u.Character.Pet.Name] = u.UserId
			petNames = append(petNames, u.Character.Pet.Name)
		}
	}

	match, closeMatch := util.FindMatchIn(searchName, petNames...)
	if match == `` {
		if closeMatch == `` {
			return 0
		}
		return petOwners[closeMatch]
	}

	return petOwners[match]
}

func (r *Room) findPlayerByName(searchName string, findTypes ...FindFlag) (int, error) {

	if len(searchName) > 1 {
		if searchName[0] == '#' {
			return 0, errors.New("user not found")
		}
		if searchName[0] == '@' {
			userIdMatch, _ := strconv.Atoi(searchName[1:])

			for _, uId := range r.GetPlayers(findTypes...) {

				if userIdMatch > 0 {
					if uId != userIdMatch {
						continue
					}
					return uId, nil
				}
			}
			return 0, errors.New("user not found")
		}
	}

	namesInRoom := []string{}
	// are they looking at a player?
	playerLookup := map[string]int{}
	for _, uId := range r.GetPlayers(findTypes...) {
		u := users.GetByUserId(uId)
		playerLookup[u.Character.Name] = u.UserId
		namesInRoom = append(namesInRoom, u.Character.Name)
	}

	closeMatch, fullMatch := util.FindMatchIn(searchName, namesInRoom...)

	if len(fullMatch) == 0 {
		fullMatch = closeMatch
	}

	if len(fullMatch) == 0 {
		return 0, errors.New("player not found")
	}

	return playerLookup[fullMatch], nil
}

func (r *Room) findMobByName(searchName string, findTypes ...FindFlag) (int, error) {

	if len(searchName) > 1 {
		if searchName[0] == '@' {
			return 0, errors.New("mob not found")
		}
		if searchName[0] == '#' {
			mobIdMatch, _ := strconv.Atoi(searchName[1:])

			for _, mId := range r.GetMobs(findTypes...) {

				if mobIdMatch > 0 {
					if mId != mobIdMatch {
						continue
					}
					return mId, nil
				}
			}
			return 0, errors.New("mob not found")
		}
	}

	namesInRoom := []string{}
	friendlyMobs := map[int]*mobs.Mob{}
	mobLookup := map[string]int{}
	for _, mId := range r.GetMobs(findTypes...) {

		m := mobs.GetInstance(mId)

		if m.Character.IsCharmed() {
			friendlyMobs[mId] = m // Put friendly mobs at the end of the list.
			continue
		}

		mobName := fmt.Sprintf(`%s#%d`, m.Character.Name, len(namesInRoom)+1) // skeleton#1, skeleton#2 etc
		mobLookup[mobName] = mId
		namesInRoom = append(namesInRoom, mobName)

	}

	// Now add the friendly mobs (at the end)
	for mId, m := range friendlyMobs {
		mobName := fmt.Sprintf(`%s#%d`, m.Character.Name, len(namesInRoom)+1)
		mobLookup[mobName] = mId
		namesInRoom = append(namesInRoom, mobName)
		delete(friendlyMobs, mId)
	}

	closeMatch, fullMatch := util.FindMatchIn(searchName, namesInRoom...)

	if len(fullMatch) == 0 {
		fullMatch = closeMatch
	}

	if len(fullMatch) == 0 {
		return 0, errors.New("mob not found")
	}

	return mobLookup[fullMatch], nil

}

// Returns exitName, RoomExit
func (r *Room) FindExitTo(roomId int) string {

	for exitName, exit := range r.Exits {
		if exit.RoomId == roomId {
			return exitName
		}
	}

	for _, exit := range r.ExitsTemp {
		if exit.RoomId == roomId {
			return exit.Title
		}
	}

	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		for exitName, exit := range spec.Exits {
			if exit.RoomId == roomId {
				return exitName
			}
		}
	}

	return ""
}

func (r *Room) FindContainerByName(containerNameSearch string) string {

	if len(r.Containers) == 0 {
		return ``
	}

	containerNames := []string{}
	for containerName, _ := range r.Containers {
		containerNames = append(containerNames, containerName)
	}

	exactMatch, closeMatch := util.FindMatchIn(containerNameSearch, containerNames...)

	if len(exactMatch) > 0 {
		return exactMatch
	}

	return closeMatch
}

func (r *Room) FindNoun(noun string) (foundNoun string, nounDescription string) {
	if len(r.Nouns) == 0 {
		return "", ""
	}

	// Flatten the room nouns and create single-word aliases for multi-word nouns
	roomNouns := map[string]string{}
	for originalNoun, originalDesc := range r.Nouns {
		roomNouns[originalNoun] = originalDesc
		if strings.Contains(originalNoun, " ") {
			for _, part := range strings.Split(originalNoun, " ") {
				if _, exists := r.Nouns[part]; exists {
					continue
				}
				if _, exists := roomNouns[part]; exists {
					continue
				}
				roomNouns[part] = ":" + originalNoun
			}
		}
	}

	// Build candidate noun list
	testNouns := util.SplitButRespectQuotes(noun)
	for i := 0; i < len(testNouns); i++ {
		if strings.Contains(testNouns[i], " ") {
			for _, part := range strings.Split(testNouns[i], " ") {
				testNouns = append(testNouns, strings.ToLower(strings.TrimSpace(part)))
			}
		}
	}
	if len(testNouns) > 1 {
		testNouns = append(testNouns, strings.ToLower(strings.TrimSpace(noun)))
	}

	// Try each candidate: exact, singular/plural, alias-aware
	for _, cand := range testNouns {
		newNoun := strings.ToLower(strings.TrimSpace(cand))

		// Direct match or single-level alias
		if desc, ok := roomNouns[newNoun]; ok {
			if strings.HasPrefix(desc, ":") {
				target := desc[1:]
				if targetDesc, ok2 := roomNouns[target]; ok2 && !strings.HasPrefix(targetDesc, ":") {
					return target, targetDesc
				}
				// alias->alias or missing target => ignore
			} else {
				return newNoun, desc
			}
		}

		// Strip "es"
		if strings.HasSuffix(newNoun, "es") {
			tn := strings.TrimSuffix(newNoun, "es")
			if desc, ok := roomNouns[tn]; ok {
				if strings.HasPrefix(desc, ":") {
					target := desc[1:]
					if targetDesc, ok2 := roomNouns[target]; ok2 && !strings.HasPrefix(targetDesc, ":") {
						return target, targetDesc
					}
				} else {
					return tn, desc
				}
			}
		} else {
			// Add "es"
			tn := newNoun + "es"
			if desc, ok := roomNouns[tn]; ok {
				if strings.HasPrefix(desc, ":") {
					target := desc[1:]
					if targetDesc, ok2 := roomNouns[target]; ok2 && !strings.HasPrefix(targetDesc, ":") {
						return target, targetDesc
					}
				} else {
					return tn, desc
				}
			}
		}

		// "ies" -> "y"
		if strings.HasSuffix(newNoun, "ies") {
			tn := strings.TrimSuffix(newNoun, "ies") + "y"
			if desc, ok := roomNouns[tn]; ok {
				if strings.HasPrefix(desc, ":") {
					target := desc[1:]
					if targetDesc, ok2 := roomNouns[target]; ok2 && !strings.HasPrefix(targetDesc, ":") {
						return target, targetDesc
					}
				} else {
					return tn, desc
				}
			}
		}
	}

	// Multi-word noun match
	for full, desc := range roomNouns {
		if strings.Contains(full, " ") {
			for _, part := range testNouns {
				if strings.Contains(full, part) {
					if strings.HasPrefix(desc, ":") {
						target := desc[1:]
						if td, ok := roomNouns[target]; ok && !strings.HasPrefix(td, ":") {
							return target, td
						}
					} else {
						return full, desc
					}
				}
			}
		}
	}

	// Single-word match for multi-word nouns
	for full, desc := range roomNouns {
		if !strings.Contains(full, " ") {
			for _, part := range testNouns {
				if part == full {
					if strings.HasPrefix(desc, ":") {
						target := desc[1:]
						if td, ok := roomNouns[target]; ok && !strings.HasPrefix(td, ":") {
							return target, td
						}
					} else {
						return full, desc
					}
				}
			}
		}
	}

	return "", ""
}

func (r *Room) FindExitByName(exitNameSearch string) (exitName string, exitRoomId int) {

	// Check for direction aliases from keywords.yaml first
	fullDirection := keywords.TryDirectionAlias(exitNameSearch)
	if fullDirection != exitNameSearch {
		// A direction alias was found, check if this exact direction exists
		if exitInfo, ok := r.Exits[fullDirection]; ok {
			return fullDirection, exitInfo.RoomId
		}
		// Check temporary exits
		if tempExit, ok := r.ExitsTemp[fullDirection]; ok {
			return fullDirection, tempExit.RoomId
		}
		// Check mutator exits
		for mut := range r.ActiveMutators {
			spec := mut.GetSpec()
			if exitInfo, ok := spec.Exits[fullDirection]; ok {
				return fullDirection, exitInfo.RoomId
			}
		}
		// Direction alias used but exit doesn't exist
		return ``, 0
	}

	// Build list of all exits for fuzzy matching
	exitNames := []string{}
	for exitName, _ := range r.Exits {
		exitNames = append(exitNames, exitName)
	}

	for exitName, _ := range r.ExitsTemp {
		exitNames = append(exitNames, exitName)
	}

	mutatorExits := map[string]exit.RoomExit{}
	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		for exitName, exitInfo := range spec.Exits {
			mutatorExits[exitName] = exitInfo
			exitNames = append(exitNames, exitName)
		}
	}

	// Use fuzzy matching for all exits
	exactMatch, closeMatch := util.FindMatchIn(exitNameSearch, exitNames...)

	if len(exactMatch) == 0 {
		portalStr := `portal`
		if strings.HasPrefix(closeMatch, exitNameSearch) {
			exactMatch = closeMatch
		} else if strings.Contains(closeMatch, portalStr) { // If has portal in the word, lets consider a partial match on "portal"
			if exitNameSearch == portalStr {
				exactMatch = closeMatch
			} else { // partial starting match on "portal"?
				searchLen := len(exitNameSearch)
				if searchLen <= len(portalStr) {
					if portalStr[:searchLen] == exitNameSearch {
						exactMatch = closeMatch
					}
				}
			}
		}
	}
	if len(closeMatch) == 0 {
		return "", 0
	}

	if exitInfo, ok := r.Exits[exactMatch]; ok {
		return exactMatch, exitInfo.RoomId
	}

	if exitInfo, ok := r.ExitsTemp[exactMatch]; ok {
		return exitInfo.Title, exitInfo.RoomId
	}

	if exitInfo, ok := mutatorExits[exactMatch]; ok {
		return exactMatch, exitInfo.RoomId
	}

	return "", 0
}

func (r *Room) FindTemporaryExitByUserId(userId int) (exit.TemporaryRoomExit, bool) {

	if r.ExitsTemp != nil {
		for _, v := range r.ExitsTemp {
			if v.UserId == userId {
				return v, true
			}
		}
	}

	return exit.TemporaryRoomExit{}, false
}

func (r *Room) GetMobs(findTypes ...FindFlag) []int {

	mobMatches := []int{}
	if len(r.mobs) == 0 {
		return mobMatches
	}

	var typeFlag FindFlag = 0
	if len(findTypes) < 1 {
		typeFlag = FindAll
	} else {
		for _, ff := range findTypes {
			typeFlag |= ff
		}
	}

	// If no filtering, just copy all mobs in the room and return it
	if typeFlag == FindAll {
		return append([]int{}, r.mobs...)
	}

	var isCharmed bool = false

	for _, mobId := range r.mobs {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if typeFlag == FindAll {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if mob.Character.Aggro != nil {
			if typeFlag&FindFightingPlayer == FindFightingPlayer && mob.Character.Aggro.UserId != 0 {
				mobMatches = append(mobMatches, mobId)
				continue
			}
			if typeFlag&FindFightingMob == FindFightingMob && mob.Character.Aggro.MobInstanceId != 0 {
				mobMatches = append(mobMatches, mobId)
				continue
			}
		}

		if typeFlag&FindNative == FindNative {
			if mob.HomeRoomId == r.RoomId {
				mobMatches = append(mobMatches, mobId)
				continue
			}
			// If not native, and that was all we were looking for, abort further tests
			if typeFlag == FindNative {
				continue
			}
		}

		if typeFlag&FindHasLight == FindHasLight && mob.Character.HasBuffFlag(buffs.EmitsLight) {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		// Useful to find any mobs that will always attack players
		if mob.Hostile && typeFlag&FindHostile == FindHostile {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		isCharmed = mob.Character.IsCharmed()

		if isCharmed && typeFlag&FindCharmed == FindCharmed {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		// If not allied with players
		// and not current aggressive to anything
		// and won't automatically attack players
		if typeFlag&FindNeutral == FindNeutral && !isCharmed && mob.Character.Aggro == nil && !mob.Hostile {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindMerchant == FindMerchant && mob.HasShop() {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindDowned == FindDowned && mob.Character.Health < 1 {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindBuffed == FindBuffed && len(mob.Character.Buffs.List) > 0 {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindHasPet == FindHasPet && mob.Character.Pet.Exists() {
			mobMatches = append(mobMatches, mobId)
			continue
		}
	}

	return mobMatches
}

func (r *Room) GetPlayers(findTypes ...FindFlag) []int {

	playerMatches := []int{}
	if len(r.players) == 0 {
		return playerMatches
	}

	var typeFlag FindFlag = 0
	if len(findTypes) < 1 {
		typeFlag = FindAll
	} else {
		for _, ff := range findTypes {
			typeFlag |= ff
		}
	}

	// If no filtering, just copy all mobs in the room and return it
	if typeFlag == FindAll {
		return append([]int{}, r.players...)
	}

	var isCharmed bool = false

	for _, userId := range r.players {

		user := users.GetByUserId(userId)
		if user == nil {
			continue
		}

		if typeFlag == FindAll {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if user.Character.Aggro != nil {
			if typeFlag&FindFightingPlayer == FindFightingPlayer && user.Character.Aggro.UserId != 0 {
				playerMatches = append(playerMatches, userId)
				continue
			}
			if typeFlag&FindFightingMob == FindFightingMob && user.Character.Aggro.MobInstanceId != 0 {
				playerMatches = append(playerMatches, userId)
				continue
			}
		}

		if typeFlag&FindHasLight == FindHasLight && user.Character.HasBuffFlag(buffs.EmitsLight) {
			playerMatches = append(playerMatches, userId)
			continue
		}

		isCharmed = user.Character.IsCharmed()

		if isCharmed && typeFlag&FindCharmed == FindCharmed {
			playerMatches = append(playerMatches, userId)
			continue
		}

		// If not allied with players
		// and not current aggressive to anything
		// and won't automatically attack players
		if typeFlag&FindNeutral == FindNeutral && !isCharmed && user.Character.Aggro == nil {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindMerchant == FindMerchant && user.HasShop() {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindDowned == FindDowned && user.Character.Health < 1 {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindBuffed == FindBuffed && len(user.Character.Buffs.List) > 0 {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindHasPet == FindHasPet && user.Character.Pet.Exists() {
			playerMatches = append(playerMatches, userId)
			continue
		}
	}

	return playerMatches
}

func (r *Room) isInRoom(mobName string, userName string) bool {

	if mobName != `` {
		for _, mobInstId := range r.mobs {
			if mob := mobs.GetInstance(mobInstId); mob != nil {
				if strings.HasPrefix(mob.Character.Name, mobName) {
					return true
				}
			}
		}
	}

	if userName != `` {
		for _, userId := range r.players {
			if user := users.GetByUserId(userId); user != nil {
				if strings.HasPrefix(user.Character.Name, userName) {
					return true
				}
			}
		}
	}

	return false

}

func (r *Room) findMobExit(mobId int, mobName string) string {

	freshestTime := float64(0)
	freshestExitName := ``

	for exitName, exitInfo := range r.Exits {

		// Skip secret exits
		if exitInfo.Secret {
			continue
		}

		exitRoom := LoadRoom(exitInfo.RoomId)
		if exitRoom == nil {
			continue
		}

		for mId, timeLeft := range exitRoom.Visitors(VisitorMob) {

			if mobId > 0 && mobId != mId {
				continue
			}

			if visitorMob := mobs.GetInstance(mId); visitorMob != nil {

				if len(mobName) > 0 && !strings.HasPrefix(visitorMob.Character.Name, mobName) {
					continue
				}

				if timeLeft > freshestTime {
					freshestTime = timeLeft
					freshestExitName = exitName
				}

			}

		}

	}

	return freshestExitName

}

func (r *Room) findUserExit(userId int, userName string) string {

	freshestTime := float64(0)
	freshestExitName := ``

	for exitName, exitInfo := range r.Exits {

		// Skip secret exits
		if exitInfo.Secret {
			continue
		}

		exitRoom := LoadRoom(exitInfo.RoomId)
		if exitRoom == nil {
			continue
		}

		for uId, timeLeft := range exitRoom.Visitors(VisitorMob) {

			if userId > 0 && userId != uId {
				continue
			}

			if visitorUser := users.GetByUserId(uId); visitorUser != nil {

				if len(userName) > 0 && !strings.HasPrefix(visitorUser.Character.Name, userName) {
					continue
				}

				if timeLeft > freshestTime {
					freshestTime = timeLeft
					freshestExitName = exitName
				}

			}

		}

	}

	return freshestExitName

}

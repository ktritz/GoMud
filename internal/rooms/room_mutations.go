package rooms

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func (r *Room) AddCorpse(c Corpse) {
	r.Corpses = append(r.Corpses, c)
}

func (r *Room) RemoveCorpse(c Corpse) bool {
	for idx, corpse := range r.Corpses {
		if corpse.MobId != c.MobId {
			continue
		}
		if corpse.UserId != c.UserId {
			continue
		}
		if corpse.Character.Name != c.Character.Name {
			continue
		}
		if corpse.RoundCreated != c.RoundCreated {
			continue
		}

		r.Corpses = append(r.Corpses[:idx], r.Corpses[idx+1:]...)

		return true
	}
	return false
}

func (r *Room) UpdateCorpses(roundNow uint64) {

	c := configs.GetGamePlayConfig()

	if !c.Death.CorpsesEnabled {
		return
	}

	removeIdx := []int{}
	for idx, corpse := range r.Corpses {
		corpse.Update(roundNow, c.Death.CorpseDecayTime.String())
		if corpse.Prunable {
			removeIdx = append(removeIdx, idx)
			if corpse.MobId > 0 {
				r.SendText(fmt.Sprintf(`A <ansi fg="mob-corpse">%s corpse</ansi> crumbles to dust.`, corpse.Character.Name))
			}
			if corpse.UserId > 0 {
				r.SendText(fmt.Sprintf(`A <ansi fg="user-corpse">%s corpse</ansi> crumbles to dust.`, corpse.Character.Name))
			}
		}
		r.Corpses[idx] = corpse
	}

	for i := len(removeIdx) - 1; i >= 0; i-- {
		r.Corpses = append(r.Corpses[:removeIdx[i]], r.Corpses[removeIdx[i]+1:]...)
	}
}

func (r *Room) AddMob(mobInstanceId int) {

	mob := mobs.GetInstance(mobInstanceId)
	if mob == nil {
		return
	}

	r.MarkVisited(mobInstanceId, VisitorMob)

	events.AddToQueue(events.RoomChange{
		MobInstanceId: mobInstanceId,
		FromRoomId:    mob.Character.RoomId,
		ToRoomId:      r.RoomId,
		Unseen:        mob.Character.HasBuffFlag(buffs.Hidden),
	})

	mob.Character.RoomId = r.RoomId
	mob.Character.Zone = r.Zone

	r.mobs = append(r.mobs, mobInstanceId)

	roomManager.roomsWithMobs[r.RoomId] = len(r.mobs)
}

func (r *Room) RemoveMob(mobInstanceId int) {

	r.MarkVisited(mobInstanceId, VisitorMob, 1)

	mobLen := len(r.mobs)
	for i := 0; i < mobLen; i++ {
		if r.mobs[i] == mobInstanceId {
			r.mobs = append(r.mobs[:i], r.mobs[i+1:]...)
			break
		}
	}

	if len(r.mobs) < 1 {
		delete(roomManager.roomsWithMobs, r.RoomId)
	}
}

func (r *Room) AddItem(item items.Item, stash bool) {

	item.Validate()

	if stash {
		r.Stash = append(r.Stash, item)
	} else {
		r.Items = append(r.Items, item)
	}

}

func (r *Room) RemoveItem(i items.Item, stash bool) {

	if stash {
		for j := len(r.Stash) - 1; j >= 0; j-- {
			if r.Stash[j].Equals(i) {
				r.Stash = append(r.Stash[:j], r.Stash[j+1:]...)
				break
			}
		}
	} else {
		for j := len(r.Items) - 1; j >= 0; j-- {
			if r.Items[j].Equals(i) {
				r.Items = append(r.Items[:j], r.Items[j+1:]...)
				break
			}
		}
	}

}

func (r *Room) SetExitLock(exitName string, locked bool) {

	if exitInfo, ok := r.Exits[exitName]; ok {
		if !exitInfo.HasLock() {
			return
		}
		if locked {
			exitInfo.Lock.SetLocked()
		} else {
			exitInfo.Lock.SetUnlocked()
		}
		r.Exits[exitName] = exitInfo

	} else {
		for mut := range r.ActiveMutators {
			spec := mut.GetSpec()
			if exitInfo, ok = spec.Exits[exitName]; ok {
				if !exitInfo.HasLock() {
					continue
				}
				if locked {
					exitInfo.Lock.SetLocked()
				} else {
					exitInfo.Lock.SetUnlocked()
				}
				spec.Exits[exitName] = exitInfo
			}
		}
	}

}

func (r *Room) AddPlayer(userId int) int {

	for _, v := range r.players {
		if v == userId {
			return len(r.players)
		}
	}

	r.players = append(r.players, userId)

	return len(r.players)
}

// true if found
func (r *Room) RemovePlayer(userId int) (int, bool) {

	for i, v := range r.players {
		if v == userId {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return len(r.players), true
		}
	}
	return len(r.players), false
}

func (r *Room) RemoveTemporaryExit(t exit.TemporaryRoomExit) bool {

	if r.ExitsTemp == nil {
		return false
	}

	for k, v := range r.ExitsTemp {
		if v.UserId == t.UserId && v.Title == t.Title && t.RoomId == v.RoomId {
			delete(r.ExitsTemp, k)
			return true
		}
	}

	return false
}

// Can't add twoof the same exitName
// Will return false if it already exists
func (r *Room) AddTemporaryExit(exitName string, t exit.TemporaryRoomExit) bool {

	t.SpawnedRound = util.GetRoundCount()

	if r.ExitsTemp == nil {
		r.ExitsTemp = make(map[string]exit.TemporaryRoomExit)
	}

	if len(t.Title) == 0 {
		t.Title = exitName
	}
	if _, ok := r.ExitsTemp[exitName]; ok {
		return false
	}
	r.ExitsTemp[exitName] = t
	return true
}

func (r *Room) PruneTemporaryExits() []exit.TemporaryRoomExit {

	rNow := util.GetRoundCount()

	prunedExits := []exit.TemporaryRoomExit{}

	for k, v := range r.ExitsTemp {
		g := gametime.GetDate(v.SpawnedRound)
		if rNow >= g.AddPeriod(v.Expires) {
			delete(r.ExitsTemp, k)
			prunedExits = append(prunedExits, v)
		}
	}
	return prunedExits
}

func (r *Room) PruneSigns() []Sign {

	prunedSigned := []Sign{}

	signCt := len(r.Signs)
	if signCt == 0 {
		return prunedSigned
	}

	for i := signCt - 1; i >= 0; i-- {
		s := r.Signs[i]
		if s.Expires.Before(time.Now()) {
			r.Signs = append(r.Signs[:i], r.Signs[i+1:]...)
			prunedSigned = append(prunedSigned, s)
		}
	}

	return prunedSigned
}

func (r *Room) GetPublicSigns() []Sign {

	visibleSigns := []Sign{}
	for _, sign := range r.Signs {
		if sign.VisibleUserId == 0 {
			visibleSigns = append(visibleSigns, sign)
		}
	}

	return visibleSigns
}

func (r *Room) GetPrivateSigns() []Sign {

	privateSigns := []Sign{}
	for _, sign := range r.Signs {
		if sign.VisibleUserId != 0 {
			privateSigns = append(privateSigns, sign)
		}
	}

	return privateSigns
}

// Returns true if a sign was replaced
func (r *Room) AddSign(displayText string, visibleUserId int, daysBeforeDecay int) bool {

	s := Sign{
		VisibleUserId: visibleUserId,
		DisplayText:   displayText,
		Expires:       time.Now().Add(time.Hour * 24 * time.Duration(daysBeforeDecay)),
	}

	// If it's a public sign and one exists, replace it.
	// If it's a private rune and one exists for this player, replace it.
	for i, sign := range r.Signs {
		if sign.VisibleUserId == visibleUserId {
			r.Signs[i] = s
			return true
		}
	}

	r.Signs = append(r.Signs, s)
	return false
}

// applies buffs to any players in the room that don't
// already have it
func (r *Room) ApplyBuffIdToPlayers(buffIds []int, source string) {

	if len(buffIds) == 0 {
		return
	}

	for _, uid := range r.GetPlayers() {

		if u := users.GetByUserId(uid); u != nil {

			for _, bId := range buffIds {
				if u.Character.HasBuff(bId) {
					continue
				}
				u.AddBuff(bId, source)
			}
		}

	}

}

// applies buffs to any mobs in the room that don't
// already have it
func (r *Room) ApplyBuffIdToMobs(buffIds []int, source string) {

	if len(buffIds) == 0 {
		return
	}

	for _, miid := range r.GetMobs() {

		if m := mobs.GetInstance(miid); m != nil {

			for _, bId := range buffIds {
				if m.Character.HasBuff(bId) {
					continue
				}
				m.AddBuff(bId, source)
			}
		}

	}

}

// applies buffs to any mobs in the room that don't
// already have it
func (r *Room) ApplyBuffIdToNativeMobs(buffIds []int, source string) {

	if len(buffIds) == 0 {
		return
	}

	for _, miid := range r.GetMobs(FindNative) {

		if m := mobs.GetInstance(miid); m != nil {

			for _, bId := range buffIds {
				if m.Character.HasBuff(bId) {
					continue
				}
				m.AddBuff(bId, source)
			}
		}

	}

}

func (r *Room) SpawnTempContainer(name string, duration string, lockDifficulty int, trapBuffIds ...int) string {

	c := Container{}

	gd := gametime.GetDate(util.GetRoundCount())
	c.DespawnRound = gd.AddPeriod(duration)

	c.Lock.Difficulty = uint8(lockDifficulty)

	if len(trapBuffIds) > 0 {
		c.Lock.TrapBuffIds = trapBuffIds
	}

	containerName := name

	// make sure name is unique
	i := 1
	_, ok := r.Containers[containerName]
	for ok {
		containerName = name + `-` + strconv.Itoa(i)
		i++
		_, ok = r.Containers[containerName]
	}

	if r.Containers == nil {
		r.Containers = make(map[string]Container)
	}
	r.Containers[containerName] = c

	return containerName
}

// Spawns an item in the room unless:
// 1. Item is already in the room
// 2. (optional) Item is currently held by someone in the room
// 3. item repeat-spawned too recently
// If containerName is provided, ony that container name will be considered
func (r *Room) RepeatSpawnItem(itemId int, roundFrequency int, containerName ...string) bool {

	roundNum := util.GetRoundCount()
	spawnKey := strconv.Itoa(itemId)

	cName := ``
	if len(containerName) > 0 {
		cName = containerName[0]
		spawnKey = cName + `-` + spawnKey
	}

	// Are we detailing with a container?
	if cName != `` {

		c, ok := r.Containers[cName]

		// Container doesn't exist? Abort.
		if !ok {

			return false
		}

		// Item in the container? Abort.
		for _, item := range c.Items {
			if item.ItemId == itemId {

				return false
			}
		}

	}

	// Check if item is already in the room
	for _, item := range r.Items {
		if item.ItemId == itemId {

			return false
		}
	}

	// Check hidden as well
	for _, item := range r.Stash {
		if item.ItemId == itemId {

			return false
		}
	}

	// unlock for further processing that will require locks

	// Check whether enough time has passed since last spawn
	if lastSpawn := r.GetTempData(spawnKey); lastSpawn != nil {
		if lastSpawn.(uint64)+uint64(roundFrequency) > roundNum {
			return false
		}
	}

	// If someone is carrying it, abort
	for _, userId := range r.GetPlayers() {

		if user := users.GetByUserId(userId); user != nil {

			for _, item := range user.Character.GetAllBackpackItems() {
				if item.ItemId == itemId {
					return false
				}
			}

			for _, item := range user.Character.GetAllWornItems() {
				if item.ItemId == itemId {
					return false
				}
			}
		}
	}

	r.SetTempData(spawnKey, roundNum)

	// Create item
	itm := items.New(itemId)

	// Add to container?
	if cName != `` {

		c := r.Containers[cName]
		c.AddItem(itm)
		r.Containers[cName] = c

	} else { // Add to room

		r.AddItem(itm, false)

	}

	return true

}

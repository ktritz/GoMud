package scripting

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/pets"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func (a ScriptActor) SendText(msg string) {
	if a.userRecord == nil {
		return
	}

	msg = userTextWrap.Wrap(msg)

	a.userRecord.SendText(msg)
}

func (a ScriptActor) GetCharacterName(wrapInTags bool) string {

	if wrapInTags {
		if a.userRecord != nil {
			return `<ansi fg="username">` + a.characterRecord.Name + `</ansi>`
		} else if a.mobRecord != nil {
			return `<ansi fg="mobname">` + a.characterRecord.Name + `</ansi>`
		}
	}

	return a.characterRecord.Name
}

func (a ScriptActor) SetCharacterName(newName string) {
	a.characterRecord.Name = newName
}

func (a ScriptActor) HasQuest(questId string) bool {
	return a.characterRecord.HasQuest(questId)
}

func (a ScriptActor) GiveQuest(questId string) {

	if a.userRecord != nil {
		// If in a party, give to all party members.
		if party := parties.Get(a.userId); party != nil {
			for _, userId := range party.GetMembers() {

				events.AddToQueue(events.Quest{
					UserId:     userId,
					QuestToken: questId,
				})

			}
			return
		} else {

			events.AddToQueue(events.Quest{
				UserId:     a.userId,
				QuestToken: questId,
			})

		}
	}
	//a.characterRecord.GiveQuestToken(questId)

}

func (a ScriptActor) GetPartyMembers() []ScriptActor {

	partyMembers := []ScriptActor{}
	partyUserId := 0

	if a.userRecord == nil {
		if a.mobRecord.Character.Charmed == nil {
			return partyMembers
		}

		partyUserId = a.mobRecord.Character.Charmed.UserId
	} else {
		partyUserId = a.userId
	}

	if partyUserId < 1 {
		return partyMembers
	}

	// If in a party, give to all party members.
	if party := parties.Get(partyUserId); party != nil {
		for _, userId := range party.GetMembers() {

			if a := GetActor(userId, 0); a != nil {
				partyMembers = append(partyMembers, *a)
			}

		}
	}

	mobPartyMembers := []ScriptActor{}

	for _, char := range partyMembers {
		for _, mobInstId := range char.characterRecord.GetCharmIds() {
			if a := GetActor(0, mobInstId); a != nil {
				mobPartyMembers = append(mobPartyMembers, *a)
			}
		}
	}

	return append(partyMembers, mobPartyMembers...)
}

func (a ScriptActor) Sleep(seconds int) {
	if a.userId == 0 {
		a.mobRecord.Sleep(seconds)
	}
}

func (a ScriptActor) Command(cmd string, waitSeconds ...float64) {
	if len(waitSeconds) < 1 {
		waitSeconds = append(waitSeconds, 0)
	}
	if a.userId > 0 {
		a.userRecord.Command(cmd, waitSeconds[0])
	} else {
		a.mobRecord.Command(cmd, waitSeconds[0])
	}
}

func (a ScriptActor) CommandFlagged(cmd string, flags events.EventFlag, waitSeconds ...float64) {
	if len(waitSeconds) < 1 {
		waitSeconds = append(waitSeconds, 0)
	}
	if a.userId > 0 {
		a.userRecord.CommandFlagged(cmd, flags, waitSeconds[0])
	} else {
		a.mobRecord.Command(cmd, waitSeconds[0])
	}
}

func (a ScriptActor) MoveRoom(destRoomId int, leaveCharmedMobs ...bool) {

	if a.userRecord != nil {

		rmNow := rooms.LoadRoom(a.characterRecord.RoomId)

		if rmNext := rooms.LoadRoom(destRoomId); rmNext != nil {

			rooms.MoveToRoom(a.userId, destRoomId)

			if len(leaveCharmedMobs) < 1 || !leaveCharmedMobs[0] {
				for _, mobInstId := range a.characterRecord.GetCharmIds() {
					rmNow.RemoveMob(mobInstId)
					rmNext.AddMob(mobInstId)
				}
			}

			if doLook, err := TryRoomScriptEvent(`onEnter`, a.userRecord.UserId, destRoomId); err != nil || doLook {
				a.userRecord.CommandFlagged(`look`, events.CmdSecretly) // Do a secret look.
			}
		}

	} else if a.mobRecord != nil {

		if mobRoom := rooms.LoadRoom(a.characterRecord.RoomId); mobRoom != nil {
			if destRoom := rooms.LoadRoom(destRoomId); destRoom != nil {
				mobRoom.RemoveMob(a.mobInstanceId)
				destRoom.AddMob(a.mobInstanceId)
			}
		}

	}
}

func (a ScriptActor) AddEventLog(category string, message string) {
	if a.userRecord != nil {
		a.userRecord.EventLog.Add(category, message)
	}
}

func (a ScriptActor) IsHome() bool {
	if a.mobRecord != nil {
		return a.mobRecord.HomeRoomId == a.characterRecord.RoomId
	}
	return false
}

func (a ScriptActor) GetCharmCount() int {
	return len(a.characterRecord.GetCharmIds())
}

func (a ScriptActor) GetMaxCharmCount() int {
	return a.characterRecord.GetMaxCharmedCreatures()
}

func (a ScriptActor) GetPet() *pets.Pet {

	if a.characterRecord.Pet.Exists() {
		return &a.characterRecord.Pet
	}
	return nil
}

func (a ScriptActor) TimerSet(name string, period string) {
	a.characterRecord.TimerSet(name, period)
}

func (a ScriptActor) TimerExpired(name string) bool {
	return a.characterRecord.TimerExpired(name)
}

func (a ScriptActor) TimerExists(name string) bool {
	return a.characterRecord.TimerExists(name)
}

// ////////////////////////////////////////////////////////
//
// Functions only really useful for mobs
//
// ////////////////////////////////////////////////////////

// Returns true if a mob is charmed by/friendly to a player.
// If userId is ommitted, it will return true if the mob is charmed by any player.
func (a ScriptActor) IsCharmed(userId ...int) bool {
	if len(userId) < 1 {
		return a.characterRecord.IsCharmed()
	}
	return a.characterRecord.IsCharmed(userId[0])
}

func (a ScriptActor) GetCharmedUserId() int {
	return a.characterRecord.GetCharmedUserId()
}

func (a ScriptActor) CharmSet(userId int, charmRounds int, onRevertCommand ...string) {

	// If the player is in a party, add the mob to their party
	if a.mobInstanceId < 1 {
		return
	}

	if len(onRevertCommand) < 1 {
		onRevertCommand = append(onRevertCommand, ``)
	}
	a.characterRecord.Charm(userId, charmRounds, onRevertCommand[0])

	if user := users.GetByUserId(userId); user != nil {
		user.Character.TrackCharmed(a.mobInstanceId, true)
	}

}

func (a ScriptActor) CharmRemove() {
	if a.characterRecord.Charmed == nil {
		return
	}
	charmUserId := a.characterRecord.RemoveCharm()

	if user := users.GetByUserId(charmUserId); user != nil {
		user.Character.TrackCharmed(a.mobInstanceId, false)
	}
}

func (a ScriptActor) CharmExpire() {
	a.characterRecord.Charmed.Expire()
}

func (a ScriptActor) GetLastInputRound() uint64 {
	if a.userRecord != nil {
		return a.userRecord.GetLastInputRound()
	}
	return 0
}

func (a ScriptActor) Pathing() bool {
	if a.mobRecord == nil {
		return false
	}
	return a.mobRecord.Path.Current() != nil || a.mobRecord.Path.Len() > 0
}

func (a ScriptActor) PathingAtWaypoint() bool {
	if a.mobRecord == nil {
		return false
	}
	pathStep := a.mobRecord.Path.Current()
	return pathStep != nil && pathStep.Waypoint()
}


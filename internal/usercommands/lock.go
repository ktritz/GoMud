package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/locksmith"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Lock(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	return lockOrUnlock(rest, user, room, true)
}

func Unlock(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	return lockOrUnlock(rest, user, room, false)
}

func lockOrUnlock(rest string, user *users.UserRecord, room *rooms.Room, doLock bool) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		if doLock {
			user.SendText("Lock what?")
		} else {
			user.SendText("Unlock what?")
		}
		return true, nil
	}

	actionVerb := `unlock`
	if doLock {
		actionVerb = `lock`
	}

	containerName := room.FindContainerByName(args[0])
	exitName, _ := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if doLock && container.Lock.IsLocked() {
			user.SendText("That's already locked.")
			return true, nil
		}
		if !doLock && !container.Lock.IsLocked() {
			user.SendText("That's not locked.")
			return true, nil
		}

		lockId := locksmith.BuildLockId(room.RoomId, containerName)
		keyResult := locksmith.FindKey(user.Character, lockId, int(container.Lock.Difficulty))

		if keyResult.HasKeyRingKey {
			if doLock {
				container.Lock.SetLocked()
			} else {
				container.Lock.SetUnlocked()
			}
			room.Containers[containerName] = container
			room.PlaySound(`change`, `other`)

			user.SendText(fmt.Sprintf(`You use a key to %s the <ansi fg="container">%s</ansi>.`, actionVerb, containerName))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to %s the <ansi fg="container">%s</ansi>.`, user.Character.Name, actionVerb, containerName), user.UserId)

		} else if keyResult.HasBackpackKey {
			itmSpec := keyResult.BackpackKey.GetSpec()

			if doLock {
				container.Lock.SetLocked()
			} else {
				container.Lock.SetUnlocked()
			}
			room.Containers[containerName] = container

			locksmith.ConsumeBackpackKey(user.Character, keyResult.BackpackKey, lockId, user.UserId)
			room.PlaySound(`change`, `other`)

			user.SendText(fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to %s the <ansi fg="container">%s</ansi>, and add it to your key ring for the future.`, itmSpec.Name, actionVerb, containerName))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to %s the <ansi fg="container">%s</ansi>.`, user.Character.Name, actionVerb, containerName), user.UserId)

		} else {
			if doLock {
				user.SendText(`You do not have the key for that.`)
			} else {
				user.SendText(`You do not have the key for that. Maybe you could <ansi fg="command">picklock</ansi> the lock.`)
			}
		}

		return true, nil

	} else if exitName != `` {

		exitInfo, _ := room.GetExitInfo(exitName)

		if doLock && exitInfo.Lock.IsLocked() {
			user.SendText("That's already locked.")
			return true, nil
		}
		if !doLock && !exitInfo.Lock.IsLocked() {
			user.SendText("That's not locked.")
			return true, nil
		}

		lockId := locksmith.BuildLockId(room.RoomId, exitName)
		keyResult := locksmith.FindKey(user.Character, lockId, int(exitInfo.Lock.Difficulty))

		if keyResult.HasKeyRingKey {
			if doLock {
				exitInfo.Lock.SetLocked()
				room.SetExitLock(exitName, true)
			} else {
				exitInfo.Lock.SetUnlocked()
				room.SetExitLock(exitName, false)
			}
			room.PlaySound(`change`, `other`)

			user.SendText(fmt.Sprintf(`You use a key to %s the <ansi fg="exit">%s</ansi> lock.`, actionVerb, exitName))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to %s the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, actionVerb, exitName), user.UserId)

		} else if keyResult.HasBackpackKey {
			itmSpec := keyResult.BackpackKey.GetSpec()

			if doLock {
				exitInfo.Lock.SetLocked()
				room.SetExitLock(exitName, true)
			} else {
				exitInfo.Lock.SetUnlocked()
				room.SetExitLock(exitName, false)
			}

			locksmith.ConsumeBackpackKey(user.Character, keyResult.BackpackKey, lockId, user.UserId)
			room.PlaySound(`change`, `other`)

			user.SendText(fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to %s the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, actionVerb, exitName))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to %s the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, actionVerb, exitName), user.UserId)

		} else {
			if doLock {
				user.SendText(`You do not have the key for that.`)
			} else {
				user.SendText(`You do not have the key for that. Maybe you could <ansi fg="command">picklock</ansi> the lock.`)
			}
		}

		return true, nil
	}

	user.SendText("There is no such exit or container.")
	return true, nil
}

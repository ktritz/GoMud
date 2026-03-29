package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actions"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/scripting"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Equip(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "all" {
		return Gearup(``, user, room, flags)
	}

	if rest == "" {
		user.SendText(`Wear WHAT?`)
		return true, nil
	}

	matchItem, found := user.Character.FindInBackpack(rest)
	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to wear.`, rest))
		return true, nil
	}

	ctx := actions.ActionContext{
		UserId:    user.UserId,
		Character: user.Character,
		Room:      room,
		SendToActor: func(msg string) { user.SendText(msg) },
		SendToRoom:  func(msg string, exclude ...int) { room.SendText(msg, exclude...) },
		ActorName: user.Character.Name,
		ActorTag:  "username",
	}

	if actions.DoEquip(ctx, matchItem) {
		// Trigger any outstanding buff onStart events
		if len(matchItem.GetSpec().WornBuffIds) > 0 {
			for _, buff := range user.Character.Buffs.List {
				if buff.OnStartWaiting {
					if _, err := scripting.TryBuffScriptEvent(`onStart`, user.UserId, 0, buff.BuffId); err == nil {
						user.Character.TrackBuffStarted(buff.BuffId)
					}
				}
			}
		}
	}

	return true, nil
}

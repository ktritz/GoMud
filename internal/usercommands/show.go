package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/scripting"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Show(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	parsed := parser.GetParsedInput(user)

	var objectName string
	var targetName string

	if parsed != nil && !parsed.Instrument.IsEmpty() {
		// "show sword to merchant" -> Target=sword, Instrument=merchant
		objectName = parsed.Target.Noun
		targetName = parsed.Instrument.Noun
	}

	if objectName == "" || targetName == "" {
		user.SendText("Show what? To whom?")
		return true, nil
	}

	var showItem items.Item = items.Item{}
	var found bool = false

	// Check whether the user has an item in their inventory that matches
	showItem, found = user.Character.FindInBackpack(objectName)

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to show.", objectName))
		return true, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// Tell the shower
			user.SendText(
				fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, showItem.DisplayName(), targetUser.Character.Name),
			)

			// Tell the Showee
			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows you their <ansi fg="item">%s</ansi>.`, user.Character.Name, showItem.DisplayName()),
			)

			targetUser.SendText(
				"\n" + showItem.GetLongDescription() + "\n",
			)

			// Tell the rest of the room
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, user.Character.Name, showItem.DisplayName(), targetUser.Character.Name),
				targetUser.UserId,
				user.UserId)

		} else {
			user.SendText("Something went wrong.")
		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				user.SendText(
					fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, showItem.DisplayName(), targetMob.Character.Name),
				)

				// Do trigger of onShow
				scripting.TryMobScriptEvent(`onShow`, targetMob.InstanceId, user.UserId, `user`, map[string]any{`gold`: 0, `item`: showItem})

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, showItem.DisplayName(), targetMob.Character.Name),
					user.UserId,
				)

			} else {
				user.SendText("Something went wrong.")
			}

		}

		return true, nil
	}

	user.SendText("Who???")

	return true, nil
}

package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Show(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	parsed := parser.GetParsedInputFrom(mob)

	var objectName string
	var targetName string

	if parsed != nil && !parsed.Instrument.IsEmpty() {
		objectName = parsed.Target.Noun
		targetName = parsed.Instrument.Noun
	} else {
		return true, nil
	}

	if objectName == "" || targetName == "" {
		return true, nil
	}

	var showItem items.Item = items.Item{}
	var found bool = false

	// Check whether the user has an item in their inventory that matches
	showItem, found = mob.Character.FindInBackpack(objectName)

	if !found {
		return true, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// Tell the Showee
			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, showItem.DisplayName()),
			)

			targetUser.SendText(
				"\n" + showItem.GetLongDescription() + "\n",
			)

			// Tell the rest of the room
			room.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, mob.Character.Name, showItem.DisplayName(), targetUser.Character.Name),
				targetUser.UserId)

		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, showItem.DisplayName(), targetMob.Character.Name),
				)

			}

		}

		return true, nil
	}

	return true, nil
}

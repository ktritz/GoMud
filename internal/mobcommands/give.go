package mobcommands

import (
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Give(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	parsed := parser.GetParsedInputFrom(mob)

	var giveWhat string
	var giveWho string

	if parsed != nil && !parsed.Instrument.IsEmpty() {
		giveWhat = parsed.Target.Noun
		if parsed.Target.Quantity > 0 {
			giveWhat = fmt.Sprintf("%d %s", parsed.Target.Quantity, parsed.Target.Noun)
		}
		giveWho = parsed.Instrument.Noun
	} else if parsed != nil && parsed.Rest != "" {
		giveWhat = parsed.Rest
	} else {
		giveWhat = rest
	}

	if giveWhat == "" || giveWho == "" {
		return true, nil
	}

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if len(giveWhat) > 4 && giveWhat[len(giveWhat)-4:] == "gold" {

		g, _ := strconv.ParseInt(giveWhat[0:len(giveWhat)-5], 10, 32)
		giveGoldAmount = int(g)

		if giveGoldAmount > mob.Character.Gold {
			return true, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = mob.Character.FindInBackpack(giveWhat)

		if !found {
			return true, nil
		}

	}

	playerId, mobId := room.FindByName(giveWho)

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			mob.Character.RemoveItem(giveItem)

			events.AddToQueue(events.ItemOwnership{
				MobInstanceId: mob.InstanceId,
				Item:          giveItem,
				Gained:        false,
			})

			events.AddToQueue(events.ItemOwnership{
				UserId: targetUser.UserId,
				Item:   giveItem,
				Gained: true,
			})

			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName()),
			)

		} else if giveGoldAmount > 0 {

			targetUser.Character.Gold += giveGoldAmount
			mob.Character.Gold -= giveGoldAmount

			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, giveGoldAmount),
			)

		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		m := mobs.GetInstance(mobId)

		if m != nil {

			// Swap the item location
			if giveItem.ItemId > 0 {
				m.Character.StoreItem(giveItem)
				mob.Character.RemoveItem(giveItem)

				events.AddToQueue(events.ItemOwnership{
					MobInstanceId: mob.InstanceId,
					Item:          giveItem,
					Gained:        false,
				})

				events.AddToQueue(events.ItemOwnership{
					MobInstanceId: m.InstanceId,
					Item:          giveItem,
					Gained:        true,
				})

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName(), m.Character.Name),
				)
			} else if giveGoldAmount > 0 {

				m.Character.Gold += giveGoldAmount
				mob.Character.Gold -= giveGoldAmount

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, m.Character.Name),
				)
			}

		}

	}

	return true, nil
}

package mobcommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Get(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	parsed := parser.GetParsedInputFrom(mob)
	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		return true, nil
	}

	// Handle "get all"
	if args[0] == "all" || (parsed != nil && parsed.Target.All) {
		parser.StoreParsedInputOn(mob, nil)
		if room.Gold > 0 {
			Get("gold", mob, room)
		}
		if len(room.Items) > 0 {
			iCopies := append([]items.Item{}, room.Items...)
			for _, item := range iCopies {
				Get(item.Name(), mob, room)
			}
		}
		return true, nil
	}

	if args[0] == "gold" {
		if room.Gold > 0 {
			mob.Character.CancelBuffsWithFlag(buffs.Hidden)
			goldAmt := room.Gold
			mob.Character.Gold += goldAmt
			room.Gold -= goldAmt
			room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> picks up <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, goldAmt))
		}
		return true, nil
	}

	// Determine source using parser
	getFromStash := false
	sourceName := ""

	if parsed != nil && !parsed.Instrument.IsEmpty() {
		sourceName = strings.ToLower(parsed.Instrument.Noun)
		rest = parsed.Target.Noun
	}

	if sourceName == "stash" {
		getFromStash = true
	} else if sourceName == "" && len(args) >= 2 {
		// Fallback for mob scripts that pass "item stash" or "item from stash"
		if args[len(args)-1] == "stash" {
			getFromStash = true
			rest = strings.Join(args[0:len(args)-1], " ")
			if len(args) >= 3 && args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			}
		}
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := room.FindOnFloor(rest, getFromStash)

	if found {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

		// Swap the item location
		room.RemoveItem(matchItem, getFromStash)
		mob.Character.StoreItem(matchItem)

		events.AddToQueue(events.ItemOwnership{
			MobInstanceId: mob.InstanceId,
			Item:          matchItem,
			Gained:        true,
		})

		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, mob.Character.Name, matchItem.DisplayName()))
	}

	return true, nil
}

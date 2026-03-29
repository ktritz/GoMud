package mobcommands

import (
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actions"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Drop(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	ctx := actions.ActionContext{
		MobInstanceId: mob.InstanceId,
		Character:     &mob.Character,
		Room:          room,
		SendToActor:   func(msg string) {},
		SendToRoom:    func(msg string, exclude ...int) { room.SendText(msg, exclude...) },
		ActorName:     mob.Character.Name,
		ActorTag:      "mobname",
	}

	if args[0] == "all" {
		if mob.Character.Gold > 0 {
			actions.DoDropGold(ctx, mob.Character.Gold)
		}
		iCopies := append([]items.Item{}, mob.Character.Items...)
		for _, item := range iCopies {
			Drop(item.Name(), mob, room)
		}
		return true, nil
	}

	// Drop N gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt >= 1 && dropAmt <= mob.Character.Gold {
			actions.DoDropGold(ctx, dropAmt)
		}
		return true, nil
	}

	if mob.Character.HasBuffFlag(buffs.PermaGear) {
		return true, nil
	}

	matchItem, found := mob.Character.FindInBackpack(rest)
	if found {
		actions.DoDropItem(ctx, matchItem)
	}

	return true, nil
}

package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/actions"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Equip(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.HasBuffFlag(buffs.PermaGear) {
		mob.Command(`emote struggles with their gear for a while, then gives up.`)
		return true, nil
	}

	if rest == "all" {
		itemCopies := append([]items.Item{}, mob.Character.Items...)
		for _, item := range itemCopies {
			iSpec := item.GetSpec()
			if iSpec.Subtype == items.Wearable || iSpec.Type == items.Weapon {
				Equip(item.Name(), mob, room)
			}
		}
		return true, nil
	}

	var matchItem items.Item
	var found bool

	if rest == `random` && len(mob.Character.Items) > 0 {
		matchItem = mob.Character.Items[util.Rand(len(mob.Character.Items))]
		found = true
	}

	if !found {
		matchItem, found = mob.Character.FindInBackpack(rest)
	}

	if !found {
		return true, nil
	}

	ctx := actions.ActionContext{
		MobInstanceId: mob.InstanceId,
		Character:     &mob.Character,
		Room:          room,
		SendToActor:   func(msg string) {},
		SendToRoom:    func(msg string, exclude ...int) { room.SendText(msg, exclude...) },
		ActorName:     mob.Character.Name,
		ActorTag:      "mobname",
	}

	actions.DoEquip(ctx, matchItem)

	return true, nil
}

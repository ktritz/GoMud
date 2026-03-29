package actions

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// DoEquip handles equipping an item from the character's backpack.
// Shared logic for both user and mob equip commands.
func DoEquip(ctx ActionContext, matchItem items.Item) bool {

	iSpec := matchItem.GetSpec()
	if iSpec.Type != items.Weapon && iSpec.Subtype != items.Wearable {
		ctx.SendToActor(
			fmt.Sprintf(`Your <ansi fg="item">%s</ansi> doesn't look very fashionable.`, matchItem.DisplayName()),
		)
		return false
	}

	oldItems, wearSuccess, failureReason := ctx.Character.Wear(matchItem)

	if !wearSuccess {
		if len(failureReason) <= 1 {
			failureReason = fmt.Sprintf(`You can't figure out how to equip the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName())
		}
		ctx.SendToActor(failureReason)
		return false
	}

	ctx.Character.CancelBuffsWithFlag(buffs.Hidden)
	ctx.Character.RemoveItem(matchItem)

	for _, oldItem := range oldItems {
		if oldItem.ItemId != 0 {
			ctx.SendToActor(
				fmt.Sprintf(`You remove your <ansi fg="item">%s</ansi> and return it to your backpack.`, oldItem.DisplayName()),
			)
			ctx.SendToRoom(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, ctx.ActorTag, ctx.ActorName, oldItem.DisplayName()),
				ctx.UserId,
			)
			ctx.Character.StoreItem(oldItem)
		}
	}

	if iSpec.Subtype == items.Wearable {
		ctx.SendToActor(
			fmt.Sprintf(`You wear your <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
		)
		ctx.SendToRoom(
			fmt.Sprintf(`<ansi fg="%s">%s</ansi> puts on their <ansi fg="item">%s</ansi>.`, ctx.ActorTag, ctx.ActorName, matchItem.DisplayName()),
			ctx.UserId,
		)
	} else {
		ctx.SendToActor(
			fmt.Sprintf(`You wield your <ansi fg="item">%s</ansi>. You're feeling dangerous.`, matchItem.DisplayName()),
		)
		ctx.SendToRoom(
			fmt.Sprintf(`<ansi fg="%s">%s</ansi> wields their <ansi fg="item">%s</ansi>.`, ctx.ActorTag, ctx.ActorName, matchItem.DisplayName()),
			ctx.UserId,
		)
	}

	ctx.Character.Validate()

	events.AddToQueue(events.EquipmentChange{
		UserId:        ctx.UserId,
		MobInstanceId: ctx.MobInstanceId,
		ItemsWorn:     []items.Item{matchItem},
		ItemsRemoved:  oldItems,
	})

	return true
}

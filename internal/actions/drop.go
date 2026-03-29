package actions

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// DoDropItem handles dropping a specific item from the character's backpack.
func DoDropItem(ctx ActionContext, matchItem items.Item) {

	ctx.Character.CancelBuffsWithFlag(buffs.Hidden)
	ctx.Character.RemoveItem(matchItem)

	events.AddToQueue(events.ItemOwnership{
		UserId:        ctx.UserId,
		MobInstanceId: ctx.MobInstanceId,
		Item:          matchItem,
		Gained:        false,
	})

	ctx.SendToActor(
		fmt.Sprintf(`You drop the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
	)
	ctx.SendToRoom(
		fmt.Sprintf(`<ansi fg="%s">%s</ansi> drops their <ansi fg="item">%s</ansi>...`, ctx.ActorTag, ctx.ActorName, matchItem.DisplayName()),
		ctx.UserId,
	)

	ctx.Room.AddItem(matchItem, false)
}

// DoDropGold handles dropping gold on the floor.
func DoDropGold(ctx ActionContext, amount int) {

	ctx.Character.CancelBuffsWithFlag(buffs.Hidden)

	ctx.Room.Gold += amount
	ctx.Character.Gold -= amount

	events.AddToQueue(events.EquipmentChange{
		UserId:        ctx.UserId,
		MobInstanceId: ctx.MobInstanceId,
		GoldChange:    -amount,
	})

	ctx.SendToActor(
		fmt.Sprintf(`You drop <ansi fg="gold">%d gold</ansi> on the floor.`, amount),
	)
	ctx.SendToRoom(
		fmt.Sprintf(`<ansi fg="%s">%s</ansi> drops <ansi fg="gold">%d gold</ansi>.`, ctx.ActorTag, ctx.ActorName, amount),
		ctx.UserId,
	)
}

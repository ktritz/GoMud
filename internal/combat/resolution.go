package combat

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// SendRoundMessages dispatches all messages and buffs from an AttackResult.
// sourceUser/targetUser may be nil for mob-only combats.
// sourceRoom/targetRoom should not be nil.
// excludeUserIds lists user IDs to exclude from room messages.
type RoundMessageContext struct {
	SourceUser     *users.UserRecord
	TargetUser     *users.UserRecord
	SourceMob      *mobs.Mob
	TargetMob      *mobs.Mob
	SourceRoom     *rooms.Room
	TargetRoom     *rooms.Room
	ExcludeUserIds []int
}

// SendRoundMessages dispatches messages and applies buffs from a round result.
func SendRoundMessages(result AttackResult, ctx RoundMessageContext) {

	// Apply buffs
	for _, buffId := range result.BuffSource {
		if ctx.SourceUser != nil {
			ctx.SourceUser.AddBuff(buffId, `combat`)
		} else if ctx.SourceMob != nil {
			ctx.SourceMob.AddBuff(buffId, `combat`)
		}
	}

	for _, buffId := range result.BuffTarget {
		if ctx.TargetUser != nil {
			ctx.TargetUser.AddBuff(buffId, `combat`)
		} else if ctx.TargetMob != nil {
			ctx.TargetMob.AddBuff(buffId, `combat`)
		}
	}

	// Send messages to source
	for _, msg := range result.MessagesToSource {
		if ctx.SourceUser != nil {
			ctx.SourceUser.SendText(msg)
		}
	}

	// Send messages to target
	for _, msg := range result.MessagesToTarget {
		if ctx.TargetUser != nil {
			ctx.TargetUser.SendText(msg)
		}
	}

	// Send messages to source room
	for _, msg := range result.MessagesToSourceRoom {
		if ctx.SourceRoom != nil {
			ctx.SourceRoom.SendText(msg, ctx.ExcludeUserIds...)
		}
	}

	// Send messages to target room
	for _, msg := range result.MessagesToTargetRoom {
		if ctx.TargetRoom != nil {
			ctx.TargetRoom.SendText(msg, ctx.ExcludeUserIds...)
		}
	}
}

// HandleEquipmentBreak tests whether the defender's offhand item breaks from combat.
// Returns true if the item broke.
func HandleEquipmentBreak(defChar *characters.Character, defRoom *rooms.Room, roundResult AttackResult, defUserId int, defMobInstanceId int) bool {

	if !roundResult.Hit {
		return false
	}

	if defChar.Equipment.Offhand.ItemId == 0 {
		return false
	}

	modifier := 0
	if roundResult.Crit {
		modifier = int(defChar.Equipment.Offhand.GetSpec().BreakChance)
	}

	if !defChar.Equipment.Offhand.BreakTest(modifier) {
		return false
	}

	brokenItemName := defChar.Equipment.Offhand.NameSimple()

	if defUserId > 0 {
		if defUser := users.GetByUserId(defUserId); defUser != nil {
			defUser.SendText(`<ansi fg="202">***</ansi>`)
			defUser.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, brokenItemName))
			defUser.SendText(`<ansi fg="202">***</ansi>`)
		}

		defRoom.SendText(
			fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, brokenItemName, defChar.Name),
			defUserId,
		)
	} else {
		defRoom.SendText(
			fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="mobname">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, brokenItemName, defChar.Name),
		)
	}

	events.AddToQueue(events.ItemOwnership{
		UserId:        defUserId,
		MobInstanceId: defMobInstanceId,
		Item:          defChar.Equipment.Offhand,
		Gained:        false,
	})

	defChar.RemoveFromBody(defChar.Equipment.Offhand)

	itm := items.New(20) // Broken item
	if !defChar.StoreItem(itm) {
		defRoom.AddItem(itm, false)
	}

	events.AddToQueue(events.ItemOwnership{
		UserId:        defUserId,
		MobInstanceId: defMobInstanceId,
		Item:          itm,
		Gained:        true,
	})

	return true
}

// TriggerCharmedMobRetaliation makes charmed mobs in the room attack on behalf of their owner.
func TriggerCharmedMobRetaliation(room *rooms.Room, ownerUserId int, attackerId int, attackerIsMob bool) {
	for _, instanceId := range room.GetMobs(rooms.FindCharmed) {
		if charmedMob := mobs.GetInstance(instanceId); charmedMob != nil {
			if charmedMob.Character.IsCharmed(ownerUserId) && charmedMob.Character.Aggro == nil {
				// Set aggro to prevent multiple attack triggers on this conditional
				charmedMob.Character.Aggro = &characters.Aggro{
					Type: characters.DefaultAttack,
				}

				if attackerIsMob {
					charmedMob.Command(fmt.Sprintf("attack #%d", attackerId))
				} else {
					charmedMob.Command(fmt.Sprintf("attack @%d", attackerId))
				}
			}
		}
	}
}

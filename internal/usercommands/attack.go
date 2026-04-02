package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/targeting"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Attack(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	attackPlayerId := 0
	attackMobInstanceId := 0

	// Use parsed target if available (strips fillers like "the")
	if parsed := parser.GetParsedInput(user); parsed != nil && parsed.Rest != "" {
		rest = parsed.Rest
	}

	if rest == `` {
		// First check who's directly attacking this player
		attackPlayerId, attackMobInstanceId = targeting.FindAutoTarget(room, user.UserId, 0)

		// Also check if anyone is attacking a party member
		if attackMobInstanceId == 0 && attackPlayerId == 0 {
			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				// Check mobs attacking party members
				for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
					m := mobs.GetInstance(mId)
					if m.Character.Aggro != nil && partyInfo.IsMember(m.Character.Aggro.UserId) {
						attackMobInstanceId = m.InstanceId
						break
					}
				}

				// Check if any party member is already in combat (glom onto their target)
				if attackMobInstanceId == 0 && attackPlayerId == 0 {
					for uId := range partyInfo.GetMembers() {
						if partyUser := users.GetByUserId(uId); partyUser != nil {
							if partyUser.Character.Aggro == nil {
								continue
							}
							if partyUser.Character.Aggro.MobInstanceId > 0 {
								attackMobInstanceId = partyUser.Character.Aggro.MobInstanceId
								break
							}
							if partyUser.Character.Aggro.UserId > 0 {
								attackPlayerId = partyUser.Character.Aggro.UserId
								break
							}
						}
					}
				}
			}
		}

	} else if rest[0] == '*' {
		attackPlayerId, attackMobInstanceId = targeting.FindRandomTarget(room, rest, user.UserId, 0)
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackPlayerId == user.UserId { // Can't attack self!
		attackPlayerId = 0
	}

	if attackMobInstanceId == 0 && attackPlayerId == 0 {
		user.SendText("You attack the darkness!")
		return true, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	/*
		combatAddlWaitRounds := user.Character.Equipment.Weapon.GetSpec().WaitRounds + user.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if user.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {
			if m.Character.IsCharmed(user.UserId) {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := users.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {

								partyUser.Command(fmt.Sprintf(`attack #%d`, attackMobInstanceId)) // # denotes a specific mob instanceId

							}
						}

					}
				}
			}

			user.Character.SetAggro(0, attackMobInstanceId, characters.DefaultAttack)

			user.SendText(
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="mobname">%s</ansi>.`, m.Character.Name),
			)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>.`, user.Character.Name, m.Character.Name),
					user.UserId,
				)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.Aggro == nil && m.Character.IsCharmed(user.UserId) { // Charmed mobs help the player

						m.Command(fmt.Sprintf(`attack #%d`, attackMobInstanceId)) // # denotes a specific mob instanceId

					}
				}
			}

		}

	} else if attackPlayerId > 0 {

		if p := users.GetByUserId(attackPlayerId); p != nil {

			if pvpErr := room.CanPvp(user, p); pvpErr != nil {
				user.SendText(pvpErr.Error())
				return true, nil
			}

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					return true, nil
				}
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := users.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {
								partyUser.Command(fmt.Sprintf(`attack @%d`, attackPlayerId)) // # denotes a specific mob instanceId
							}
						}
					}
				}
			}

			user.Character.SetAggro(attackPlayerId, 0, characters.DefaultAttack)

			user.SendText(
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="username">%s</ansi>.`, p.Character.Name),
			)

			if !isSneaking {

				p.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight you!`, user.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>.`, user.Character.Name, p.Character.Name),
					user.UserId, attackPlayerId)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.Aggro == nil && m.Character.IsCharmed(user.UserId) { // Charmed mobs help the player

						m.Command(fmt.Sprintf(`attack @%d`, attackPlayerId)) // @ denotes a specific user id

					}
				}
			}

		}

	}

	return true, nil
}

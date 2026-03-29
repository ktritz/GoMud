package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/targeting"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Backstab(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if !isSneaking {
		return true, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
		if mob.Character.Aggro != nil {
			mob.Character.Aggro.Type = characters.BackStab
			return true, nil
		}
		attackPlayerId, attackMobInstanceId = targeting.FindAutoTarget(room, 0, mob.InstanceId)
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackMobInstanceId == mob.InstanceId {
		attackMobInstanceId = 0
	}

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {
			mob.Character.SetAggro(0, attackMobInstanceId, characters.BackStab)
		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {
			mob.Character.SetAggro(attackPlayerId, 0, characters.BackStab)
		}

	}

	return true, nil
}

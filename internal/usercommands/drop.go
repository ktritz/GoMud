package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actions"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Drop(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		user.SendText(`Drop what?`)
		return true, nil
	}

	ctx := actions.ActionContext{
		UserId:      user.UserId,
		Character:   user.Character,
		Room:        room,
		SendToActor: func(msg string) { user.SendText(msg) },
		SendToRoom:  func(msg string, exclude ...int) { room.SendText(msg, exclude...) },
		ActorName:   user.Character.Name,
		ActorTag:    "username",
	}

	if args[0] == "all" {
		if user.Character.Gold > 0 {
			actions.DoDropGold(ctx, user.Character.Gold)
		}
		iCopies := append([]items.Item{}, user.Character.Items...)
		for _, item := range iCopies {
			actions.DoDropItem(ctx, item)
		}
		return true, nil
	}

	// Drop N gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt < 1 {
			user.SendText("Oops!")
			return true, nil
		}
		if dropAmt > user.Character.Gold {
			user.SendText(fmt.Sprintf("You don't have a %d gold to drop.", dropAmt))
			return true, nil
		}
		actions.DoDropGold(ctx, dropAmt)
		return true, nil
	}

	matchItem, found := user.Character.FindInBackpack(rest)
	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to drop.", rest))
		return true, nil
	}

	actions.DoDropItem(ctx, matchItem)

	iSpec := matchItem.GetSpec()
	if iSpec.Type == items.Grenade {
		user.SendText(`Todo. Grenades disabled for now.`)
	}

	return true, nil
}

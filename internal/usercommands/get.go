package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/parser"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Get(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	parsed := parser.GetParsedInput(user)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		user.SendText("Get what?")
		return true, nil
	}

	// Handle "get all" — check args first since recursive calls pass raw "rest"
	// and the parsed data in tempStore would be stale
	isAll := args[0] == "all"
	if !isAll && parsed != nil && parsed.Target.All {
		isAll = true
	}
	if isAll {
		// Clear parsed data so recursive Get calls don't re-trigger "all"
		parser.StoreParsedInput(user, nil)
		if room.Gold > 0 {
			Get(`gold`, user, room, flags)
		}
		if len(room.Items) > 0 {
			iCopies := append([]items.Item{}, room.Items...)
			for _, item := range iCopies {
				Get(item.Name(), user, room, flags)
			}
		}
		return true, nil
	}

	// Determine source using parser if available
	getFromStash := false
	containerName := ``
	petUserId := 0
	sourceName := ``

	if parsed != nil && !parsed.Instrument.IsEmpty() {
		sourceName = strings.ToLower(parsed.Instrument.Noun)
		rest = parsed.Target.Noun
		if len(parsed.Target.Adjectives) > 0 {
			rest = strings.Join(parsed.Target.Adjectives, " ") + " " + rest
		}
	}

	// Check source type
	if sourceName == "stash" {
		getFromStash = true
	} else if sourceName == "ground" {
		getFromStash = false
	} else if sourceName != "" {
		// Check container
		containerName = room.FindContainerByName(sourceName)

		// Check pet
		if containerName == "" {
			petUserId = room.FindByPetName(sourceName)
			if petUserId == 0 && sourceName == "pet" && user.Character.Pet.Exists() {
				petUserId = user.UserId
			}
		}
	}

	// Fallback: if parser didn't find a source, try old-style last-arg detection
	if sourceName == "" && len(args) >= 2 {
		lastArg := args[len(args)-1]
		if lastArg == "stash" {
			getFromStash = true
			rest = strings.Join(args[0:len(args)-1], " ")
			if len(args) >= 3 && args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			}
		} else if lastArg == "ground" {
			if len(args) >= 3 && args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		} else {
			cn := room.FindContainerByName(lastArg)
			if cn != "" {
				containerName = cn
				rest = strings.Join(args[0:len(args)-1], " ")
				if len(args) >= 3 && args[len(args)-2] == "from" {
					rest = strings.Join(args[0:len(args)-2], " ")
				}
			} else {
				pId := room.FindByPetName(lastArg)
				if pId == 0 && lastArg == "pet" && user.Character.Pet.Exists() {
					pId = user.UserId
				}
				if pId > 0 {
					petUserId = pId
					rest = strings.Join(args[0:len(args)-1], " ")
					if len(args) >= 3 && args[len(args)-2] == "from" {
						rest = strings.Join(args[0:len(args)-2], " ")
					}
				}
			}
		}
	}

	if petUserId > 0 && petUserId != user.UserId {
		user.SendText(`You can't do that!`)
		return true, nil
	}

	if petUserId == user.UserId {

		matchItem, found := user.Character.Pet.FindItem(rest)
		if !found {
			user.SendText(fmt.Sprintf(`You don't see a %s carried by %s.`, rest, user.Character.Pet.DisplayName()))
		} else {

			if user.Character.Pet.RemoveItem(matchItem) {
				if !user.Character.StoreItem(matchItem) {
					user.Character.Pet.StoreItem(matchItem)
				} else {

					events.AddToQueue(events.ItemOwnership{
						UserId: user.UserId,
						Item:   matchItem,
						Gained: true,
					})
				}

				user.SendText(
					fmt.Sprintf(`You remove a <ansi fg="itemname">%s</ansi> from %s.`, matchItem.DisplayName(), user.Character.Pet.DisplayName()),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> removes a <ansi fg="itemname">%s</ansi> from %s...`, user.Character.Name, matchItem.DisplayName(), user.Character.Pet.DisplayName()),
					user.UserId,
				)

			}
		}

		return true, nil

	}

	if containerName != `` {
		container := room.Containers[containerName]

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) < 5 && goldName[0:len(args[0])-1] == args[0]) {

			if container.Gold < 1 {
				user.SendText("There's no gold to grab.")
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := container.Gold
				user.Character.Gold += goldAmt
				container.Gold -= goldAmt
				room.Containers[containerName] = container

				events.AddToQueue(events.EquipmentChange{
					UserId:     user.UserId,
					GoldChange: -goldAmt,
				})

				user.SendText(
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi> from the <ansi fg="container">%s</ansi>.`, goldAmt, containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi> from the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName),
					user.UserId,
				)
			}

			return true, nil
		}

		matchItem, found := container.FindItem(rest)

		if !found {
			user.SendText(fmt.Sprintf(`You don't see a %s in the <ansi fg="container">%s</ansi>.`, rest, containerName))
		} else {

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			// Trigger onFound event
			if user.Character.StoreItem(matchItem) {

				events.AddToQueue(events.ItemOwnership{
					UserId: user.UserId,
					Item:   matchItem,
					Gained: true,
				})

				// Swap the item location
				container.RemoveItem(matchItem)
				room.Containers[containerName] = container

				user.SendText(
					fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>.`, matchItem.DisplayName(), containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>...`, user.Character.Name, matchItem.DisplayName(), containerName),
					user.UserId,
				)

				return true, nil

			} else {
				user.SendText(
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}

		}

	} else {

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) < 5 && goldName[0:len(args[0])-1] == args[0]) {

			if room.Gold < 1 {
				user.SendText("There's no gold to grab.")
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := room.Gold
				user.Character.Gold += goldAmt
				room.Gold -= goldAmt

				events.AddToQueue(events.EquipmentChange{
					UserId:     user.UserId,
					GoldChange: -goldAmt,
				})

				user.SendText(
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi>.`, goldAmt),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi>.`, user.Character.Name),
					user.UserId,
				)
			}

			return true, nil
		}

		// Check whether the user has an item in their inventory that matches
		matchItem, found := room.FindOnFloor(rest, getFromStash)

		// Check if user is specifying an item they stashed
		if !found && !getFromStash {
			stashItemMatch, stashFound := room.FindOnFloor(rest, true)
			if stashFound && stashItemMatch.StashedBy == user.UserId {
				found = true
				getFromStash = true
				matchItem = stashItemMatch
			}
		}

		if found {

			if matchItem.HasAdjective(`exploding`) {
				user.SendText(`You can't pick that up, it's about to explode!`)
				return true, nil
			}

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			// If it was in the stash, remove the stash owner tag
			if getFromStash {
				matchItem.StashedBy = 0
			}

			if user.Character.StoreItem(matchItem) {

				// Swap the item location
				room.RemoveItem(matchItem, getFromStash)

				events.AddToQueue(events.ItemOwnership{
					UserId: user.UserId,
					Item:   matchItem,
					Gained: true,
				})

				if getFromStash {
					user.SendText(
						fmt.Sprintf(`You dig out the <ansi fg="itemname">%s</ansi> from where it was stashed.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> digs around in the area and picks something up...`, user.Character.Name),
						user.UserId,
					)
				} else {
					user.SendText(
						fmt.Sprintf(`You pick up the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
						user.UserId,
					)
				}

			} else {
				user.SendText(
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}

			return true, nil
		}

		//
		// Look for any nouns in the room info
		//
		foundNoun, _ := room.FindNoun(rest)
		if len(foundNoun) > 0 {

			user.SendText(fmt.Sprintf(`You can't get the <ansi fg="noun">%s</ansi>`, foundNoun))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is grasping at the air.`, user.Character.Name), user.UserId)

			return true, nil
		}

	}

	if _, corpseFound := room.FindCorpse(rest); corpseFound {
		user.SendText(`You can't pick up corpses. What would people think?`)
		return true, nil
	}

	containerName = room.FindContainerByName(rest)
	if containerName != `` {
		user.SendText(fmt.Sprintf(`You can't pick up the <ansi fg="container">%s</ansi>. Try looking at it.`, containerName))
	} else {
		user.SendText(fmt.Sprintf("You don't see a %s around.", rest))
	}

	return true, nil
}

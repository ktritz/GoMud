package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"sort"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/gamelock"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func room_Edit_Exits(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// This basic struct will be used to keep track of what we're editing
	type ExitEdit struct {
		Name    string
		NameNew string
		Exit    exit.RoomExit
		Exists  bool
	}

	exitOptions := []templates.NameDescription{}

	for name, c := range room.Exits {

		exitOpt := templates.NameDescription{Name: name}

		if c.Lock.Difficulty > 0 {
			exitOpt.Description += fmt.Sprintf(`[Lvl %d Lock] `, c.Lock.Difficulty)
		}

		if c.Secret {
			exitOpt.Description += `[hidden] `
		}

		exitOptions = append(exitOptions, exitOpt)

	}

	// Must sort since maps will often change between iterations
	sort.SliceStable(exitOptions, func(i, j int) bool {
		return exitOptions[i].Name < exitOptions[j].Name
	})

	//
	// Create a holder for exit editing data
	//
	currentlyEditing := ExitEdit{}

	cmdPrompt, _ := user.StartPrompt(`room edit exits`, rest)

	question := cmdPrompt.Ask(`Choose one:`, []string{`new`}, `new`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", exitOptions, user.UserId)
		user.SendText(tplTxt)
		return true, nil
	}

	currentlyEditing.Name = question.Response

	if restNum, err := strconv.Atoi(currentlyEditing.Name); err == nil {
		if restNum > 0 && restNum <= len(exitOptions) {
			currentlyEditing.Name = exitOptions[restNum-1].Name
		}
	}

	for _, o := range exitOptions {
		if strings.EqualFold(o.Name, currentlyEditing.Name) {
			currentlyEditing.Name = o.Name
			break
		}
	}

	// Load the (possible) existing exit
	currentlyEditing.Exit, currentlyEditing.Exists = room.Exits[currentlyEditing.Name]

	// If they entered a exit name...
	if currentlyEditing.Name != `new` {

		// Does the exit name they entered not exist? Failure!
		if !currentlyEditing.Exists {
			user.SendText("Invalid option selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Since they picked a exit that exists, lets get the question of delete out of the way immediately.
		question := cmdPrompt.Ask(`Delete this exit?`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		// Delete the exit if that's what they want!
		if question.Response == `yes` {

			delete(room.Exits, currentlyEditing.Name)
			rooms.SaveRoomTemplate(*room)

			user.SendText(``)
			user.SendText(fmt.Sprintf(`<ansi fg="exit">%s</ansi> deleted from the room.`, currentlyEditing.Name))
			user.SendText(``)

			user.ClearPrompt()
			return true, nil
		}

	}

	//
	// Name Selection
	//
	{
		// If they are creating a new exit, we don't want that to become a viable exit name, lets empty it
		if currentlyEditing.Name == `new` {
			currentlyEditing.Name = ``
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Choose a name for this exit:`, []string{currentlyEditing.Name}, currentlyEditing.Name)
		if !question.Done {
			return true, nil
		}
		currentlyEditing.NameNew = question.Response

		// Make sure they aren't using any reserved names.
		if currentlyEditing.NameNew == `quit` || currentlyEditing.NameNew == `new` {
			user.SendText("Invalid new name selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Make sure the new name isn't a duplicate
		if currentlyEditing.Name != currentlyEditing.NameNew {
			if _, ok := room.Exits[currentlyEditing.NameNew]; ok {

				user.SendText(`<ansi fg="red">An exit with that name already exists!</ansi>`)
				question.RejectResponse()
				return true, nil

			}
		}

	}

	//
	// Target RoomId
	//
	{
		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`What RoomId will this exit lead to?`, []string{strconv.Itoa(currentlyEditing.Exit.RoomId)}, strconv.Itoa(currentlyEditing.Exit.RoomId))
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.RoomId, _ = strconv.Atoi(question.Response)

		// Make sure they aren't using any reserved names.
		if rooms.LoadRoom(currentlyEditing.Exit.RoomId) == nil {
			user.SendText("Invalid RoomId provided.")
			question.RejectResponse()
			return true, nil
		}

	}

	//
	// Exit message?
	//
	{
		secretExitDefault := `no`
		if currentlyEditing.Exit.Secret {
			secretExitDefault = `yes`
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Is this a hidden exit?`, []string{`yes`, `no`}, secretExitDefault)
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.Secret = question.Response == `yes`
	}

	//
	// Secret exit?
	//
	{
		secretExitDefault := `no`
		if currentlyEditing.Exit.Secret {
			secretExitDefault = `yes`
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Is this a hidden exit?`, []string{`yes`, `no`}, secretExitDefault)
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.Secret = question.Response == `yes`
	}

	//
	// Special message when using the exit?
	//
	{
		defaultMessage := currentlyEditing.Exit.ExitMessage
		if defaultMessage == `` {
			defaultMessage = `none`
		}
		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Special message when using the exit?`, []string{defaultMessage}, defaultMessage)
		if !question.Done {
			return true, nil
		}

		if question.Response != `none` {
			currentlyEditing.Exit.ExitMessage = question.Response
		}

	}

	//
	// Lock Options
	//
	{
		question := cmdPrompt.Ask(`Will this exit be locked?`, []string{`yes`, `no`}, util.BoolYN(currentlyEditing.Exit.Lock.Difficulty > 0))
		if !question.Done {
			return true, nil
		}

		if question.Response == `yes` {

			defaultDifficultyAnswer := ``
			if currentlyEditing.Exit.Lock.Difficulty > 0 {
				defaultDifficultyAnswer = strconv.Itoa(int(currentlyEditing.Exit.Lock.Difficulty))
			}

			question := cmdPrompt.Ask(`What difficulty will the lock be (2-32)?`, []string{defaultDifficultyAnswer}, defaultDifficultyAnswer)
			if !question.Done {
				return true, nil
			}

			difficultyInt, _ := strconv.Atoi(question.Response)

			// Make sure the provided difficulty is within acceptable range.
			if difficultyInt < 2 || difficultyInt > 32 {
				user.SendText("Difficulty must between 2 and 32, inclusive.")
				question.RejectResponse()
				return true, nil
			}

			currentlyEditing.Exit.Lock.Difficulty = uint8(difficultyInt)

		} else {
			// reset the lock state if there is no lock.
			currentlyEditing.Exit.Lock = gamelock.Lock{}
		}

		if currentlyEditing.Exit.Lock.Difficulty > 0 {
			//
			// Lock Trap Options
			//
			question = cmdPrompt.Ask(`Will this lock have a trap?`, []string{`yes`, `no`}, util.BoolYN(len(currentlyEditing.Exit.Lock.TrapBuffIds) > 0))
			if !question.Done {
				return true, nil
			}

			if question.Response == `yes` {

				selectedBuffList := []int{}
				if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
					selectedBuffList = cb.([]int)
				}

				if len(selectedBuffList) == 0 {
					selectedBuffList = append(selectedBuffList, currentlyEditing.Exit.Lock.TrapBuffIds...)
				}

				// Keep track of the state
				cmdPrompt.Store(`trapBuffs`, selectedBuffList)

				selectedBuffLookup := map[int]bool{}
				for _, bId := range selectedBuffList {
					selectedBuffLookup[bId] = true
				}

				buffOptions := []templates.NameDescription{}

				for _, buffId := range buffs.GetAllBuffIds() {
					if b := buffs.GetBuffSpec(buffId); b != nil {

						if b.Name == `empty` {
							continue
						}

						marked := false
						if _, ok := selectedBuffLookup[buffId]; ok {
							marked = true
						}

						buffOptions = append(buffOptions, templates.NameDescription{Id: buffId, Marked: marked, Name: b.Name})
					}
				}

				sort.SliceStable(buffOptions, func(i, j int) bool {
					return buffOptions[i].Name < buffOptions[j].Name
				})

				question := cmdPrompt.Ask(`Select a buff to add to the trap, or nothing to continue:`, []string{}, `0`)
				if !question.Done {
					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions, user.UserId)
					user.SendText(tplTxt)
					return true, nil
				}

				buffSelected := question.Response

				if buffSelected != `0` {

					buffSelectedInt := 0

					if restNum, err := strconv.Atoi(buffSelected); err == nil {
						if restNum > 0 && restNum <= len(buffOptions) {
							buffSelectedInt = buffOptions[restNum-1].Id.(int)
						}
					}

					if buffSelectedInt == 0 {
						for _, b := range buffOptions {
							if strings.EqualFold(b.Name, buffSelected) {
								buffSelectedInt = b.Id.(int)
								break
							}
						}
					}

					if buffSelectedInt == 0 {

						user.SendText("Invalid selection.")
						question.RejectResponse()

						tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions, user.UserId)
						user.SendText(tplTxt)
						return true, nil
					}

					if _, ok := selectedBuffLookup[buffSelectedInt]; ok {

						delete(selectedBuffLookup, buffSelectedInt)
						for idx, buffId := range selectedBuffList {
							if buffId == buffSelectedInt {
								selectedBuffList = append(selectedBuffList[0:idx], selectedBuffList[idx+1:]...)
								break
							}
						}

					} else {

						selectedBuffList = append(selectedBuffList, buffSelectedInt)
						selectedBuffLookup[buffSelectedInt] = true

					}

					cmdPrompt.Store(`trapBuffs`, selectedBuffList)

					question.RejectResponse()

					for idx, data := range buffOptions {
						_, data.Marked = selectedBuffLookup[data.Id.(int)]
						buffOptions[idx] = data
					}

					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions, user.UserId)
					user.SendText(tplTxt)
					return true, nil

				}

			}

			if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
				currentlyEditing.Exit.Lock.TrapBuffIds = cb.([]int)
			}

			if currentlyEditing.Exit.Lock.RelockInterval == `` {
				currentlyEditing.Exit.Lock.RelockInterval = gamelock.DefaultRelockTime
			}

			question = cmdPrompt.Ask(`How long until it automatically relocks?`, []string{currentlyEditing.Exit.Lock.RelockInterval}, currentlyEditing.Exit.Lock.RelockInterval)
			if !question.Done {
				return true, nil
			}

			currentlyEditing.Exit.Lock.RelockInterval = question.Response

			// If the default time is chosen, can just leave it blank.
			if currentlyEditing.Exit.Lock.RelockInterval == gamelock.DefaultRelockTime {
				currentlyEditing.Exit.Lock.RelockInterval = ``
			}

		}
	}

	//
	// Done editing. Save results
	//
	if currentlyEditing.Name != `` {
		delete(room.Exits, currentlyEditing.Name)
	}

	room.Exits[currentlyEditing.NameNew] = currentlyEditing.Exit
	rooms.SaveRoomTemplate(*room)

	user.SendText(``)

	if currentlyEditing.Exit.Lock.Difficulty > 0 {
		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, currentlyEditing.NameNew)
		user.SendText(fmt.Sprintf(`<ansi fg="red">To Create Key -  LockId: <ansi fg="231" bg="5">%s</ansi></ansi>`, lockId))

		seqString := ``
		for _, dir := range util.GetLockSequence(lockId, int(currentlyEditing.Exit.Lock.Difficulty), string(configs.GetServerConfig().Seed)) {
			seqString += string(dir) + " "
		}
		user.SendText(fmt.Sprintf(`<ansi fg="red">To pick lock - Sequence: <ansi fg="green">%s</ansi></ansi>`, seqString))
	}

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}


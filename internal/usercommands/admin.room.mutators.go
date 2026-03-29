package usercommands

import (
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func room_Edit_Mutators(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	allRoomMutators := []string{}
	for _, roomMut := range room.Mutators {
		allRoomMutators = append(allRoomMutators, roomMut.MutatorId)
	}

	cmdPrompt, _ := user.StartPrompt(`room edit mutators`, rest)

	selectedMutatorList := []string{}
	if muts, ok := cmdPrompt.Recall(`mutators`); ok {
		selectedMutatorList = muts.([]string)
	} else {
		if len(selectedMutatorList) == 0 {
			selectedMutatorList = append(selectedMutatorList, allRoomMutators...)
		}
	}

	// Keep track of the state
	cmdPrompt.Store(`mutators`, selectedMutatorList)

	selectedMutatorLookup := map[string]bool{}
	for _, mutId := range selectedMutatorList {
		selectedMutatorLookup[mutId] = true
	}

	mutatorOptions := []templates.NameDescription{}

	for _, mutId := range mutators.GetAllMutatorIds() {
		marked := false
		if _, ok := selectedMutatorLookup[mutId]; ok {
			marked = true
		}

		mutatorOptions = append(mutatorOptions, templates.NameDescription{Id: mutId, Marked: marked, Name: mutId})

	}

	sort.SliceStable(mutatorOptions, func(i, j int) bool {
		return mutatorOptions[i].Name < mutatorOptions[j].Name
	})

	question := cmdPrompt.Ask(`Select a mutator to add to the room, or nothing to continue:`, []string{}, `0`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions, user.UserId)
		user.SendText(tplTxt)
		return true, nil
	}

	if question.Response != `0` {

		mutatorSelected := ``

		if restNum, err := strconv.Atoi(question.Response); err == nil {
			if restNum > 0 && restNum <= len(mutatorOptions) {
				mutatorSelected = mutatorOptions[restNum-1].Id.(string)
			}
		}

		if mutatorSelected == `` {
			for _, b := range mutatorOptions {
				if strings.EqualFold(b.Name, question.Response) {
					mutatorSelected = b.Id.(string)
					break
				}
			}
		}

		if mutatorSelected == `` {

			user.SendText("Invalid selection.")
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions, user.UserId)
			user.SendText(tplTxt)
			return true, nil
		}

		if _, ok := selectedMutatorLookup[mutatorSelected]; ok {

			delete(selectedMutatorLookup, mutatorSelected)
			for idx, mutId := range selectedMutatorList {
				if mutId == mutatorSelected {
					selectedMutatorList = append(selectedMutatorList[0:idx], selectedMutatorList[idx+1:]...)
					break
				}
			}

		} else {

			selectedMutatorList = append(selectedMutatorList, mutatorSelected)
			selectedMutatorLookup[mutatorSelected] = true

		}

		cmdPrompt.Store(`mutators`, selectedMutatorList)

		question.RejectResponse()

		for idx, data := range mutatorOptions {
			_, data.Marked = selectedMutatorLookup[data.Id.(string)]
			mutatorOptions[idx] = data
		}

		tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions, user.UserId)
		user.SendText(tplTxt)
		return true, nil

	}

	//
	// Done editing. Save results
	//
	room.Mutators = mutators.MutatorList{}
	for _, mutId := range selectedMutatorList {
		room.Mutators = append(room.Mutators, mutators.Mutator{MutatorId: mutId})
	}
	rooms.SaveRoomTemplate(*room)

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}

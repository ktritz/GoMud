package rooms

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/audio"
	"github.com/GoMudEngine/GoMud/internal/events"
)

func (r *Room) SendTextCommunication(txt string, excludeUserIds ...int) {

	events.AddToQueue(events.Message{
		RoomId:          r.RoomId,
		Text:            txt + "\n",
		ExcludeUserIds:  excludeUserIds,
		IsQuiet:         false,
		IsCommunication: true,
	})

}

func (r *Room) SendText(txt string, excludeUserIds ...int) {

	events.AddToQueue(events.Message{
		RoomId:         r.RoomId,
		Text:           txt + "\n",
		ExcludeUserIds: excludeUserIds,
		IsQuiet:        false,
	})

}

func (r *Room) PlaySound(soundId string, category string, excludeUserIds ...int) {

	volume := 100
	if soundConfig := audio.GetFile(soundId); soundConfig.FilePath != `` {
		soundId = soundConfig.FilePath
		if soundConfig.Volume > 0 && soundConfig.Volume <= 100 {
			volume = soundConfig.Volume
		}
	}

	for _, userId := range r.players {

		skip := false

		exLen := len(excludeUserIds)
		if exLen > 0 {
			for _, excludeId := range excludeUserIds {
				if excludeId == userId {
					skip = true
					break
				}
			}
		}

		if skip {
			continue
		}

		events.AddToQueue(events.MSP{
			UserId:    userId,
			SoundType: `SOUND`,
			SoundFile: soundId,
			Volume:    volume,
			Category:  category,
		})
	}

}

func (r *Room) SendTextToExits(txt string, isQuiet bool, excludeUserIds ...int) {

	testExitIds := []int{}
	for _, rExit := range r.Exits {
		testExitIds = append(testExitIds, rExit.RoomId)
	}
	for _, tExit := range r.ExitsTemp {
		testExitIds = append(testExitIds, tExit.RoomId)
	}

	for _, roomId := range testExitIds {

		tgtRoom := LoadRoom(roomId)
		if tgtRoom == nil {
			continue
		}

		for exitName, tExit := range tgtRoom.Exits {
			if tExit.RoomId != r.RoomId {
				continue
			}

			events.AddToQueue(events.Message{
				RoomId:         tgtRoom.RoomId,
				Text:           fmt.Sprintf(`(From <ansi fg="exit">%s</ansi>) `, exitName) + txt + "\n",
				IsQuiet:        isQuiet,
				ExcludeUserIds: excludeUserIds,
			})
		}

	}

}

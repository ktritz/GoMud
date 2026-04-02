package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/transport"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

type messageRecipient struct {
	ConnectionId connections.ConnectionId
	ScreenReader bool
}

type messageDeliveryPlan struct {
	Text       string
	Recipients []messageRecipient
}

func (m messageDeliveryPlan) Type() string { return `MessageDeliveryPlan` }

func Message_PlanDelivery(e events.Event) events.ListenerReturn {

	message, typeOk := e.(events.Message)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
		return events.Continue
	}

	recipients := []messageRecipient{}

	if message.UserId > 0 {

		if user := users.GetByUserId(message.UserId); user != nil {

			// If they are deafened, they cannot hear user communications
			if message.IsCommunication && user.Deafened {
				return events.Continue
			}

			recipients = append(recipients, messageRecipient{
				ConnectionId: user.ConnectionId(),
				ScreenReader: user.ScreenReader,
			})

			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

		}
	}

	if message.RoomId > 0 {

		room := rooms.LoadRoom(message.RoomId)
		if room == nil {
			return events.Continue
		}

		for _, userId := range room.GetPlayers() {
			skip := false

			if message.UserId == userId {
				continue
			}

			exLen := len(message.ExcludeUserIds)
			if exLen > 0 {
				for _, excludeId := range message.ExcludeUserIds {
					if excludeId == userId {
						skip = true
						break
					}
				}
			}

			if skip {
				continue
			}

			if user := users.GetByUserId(userId); user != nil {

				// If they are deafened, they cannot hear user communications
				if message.IsCommunication && user.Deafened {
					continue
				}

				// If this is a quiet message, make sure the player can hear it
				if message.IsQuiet {
					if !user.Character.HasBuffFlag(buffs.SuperHearing) {
						continue
					}
				}

				recipients = append(recipients, messageRecipient{
					ConnectionId: user.ConnectionId(),
					ScreenReader: user.ScreenReader,
				})

				events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

			}
		}
	}

	if len(recipients) > 0 {
		events.AddToQueue(messageDeliveryPlan{
			Text:       message.Text,
			Recipients: recipients,
		})
	}

	return events.Continue

}

// Checks whether their level is too high for a guide
func Message_SendMessage(e events.Event) events.ListenerReturn {

	plan, typeOk := e.(messageDeliveryPlan)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "MessageDeliveryPlan", "Actual Type", e.Type())
		return events.Continue
	}

	if len(plan.Recipients) == 0 {
		return events.Continue
	}

	textOut := templates.AnsiParse(plan.Text)
	textOutSR := ``
	prefix := term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String()
	deliveries := make([]transport.Delivery, 0, len(plan.Recipients))

	for _, recipient := range plan.Recipients {
		payloadText := textOut
		if recipient.ScreenReader {
			if textOutSR == `` {
				textOutSR = util.StripCharsForScreenReaders(textOut)
			}
			payloadText = textOutSR
		}

		deliveries = append(deliveries, transport.Delivery{
			ConnectionIds: []connections.ConnectionId{recipient.ConnectionId},
			Payload:       []byte(prefix + payloadText),
		})
	}

	transport.Queue(deliveries...)

	return events.Continue
}

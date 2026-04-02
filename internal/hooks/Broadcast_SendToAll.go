package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/transport"
	"github.com/GoMudEngine/GoMud/internal/users"
)

type broadcastRecipient struct {
	ConnectionId connections.ConnectionId
	ScreenReader bool
}

type broadcastDeliveryPlan struct {
	Text             string
	TextScreenReader string
	SkipLineRefresh  bool
	Recipients       []broadcastRecipient
}

func (b broadcastDeliveryPlan) Type() string { return `BroadcastDeliveryPlan` }

func Broadcast_PlanDelivery(e events.Event) events.ListenerReturn {

	broadcast, typeOk := e.(events.Broadcast)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
		return events.Continue
	}

	recipients := []broadcastRecipient{}

	for _, u := range users.GetAllActiveUsers() {

		if broadcast.IsCommunication {
			if u.Deafened && !broadcast.SourceIsMod {
				continue
			}
		}

		events.AddToQueue(events.RedrawPrompt{UserId: u.UserId}, 100)
		recipients = append(recipients, broadcastRecipient{
			ConnectionId: u.ConnectionId(),
			ScreenReader: u.ScreenReader,
		})
	}

	if len(recipients) > 0 {
		events.AddToQueue(broadcastDeliveryPlan{
			Text:             broadcast.Text,
			TextScreenReader: broadcast.TextScreenReader,
			SkipLineRefresh:  broadcast.SkipLineRefresh,
			Recipients:       recipients,
		})
	}

	return events.Continue
}

// Checks whether their level is too high for a guide
func Broadcast_SendToAll(e events.Event) events.ListenerReturn {

	plan, typeOk := e.(broadcastDeliveryPlan)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "BroadcastDeliveryPlan", "Actual Type", e.Type())
		return events.Continue
	}

	if len(plan.Recipients) == 0 {
		return events.Continue
	}

	textOut := ``
	if len(plan.Text) > 0 {
		textOut = templates.AnsiParse(plan.Text)
	}

	textOutSR := ``
	if len(plan.TextScreenReader) > 0 {
		textOutSR = templates.AnsiParse(plan.TextScreenReader)
	}

	prefix := term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String()
	deliveries := make([]transport.Delivery, 0, len(plan.Recipients))

	for _, recipient := range plan.Recipients {
		payloadText := textOut
		if recipient.ScreenReader && len(textOutSR) > 0 {
			payloadText = textOutSR
		}

		if !plan.SkipLineRefresh {
			payloadText = prefix + payloadText
		}

		deliveries = append(deliveries, transport.Delivery{
			ConnectionIds: []connections.ConnectionId{recipient.ConnectionId},
			Payload:       []byte(payloadText),
		})
	}

	transport.Queue(deliveries...)

	return events.Continue
}

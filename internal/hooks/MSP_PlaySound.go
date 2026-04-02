package hooks

import (
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/transport"
	"github.com/GoMudEngine/GoMud/internal/users"
)

//
// Checks for quests on the item
//

type mspPayloadPlan struct {
	ConnectionId connections.ConnectionId
	Payloads     [][]byte
}

func (m mspPayloadPlan) Type() string { return `MSPPayloadPlan` }

func PlaySound_PlanDelivery(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.MSP)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "MSP", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId < 1 {
		return events.Continue
	}

	if evt.SoundFile == `` {
		return events.Continue
	}

	if user := users.GetByUserId(evt.UserId); user != nil {
		connId := user.ConnectionId()
		if connId == 0 {
			return events.Continue
		}

		isWebsocket := connections.IsWebsocket(connId)
		wrapPayload := func(msg []byte) []byte {
			if isWebsocket {
				return msg
			}
			return term.MspCommand.BytesWithPayload(msg)
		}

		payloads := [][]byte{}

		if evt.SoundType == `MUSIC` {

			if user.LastMusic != evt.SoundFile {

				msg := []byte("!!MUSIC(Off)")
				payloads = append(payloads, wrapPayload(msg))
			}

			user.LastMusic = evt.SoundFile

			msg := []byte("!!MUSIC(" + evt.SoundFile + " V=" + strconv.Itoa(evt.Volume) + " L=-1 C=1)")
			payloads = append(payloads, wrapPayload(msg))
		} else {

			msg := []byte("!!SOUND(" + evt.SoundFile + " T=" + evt.Category + " V=" + strconv.Itoa(evt.Volume) + ")")
			payloads = append(payloads, wrapPayload(msg))

		}

		if len(payloads) > 0 {
			events.AddToQueue(mspPayloadPlan{
				ConnectionId: connId,
				Payloads:     payloads,
			})
		}
	}

	return events.Continue
}

func PlaySound(e events.Event) events.ListenerReturn {
	plan, typeOk := e.(mspPayloadPlan)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "MSPPayloadPlan", "Actual Type", e.Type())
		return events.Cancel
	}

	if plan.ConnectionId == 0 || len(plan.Payloads) == 0 {
		return events.Continue
	}

	deliveries := make([]transport.Delivery, 0, len(plan.Payloads))
	for _, payload := range plan.Payloads {
		deliveries = append(deliveries, transport.Delivery{
			ConnectionIds: []connections.ConnectionId{plan.ConnectionId},
			Payload:       append([]byte(nil), payload...),
		})
	}

	transport.Queue(deliveries...)

	return events.Continue
}

package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/transport"
	"github.com/GoMudEngine/GoMud/internal/users"
)

type promptDeliveryPlan struct {
	ConnectionId connections.ConnectionId
	Prompt       string
}

func (p promptDeliveryPlan) Type() string { return `PromptDeliveryPlan` }

func RedrawPrompt_PlanDelivery(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.RedrawPrompt)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RedrawPrompt", "Actual Type", e.Type())
		return events.Cancel
	}

	if user := users.GetByUserId(evt.UserId); user != nil {

		newCmdPrompt := user.GetCommandPrompt()

		if evt.OnlyIfChanged {

			oldCmdPrompt := user.GetTempData(`cmdprompt`)

			// If the prompt hasn't changed, skip redrawing
			if oldCmdPrompt != nil && oldCmdPrompt.(string) == newCmdPrompt {
				return events.Continue
			}

			// save the new prompt for next time we want to check
			user.SetTempData(`cmdprompt`, newCmdPrompt)

		}

		events.AddToQueue(promptDeliveryPlan{
			ConnectionId: user.ConnectionId(),
			Prompt:       newCmdPrompt,
		})

	}

	return events.Continue
}

// Checks whether their level is too high for a guide
func RedrawPrompt_SendRedraw(e events.Event) events.ListenerReturn {

	plan, typeOk := e.(promptDeliveryPlan)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PromptDeliveryPlan", "Actual Type", e.Type())
		return events.Cancel
	}

	if plan.ConnectionId == 0 {
		return events.Continue
	}

	transport.Queue(transport.Delivery{
		ConnectionIds: []connections.ConnectionId{plan.ConnectionId},
		Payload:       []byte(templates.AnsiParse(plan.Prompt)),
	})

	return events.Continue
}

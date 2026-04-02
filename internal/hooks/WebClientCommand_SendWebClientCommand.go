package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/transport"
)

// Checks whether their level is too high for a guide
func WebClientCommand_SendWebClientCommand(e events.Event) events.ListenerReturn {

	cmd, typeOk := e.(events.WebClientCommand)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "WebClientCommand", "Actual Type", e.Type())
		return events.Continue
	}

	if !connections.IsWebsocket(cmd.ConnectionId) {
		return events.Cancel
	}

	transport.Queue(transport.Delivery{
		ConnectionIds: []connections.ConnectionId{cmd.ConnectionId},
		Payload:       []byte(cmd.Text),
	})

	return events.Continue
}

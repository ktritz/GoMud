package transport

import (
	"slices"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

const eventPriority = -1000

type TransportEvent interface {
	events.Event
	TransportEvent() bool
}

type Delivery struct {
	ConnectionIds []connections.ConnectionId
	Payload       []byte
}

type Batch struct {
	Deliveries []Delivery
}

func (b Batch) Type() string { return "TransportBatch" }

func (b Batch) TransportEvent() bool { return true }

func RegisterListeners() {
	events.RegisterTransportListener(Batch{}, handleBatch)
}

func Queue(deliveries ...Delivery) {
	if len(deliveries) == 0 {
		return
	}

	batch := Batch{Deliveries: make([]Delivery, 0, len(deliveries))}
	for _, delivery := range deliveries {
		if len(delivery.Payload) == 0 || len(delivery.ConnectionIds) == 0 {
			continue
		}

		batch.Deliveries = append(batch.Deliveries, Delivery{
			ConnectionIds: slices.Clone(delivery.ConnectionIds),
			Payload:       slices.Clone(delivery.Payload),
		})
	}

	if len(batch.Deliveries) == 0 {
		return
	}

	events.AddToQueue(batch, eventPriority)
}

func handleBatch(e events.Event) events.ListenerReturn {
	batch, ok := e.(Batch)
	if !ok {
		mudlog.Error("Transport", "Expected Type", "TransportBatch", "Actual Type", e.Type())
		return events.Cancel
	}

	for _, delivery := range batch.Deliveries {
		connections.SendTo(delivery.Payload, delivery.ConnectionIds...)
	}

	return events.Continue
}

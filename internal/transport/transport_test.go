package transport

import (
	"net"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func drainEvents() {
	for events.ProcessSingle(func(e events.Event) events.ListenerReturn {
		return events.Continue
	}) {
	}
}

func TestTransportBatchDeliversQueuedPayload(t *testing.T) {
	mudlog.SetupLogger(nil, "debug", "", false)
	events.ClearListeners()
	drainEvents()
	RegisterListeners()

	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		clientConn.Close()
		connections.Cleanup()
		events.ClearListeners()
	})

	cd := connections.Add(serverConn, nil)

	Queue(Delivery{
		ConnectionIds: []connections.ConnectionId{cd.ConnectionId()},
		Payload:       []byte("transport hello"),
	})

	processed := make(chan bool, 1)
	go func() {
		processed <- events.ProcessSingle(func(e events.Event) events.ListenerReturn {
			return events.DoListeners(e)
		})
	}()

	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}

	buf := make([]byte, 64)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if got := string(buf[:n]); got != "transport hello" {
		t.Fatalf("Read() = %q, want %q", got, "transport hello")
	}

	if ok := <-processed; !ok {
		t.Fatal("ProcessSingle() = false, want queued transport batch")
	}
}

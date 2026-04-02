package gmcp

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/transport"
	lru "github.com/hashicorp/golang-lru/v2"
)

var registerTransportListenersOnce sync.Once

func resetGMCPForTest(t *testing.T) {
	t.Helper()

	mudlog.SetupLogger(nil, "debug", "", false)

	oldCache := gmcpModule.cache
	newCache, err := lru.New[uint64, GMCPSettings](128)
	if err != nil {
		t.Fatalf("lru.New() error = %v", err)
	}
	gmcpModule.cache = newCache
	registerTransportListenersOnce.Do(transport.RegisterListeners)

	connections.Cleanup()

	t.Cleanup(func() {
		connections.Cleanup()
		gmcpModule.cache = oldCache
	})
}

func addGMCPTestConnection(t *testing.T) (connections.ConnectionId, net.Conn) {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	cd := connections.Add(serverConn, nil)

	t.Cleanup(func() {
		clientConn.Close()
	})

	return cd.ConnectionId(), clientConn
}

func readGMCPPayload(t *testing.T, clientConn net.Conn) []byte {
	t.Helper()

	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}

	buf := make([]byte, 256)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	return buf[:n]
}

func processQueuedTransportEvent(t *testing.T) {
	t.Helper()

	processed := make(chan bool, 1)
	go func() {
		processed <- events.ProcessSingle(events.DoListeners)
	}()

	if ok := <-processed; !ok {
		t.Fatal("ProcessSingle() = false, want queued TransportBatch event")
	}
}

func TestDispatchGMCPUsesSnapshottedConnection(t *testing.T) {
	resetGMCPForTest(t)

	connID, clientConn := addGMCPTestConnection(t)
	settings := GMCPSettings{}
	settings.GMCPAccepted = true
	gmcpModule.cache.Add(connID, settings)

	gmcpModule.dispatchGMCP(GMCPOut{
		UserId:       999,
		ConnectionId: connID,
		Module:       "Core.Test",
		Payload:      `{"ok":true}`,
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readGMCPPayload(t, clientConn)
	}()
	processQueuedTransportEvent(t)
	got := <-gotCh
	want := GmcpPayload.BytesWithPayload([]byte(`Core.Test {"ok":true}`))
	if !bytes.Equal(got, want) {
		t.Fatalf("dispatchGMCP() payload = %q, want %q", got, want)
	}
}

func TestDispatchGMCPUsesSnapshottedWebWrapper(t *testing.T) {
	resetGMCPForTest(t)

	connID, clientConn := addGMCPTestConnection(t)
	settings := GMCPSettings{}
	settings.GMCPAccepted = true
	gmcpModule.cache.Add(connID, settings)

	gmcpModule.dispatchGMCP(GMCPOut{
		UserId:        999,
		ConnectionId:  connID,
		UseWebPayload: true,
		Module:        "Core.Test",
		Payload:       `{"ok":true}`,
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readGMCPPayload(t, clientConn)
	}()
	processQueuedTransportEvent(t)
	got := <-gotCh
	want := GmcpWebPayload.BytesWithPayload([]byte(`Core.Test {"ok":true}`))
	if !bytes.Equal(got, want) {
		t.Fatalf("dispatchGMCP() web payload = %q, want %q", got, want)
	}
}

func TestMudletOutboundPlanDispatchesSnapshottedMessages(t *testing.T) {
	resetGMCPForTest(t)

	connID, clientConn := addGMCPTestConnection(t)
	settings := GMCPSettings{}
	settings.GMCPAccepted = true
	gmcpModule.cache.Add(connID, settings)

	mudlet := GMCPMudletModule{}
	mudlet.dispatchOutboundPlan(mudletOutboundPlan{
		Messages: []GMCPOut{{
			UserId:       999,
			ConnectionId: connID,
			Module:       "Core.Test",
			Payload:      `{"ok":true}`,
		}},
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readGMCPPayload(t, clientConn)
	}()
	processQueuedTransportEvent(t)
	got := <-gotCh
	want := GmcpPayload.BytesWithPayload([]byte(`Core.Test {"ok":true}`))
	if !bytes.Equal(got, want) {
		t.Fatalf("dispatchOutboundPlan() payload = %q, want %q", got, want)
	}
}

func TestDiscordStatusRequestDispatchesPreparedMessages(t *testing.T) {
	resetGMCPForTest(t)

	connID, clientConn := addGMCPTestConnection(t)
	settings := GMCPSettings{}
	settings.GMCPAccepted = true
	gmcpModule.cache.Add(connID, settings)

	mudlet := GMCPMudletModule{}
	mudlet.discordStatusRequestHandler(GMCPDiscordStatusRequest{
		UserId: 999,
		Messages: []GMCPOut{{
			UserId:       999,
			ConnectionId: connID,
			Module:       "Core.Test",
			Payload:      `{"ok":true}`,
		}},
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readGMCPPayload(t, clientConn)
	}()
	processQueuedTransportEvent(t)
	got := <-gotCh
	want := GmcpPayload.BytesWithPayload([]byte(`Core.Test {"ok":true}`))
	if !bytes.Equal(got, want) {
		t.Fatalf("discordStatusRequestHandler() payload = %q, want %q", got, want)
	}
}

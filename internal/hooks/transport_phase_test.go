package hooks

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/transport"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var registerTransportListenersOnce sync.Once

func resetTransportHookTest(t *testing.T) {
	t.Helper()

	mudlog.SetupLogger(nil, "debug", "", false)
	registerTransportListenersOnce.Do(transport.RegisterListeners)
	connections.Cleanup()

	t.Cleanup(connections.Cleanup)
}

func addHookTestConnection(t *testing.T) (connections.ConnectionId, net.Conn) {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	cd := connections.Add(serverConn, nil)

	t.Cleanup(func() {
		clientConn.Close()
	})

	return cd.ConnectionId(), clientConn
}

func processQueuedTransportBatch(t *testing.T) {
	t.Helper()

	processed := make(chan bool, 1)
	go func() {
		processed <- events.ProcessSingle(events.DoListeners)
	}()

	if ok := <-processed; !ok {
		t.Fatal("ProcessSingle() = false, want queued TransportBatch event")
	}
}

func readHookPayload(t *testing.T, clientConn net.Conn) []byte {
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

func TestMessageSendMessageUsesSnapshottedRecipients(t *testing.T) {
	resetTransportHookTest(t)

	connID, clientConn := addHookTestConnection(t)
	text := `<ansi fg="red">danger</ansi>`

	Message_SendMessage(messageDeliveryPlan{
		Text: text,
		Recipients: []messageRecipient{{
			ConnectionId: connID,
			ScreenReader: true,
		}},
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readHookPayload(t, clientConn)
	}()
	processQueuedTransportBatch(t)
	got := <-gotCh
	wantText := util.StripCharsForScreenReaders(templates.AnsiParse(text))
	want := []byte(term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + wantText)
	if !bytes.Equal(got, want) {
		t.Fatalf("Message_SendMessage() payload = %q, want %q", got, want)
	}
}

func TestBroadcastSendToAllUsesScreenReaderVariant(t *testing.T) {
	resetTransportHookTest(t)

	connID, clientConn := addHookTestConnection(t)

	Broadcast_SendToAll(broadcastDeliveryPlan{
		Text:             `<ansi fg="green">formatted</ansi>`,
		TextScreenReader: `plain`,
		Recipients: []broadcastRecipient{{
			ConnectionId: connID,
			ScreenReader: true,
		}},
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readHookPayload(t, clientConn)
	}()
	processQueuedTransportBatch(t)
	got := <-gotCh
	want := []byte(term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + templates.AnsiParse(`plain`))
	if !bytes.Equal(got, want) {
		t.Fatalf("Broadcast_SendToAll() payload = %q, want %q", got, want)
	}
}

func TestRedrawPromptSendRedrawUsesPromptPlan(t *testing.T) {
	resetTransportHookTest(t)

	connID, clientConn := addHookTestConnection(t)

	RedrawPrompt_SendRedraw(promptDeliveryPlan{
		ConnectionId: connID,
		Prompt:       `<ansi fg="yellow">hp:100</ansi>`,
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readHookPayload(t, clientConn)
	}()
	processQueuedTransportBatch(t)
	got := <-gotCh
	want := []byte(templates.AnsiParse(`<ansi fg="yellow">hp:100</ansi>`))
	if !bytes.Equal(got, want) {
		t.Fatalf("RedrawPrompt_SendRedraw() payload = %q, want %q", got, want)
	}
}

func TestPlaySoundUsesSnapshottedWrapper(t *testing.T) {
	resetTransportHookTest(t)

	connID, clientConn := addHookTestConnection(t)

	PlaySound(mspPayloadPlan{
		ConnectionId: connID,
		Payloads:     [][]byte{term.MspCommand.BytesWithPayload([]byte("!!SOUND(chime T=other V=50)"))},
	})
	gotCh := make(chan []byte, 1)
	go func() {
		gotCh <- readHookPayload(t, clientConn)
	}()
	processQueuedTransportBatch(t)
	got := <-gotCh
	want := term.MspCommand.BytesWithPayload([]byte("!!SOUND(chime T=other V=50)"))
	if !bytes.Equal(got, want) {
		t.Fatalf("PlaySound() payload = %q, want %q", got, want)
	}
}

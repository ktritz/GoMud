package connections

import (
	"net"
	"slices"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func resetConnectionsForTest(t *testing.T) {
	t.Helper()
	mudlog.SetupLogger(nil, "debug", "", false)

	lock.Lock()
	oldConnections := netConnections
	oldConnectCounter := connectCounter
	oldDisconnectCounter := disconnectCounter
	oldAsyncSender := asyncSender
	netConnections = map[ConnectionId]*ConnectionDetails{}
	connectCounter = 0
	disconnectCounter = 0
	asyncSender = nil
	lock.Unlock()

	if oldAsyncSender != nil {
		oldAsyncSender.stopAndWait()
	}

	t.Cleanup(func() {
		lock.Lock()
		for _, cd := range netConnections {
			cd.Close()
		}
		netConnections = oldConnections
		connectCounter = oldConnectCounter
		disconnectCounter = oldDisconnectCounter
		currentAsyncSender := asyncSender
		asyncSender = nil
		lock.Unlock()

		if currentAsyncSender != nil {
			currentAsyncSender.stopAndWait()
		}
	})
}

func addTestConnection(t *testing.T, id ConnectionId) net.Conn {
	t.Helper()

	serverConn, clientConn := net.Pipe()

	lock.Lock()
	netConnections[id] = NewConnectionDetails(id, serverConn, nil, nil)
	lock.Unlock()

	return clientConn
}

func TestGetAllConnectionIdsHasNoZeroPadding(t *testing.T) {
	resetConnectionsForTest(t)
	clientA := addTestConnection(t, 11)
	clientB := addTestConnection(t, 22)
	t.Cleanup(func() {
		clientA.Close()
		clientB.Close()
	})

	ids := GetAllConnectionIds()
	if len(ids) != 2 {
		t.Fatalf("GetAllConnectionIds() len = %d, want 2", len(ids))
	}

	if slices.Contains(ids, ConnectionId(0)) {
		t.Fatalf("GetAllConnectionIds() = %v, should not contain zero id", ids)
	}
}

func TestKickRemovesConnectionFromRegistry(t *testing.T) {
	resetConnectionsForTest(t)
	clientConn := addTestConnection(t, 7)
	t.Cleanup(func() {
		clientConn.Close()
	})

	if err := Kick(7, "test"); err != nil {
		t.Fatalf("Kick() error = %v", err)
	}

	if got := Get(7); got != nil {
		t.Fatalf("Get(7) = %#v, want nil after Kick()", got)
	}
}

func TestSendToRemovesBrokenConnection(t *testing.T) {
	resetConnectionsForTest(t)
	clientConn := addTestConnection(t, 9)
	clientConn.Close()

	SendTo([]byte("hello"), 9)

	if got := Get(9); got != nil {
		t.Fatalf("Get(9) = %#v, want nil after failed SendTo()", got)
	}
}

func TestSendToQueuedWritesToConnection(t *testing.T) {
	resetConnectionsForTest(t)
	clientConn := addTestConnection(t, 13)
	t.Cleanup(func() {
		clientConn.Close()
	})

	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}

	SendToQueued([]byte("queued hello"), 13)

	buf := make([]byte, 32)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if got := string(buf[:n]); got != "queued hello" {
		t.Fatalf("Read() = %q, want %q", got, "queued hello")
	}
}

func TestSendToQueuedRemovesBrokenConnection(t *testing.T) {
	resetConnectionsForTest(t)
	clientConn := addTestConnection(t, 17)
	clientConn.Close()

	SendToQueued([]byte("hello"), 17)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if Get(17) == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := Get(17); got != nil {
		t.Fatalf("Get(17) = %#v, want nil after failed queued send", got)
	}
}

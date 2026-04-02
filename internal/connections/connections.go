package connections

import (
	"errors"
	"net"
	"os"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/gorilla/websocket"
)

const ReadBufferSize = 1024

const (
	asyncSendWorkerCount = 8
	asyncSendQueueSize   = 256
)

type ConnectionId = uint64

var (

	//
	// Mutex
	//
	lock sync.RWMutex = sync.RWMutex{}
	//
	// Counters
	//
	connectCounter    uint64 = 0 // a counter for each time a connection is accepted
	disconnectCounter uint64 = 0 // a counter for each tim ea connection is dropped
	//
	// Track connections
	//
	netConnections map[ConnectionId]*ConnectionDetails = map[ConnectionId]*ConnectionDetails{} // a mapping of unique id's to connections
	//
	// Channel to send a shutdown signal to
	//
	shutdownChannel chan os.Signal // channel to receive shutdown signals
	asyncSender     *sendDispatcher
)

type connectionSnapshot struct {
	id ConnectionId
	cd *ConnectionDetails
}

type outboundMessage struct {
	id      ConnectionId
	payload []byte
}

type sendDispatcher struct {
	queues []chan outboundMessage
	stop   chan struct{}
	wg     sync.WaitGroup
}

func newSendDispatcher() *sendDispatcher {
	d := &sendDispatcher{
		queues: make([]chan outboundMessage, asyncSendWorkerCount),
		stop:   make(chan struct{}),
	}

	for i := range d.queues {
		d.queues[i] = make(chan outboundMessage, asyncSendQueueSize)
		d.wg.Add(1)
		go d.runQueue(d.queues[i])
	}

	return d
}

func (d *sendDispatcher) runQueue(queue <-chan outboundMessage) {
	defer d.wg.Done()

	for {
		select {
		case <-d.stop:
			return
		case msg := <-queue:
			SendTo(msg.payload, msg.id)
		}
	}
}

func (d *sendDispatcher) stopAndWait() {
	close(d.stop)
	d.wg.Wait()
}

func getSendDispatcher() *sendDispatcher {
	lock.Lock()
	defer lock.Unlock()

	if asyncSender == nil {
		asyncSender = newSendDispatcher()
	}

	return asyncSender
}

func cloneBytes(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}

	out := make([]byte, len(b))
	copy(out, b)
	return out
}

func SignalShutdown(s os.Signal) {
	if shutdownChannel != nil {
		shutdownChannel <- s
	}
}

func Add(conn net.Conn, wsConn *websocket.Conn) *ConnectionDetails {

	lock.Lock()
	defer lock.Unlock()

	connectCounter++

	connDetails := NewConnectionDetails(
		connectCounter,
		conn,
		wsConn,
		nil, // use default settings for now TODO: add into overall config pattern?
	)

	netConnections[connDetails.ConnectionId()] = connDetails

	// return the unique ID to find this connection later
	return connDetails
}

// Returns the total number of connections
func Get(id ConnectionId) *ConnectionDetails {
	lock.RLock()
	defer lock.RUnlock()

	return netConnections[id]
}

func IsWebsocket(id ConnectionId) bool {
	lock.RLock()
	defer lock.RUnlock()

	if cd, ok := netConnections[id]; ok {
		return cd.IsWebSocket()
	}

	return false
}

func GetAllConnectionIds() []ConnectionId {

	lock.RLock()
	defer lock.RUnlock()

	ids := make([]ConnectionId, 0, len(netConnections))

	for id := range netConnections {
		ids = append(ids, id)
	}

	return ids
}

func Cleanup() {
	for _, id := range GetAllConnectionIds() {
		Remove(id)
	}

	lock.Lock()
	dispatcher := asyncSender
	asyncSender = nil
	lock.Unlock()

	if dispatcher != nil {
		dispatcher.stopAndWait()
	}
}

func Kick(id ConnectionId, reason string) (err error) {

	lock.Lock()
	defer lock.Unlock()

	// Try to retrieve the value
	if cd, ok := netConnections[id]; ok {

		// close the connection, no longer useful.
		cd.Close()
		// keep track of the number of disconnects
		disconnectCounter++
		// remove the connection from the map
		delete(netConnections, id)
		mudlog.Info("connection kicked", "connectionId", id, "remoteAddr", cd.RemoteAddr().String(), `reason`, reason)

		return nil

	}

	return errors.New("connection not found")
}

func Remove(id ConnectionId) (err error) {

	lock.Lock()
	defer lock.Unlock()

	// Try to retrieve the value
	if cd, ok := netConnections[id]; ok {

		// close the connection, no longer useful.
		cd.Close()
		// keep track of the number of disconnects
		disconnectCounter++
		// Remove the entry
		delete(netConnections, id)

		return nil

	}

	return errors.New("connection not found")
}

func Broadcast(colorizedText []byte, skipConnectionIds ...ConnectionId) []ConnectionId {
	lock.RLock()
	snapshots := make([]connectionSnapshot, 0, len(netConnections))
	for id, cd := range netConnections {
		snapshots = append(snapshots, connectionSnapshot{id: id, cd: cd})
	}
	lock.RUnlock()

	removeIds := []ConnectionId{}
	sentToIds := []ConnectionId{}

	for _, snapshot := range snapshots {
		id, cd := snapshot.id, snapshot.cd

		if cd.State() == Login {
			continue
		}

		if len(skipConnectionIds) > 0 {
			skip := false
			for _, cId := range skipConnectionIds {
				if cId == id {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// Write the message to the connection
		var err error

		_, err = cd.Write(colorizedText)

		if err != nil {
			mudlog.Warn("Broadcast()", "connectionId", id, "remoteAddr", cd.RemoteAddr().String(), "error", err)
			// Remove from the connections
			removeIds = append(removeIds, id)
			continue
		}

		sentToIds = append(sentToIds, id)
	}

	for _, id := range removeIds {
		Remove(id)
	}

	return sentToIds
}

func SendTo(b []byte, ids ...ConnectionId) {
	lock.RLock()
	snapshots := make([]connectionSnapshot, 0, len(ids))
	for _, id := range ids {
		if cd, ok := netConnections[id]; ok {
			snapshots = append(snapshots, connectionSnapshot{id: id, cd: cd})
		}
	}
	lock.RUnlock()

	removeIds := []ConnectionId{}

	sentCt := 0
	// iterate through all provided id's and attempt to send

	for _, snapshot := range snapshots {
		id, cd := snapshot.id, snapshot.cd
		if _, err := cd.Write(b); err != nil {
			mudlog.Warn("SendTo()", "connectionId", id, "remoteAddr", cd.RemoteAddr().String(), "error", err)
			// Remove from the connections
			removeIds = append(removeIds, id)
			continue
		}

		sentCt++
	}

	if sentCt < 1 {
		//mudlog.Info("message sent to nobody", "message", strings.Replace(string(b), "\033", "ESC", -1))
	}

	for _, id := range removeIds {
		Remove(id)
	}
}

func SendToQueued(b []byte, ids ...ConnectionId) {
	if len(b) == 0 || len(ids) == 0 {
		return
	}

	dispatcher := getSendDispatcher()
	payload := cloneBytes(b)

	for _, id := range ids {
		msg := outboundMessage{
			id:      id,
			payload: payload,
		}

		queue := dispatcher.queues[int(id%ConnectionId(len(dispatcher.queues)))]

		select {
		case queue <- msg:
		default:
			mudlog.Warn("SendToQueued()", "connectionId", id, "warning", "send queue full, falling back to synchronous write")
			SendTo(payload, id)
		}
	}
}

// make this more efficient later
func ActiveConnectionCount() int {
	lock.RLock()
	defer lock.RUnlock()

	return len(netConnections)
}

// make this more efficient later
func SetShutdownChan(osSignalChan chan os.Signal) {
	lock.Lock()
	defer lock.Unlock()

	if shutdownChannel != nil {
		panic("Can't set shutdown channel a second time!")
	}
	shutdownChannel = osSignalChan
}

func Stats() (connections uint64, disconnections uint64) {
	lock.RLock()
	defer lock.RUnlock()

	return connectCounter, disconnectCounter
}

func GetClientSettings(id ConnectionId) ClientSettings {
	lock.RLock()
	defer lock.RUnlock()

	if cd, ok := netConnections[id]; ok {
		return cd.clientSettings
	}

	return ClientSettings{}
}

func OverwriteClientSettings(id ConnectionId, cs ClientSettings) {
	lock.Lock()
	defer lock.Unlock()

	if cd, ok := netConnections[id]; ok {
		cd.clientSettings = cs
	}
}

package events

import (
	"container/heap"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/util"
)

type EventType string

var (
	qLock     = sync.Mutex{}
	allQueues = map[string]*Queue[Event]{}

	requeues = []requeue{}

	globalQueue  priorityQueue
	orderCounter uint64                      // global counter to maintain insertion order.
	uniqueMap    = make(map[string]struct{}) // map to enforce uniqueness

	eventDebugging bool
	eventReady     = make(chan struct{}, 1)
)

type requeue struct {
	evt      Event
	priority int
}

// Event is the common interface for events.
type Event interface {
	Type() string
}

// Generic events are mostly used for plugins
type GenericEvent interface {
	Event
	Data(name string) any
}

// prioritizedEvent wraps an Event with a priority and an order field.
type prioritizedEvent struct {
	event    Event
	priority int    // Lower numbers indicate higher priority. Default is 0.
	order    uint64 // Used to preserve FIFO order among events with the same priority.
}

// UniqueEvent is implemented by events that should be unique in the queue.
type uniqueEvent interface {
	Event
	UniqueID() string
}

// PriorityQueue implements heap.Interface for *PrioritizedEvent.
type priorityQueue []*prioritizedEvent

func (pq priorityQueue) Len() int { return len(pq) }

// Less returns true if element i has a higher priority than element j.
// Here, "higher priority" means a lower numeric value. If priorities are equal,
// the one with the lower order (i.e. inserted earlier) is considered higher.
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].priority == pq[j].priority {
		return pq[i].order < pq[j].order
	}
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*prioritizedEvent)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Enqueue adds an event to the global queue.
// The caller can optionally pass a priority value.
// If omitted, the default priority is 0.
func AddToQueue(e Event, priority ...int) {

	qLock.Lock()
	defer qLock.Unlock()

	prio := 0
	if len(priority) > 0 {
		prio = priority[0]
	}

	// Check for uniqueness if the event implements UniqueEvent.
	if ue, ok := e.(uniqueEvent); ok {
		uid := ue.UniqueID()
		// If we already have an entry for this uniqueID, skip it!
		if _, exists := uniqueMap[uid]; exists {
			return
		}

		uniqueMap[ue.UniqueID()] = struct{}{}
	}

	orderCounter++
	pe := &prioritizedEvent{
		event:    e,
		priority: prio,
		order:    orderCounter,
	}

	if eventDebugging {
		fmt.Println(`events.AddToQueue`, "type:", e.Type(), `priority`, prio, `order`, orderCounter)
	}

	heap.Push(&globalQueue, pe)
	signalProcessor()
}

// Same as AddToQueue but avoids a mutex lock for optimization purposes
// Should only be used when a mutex lock is already held
func reAddToQueue(e Event, priority ...int) {

	prio := 0
	if len(priority) > 0 {
		prio = priority[0]
	}

	// Check for uniqueness if the event implements UniqueEvent.
	if ue, ok := e.(uniqueEvent); ok {
		uid := ue.UniqueID()
		// If we already have an entry for this uniqueID, skip it!
		if _, exists := uniqueMap[uid]; exists {
			return
		}

		uniqueMap[ue.UniqueID()] = struct{}{}
	}

	orderCounter++
	pe := &prioritizedEvent{
		event:    e,
		priority: prio,
		order:    orderCounter,
	}

	if eventDebugging {
		fmt.Println(`events.reAddToQueue`, "type:", e.Type(), `priority:`, prio, `order:`, orderCounter)
	}

	heap.Push(&globalQueue, pe)
	signalProcessor()
}

func addToRequeue(e Event, priority ...int) {
	qLock.Lock()
	defer qLock.Unlock()

	prio := 0
	if len(priority) > 0 {
		prio = priority[0]
	}
	requeues = append(requeues, requeue{
		evt:      e,
		priority: prio,
	})
	signalProcessor()
}

func activateRequeuesLocked() {
	for _, itm := range requeues {
		reAddToQueue(itm.evt, itm.priority)
	}
	requeues = requeues[:0]
}

func popNextEventLocked() *prioritizedEvent {
	if globalQueue.Len() < 1 {
		return nil
	}

	pe := heap.Pop(&globalQueue).(*prioritizedEvent)
	if ue, ok := pe.event.(uniqueEvent); ok {
		delete(uniqueMap, ue.UniqueID())
	}

	return pe
}

// ActivateRequeues moves deferred events back into the main queue.
// Should be called once per EventLoop invocation, not per-event.
func ActivateRequeues() {
	qLock.Lock()
	defer qLock.Unlock()
	activateRequeuesLocked()
}

func signalProcessor() {
	select {
	case eventReady <- struct{}{}:
	default:
	}
}

func NotifyChan() <-chan struct{} {
	return eventReady
}

func HasPending() bool {
	qLock.Lock()
	defer qLock.Unlock()

	return globalQueue.Len() > 0 || len(requeues) > 0
}

// ProcessEvents runs the event loop until the queue is empty.
// It processes events one at a time in order of priority.
// Any events enqueued (even from within a handler) will be picked up in order.
func ProcessEvents() {

	// Since this is intended to run frequently and quickly
	// Only sample the runtime 1 in 100 times
	eventCounter := 0
	if eventDebugging || rand.Intn(100) == 0 {
		start := time.Now()
		defer func() {
			util.TrackTime(`events.ProcessEvents()`, time.Since(start).Seconds())
			if eventDebugging {
				if time.Since(start).Seconds() > 0.00125 {
					fmt.Println(`events.ProcessEvents`, "events handled:", eventCounter, "time taken:", time.Since(start).Seconds())
				}
			}
		}()
	}

	qLock.Lock()

	var evtResult ListenerReturn
	for {
		pe := popNextEventLocked()
		if pe == nil {
			break
		}

		if eventDebugging {
			eventCounter++
			fmt.Println(`events.ProcessEvents`, "type:", pe.event.Type(), `remain:`, globalQueue.Len())
		}

		qLock.Unlock()

		evtResult = DoListeners(pe.event)
		if evtResult == CancelAndRequeue {
			addToRequeue(pe.event, pe.priority)
		}

		qLock.Lock()

	}

	qLock.Unlock()
}

func ProcessSingle(dispatch func(Event) ListenerReturn) bool {
	qLock.Lock()
	pe := popNextEventLocked()
	qLock.Unlock()

	if pe == nil {
		return false
	}

	evtResult := dispatch(pe.event)
	if evtResult == CancelAndRequeue {
		addToRequeue(pe.event, pe.priority)
	}

	return true
}

func SetDebug(on bool) {
	qLock.Lock()
	defer qLock.Unlock()
	eventDebugging = on
}

// Initialize the priority queue.
func init() {
	heap.Init(&globalQueue)
}

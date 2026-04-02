package events

import (
	"container/heap"
	"testing"
	"time"
)

type runtimeTestEvent struct{}

func (runtimeTestEvent) Type() string {
	return "runtime-test"
}

func resetEventStateForTest(t *testing.T) {
	t.Helper()

	ClearListeners()

	qLock.Lock()
	allQueues = map[string]*Queue[Event]{}
	requeues = nil
	globalQueue = priorityQueue{}
	heap.Init(&globalQueue)
	orderCounter = 0
	uniqueMap = make(map[string]struct{})
	eventDebugging = false
	qLock.Unlock()

	for {
		select {
		case <-NotifyChan():
		default:
			return
		}
	}
}

func TestDoListenersAllowsUnregisterDuringDispatch(t *testing.T) {
	resetEventStateForTest(t)

	callCount := 0
	var id ListenerId
	id = RegisterListener(runtimeTestEvent{}, func(Event) ListenerReturn {
		callCount++
		if !UnregisterListener(runtimeTestEvent{}, id) {
			t.Fatalf("UnregisterListener() = false, want true")
		}
		return Continue
	})

	done := make(chan struct{})
	go func() {
		DoListeners(runtimeTestEvent{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("DoListeners() blocked when listener unregistered itself")
	}

	DoListeners(runtimeTestEvent{})

	if callCount != 1 {
		t.Fatalf("listener callCount = %d, want 1 after self-unregister", callCount)
	}
}

func TestAddToQueueSignalsProcessor(t *testing.T) {
	resetEventStateForTest(t)

	AddToQueue(runtimeTestEvent{})

	select {
	case <-NotifyChan():
	case <-time.After(time.Second):
		t.Fatal("AddToQueue() did not signal processor")
	}
}

func TestRequeueSignalsProcessor(t *testing.T) {
	resetEventStateForTest(t)

	callCount := 0
	RegisterListener(runtimeTestEvent{}, func(Event) ListenerReturn {
		callCount++
		if callCount == 1 {
			return CancelAndRequeue
		}
		return Continue
	})

	AddToQueue(runtimeTestEvent{})

	select {
	case <-NotifyChan():
	case <-time.After(time.Second):
		t.Fatal("initial AddToQueue() did not signal processor")
	}

	ProcessEvents()

	select {
	case <-NotifyChan():
	case <-time.After(time.Second):
		t.Fatal("CancelAndRequeue did not signal processor")
	}

	ProcessEvents()

	if callCount != 2 {
		t.Fatalf("listener callCount = %d, want 2 after requeue", callCount)
	}
}

func TestRunListenersForPhaseSeparatesSimulationAndTransport(t *testing.T) {
	resetEventStateForTest(t)

	order := []string{}

	RegisterListener(runtimeTestEvent{}, func(Event) ListenerReturn {
		order = append(order, "simulation")
		return Continue
	})
	RegisterTransportListener(runtimeTestEvent{}, func(Event) ListenerReturn {
		order = append(order, "transport")
		return Continue
	})

	if result, found := RunListenersForPhase(runtimeTestEvent{}, false); result != Continue || !found {
		t.Fatalf("RunListenersForPhase(simulation) = (%v, %v), want (Continue, true)", result, found)
	}

	if len(order) != 1 || order[0] != "simulation" {
		t.Fatalf("simulation phase order = %v, want [simulation]", order)
	}

	if result, found := RunListenersForPhase(runtimeTestEvent{}, true); result != Continue || !found {
		t.Fatalf("RunListenersForPhase(transport) = (%v, %v), want (Continue, true)", result, found)
	}

	if len(order) != 2 || order[1] != "transport" {
		t.Fatalf("combined phase order = %v, want simulation then transport", order)
	}
}

func TestPlayerSpawnCarriesSnapshotData(t *testing.T) {
	evt := PlayerSpawn{
		UserId: 1,
		OnlinePlayers: []OnlinePlayerSnapshot{
			{
				UserId: 7,
				Name:   "Alice",
				Role:   "admin",
				Level:  12,
			},
		},
	}

	if len(evt.OnlinePlayers) != 1 {
		t.Fatalf("len(PlayerSpawn.OnlinePlayers) = %d, want 1", len(evt.OnlinePlayers))
	}

	if evt.OnlinePlayers[0].Name != "Alice" || evt.OnlinePlayers[0].Level != 12 {
		t.Fatalf("PlayerSpawn.OnlinePlayers[0] = %+v, want snapshot data preserved", evt.OnlinePlayers[0])
	}
}

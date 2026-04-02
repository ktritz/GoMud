package discord

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func resetDiscordClientState(t *testing.T) {
	t.Helper()
	mudlog.SetupLogger(nil, "debug", "", false)

	oldWebhookURL := WebhookUrl
	oldInitialized := initialized
	oldWaitUntil := waitUntil
	oldHTTP := discordHTTP
	oldQueue := sendQueue

	WebhookUrl = ""
	initialized = false
	waitUntil = time.Time{}
	discordHTTP = &http.Client{}
	sendQueue = nil
	events.ClearListeners()
	for events.ProcessSingle(func(e events.Event) events.ListenerReturn {
		return events.Continue
	}) {
	}

	t.Cleanup(func() {
		WebhookUrl = oldWebhookURL
		initialized = oldInitialized
		waitUntil = oldWaitUntil
		discordHTTP = oldHTTP
		sendQueue = oldQueue
		events.ClearListeners()
	})
}

func TestSendNowBacksOffOnInvalidRequest(t *testing.T) {
	resetDiscordClientState(t)

	WebhookUrl = "://bad-url"
	sendNow([]byte(`{}`))

	if !isRequestBackoff() {
		t.Fatal("expected request backoff after invalid request URL")
	}
}

func TestSendNowBacksOffOnNon204(t *testing.T) {
	resetDiscordClientState(t)

	WebhookUrl = "https://example.invalid/webhook"
	discordHTTP = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("bad gateway")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	sendNow([]byte(`{}`))

	if !isRequestBackoff() {
		t.Fatal("expected request backoff after non-204 response")
	}
}

func TestSendPayloadQueuesWebhookEvent(t *testing.T) {
	resetDiscordClientState(t)

	sendQueue = make(chan []byte, 1)
	initialized = true
	registerListeners()

	SendPayload(webHookPayload{Content: "hello"})

	select {
	case <-sendQueue:
		t.Fatal("SendPayload() wrote to sendQueue before webhook event was processed")
	default:
	}

	if !events.ProcessSingle(func(e events.Event) events.ListenerReturn {
		return events.DoListeners(e)
	}) {
		t.Fatal("ProcessSingle() = false, want queued DiscordWebhook event")
	}

	select {
	case payload := <-sendQueue:
		if !strings.Contains(string(payload), `"content":"hello"`) {
			t.Fatalf("queued payload = %s, want marshaled content", string(payload))
		}
	case <-time.After(time.Second):
		t.Fatal("expected DiscordWebhook processing to enqueue payload")
	}
}

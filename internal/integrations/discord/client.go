// Package discord
//
// This package provides programatic access to discord messages integrated with GoMud.
//
// References:
// https://leovoel.github.io/embed-visualizer/
// https://birdie0.github.io/discord-webhooks-guide/discord_webhook.html
// https://gist.github.com/rxaviers/7360908
package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	WebhookUrl  string
	initialized bool
	waitMutex   sync.RWMutex
	waitUntil   time.Time
	discordHTTP = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	sendQueue chan []byte
)

const (
	RequestFailureBackoffSeconds = 30
	sendQueueSize                = 64
)

// Initializes and sets the webhook so we can send messages to discord
// and registers listeners to listen for events
func Init(webhookUrl string) {
	if initialized {
		return
	}

	WebhookUrl = webhookUrl
	sendQueue = make(chan []byte, sendQueueSize)
	go sendWorker()
	registerListeners()
	initialized = true
}

func registerListeners() {
	events.RegisterTransportListener(events.PlayerSpawn{}, HandlePlayerSpawn)
	events.RegisterTransportListener(events.PlayerDespawn{}, HandlePlayerDespawn)
	events.RegisterTransportListener(events.Log{}, HandleLogs)
	events.RegisterTransportListener(events.LevelUp{}, HandleLevelup)
	events.RegisterTransportListener(events.PlayerDeath{}, HandleDeath)
	events.RegisterTransportListener(events.Broadcast{}, HandleBroadcast)
	events.RegisterTransportListener(`AuctionUpdate`, HandleAuctionUpdate)
	events.RegisterTransportListener(webhookEvent{}, handleWebhookEvent)
}

// Sends an embed message to discord which includes a colored bar to the left
// hexColor should be specified as a string in this format "#000000"
func SendRichMessage(message string, color Color) {
	if !initialized {
		mudlog.Error(`discord`, `error`, "Discord client was not initialized.")
		return
	}

	payload := webHookPayload{
		Embeds: []embed{
			{
				Description: message,
				Color:       color,
			},
		},
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		mudlog.Error(`discord`, `error`, fmt.Sprintf("Couldn't marshal discord message"))
		return
	}

	queueWebhookPayload(marshalled)

}

// Sends a simple message to discord
func SendMessage(message string) {
	if !initialized {
		mudlog.Error(`discord`, `error`, errors.New("Discord client was not initialized."))
		return
	}

	payload := webHookPayload{
		Content: message,
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		mudlog.Error(`discord`, `error`, fmt.Sprintf("Couldn't marshal discord message"))
		return
	}

	queueWebhookPayload(marshalled)
}

// Sends a simple message to discord
func SendPayload(payload webHookPayload) {
	if !initialized {
		mudlog.Error(`discord`, `error`, errors.New("Discord client was not initialized."))
		return
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		mudlog.Error(`discord`, `error`, fmt.Sprintf("Couldn't marshal discord message"))
		return
	}

	queueWebhookPayload(marshalled)
}

func queueWebhookPayload(marshalled []byte) {
	events.AddToQueue(webhookEvent{Payload: append([]byte(nil), marshalled...)})
}

func handleWebhookEvent(e events.Event) events.ListenerReturn {
	evt, ok := e.(webhookEvent)
	if !ok {
		mudlog.Error(`discord`, `error`, "Expected DiscordWebhook event")
		return events.Cancel
	}

	send(evt.Payload)
	return events.Continue
}

func send(marshalled []byte) {

	if isRequestBackoff() {
		return
	}

	payload := append([]byte(nil), marshalled...)

	select {
	case sendQueue <- payload:
	default:
		mudlog.Warn(`discord`, `error`, "Discord webhook queue full, dropping payload.")
	}

}

func sendWorker() {
	for payload := range sendQueue {
		sendNow(payload)
	}
}

func sendNow(marshalled []byte) {
	request, err := http.NewRequest("POST", WebhookUrl, bytes.NewReader(marshalled))
	if err != nil {
		doRequestBackoff()
		mudlog.Error(`discord`, `error`, err)
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response, err := discordHTTP.Do(request)
	if err != nil {

		doRequestBackoff()

		mudlog.Error(`discord`, `error`, err)
		return
	}
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)

	// Expect 204 No Content reply
	if response.StatusCode != 204 {

		doRequestBackoff()

		mudlog.Error(`discord`, `error`, fmt.Sprintf("Expected discord to send status code 204, got %v.", response.StatusCode))
		return
	}

}

// Returns true if requests are in a penalty box
func isRequestBackoff() bool {
	waitMutex.RLock()
	defer waitMutex.RUnlock()

	return waitUntil.After(time.Now())
}

// Sets a time for requests to resume
func doRequestBackoff() {
	waitMutex.Lock()
	waitUntil = time.Now().Add(RequestFailureBackoffSeconds * time.Second)
	waitMutex.Unlock()
}

func hexToColor(hexColor string) Color {
	if strings.HasPrefix(hexColor, "#") {
		hexColor = hexColor[1:]
	}

	color, err := strconv.ParseInt(hexColor, 16, 32)
	if err != nil {
		mudlog.Error(`discord`, `error`, fmt.Sprintf("Invalid color specified, expected format #000000"))
		return Default
	}
	return Color(color)
}

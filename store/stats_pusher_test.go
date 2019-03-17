package store_test

import (
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/smartcontractkit/chainlink/store"
	"github.com/stretchr/testify/require"
)

func TestWebsocketStatsPusher_StartCloseStart(t *testing.T) {
	wsserver, cleanup := cltest.NewEventWebsocketServer(t)
	defer cleanup()

	pusher := store.NewWebsocketStatsPusher(wsserver.URL)
	require.NoError(t, pusher.Start())
	cltest.CallbackOrTimeout(t, "stats pusher connects", func() {
		<-wsserver.Connected
	})
	require.NoError(t, pusher.Close())

	// restart after client disconnect
	require.NoError(t, pusher.Start())
	cltest.CallbackOrTimeout(t, "stats pusher restarts", func() {
		<-wsserver.Connected
	}, 3*time.Second)
	require.NoError(t, pusher.Close())
}

func TestWebsocketStatsPusher_ReconnectLoop(t *testing.T) {
	wsserver, cleanup := cltest.NewEventWebsocketServer(t)
	defer cleanup()

	pusher := store.NewWebsocketStatsPusher(wsserver.URL)
	require.NoError(t, pusher.Start())
	cltest.CallbackOrTimeout(t, "stats pusher connects", func() {
		<-wsserver.Connected
	})

	// reconnect after server disconnect
	wsserver.WriteCloseMessage()
	cltest.CallbackOrTimeout(t, "stats pusher reconnects", func() {
		<-wsserver.Connected
	}, 3*time.Second)
	require.NoError(t, pusher.Close())
}

func TestWebsocketStatsPusher_Send(t *testing.T) {
	wsserver, cleanup := cltest.NewEventWebsocketServer(t)
	defer cleanup()

	pusher := store.NewWebsocketStatsPusher(wsserver.URL)
	require.NoError(t, pusher.Start())
	defer pusher.Close()

	expectation := `{"hello": "world"}`
	pusher.Send([]byte(expectation))
	cltest.CallbackOrTimeout(t, "receive stats", func() {
		require.Equal(t, expectation, <-wsserver.Received)
	})
}

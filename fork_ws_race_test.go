package ccxt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Race guards for the WS path a watcher uses (watchOHLCV over a WSClient).
// They are meant to run with -race: without it most of them pass by luck.

// Append holds c.Mu; GetLimit must too, as the watcher calls it on wake-up
// while the read loop keeps appending.
func TestArrayCacheByTimestampGetLimitConcurrentAppend(t *testing.T) {
	c := NewArrayCacheByTimestamp(int64(10))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c.Append([]any{int64(1000 + i%3), 1.0, 1.0, 1.0, float64(i), 1.0})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = c.GetLimit("BTC", nil)
		}
	}()
	wg.Wait()
}

// Rows handed out by ToArray are read without c.Mu (OHLCVsFromList), so an
// update of the same timestamp must replace the row, never write into it.
func TestArrayCacheByTimestampMergedRowConcurrentReader(t *testing.T) {
	c := NewArrayCacheByTimestamp(int64(10))
	c.Append([]any{int64(1000), 1.0, 1.0, 1.0, 1.0, 1.0})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c.Append([]any{int64(1000), 1.0, 1.0, 1.0, float64(i), float64(i)})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = OHLCVsFromList(c.ToArray())
		}
	}()
	wg.Wait()
}

// The same contract checked deterministically: a snapshot keeps its values.
func TestArrayCacheByTimestampSnapshotIsNotRewritten(t *testing.T) {
	c := NewArrayCacheByTimestamp(int64(10))
	c.Append([]any{int64(1000), 1.0, 2.0, 0.5, 1.5, 10.0})
	before := c.ToArray()
	c.Append([]any{int64(1000), 1.0, 3.0, 0.5, 2.5, 20.0})

	if got := before[0].([]any)[4]; got != 1.5 {
		t.Fatalf("snapshot row was rewritten in place: close = %v, want 1.5", got)
	}
	after := c.ToArray()
	if len(after) != 1 || after[0].([]any)[4] != 2.5 {
		t.Fatalf("update of the same timestamp not applied: %v", after)
	}
	// a later update must still find the row through the hashmap
	c.Append([]any{int64(1000), 1.0, 3.0, 0.5, 3.5, 30.0})
	if last := c.ToArray(); len(last) != 1 || last[0].([]any)[4] != 3.5 {
		t.Fatalf("second update of the same timestamp not applied: %v", last)
	}
}

func wsURL(srv *httptest.Server) string { return "ws" + strings.TrimPrefix(srv.URL, "http") }

func newTestWSClient(t *testing.T, url string, keepAliveMs int, onError func(any, any)) *WSClient {
	t.Helper()
	noop := func(any, any) {}
	if onError == nil {
		onError = noop
	}
	c := NewWSClient(url, noop, onError, noop, noop, "", map[string]any{"keepAlive": keepAliveMs})
	fut, err := c.Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	select {
	case <-fut.Await():
	case <-time.After(5 * time.Second):
		t.Fatal("connect: timed out")
	}
	return c
}

// A server that never reads never answers pings: the pinger times out and
// calls OnError, whose Reset closes the socket, so the read loop calls OnError
// at the same time. Both write and read the client's Error.
func TestWSClientPingTimeoutAndReadLoopErrorConcurrently(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		<-release
		conn.Close()
	}))
	defer srv.Close()
	defer close(release)

	var errors atomic.Int32
	c := newTestWSClient(t, wsURL(srv), 10, func(any, any) { errors.Add(1) })

	deadline := time.Now().Add(5 * time.Second)
	for errors.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if errors.Load() == 0 {
		t.Fatal("ping-pong timeout never reported")
	}
	if c.GetError() == nil {
		t.Fatal("client error not recorded")
	}
	time.Sleep(50 * time.Millisecond) // let both OnError paths finish under -race
}

// Close must stop the keep-alive pinger: no ping may reach the server after it.
func TestWSClientCloseStopsPinger(t *testing.T) {
	var pings atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetPingHandler(func(data string) error {
			pings.Add(1)
			return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	var errorsAfterClose atomic.Int32
	var closed atomic.Bool
	// the read loop reports the closed socket as a NetworkError; only the
	// pinger's pong-miss timeout counts here
	c := newTestWSClient(t, wsURL(srv), 50, func(_ any, err any) {
		if closed.Load() && strings.Contains(fmt.Sprint(err), "keepalive") {
			errorsAfterClose.Add(1)
		}
	})

	deadline := time.Now().Add(5 * time.Second)
	for pings.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if pings.Load() < 2 {
		t.Fatal("pinger never ran")
	}

	closed.Store(true)
	select {
	case <-c.Close().Await():
	case <-time.After(5 * time.Second):
		t.Fatal("close: disconnection never signalled")
	}
	seen := pings.Load()
	time.Sleep(500 * time.Millisecond) // 10 keep-alive periods
	if extra := pings.Load() - seen; extra > 0 {
		t.Fatalf("pinger kept running after Close: %d more pings", extra)
	}
	if n := errorsAfterClose.Load(); n > 0 {
		t.Fatalf("pinger reported %d keep-alive timeouts after Close", n)
	}
}

package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type wsConn struct {
	conn   *websocket.Conn
	send   chan []byte
	closed atomic.Bool
}

func (wc *wsConn) close() {
	wc.closed.Store(true)
	close(wc.send)
	if wc.conn != nil {
		wc.conn.Close()
	}
}

func (wc *wsConn) writePump() {
	defer wc.close()

	for msg := range wc.send {
		msgType := websocket.BinaryMessage
		if len(msg) < 1 {
			msgType = websocket.TextMessage
		} else if msg[0] == '{' {
			msgType = websocket.TextMessage
		}
		err := wc.conn.WriteMessage(msgType, msg)
		if err != nil {
			return
		}
	}
}

func (wc *wsConn) readPump(recv chan<- []byte) {
	defer wc.close()

	wc.conn.SetReadLimit(512 << 10) // 512kb
	wc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wc.conn.SetPongHandler(func(string) error { wc.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		msgType, msg, err := wc.conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType == websocket.PingMessage {
			wc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			continue
		}
		recv <- msg
	}
}

type SessionActor struct {
	sid        string
	tid        string
	mu         sync.Mutex
	host       *wsConn
	client     *wsConn
	seq        uint64
	ring       *Ring
	rate       *RateLimiter
	lastActive atomic.Int64
	idleTO     time.Duration
	ttl        time.Duration
	closed     atomic.Bool
	redisURL   string
	rdb        *redis.Client
}

func NewSessionActor(sid, tid string, ttl, idleTO time.Duration, ringSize, rateBPS int, redisURL string, rdb *redis.Client) *SessionActor {
	sa := &SessionActor{
		sid:        sid,
		tid:        tid,
		seq:        0,
		ring:       NewRing(ringSize),
		rate:       NewRateLimiter(rateBPS),
		idleTO:     idleTO,
		ttl:        ttl,
		lastActive: atomic.Int64{},
		redisURL:   redisURL,
		rdb:        rdb,
	}

	sa.lastActive.Store(time.Now().UnixMilli())

	// Start idle timer
	go sa.idleTimer()

	if rdb != nil {
		go sa.saveMetadataPeriodic()
	}

	return sa
}

func (sa *SessionActor) AttachHost(conn *websocket.Conn) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.closed.Load() {
		return fmt.Errorf("session closed")
	}

	if sa.host != nil {
		sa.host.close()
	}

	sa.host = &wsConn{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Start pumps
	hostRecv := make(chan []byte, 256)
	go sa.host.readPump(hostRecv)
	go sa.host.writePump()

	// Process incoming from host
	go func() {
		for msg := range hostRecv {
			sa.processHostToClient(msg)
		}
		sa.mu.Lock()
		if sa.host != nil {
			sa.host.close()
			sa.host = nil
		}
		sa.mu.Unlock()
		IncWsHostConnected()
	}()

	return nil
}

func (sa *SessionActor) AttachClient(conn *websocket.Conn, resumeFrom uint64) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.closed.Load() {
		return fmt.Errorf("session closed")
	}

	if sa.client != nil {
		sa.client.close()
	}

	sa.client = &wsConn{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Replay if resume
	if resumeFrom > 0 {
		frames := sa.ring.ReplayFrom(resumeFrom)
		if len(frames) == 0 {
			errMsg := map[string]interface{}{
				"t":      "err",
				"message": "out_of_history",
				"sid":    sa.sid,
			}
			data, _ := json.Marshal(errMsg)
			sa.client.send <- data
			IncFramesJsonRingReplayed(len(frames))
		} else {
			for _, f := range frames {
				sa.client.send <- f
			}
			IncFramesJsonRingReplayed(len(frames))
		}
	}

	// Start pumps
	clientRecv := make(chan []byte, 256)
	go sa.client.readPump(clientRecv)
	go sa.client.writePump()

	// Process incoming from client
	go func() {
		for msg := range clientRecv {
			sa.processClientToHost(msg)
		}
		sa.mu.Lock()
		if sa.client != nil {
			sa.client.close()
			sa.client = nil
		}
		sa.mu.Unlock()
		IncWsClientConnected()
	}()

	return nil
}

func (sa *SessionActor) processHostToClient(msg []byte) {
	sa.lastActive.Store(time.Now().UnixMilli())

	isJSON := len(msg) > 0 && msg[0] == '{'

	if isJSON {
		// Add seq, push to ring, send to client
		sa.mu.Lock()
		if sa.client == nil || sa.client.closed.Load() {
			sa.mu.Unlock()
			return
		}
		seq := atomic.AddUint64(&sa.seq, 1)
		sa.ring.Push(seq, msg)
		sa.client.send <- msg // todo: add seq if missing
		sa.mu.Unlock()
	} else {
		// PTY, rate limit
		sa.mu.Lock()
		if sa.client == nil || sa.client.closed.Load() {
			sa.mu.Unlock()
			return
		}
		if !sa.rate.Allow(len(msg)) {
			// Throttle event
			throttleMsg := map[string]interface{}{
				"t":    "evt",
				"kind": "throttle",
				"bps":  sa.rate.maxTokens,
				"sid":  sa.sid,
				"seq":  atomic.LoadUint64(&sa.seq),
			}
			data, _ := json.Marshal(throttleMsg)
			sa.client.send <- data
			IncThrottleEvents()
			sa.mu.Unlock()
			return
		}
		sa.client.send <- msg
		AddBytesHostToClient(len(msg))
		sa.mu.Unlock()
	}
}

func (sa *SessionActor) processClientToHost(msg []byte) {
	sa.lastActive.Store(time.Now().UnixMilli())

	if len(msg) > 0 && msg[0] != '{' {
		// Binary from client, ignore
		errMsg := map[string]interface{}{
			"t":      "err",
			"message": "binary not expected from client",
			"sid":    sa.sid,
		}
		data, _ := json.Marshal(errMsg)
		sa.sendToHost(data)
		return
	}

	// JSON, add seq, ring, send to host
	sa.mu.Lock()
	seq := atomic.AddUint64(&sa.seq, 1)
	sa.ring.Push(seq, msg)
	sa.sendToHost(msg)
	sa.mu.Unlock()
}

func (sa *SessionActor) sendToHost(msg []byte) {
	if sa.host != nil && !sa.host.closed.Load() {
		sa.host.send <- msg
	}
}

func (sa *SessionActor) sendToClient(msg []byte) {
	if sa.client != nil && !sa.client.closed.Load() {
		sa.client.send <- msg
	}
}

func (sa *SessionActor) Close(reason string) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.closed.Load() {
		return
	}
	sa.closed.Store(true)

	// Send close reason if possible
	closeMsg := map[string]interface{}{
		"t":      "err",
		"message": reason,
		"sid":    sa.sid,
	}
	data, _ := json.Marshal(closeMsg)
	if sa.host != nil {
		sa.host.send <- data
		sa.host.close()
	}
	if sa.client != nil {
		sa.client.send <- data
		sa.client.close()
	}

	// Save last seq to redis if available
	if sa.rdb != nil {
		exp := int64(time.Now().Add(sa.ttl).Unix())
		data, _ := json.Marshal(map[string]interface{}{"tid": sa.tid, "lastSeq": sa.seq, "expiresAt": exp})
		err := sa.rdb.Set(context.Background(), fmt.Sprintf("relay:sess:%s", sa.sid), data, sa.ttl).Err()
		if err != nil {
			// log but ignore
		}
	}
}

func (sa *SessionActor) idleTimer() {
	for {
		time.Sleep(sa.idleTO)
		if sa.closed.Load() {
			return
		}
		last := sa.lastActive.Load()
		if time.Now().UnixMilli()-last > int64(sa.idleTO.Milliseconds()) {
			sa.Close("idle_timeout")
			return
		}
	}
}

func (sa *SessionActor) saveMetadataPeriodic() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if sa.closed.Load() {
			return
		}
		exp := int64(time.Now().Add(sa.ttl).Unix())
		data, _ := json.Marshal(map[string]interface{}{"tid": sa.tid, "lastSeq": sa.seq, "expiresAt": exp})
		sa.rdb.Set(context.Background(), fmt.Sprintf("relay:sess:%s", sa.sid), data, sa.ttl).Err()
	}
}

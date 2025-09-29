package relay

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	activeSessions = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_sessions",
		Help: "Number of active sessions",
	})

	wsHostConnected = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_host_connected_total",
		Help: "Total number of host WebSocket connections",
	})

	wsClientConnected = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_client_connected_total",
		Help: "Total number of client WebSocket connections",
	})

	bytesHostToClient = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bytes_host_to_client_total",
		Help: "Total bytes forwarded from host to client",
	})

	framesJsonRingReplayed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "frames_json_ring_replayed_total",
		Help: "Total JSON frames replayed from ring buffer",
	})

	throttleEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "throttle_events_total",
		Help: "Total throttle events emitted",
	})
)

func init() {
	prometheus.MustRegister(activeSessions)
	prometheus.MustRegister(wsHostConnected)
	prometheus.MustRegister(wsClientConnected)
	prometheus.MustRegister(bytesHostToClient)
	prometheus.MustRegister(framesJsonRingReplayed)
	prometheus.MustRegister(throttleEvents)
}

func IncActiveSessions(delta int) {
	activeSessions.Add(float64(delta))
}

func IncWsHostConnected() {
	wsHostConnected.Inc()
}

func IncWsClientConnected() {
	wsClientConnected.Inc()
}

func AddBytesHostToClient(n int) {
	bytesHostToClient.Add(float64(n))
}

func IncFramesJsonRingReplayed(n int) {
	framesJsonRingReplayed.Add(float64(n))
}

func IncThrottleEvents() {
	throttleEvents.Inc()
}

// MetricsHandler returns the Prometheus handler, optionally protected
func MetricsHandler(adminToken string) http.Handler {
	if adminToken == "" {
		return http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) < 7 || token[:7] != "Bearer " {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token = token[7:]
		if token != adminToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		promhttp.Handler().ServeHTTP(w, r)
	})
}

// Stub JSON metrics if no prom, but since prom is used, ok.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// APIRequestDuration tracks HTTP request duration by endpoint and method
	APIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chemistryuno_api_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // Default: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"endpoint", "method", "status"},
	)

	// DBQueryDuration tracks database query duration by table and operation
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chemistryuno_db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"table", "operation"},
	)

	// CacheHitsTotal counts cache hits by cache name
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chemistryuno_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache"},
	)

	// CacheMissesTotal counts cache misses by cache name
	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chemistryuno_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache"},
	)

	// WebSocketMessagesTotal counts WebSocket messages by room
	WebSocketMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chemistryuno_websocket_messages_total",
			Help: "Total number of WebSocket messages broadcast",
		},
		[]string{"room_id"},
	)

	// WebSocketConnectionsActive tracks active WebSocket connections
	WebSocketConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "chemistryuno_websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	// WebSocketBroadcastDuration tracks broadcast operation duration
	WebSocketBroadcastDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chemistryuno_websocket_broadcast_duration_seconds",
			Help:    "Duration of WebSocket broadcast operations",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"room_id"},
	)
)

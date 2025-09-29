# Relay Layer (Switchboard)

The relay is a standalone Go service that pairs WebSocket connections between a mobile client and a local backend/agent host across NATs and networks. It forwards PTY bytes and JSON control/events with seq/ack, replay on reconnect, rate limiting, and auth via JWT tokens. It does not run agents or touch repositories.

## Run Locally

```bash
cd relay
export PORT=8081 JWT_SECRET=devsecret RELAY_MINT=true
go run ./cmd/relay
```

## Mint a Session (for dev/testing)

```bash
curl -X POST http://localhost:8081/api/session \
  -H "Authorization: Bearer anything" \
  -d '{"tenantId":"t_demo","ttlSeconds":3600}'
```

## Connect Host (backend should dial)

```bash
wscat -c "ws://localhost:8081/ws/host?sessionId=...&token=..."
```

## Connect Client

```bash
wscat -c "ws://localhost:8081/ws/client?sessionId=...&token=..."
```

Reconnect with `&resumeSeq=NNN` and observe replay of JSON frames.

## Docker

```bash
docker build -t cockpit-relay .
docker run --rm -p 8081:8081 -e JWT_SECRET=devsecret cockpit-relay
```

## Environment Variables

- `PORT`: HTTP/WS port (default 8081)
- `JWT_SECRET`: HS256 secret (required)
- `SESSION_TTL_SECONDS`: Session lifetime (default 86400)
- `IDLE_TIMEOUT_SECONDS`: Close idle sessions (default 1800)
- `RING_BUFFER_BYTES`: JSON replay buffer (default 131072)
- `RATE_LIMIT_BPS`: PTY rate from host->client (default 65536)
- `CORS_ORIGINS`: Comma-separated allowed origins
- `REDIS_URL`: Optional Redis for resume metadata
- `RELAY_MINT`: Enable POST /api/session (default false)
- `ADMIN_TOKEN`: Required for /metrics (optional)

## Endpoints

- GET /healthz: `{"ok":true}`
- GET /metrics: Prometheus (if ADMIN_TOKEN set)
- POST /api/session: Mint session (if RELAY_MINT=true)
- GET /ws/host?sessionId=...&token=...: Host WS
- GET /ws/client?sessionId=...&token=...&resumeSeq=...: Client WS

## Integration Notes

The backend mints tokens and dials /ws/host. The mobile app connects to /ws/client with resumeSeq for replay.

## Testing

Run `go test ./internal/relay/...` for ring buffer, auth, and session actor tests.

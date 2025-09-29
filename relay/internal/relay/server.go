package relay

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	hub *Hub
	mux *mux.Router

	corsOrigins []string
	relayMint  bool
	adminToken string
	jwtSecret  []byte
}

func NewServer(hub *Hub, corsOrigins []string, relayMint bool, adminToken, jwtSecret string) *Server {
	s := &Server{
		hub:         hub,
		corsOrigins: corsOrigins,
		relayMint:   relayMint,
		adminToken:  adminToken,
		jwtSecret:   []byte(jwtSecret),
	}

	s.mux = mux.NewRouter()
	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/healthz", s.healthz).Methods("GET")

	if s.adminToken != "" {
		s.mux.Handle("/metrics", MetricsHandler(s.adminToken)).Methods("GET")
	} else {
		s.mux.HandleFunc("/metrics", s.noop).Methods("GET")
	}

	if s.relayMint {
		s.mux.HandleFunc("/api/session", s.mintSession).Methods("POST")
	} else {
		s.mux.HandleFunc("/api/session", s.mintDisabled).Methods("POST")
	}

	s.mux.HandleFunc("/ws/host", s.wsHost).Methods("GET")
	s.mux.HandleFunc("/ws/client", s.wsClient).Methods("GET")

	// CORS
	s.mux.Use(s.corsMiddleware)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) noop(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s *Server) mintDisabled(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "minting disabled; use backend’s /api/session", http.StatusBadRequest)
}

type MintReq struct {
	TenantId   string `json:"tenantId"`
	TTLSeconds int64  `json:"ttlSeconds"`
}

type MintRes struct {
	SessionId string `json:"sessionId"`
	Token     string `json:"token"`
	WS        struct {
		Host   string `json:"host"`
		Client string `json:"client"`
	} `json:"ws"`
	ExpiresAt string `json:"expiresAt"`
}

func (s *Server) mintSession(w http.ResponseWriter, r *http.Request) {
	var req MintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.TenantId == "" || req.TTLSeconds <= 0 {
		http.Error(w, "invalid tenantId or ttlSeconds", http.StatusBadRequest)
		return
	}

	sid := fmt.Sprintf("sess_%s", s.randomString(16))
	exp := time.Now().Add(time.Duration(req.TTLSeconds) * time.Second)

	claims := NewClaims(sid, req.TenantId, exp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		log.Printf("Mint token error: %v", err)
		http.Error(w, "failed to mint token", http.StatusInternalServerError)
		return
	}

	u := r.Host
	if r.TLS != nil {
		u = "wss://" + u
	} else {
		u = "ws://" + u
	}

	clientWS := fmt.Sprintf("%s/ws/client", u)
	hostWS := fmt.Sprintf("%s/ws/host", u)

	res := MintRes{
		SessionId: sid,
		Token:     tokenStr,
		WS: struct {
			Host   string `json:"host"`
			Client string `json:"client"`
		}{
			Host:   hostWS,
			Client: clientWS,
		},
		ExpiresAt: exp.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) wsHost(w http.ResponseWriter, r *http.Request) {
	s.handleWS(w, r, true, 0)
}

func (s *Server) wsClient(w http.ResponseWriter, r *http.Request) {
	resumeStr := r.URL.Query().Get("resumeSeq")
	var resume uint64
	if resumeStr != "" {
		var err error
		resume, err = strconv.ParseUint(resumeStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid resumeSeq", http.StatusBadRequest)
			return
		}
	}
	s.handleWS(w, r, false, resume)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request, isHost bool, resume uint64) {
	sessionId := r.URL.Query().Get("sessionId")
	tokenStr := r.URL.Query().Get("token")

	if sessionId == "" || tokenStr == "" {
		http.Error(w, "missing sessionId or token", http.StatusBadRequest)
		return
	}

	claims, err := VerifyToken(s.jwtSecret, tokenStr)
	if err != nil {
		log.Printf("Token verify error: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	if claims.SID != sessionId {
		http.Error(w, "token sid mismatch", http.StatusUnauthorized)
		return
	}

	sa := s.hub.GetOrCreate(sessionId, claims.TID)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // todo: validate against cors
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	if isHost {
		err = sa.AttachHost(conn)
	} else {
		err = sa.AttachClient(conn, resume)
	}
	if err != nil {
		conn.Close()
		log.Printf("Attach error: %v", err)
	}

	// The attach starts the pumps, so no need for defer close here, as it's handled in attach
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if s.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) isAllowedOrigin(origin string) bool {
	if len(s.corsOrigins) == 0 {
		return true
	}
	for _, allowed := range s.corsOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (s *Server) randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[num.Int64()]
	}
	return string(b)
}

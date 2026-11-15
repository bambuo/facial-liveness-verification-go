package web

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"facial-liveness-verification-go/internal/domain"
)

// withMiddleware 串起异常信封、request_id、限流与 CORS。
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return s.recoverMiddleware(s.requestIDMiddleware(s.rateLimitMiddleware(s.corsMiddleware(next))))
}

func (s *Server) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeJSON(w, http.StatusServiceUnavailable, failEnvelope(domain.CodeServiceUnavailable, "服务暂不可用", requestIDFrom(r)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := normalizeRequestID(r.Header.Get("X-Request-Id"))
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	if s.limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !s.limiter.allow(ip, time.Now()) {
			writeJSON(w, http.StatusTooManyRequests, failEnvelope(domain.CodeQuotaExceeded, "超过请求频率限制", requestIDFrom(r)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	if len(s.cfg.Security.AllowedOrigins) == 0 {
		return next
	}
	allowed := make(map[string]bool, len(s.cfg.Security.AllowedOrigins))
	for _, origin := range s.cfg.Security.AllowedOrigins {
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-Id")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ipLimiter 是每 IP 固定窗口限流器。
type ipLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]limitEntry
}

type limitEntry struct {
	start time.Time
	count int
}

func newIPLimiter(limit int) *ipLimiter {
	return &ipLimiter{limit: limit, window: time.Minute, entries: make(map[string]limitEntry)}
}

func (l *ipLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	if !ok || now.Sub(e.start) >= l.window {
		l.entries[key] = limitEntry{start: now, count: 1}
		if len(l.entries) > 10000 {
			l.prune(now)
		}
		return true
	}
	e.count++
	l.entries[key] = e
	return e.count <= l.limit
}

func (l *ipLimiter) prune(now time.Time) {
	for k, e := range l.entries {
		if now.Sub(e.start) >= l.window {
			delete(l.entries, k)
		}
	}
}

func bearerToken(header string) (string, bool) {
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", false
	}
	token := strings.TrimSpace(header[len("Bearer "):])
	return token, token != ""
}

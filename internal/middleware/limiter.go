package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// ConnectionsLimiter is a middleware that limits the number of simultaneous connections per IP.
type ConnectionsLimiter struct {
	mu          sync.Mutex
	connections map[string]int
	max         int
}

func NewConnectionLimiter(i int) *ConnectionsLimiter {
	return &ConnectionsLimiter{
		connections: map[string]int{},
		max:         i,
	}
}

// leaseConnection increases a number of connections per given token and
// returns a release function to be called once a request is finished.
// If the token reaches the limit of max simultaneous connections, leaseConnection returns an error.
func (auth *ConnectionsLimiter) LeaseConnection(request *http.Request) (release func(), err error) {
	key := fmt.Sprintf("ip-%v", realIP(request))
	auth.mu.Lock()
	defer auth.mu.Unlock()

	if auth.connections[key] >= auth.max {
		return nil, fmt.Errorf("you have reached the limit of streaming connections: %v max", auth.max)
	}
	auth.connections[key] += 1

	return func() {
		auth.mu.Lock()
		defer auth.mu.Unlock()
		auth.connections[key] -= 1
		if auth.connections[key] == 0 {
			delete(auth.connections, key)
		}
	}, nil
}

func realIP(request *http.Request) string {
	// Fall back to legacy behavior
	if ip := request.Header.Get("X-Forwarded-For"); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			return strings.Trim(ip[:i], "[] \t")
		}
		return ip
	}
	if ip := request.Header.Get("X-Real-Ip"); ip != "" {
		return strings.Trim(ip, "[]")
	}
	ra, _, _ := net.SplitHostPort(request.RemoteAddr)
	return ra
}

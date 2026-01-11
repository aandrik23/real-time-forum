package authutils

import (
	"fmt"
	"forum/internal/logger"
	"sync"
	"time"
)

type loginKey struct {
	IP    string
	Email string
}

var (
	loginAttempts = make(map[loginKey][]time.Time)
	blockedIPs    = make(map[string]time.Time)
	loginMu       sync.Mutex
)

func TooManyAttempts(ip, email string) bool {
	const maxAttempts = 5
	const window = 5 * time.Minute
	const banDuration = 15 * time.Minute

	now := time.Now()
	key := loginKey{IP: ip, Email: email}

	loginMu.Lock()
	defer loginMu.Unlock()

	// Check if IP is blocked
	if banUntil, blocked := blockedIPs[ip]; blocked {
		if now.Before(banUntil) {
			return true
		}
		delete(blockedIPs, ip) // Unban if time passed
	}

	// Update attempts
	attempts := loginAttempts[key]
	var recent []time.Time
	for _, t := range attempts {
		if now.Sub(t) < window {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	loginAttempts[key] = recent

	// 6th failed attempt triggers block
	if len(recent) > maxAttempts {
		blockedIPs[ip] = now.Add(banDuration)
		delete(loginAttempts, key)
		logger.Log(fmt.Sprintf("IP %s temporarily banned for excessive login attempts", ip), logger.WarnLevel)
		return true
	}

	return false
}

func cleanupLoginAttempts() {
	const window = 5 * time.Minute
	now := time.Now()

	loginMu.Lock()
	defer loginMu.Unlock()

	// Clean old login attempts
	for key, attempts := range loginAttempts {
		var recent []time.Time
		for _, t := range attempts {
			if now.Sub(t) < window {
				recent = append(recent, t)
			}
		}
		if len(recent) > 0 {
			loginAttempts[key] = recent
		} else {
			delete(loginAttempts, key)
		}
	}

	// Clean expired blocks
	for ip, expiry := range blockedIPs {
		if now.After(expiry) {
			delete(blockedIPs, ip)
		}
	}
}

func StartLoginAttemptCleanup() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cleanupLoginAttempts()
		}
	}()
}

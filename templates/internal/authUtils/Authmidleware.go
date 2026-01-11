package authutils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"forum/internal/database"
	"forum/internal/logger"
	"forum/internal/realtime"
	"forum/internal/utils"
	"net/http"
	"strings"
	"time"
)

// =========================
// AuthMiddleware
// =========================

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		clearSessionKilledIfPresent(w, r)

		// 1) Grab auth cookie
		tok, ok := getCookieValue(r, "auth_token")
		if !ok {
			if handledNoAccessCookie(w, r, next) {
				return
			}
			// handledNoAccessCookie should always return true, but keep defensive
			return
		}

		// 2) Validate/refresh access token and attach payload
		payload, r2, ok := authenticateRequest(w, r, tok)
		if !ok {
			return // authenticateRequest already responded (401/redirect)
		}

		// 3) Enforce DB JTI
		if !checkAccessJTI(w, r2, payload) {
			return
		}

		// 4) SPA rule: anon token can’t access API data
		if !enforceAnonAPIRule(w, r2, payload) {
			return
		}

		next.ServeHTTP(w, r2)
	}
}

// =========================
// helpers (private)
// =========================
func writeAPIUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "unauthorized",
	})
}

// only to actively reset the session (expired refresh, revoked token, invalid signature). , we issue anon token and redirect user to home page
func redirectToHomeAsAnon(w http.ResponseWriter, r *http.Request, payload *JWTPayload) {
	isAPI := strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	// API/fetch requests: do NOT rotate/expire cookies here — just deny
	if isAPI {
		writeAPIUnauthorized(w)
		return
	}

	// Browser navigation: reset session + issue anon + redirect
	ExpireTokens(w, r)
	if payload != nil {
		realtime.DM.ForceDisconnectUser(payload.UserID)
	}
	uuid := utils.GenerateUUID()
	createAnonymousToken(w, uuid)

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func safeLogWithRequest(r *http.Request, msg string, payload *JWTPayload, level logger.LogLevel) {
	meta := fmt.Sprintf(" | Method:%s | IP:%s", r.Method, r.RemoteAddr)
	if payload != nil {
		msg += " UUID:" + payload.UUID
	} else {
		msg += " UUID: ''"
	}
	logger.Log(msg+meta, level)
}

func clearSessionKilledIfPresent(w http.ResponseWriter, r *http.Request) {
	if ck, err := r.Cookie("session_killed"); err == nil && ck.Value == "true" {
		expireSessionKilledToken(w)
	}
}

func getCookieValue(r *http.Request, name string) (string, bool) {
	c, err := r.Cookie(name)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// returns true if request was fully handled (response written)
func handledNoAccessCookie(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) bool {
	// allow auth endpoints without cookie
	if r.URL.Path == "/api/login" || r.URL.Path == "/api/register" {
		next.ServeHTTP(w, r)
		return true
	}

	// API calls without token: just 401
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeAPIUnauthorized(w)
		return true
	}

	// shell route -> issue anon and allow
	uuid := utils.GenerateUUID()
	createAnonymousToken(w, uuid)
	next.ServeHTTP(w, r)
	return true
}

// returns payload, updated request with context, ok=true if authenticated
func authenticateRequest(w http.ResponseWriter, r *http.Request, accessToken string) (*JWTPayload, *http.Request, bool) {
	payload, err := VerifyJWT(accessToken, TokenTypeAccess)
	if err == nil {
		ctx := context.WithValue(r.Context(), "jwtPayload", payload)
		r2 := r.WithContext(ctx)
		logger.Log(fmt.Sprint("Verified token UUID:", payload.UUID, "| Role:", payload.Role), logger.InfoLevel)
		return payload, r2, true
	}

	// expired
	if errors.Is(err, ErrTokenExpired) {
		// if expired anon -> mint new anon + continue as anon
		if anonPayload := checkForAnonymousPayload(accessToken); anonPayload != nil {
			ctx := context.WithValue(r.Context(), "jwtPayload", anonPayload)
			r2 := r.WithContext(ctx)

			uuid := utils.GenerateUUID()
			createAnonymousToken(w, uuid)
			logger.Log(fmt.Sprint("Expired Anonymous Token. NEW JWT issued for anonymous user new uuid:", uuid), logger.InfoLevel)

			return anonPayload, r2, true
		}

		// try refresh
		refreshTok, ok := getCookieValue(r, "refresh_token")
		if !ok {
			safeLogWithRequest(r, "Missing refresh token", payload, logger.ErrorLevel)
			redirectToHomeAsAnon(w, r, payload)
			return nil, r, false
		}

		newAccess, err := refreshAccessToken(refreshTok, w, r)
		if err != nil {
			safeLogWithRequest(r, "Session expired", payload, logger.ErrorLevel)
			redirectToHomeAsAnon(w, r, payload)
			return nil, r, false
		}

		payload, err = VerifyJWT(newAccess, TokenTypeAccess)
		if err != nil {
			safeLogWithRequest(r, "Token refresh failed", payload, logger.ErrorLevel)
			redirectToHomeAsAnon(w, r, payload)
			return nil, r, false
		}

		ctx := context.WithValue(r.Context(), "jwtPayload", payload)
		r2 := r.WithContext(ctx)
		logger.Log(fmt.Sprint("Verified token UUID:", payload.UUID, "| Role:", payload.Role), logger.InfoLevel)
		return payload, r2, true
	}

	// invalid signature/structure/etc
	safeLogWithRequest(r, "Invalid token", payload, logger.ErrorLevel)
	redirectToHomeAsAnon(w, r, payload)
	return nil, r, false
}

func checkAccessJTI(w http.ResponseWriter, r *http.Request, payload *JWTPayload) bool {
	if payload.JTI != "" {
		exists, err := database.TokenExists(payload.JTI)
		if err != nil || !exists {
			safeLogWithRequest(r, "Access token revoked or invalid", payload, logger.ErrorLevel)
			redirectToHomeAsAnon(w, r, payload)
			return false
		}
		return true
	}

	if payload.Role != RoleAnonymous {
		http.Error(w, "Invalid token structure", http.StatusUnauthorized)
		return false
	}
	return true
}

func enforceAnonAPIRule(w http.ResponseWriter, r *http.Request, payload *JWTPayload) bool {
	if payload.Role == RoleAnonymous && strings.HasPrefix(r.URL.Path, "/api/") {
		switch r.URL.Path {
		case "/api/login", "/api/register", "/api/logout":
			return true
		default:
			writeAPIUnauthorized(w)
			return false
		}
	}
	return true
}

func checkForAnonymousPayload(token string) *JWTPayload {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	var tempPayload JWTPayload
	if err := json.Unmarshal(payloadRaw, &tempPayload); err != nil {
		return nil
	}

	if tempPayload.Role == RoleAnonymous {
		return &tempPayload
	}
	return nil
}

func refreshAccessToken(refreshToken string, w http.ResponseWriter, r *http.Request) (string, error) {

	// 1. Verify the refresh token
	payload, err := VerifyJWT(refreshToken, TokenTypeRefresh)
	if err != nil {
		msg := "invalid or expired refresh token"
		safeLogWithRequest(r, msg, payload, logger.ErrorLevel)
		return "", errors.New(msg)
	}
	logger.Log("Refresh Token verrified "+"UUID:"+payload.UUID, logger.DebugLevel)

	// 1.1 Anonymous users can't refresh
	if payload.Role == RoleAnonymous {
		msg := "anonymous users cannot refresh token"
		return "", errors.New(msg)
	}

	// 2. Check that refresh token exists in DB
	exists, err := database.TokenExists(payload.JTI)
	if err != nil || !exists {
		logger.Log(fmt.Sprintf("Refresh token not recognized: JTI=%s | Exists=%v | Err=%v"+" | UUID=%s", payload.JTI, exists, err, payload.UUID), logger.DebugLevel)
		return "", errors.New("refresh token not recognized")
	}
	logger.Log("Refresh Token ID exists in DB ID:"+payload.JTI, logger.DebugLevel)

	// 3. Delete old Access and refresh token ID (rotation)
	if err := database.DeleteToken(payload.JTI); err != nil {
		msg := "failed to revoke old refresh token"
		safeLogWithRequest(r, msg, payload, logger.ErrorLevel)
		return "", errors.New(msg)
	}
	logger.Log("Refresh Token ID deleted from DB", logger.DebugLevel)

	if payload.AccessJTI != "" {
		if err := database.DeleteToken(payload.AccessJTI); err != nil {
			msg := "failed to revoke old access token"
			safeLogWithRequest(r, msg, payload, logger.ErrorLevel)
			return "", errors.New(msg)
		}
	}
	logger.Log("Access Token ID deleted from DB", logger.DebugLevel)

	// 5. Generate new refresh & access token ID and store it
	new_refresh_JTI := utils.GenerateUUID()
	new_access_JTI := utils.GenerateUUID()
	refresh_expiry := time.Now().Add(RefreshTokenTime).Unix()
	access_expiry := time.Now().Add(AccessTokenTime).Unix()

	if err := database.SaveToken(payload.UUID, new_refresh_JTI, TokenTypeRefresh, refresh_expiry); err != nil {
		msg := "failed to store new refresh token"
		safeLogWithRequest(r, msg, payload, logger.ErrorLevel)
		return "", errors.New(msg)
	}
	logger.Log("Refresh Token ID added to DB"+" \n[ UUID:"+payload.UUID+" \nNEW_Refresh_jti: "+new_refresh_JTI+" \nToken Type:"+TokenTypeRefresh+" ]", logger.DebugLevel)

	if err := database.SaveToken(payload.UUID, new_access_JTI, TokenTypeAccess, access_expiry); err != nil {
		msg := "failed to store new refresh token"
		safeLogWithRequest(r, msg, payload, logger.ErrorLevel)
		return "", errors.New(msg)
	}
	logger.Log("Access Token ID added to DB"+" \n[ UUID:"+payload.UUID+" \nNEW_Access_jti: "+new_access_JTI+" \nToken Type:"+TokenTypeAccess+" ]", logger.DebugLevel)

	// 6. Generate new access
	token := createAccessToken(w, payload.Username, payload.Role, payload.UUID, new_access_JTI, payload.Bio, payload.Avatar, payload.UserID)
	//logger.Log("New Access Token created"+"\n[ UUID:"+payload.UUID+" \nRole:"+payload.Role+" \nNEW_Access_jti: "+new_access_JTI+" \nToken Type:"+TokenTypeAccess+" \nExpire:"+AccessTokenTime.String()+" ]", logger.DebugLevel)

	// Expire old cookies
	expireRefreshToken(w)
	expireCsrfToken(w)

	// Set new tokens
	createRefreshToken(w, payload.Username, payload.Role, payload.UUID, new_refresh_JTI, new_access_JTI, payload.Bio, payload.Avatar, payload.UserID)
	createCsrfToken(w)

	return token, nil
}

//----------------------------------------------------------------

func RequireRoleMiddleware(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	roleSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			payload, ok := r.Context().Value("jwtPayload").(*JWTPayload)

			wantsJSON := strings.HasPrefix(r.URL.Path, "/api/") ||
				strings.Contains(r.Header.Get("Accept"), "application/json") ||
				r.Header.Get("X-Requested-With") == "XMLHttpRequest"

			if !ok || payload == nil {
				msg := "Unauthorized"
				safeLogWithRequest(r, msg, payload, logger.ErrorLevel)

				if wantsJSON {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "unauthorized",
					})
					return
				}

				http.Redirect(w, r, "/?show=register", http.StatusSeeOther)
				return
			}

			if _, ok := roleSet[payload.Role]; !ok {
				msg := "Forbidden: insufficient role"
				safeLogWithRequest(r, msg, payload, logger.ErrorLevel)

				if wantsJSON {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "forbidden",
					})
					return
				}

				http.Redirect(w, r, "/?show=register", http.StatusSeeOther)
				return
			}

			safeLogWithRequest(r,
				fmt.Sprintf("Access granted: Role: %s", payload.Role),
				payload,
				logger.InfoLevel,
			)
			next.ServeHTTP(w, r)
		}
	}
}

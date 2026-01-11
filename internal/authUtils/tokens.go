package authutils

import (
	"forum/internal/database"
	"forum/internal/logger"
	"forum/internal/utils"
	"net/http"
	"time"
)

const (
	RoleAnonymous = "anonymous"
	RoleUser      = "user"
	RoleAdmin     = "admin"
)

var secure_value = false //  Set to true in production
const (
	AccessTokenTime    = 10 * time.Minute
	AnonymousTokenTime = 5 * time.Minute
	RefreshTokenTime   = 24 * time.Hour
)

func createAnonymousToken(w http.ResponseWriter, uuid string) {
	token, _ := GenerateJWT(uuid, "guest", RoleAnonymous, TokenTypeAccess, "", AnonymousTokenTime, "", "", -1)

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
	logger.Log("Anonymous access token issued for UUID: "+uuid, logger.DebugLevel)
}

func createAccessToken(w http.ResponseWriter, username, role, uuid, jti, bio, avatar string, userID int) string {
	token, err := GenerateJWT(uuid, username, role, TokenTypeAccess, jti, AccessTokenTime, bio, avatar, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return ""
	}
	logger.Log("Access Token generated"+" \n[ UUID:"+uuid+" \n Role:"+role+"\nJti: "+jti+" \nToken Type:"+TokenTypeAccess+" \nExpire:"+AccessTokenTime.String()+" ]", logger.DebugLevel)

	// Set JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})

	return token
}

func createRefreshToken(w http.ResponseWriter, username, role, uuid, jti, accessJTI, bio, avatar string, userID int) {
	refreshToken, _ := GenerateJWT(uuid, username, role, TokenTypeRefresh, jti, RefreshTokenTime, bio, avatar, userID, accessJTI)
	logger.Log("Refresh Token generated"+" \n[ UUID:"+uuid+" \nRole:"+role+"\nJti: "+jti+" \nToken Type:"+TokenTypeRefresh+" \nExpire:"+RefreshTokenTime.String()+" ]", logger.DebugLevel)

	// Set Refresh Token
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func createCsrfToken(w http.ResponseWriter) {
	csrfToken := utils.GenerateCSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false, // Must be readable by frontend (form or JS)
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func CreateTokens(w http.ResponseWriter, username, role, bio, avatar string, userID int) {
	// Generate UUID + JTI
	uuid := utils.GenerateUUID()
	refresh_jti := utils.GenerateUUID()
	access_jti := utils.GenerateUUID()
	expiry := time.Now().Add(RefreshTokenTime).Unix()

	// Store refresh token & access JTI in DB
	if err := database.SaveToken(uuid, refresh_jti, TokenTypeRefresh, expiry); err != nil {
		logger.Log("Failed to add Refresh Token ID to DB "+" \n[ UUID:"+uuid+"\nRefresh_jti: "+refresh_jti+" \nToken Type:"+TokenTypeRefresh+" \nExpire:"+RefreshTokenTime.String()+" ]", logger.ErrorLevel)
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}
	logger.Log("Refresh Token ID added to DB "+" \n[ UUID:"+uuid+"\nRefresh_jti: "+refresh_jti+" \nToken Type:"+TokenTypeRefresh+" \nExpire:"+RefreshTokenTime.String()+" ]", logger.DebugLevel)

	if err := database.SaveToken(uuid, access_jti, TokenTypeAccess, expiry); err != nil {
		logger.Log("Failed to add Access Token ID to DB "+" \n[ UUID:"+uuid+"\nAccess_jti: "+access_jti+" \nToken Type:"+TokenTypeAccess+" \nExpire:"+AccessTokenTime.String()+" ]", logger.ErrorLevel)
		http.Error(w, "Failed to store access token", http.StatusInternalServerError)
		return
	}
	logger.Log("Access Token ID added to DB "+" \n[ UUID:"+uuid+"\nAccess_jti: "+access_jti+" \nToken Type:"+TokenTypeAccess+" \nExpire:"+AccessTokenTime.String()+" ]", logger.DebugLevel)

	//expire anonymous token
	expireAccessToken(w)
	// Generate Access Token
	_ = createAccessToken(w, username, role, uuid, access_jti, bio, avatar, userID)
	// Generate Refresh Token
	createRefreshToken(w, username, role, uuid, refresh_jti, access_jti, bio, avatar, userID)
	// Set CSRF Token
	createCsrfToken(w)
}

//----------------------------------------------------------------

func expireAccessToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire immediately
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func expireRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func expireCsrfToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func expireSessionKilledToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_killed",
		Value:    "true",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

func ExpireTokens(w http.ResponseWriter, r *http.Request) {
	// Expire the auth token cookie
	expireAccessToken(w)

	// Expire the refresh token cookie
	expireRefreshToken(w)

	// Expire the CSRF token too (if needed)
	expireCsrfToken(w)

	// Revoke refresh token
	if refreshCookie, err := r.Cookie("refresh_token"); err == nil && refreshCookie.Value != "" {
		if payload, err := VerifyJWT(refreshCookie.Value, TokenTypeRefresh); err == nil {
			err = database.DeleteToken(payload.JTI) // Revoke refresh token
			if err != nil {
				logger.Log("Failed to delete Refresh Token ID from DB"+" JTI:"+payload.JTI, logger.ErrorLevel)
			}
			logger.Log("Refresh Token ID deleted from DB"+" JTI:"+payload.JTI, logger.DebugLevel)
			if payload.AccessJTI != "" {
				err = database.DeleteToken(payload.AccessJTI) // Revoke tied access token
				if err != nil {
					logger.Log("Failed to delete Access Token from DB"+" Access_JTI:"+payload.AccessJTI, logger.ErrorLevel)
				}
				logger.Log("Access Token deleted from DB"+" Access_JTI:"+payload.AccessJTI, logger.DebugLevel)
			}
		}
	}

	// logout flag to block refresh
	http.SetCookie(w, &http.Cookie{
		Name:     "session_killed",
		Value:    "true",
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
		Secure:   secure_value,
		SameSite: http.SameSiteStrictMode,
	})
}

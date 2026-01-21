package handlers

import (
	"encoding/json"
	"fmt"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/logger"
	"forum/internal/realtime"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// writeJSONError writes a JSON {"error": "..."} response with the given status code.
func WriteJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteJSONError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	identifier := strings.ToLower(strings.TrimSpace(r.FormValue("identifier")))
	password := r.FormValue("password")

	if identifier == "" || password == "" {
		WriteJSONError(w, "Identifier and password required", http.StatusBadRequest)
		return
	}

	username, role, bio, avatar, userID, status, err :=
		database.ValidateCredsByIdentifier(identifier, password)

	if err != nil {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		if authutils.TooManyAttempts(ip, identifier) {
			WriteJSONError(w, "Too many login attempts. Try again later.", http.StatusTooManyRequests)
			logger.Log(fmt.Sprintf(
				"Login rate limit triggered | IP: %s | Identifier: %s",
				ip, identifier,
			), logger.WarnLevel)
			return
		}
		WriteJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if status == "inactive" {
		if err := database.UpdateUserStatus(userID, "active"); err != nil {
			WriteJSONError(w, "Unable to update user status", http.StatusInternalServerError)
			return
		}
	} else {
		WriteJSONError(w, "User already logged in", http.StatusInternalServerError)
		return
	}

	authutils.CreateTokens(w, username, role, bio, avatar, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/home"})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())

	// Always expire cookies even if payload is nil
	authutils.ExpireTokens(w, r)

	if payload != nil {
		realtime.DM.ForceDisconnectUser(payload.UserID)
		_ = database.UpdateUserStatus(payload.UserID, "inactive")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteJSONError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	firstName := strings.TrimSpace(r.FormValue("first_name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	age := strings.TrimSpace(r.FormValue("age"))
	gender := strings.TrimSpace(r.FormValue("gender"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password2")

	// required fields
	if username == "" || firstName == "" || lastName == "" ||
		age == "" || gender == "" || email == "" ||
		password == "" || passwordConfirm == "" {
		WriteJSONError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if password != passwordConfirm {
		WriteJSONError(w, "Passwords don't match", http.StatusBadRequest)
		return
	}

	if !validateEmail(email) {
		WriteJSONError(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		WriteJSONError(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = database.AddUserToDb(
		username,
		firstName,
		lastName,
		age,
		gender,
		email,
		string(hashedPassword),
		"inactive",
	)

	if err != nil {
		if err.Error() == "username already exists" {
			WriteJSONError(w, "Username already exists", http.StatusBadRequest)
			return
		}
		if err.Error() == "email already exists" {
			WriteJSONError(w, "Email already exists", http.StatusBadRequest)
			return
		}
		WriteJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	logger.Log(fmt.Sprintf(
		"New user registered: %s | email: %s",
		username, email,
	), logger.InfoLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/home"})
}

func validateEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	if !strings.Contains(domain, ".") {
		return false
	}

	tld := domain[strings.LastIndex(domain, ".")+1:]
	if len(tld) < 2 {
		return false // TLD too short
	}

	tlds := []string{"com", "gr", "org", "info", "net", "edu", "gov"}
	tldList := make(map[string]struct{}, len(tlds))
	for _, v := range tlds {
		tldList[v] = struct{}{}
	}

	if _, ok := tldList[tld]; !ok {
		return false
	}

	return true
}

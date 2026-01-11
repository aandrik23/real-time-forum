package handlers

import (
	"bytes"
	authutils "forum/internal/authUtils"
	"net/http"
	"strings"
	"text/template"
)

type BaseViewModel struct {
	IsLoggedIn bool
	UserID     int
}

func SPAShellHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/static/") {
		http.NotFound(w, r)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	isLoggedIn := payload != nil && payload.Role != authutils.RoleAnonymous

	uid := 0
	if isLoggedIn {
		uid = payload.UserID
	}

	renderBaseOnly(w, BaseViewModel{
		IsLoggedIn: isLoggedIn,
		UserID:     uid,
	})

}

func renderBaseOnly(w http.ResponseWriter, vm BaseViewModel) {
	t, err := template.ParseFiles("templates/base.html")
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", vm); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

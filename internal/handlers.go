package internal

import (
	authutils "forum/internal/authUtils"
	"forum/internal/handlers"
	"net/http"
)

// apiProtected wraps a handler with Auth + CSRF + Role checks
func apiAuth(h http.HandlerFunc) http.HandlerFunc {
	return authutils.AuthMiddleware(h)
}

func apiProtected(h http.HandlerFunc) http.HandlerFunc {
	return authutils.AuthMiddleware(
		authutils.CSRFMiddleware(
			authutils.RequireRoleMiddleware(authutils.RoleUser, authutils.RoleAdmin)(h),
		),
	)
}

func Handlers() {
	mux := http.NewServeMux()

	// 1) static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 2) API routes
	mux.HandleFunc("/api/home", apiAuth(handlers.HomeAPIHandler))
	mux.HandleFunc("/api/post", apiAuth(handlers.PostDetailAPIHandler))
	mux.HandleFunc("/api/profile", apiAuth(handlers.ProfileAPIHandler))

	mux.HandleFunc("/api/profile/update", apiProtected(handlers.UpdateProfileHandler))

	mux.HandleFunc("/api/posts", apiProtected(handlers.CreatePostAPIHandler))
	mux.HandleFunc("/api/posts/delete", apiProtected(handlers.DeleteAPIHandler))
	mux.HandleFunc("/api/posts/new", apiProtected(handlers.PostCreateInitAPIHandler))
	mux.HandleFunc("/api/posts/react", apiProtected(handlers.LikesHandler))
	mux.HandleFunc("/api/posts/comments", apiAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.CommentsGetHandler(w, r)
			return
		}
		if r.Method == http.MethodPost {
			apiProtected(handlers.CommentsHandler)(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}))

	// ws endopoints
	mux.HandleFunc("/api/dm/threads", apiProtected(handlers.DMThreadsHandler))
	mux.HandleFunc("/api/dm/messages", apiProtected(handlers.DMMessagesHandler))

	// 3) auth endpoints (API)
	mux.HandleFunc("/api/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/login", handlers.LoginHandler)

	// (Recommended) make logout API too:
	mux.HandleFunc("/api/logout", apiAuth(handlers.LogoutHandler))

	mux.HandleFunc("/ws/dm", apiProtected(handlers.DMWebSocketHandler))
	// 4) SPA shell catch-all (PUBLIC)
	mux.HandleFunc("/", apiAuth(handlers.SPAShellHandler))

	// Make this mux the one used by the server
	http.Handle("/", mux)
}

package cmd

import (
	"context"
	"fmt"
	"forum/internal"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const address = "localhost:8080"

func StartServer() {
	internal.CheckArguments()
	database.InitDB()
	internal.Handlers()
	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Could not initialize logger: %v", err)
	}
	defer logger.CloseLogger()

	server := &http.Server{
		Addr:    address,
		Handler: nil,
	}

	fmt.Println("Server running on http://localhost:8080")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()
	// Periodical Login cleanup
	authutils.StartLoginAttemptCleanup()

	// catch interrupt (Ctrl+C or SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop // wait for shutdown signal

	fmt.Println("\nShutting down, cleaning up tokens...")

	if err := database.ClearAllTokens(); err != nil {
		log.Printf("Failed to clear tokens: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	fmt.Println("Server exited cleanly.")
}

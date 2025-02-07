package router

import (
	"net/http"

	"github.com/MogboPython/video-audio-mp3-converter/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func SetupRouter(streamHandler *handlers.StreamHandler, allowedOrigin string) http.Handler {
	r := mux.NewRouter()

	// Add routes
	r.HandleFunc("/convert", streamHandler.HandleUpload).Methods("POST", "OPTIONS")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Return the router with CORS middleware
	return c.Handler(r)
}

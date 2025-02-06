package router

import (
	"net/http"

	"github.com/MogboPython/video-audio-mp3-converter/internal/config"
	"github.com/MogboPython/video-audio-mp3-converter/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func SetupRouter(handler *handlers.StreamHandler, cfg *config.Config) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/convert", handler.HandleUpload).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{cfg.AllowedOrigin},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         86400,
	})

	return c.Handler(r)
}

package text

import (
	"chat-app/internal/storage"
	"chat-app/internal/text"
	"chat-app/internal/utils"
	"log"
	"net/http"
	"time"
)

func RunTextServer() error {
	db, err := storage.SetupDatabaseConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	http.HandleFunc("/text/", func(w http.ResponseWriter, r *http.Request) {
		text.HandleTextRoom(w, r, db)
	})

	server := &http.Server{
		Addr:         ":8082",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
		// TODO: REMOVE THE CORS MIDDLEWARE FOR PROD ENVIRONMENT
		Handler: utils.CorsMiddleware(http.DefaultServeMux),
	}

	log.Println("Text server is running on http://localhost:8082")
	return server.ListenAndServe()
}

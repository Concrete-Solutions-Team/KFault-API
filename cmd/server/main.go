package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/db"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	if err := db.Connect(); err != nil {
		log.Fatal("db connection failed", err)
	}
	log.Println("db connected")
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Whatever"))
	})
	r.Get("/ws", handler.HandleWS)
	fmt.Println("kfault server on port", port)

	http.ListenAndServe(":"+port, r)
}

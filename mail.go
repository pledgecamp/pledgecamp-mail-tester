package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/pledgecamp/mail-tester/controller"
	"github.com/pledgecamp/mail-tester/db"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func setupRouter() *httprouter.Router {
	router := httprouter.New()

	// Mail viewer routes
	router.GET("/", controller.HomeHandler)
	router.GET("/mail/:id", controller.EmailHandler)

	// API routes
	router.GET("/api/messages", controller.GetAllMail)
	router.GET("/api/messages/:id", controller.GetMail)
	router.POST("/api/messages", controller.PostMail)

	router.ServeFiles("/static/*filepath", http.Dir("static"))

	return router
}

func main() {
	log.SetPrefix("Mail ")
	godotenv.Load()
	port := getEnv("PORT", "4020")
	dbSuffix := getEnv("DB_SUFFIX", "dev")
	fmt.Println(fmt.Sprintf("Listening on port %s, db = app-%v.db", port, dbSuffix))

	db.InitDb(false)
	router := setupRouter()
	log.Fatal(http.ListenAndServe(":"+port, router))
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/pledgecamp/mail-tester/controller"
	"github.com/pledgecamp/mail-tester/db"
)

func getMailRouter(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	param := p.ByName("id")
	if param == "latest" {
		controller.GetLatestMail(w)
		return
	}
	id, err := strconv.Atoi(param)
	if err != nil {
		controller.ErrorHandler(w)
		return
	}
	controller.GetMail(w, id)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	dev := (os.Getenv("DEV") == "1")
	fmt.Println(fmt.Sprintf("Listening on port %s, dev = %v", port, dev))

	db.InitDb(false)

	router := httprouter.New()

	// Mail viewer routes
	router.GET("/", controller.HomeHandler)
	router.GET("/mail/:id", controller.EmailHandler)

	// API routes
	router.GET("/api/messages", controller.GetAllMail)
	router.GET("/api/messages/:id", getMailRouter)
	router.POST("/api/messages", controller.PostMail)

	router.ServeFiles("/static/*filepath", http.Dir("static"))
	log.Fatal(http.ListenAndServe(":"+port, router))
}

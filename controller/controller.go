package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/pledgecamp/mail-tester/db"
)

// ViewEmail contains email data and raw email HTML suitable for templates
type ViewEmail struct {
	Email   db.Email
	RawHTML template.HTML
}

func parseId(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

// ErrorHandler is a catchall for displaying 404 responses
func ErrorHandler(w http.ResponseWriter) {
	w.WriteHeader(404)
	t, _ := template.ParseFiles("templates/404.html")
	t.Execute(w, nil)
}

// HomeHandler displays the home page
func HomeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	emails := db.GetAllMail()
	t := template.Must(template.ParseFiles("templates/home.html", "templates/head.html"))
	t.ExecuteTemplate(w, "home", emails)
}

// EmailHandler displays a single Email view
func EmailHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := parseId(p.ByName("id"))
	if err != nil {
		ErrorHandler(w)
		return
	}
	email, err := db.GetMail(id)
	if err != nil {
		ErrorHandler(w)
		return
	}
	viewEmail := ViewEmail{
		Email:   email,
		RawHTML: template.HTML(email.HTML),
	}
	t := template.Must(template.ParseFiles("templates/mail.html", "templates/head.html"))
	t.ExecuteTemplate(w, "mail", viewEmail)
}

func printMail(e *db.Email) {
	if strings.Contains(log.Prefix(), "Mail") {
		log.Println(fmt.Sprintf("From: %s    |    To: %s", e.From, e.To))
		log.Println(fmt.Sprintf("Subject: %s", e.Subject))
		log.Println(fmt.Sprintf("%s\n", e.Text))
	}
}

// PostMail writes new mail to the database and returns the JSON object
func PostMail(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	mail := db.Email{
		To:      r.FormValue("to"),
		From:    r.FormValue("from"),
		Subject: r.FormValue("subject"),
		Text:    r.FormValue("text"),
		HTML:    r.FormValue("html"),
	}

	printMail(&mail)
	mail.ID = db.AddMail(&mail)
	json.NewEncoder(w).Encode(mail)
}

// GetMail writes a JSON object corresponding to the input ID
func GetMail(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	idParam := p.ByName("id")
	if idParam == "latest" {
		GetLatestMail(w)
		return
	}
	id, err := parseId(idParam)
	if err != nil {
		ErrorHandler(w)
		return
	}
	mail, err := db.GetMail(id)
	if err != nil {
		ErrorHandler(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mail)
}

// GetAllMail writes a JSON array of all saved mail
func GetAllMail(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	mail := db.GetAllMail()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mail)
}

// GetLatestMail writes a JSON array of the most recent message
func GetLatestMail(w http.ResponseWriter) {
	mail := db.GetLatestMail()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mail)
}

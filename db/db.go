package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Database driver
)

// Email is a standard email
type Email struct {
	ID      int
	To      string
	From    string
	Subject string
	Text    string
	HTML    string
}

// ErrNoEmail is returned when mail doesn't exist
type ErrNoEmail int

func (e ErrNoEmail) Error() string {
	return fmt.Sprintf("Mail with id=%d doesn't exist", e)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// InitDb initializes the application database
func InitDb(drop bool) {
	db, err := sql.Open("sqlite3", "./app-dev.db")
	if drop {
		statement, err := db.Prepare("DROP TABLE IF EXISTS mail")
		checkError(err)
		statement.Exec()
	}
	statement, err := db.Prepare(
		"CREATE TABLE IF NOT EXISTS mail (id INTEGER PRIMARY KEY, sender TEXT, receiver TEXT, subject TEXT, text TEXT, html TEXT)",
	)
	checkError(err)
	statement.Exec()
	db.Close()
}

// AddMail saves mail to the database
func AddMail(mail *Email) {
	db, err := sql.Open("sqlite3", "./app-dev.db")
	checkError(err)
	statement, err := db.Prepare("INSERT INTO mail (sender, receiver, subject, text, html) VALUES (?, ?, ?, ?, ?)")
	checkError(err)
	statement.Exec(mail.To, mail.From, mail.Subject, mail.Text, mail.HTML)
}

// GetMail gets mail by database ID
func GetMail(id int) (Email, error) {
	db, err := sql.Open("sqlite3", "./app-dev.db")
	checkError(err)
	statement, err := db.Prepare("SELECT sender, receiver, subject, text, html FROM mail WHERE id=?")
	checkError(err)

	mail := Email{ID: id}
	row := statement.QueryRow(id)
	switch err := row.Scan(&mail.To, &mail.From, &mail.Subject, &mail.Text, &mail.HTML); err {
	case sql.ErrNoRows:
		return mail, ErrNoEmail(id)
	case nil:
		return mail, nil
	default:
		checkError(err)
	}
	return mail, err
}

// GetAllMail gets all stored mail
func GetAllMail() []Email {
	db, err := sql.Open("sqlite3", "./app-dev.db")
	checkError(err)
	statement, err := db.Prepare("SELECT id, sender, receiver, subject, SUBSTR(text, 0, 100) FROM mail")
	checkError(err)
	rows, err := statement.Query()
	checkError(err)
	defer rows.Close()
	mailList := make([]Email, 0)
	for rows.Next() {
		mail := Email{}
		err := rows.Scan(&mail.ID, &mail.To, &mail.From, &mail.Subject, &mail.Text)
		checkError(err)
		mailList = append(mailList, mail)
	}
	return mailList
}

// GetLatestMail gets the most recent stored mail
func GetLatestMail() Email {
	db, err := sql.Open("sqlite3", "./app-dev.db")
	checkError(err)
	statement, err := db.Prepare("SELECT id, sender, receiver, subject, text, html FROM mail WHERE id = (SELECT MAX(id) FROM mail)")
	checkError(err)
	mail := Email{}
	row := statement.QueryRow()
	switch err := row.Scan(&mail.ID, &mail.To, &mail.From, &mail.Subject, &mail.Text, &mail.HTML); err {
	case sql.ErrNoRows:
		return Email{ID: 0}
	case nil:
		return mail
	default:
		checkError(err)
	}
	return mail
}

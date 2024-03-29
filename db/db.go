package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // Database driver
)

// Email is a standard email
type Email struct {
	ID      int64
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

func dbOpen() (*sql.DB, error) {
	dbSuffix, _ := os.LookupEnv("DB_SUFFIX")
	return sql.Open("sqlite3", fmt.Sprintf("./app-%v.db", dbSuffix))
}

// InitDb initializes the application database
func InitDb(drop bool) {
	db, err := dbOpen()
	defer db.Close()
	if drop {
		statement, err := db.Prepare("DROP TABLE IF EXISTS mail")
		defer statement.Close()
		checkError(err)
		statement.Exec()
	}
	statement, err := db.Prepare(
		"CREATE TABLE IF NOT EXISTS mail (id INTEGER PRIMARY KEY, sender TEXT, receiver TEXT, subject TEXT, text TEXT, html TEXT)",
	)
	defer statement.Close()
	checkError(err)
	statement.Exec()
}

func DeleteMail(id int64) bool {
	db, err := dbOpen()
	defer db.Close()
	checkError(err)
	statement, err := db.Prepare("DELETE FROM mail WHERE id=?")
	defer statement.Close()
	checkError(err)
	result, err := statement.Exec(id)
	checkError(err)
	rowsAffected, err := result.RowsAffected()
	checkError(err)
	return rowsAffected > 0
}

// AddMail saves mail to the database
func AddMail(mail *Email) int64 {
	db, err := dbOpen()
	defer db.Close()
	checkError(err)
	statement, err := db.Prepare("INSERT INTO mail (sender, receiver, subject, text, html) VALUES (?, ?, ?, ?, ?)")
	defer statement.Close()
	checkError(err)
	result, err := statement.Exec(mail.To, mail.From, mail.Subject, mail.Text, mail.HTML)
	checkError(err)
	id, _ := result.LastInsertId()
	return id
}

// GetMail gets mail by database ID
func GetMail(id int64) (Email, error) {
	db, err := dbOpen()
	defer db.Close()
	checkError(err)
	statement, err := db.Prepare("SELECT sender, receiver, subject, text, html FROM mail WHERE id=?")
	defer statement.Close()
	checkError(err)

	mail := Email{ID: id}
	row := statement.QueryRow(id)
	err = row.Scan(&mail.To, &mail.From, &mail.Subject, &mail.Text, &mail.HTML)
	switch err {
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
	db, err := dbOpen()
	defer db.Close()
	checkError(err)
	statement, err := db.Prepare("SELECT id, sender, receiver, subject, SUBSTR(text, 0, 100) FROM mail ORDER BY id DESC")
	defer statement.Close()
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
	db, err := dbOpen()
	defer db.Close()
	checkError(err)
	statement, err := db.Prepare("SELECT id, sender, receiver, subject, text, html FROM mail WHERE id = (SELECT MAX(id) FROM mail)")
	defer statement.Close()
	checkError(err)
	mail := Email{}
	row := statement.QueryRow()
	err = row.Scan(&mail.ID, &mail.To, &mail.From, &mail.Subject, &mail.Text, &mail.HTML)
	switch err {
	case sql.ErrNoRows:
		return Email{ID: 0}
	case nil:
		return mail
	default:
		checkError(err)
	}
	return mail
}

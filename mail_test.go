package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/pledgecamp/mail-tester/db"
	"github.com/stretchr/testify/assert"
)

func TestHome(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "<title>Mail Tester</title>")
}

func TestGetMail(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	testSubject := "Test subject"
	db.AddMail(&db.Email{
		Subject: testSubject,
	})
	t.Run("Get mail with id=1", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/mail/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), fmt.Sprintf("<div class=\"r2\">%v</div>", testSubject))
	})

	t.Run("Get mail 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/mail/2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})
}

func TestGetMailsAPI(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	t.Run("Get empty message list", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/messages", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 0)
	})

	for i := 1; i <= 3; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		db.AddMail(&db.Email{
			Subject: testSubject,
		})
	}

	t.Run("Get the message list with 3 email messages", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/messages", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 3)
		for i, email := range mailList {
			assert.Equal(t, fmt.Sprintf("Test subject %v", i+1), email.Subject)
		}
	})
}

func TestGetMailAPI(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	for i := 1; i <= 3; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		db.AddMail(&db.Email{
			Subject: testSubject,
		})
	}

	t.Run("Get mail with id=1", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/messages/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		email := db.Email{}
		json.NewDecoder(w.Body).Decode(&email)
		assert.Equal(t, fmt.Sprintf("Test subject 1"), email.Subject)
	})

	t.Run("Get mail with id=99 (404)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/messages/99", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})
}

func TestPostMailAPI(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	testCases := []struct {
		subject string
	}{
		{"Test subject 1"},
		{"Test subject 2"},
		{"Test subject 3"},
	}

	for _, tc := range testCases {
		t.Run("Post mail", func(t *testing.T) {
			w := httptest.NewRecorder()
			data := url.Values{
				"subject": {tc.subject},
			}
			req, _ := http.NewRequest("POST", "/api/messages", bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			email := db.Email{}
			json.NewDecoder(w.Body).Decode(&email)
			assert.Equal(t, tc.subject, email.Subject)
		})
	}
	t.Run("Get the message list with 3 email messages", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/messages", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 3)
	})
}

func TestPostMailAPIDbConnection(t *testing.T) {
	db.InitDb(true)
	router := setupRouter()

	for i := 1; i <= 2<<7; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		t.Run("Post mail", func(t *testing.T) {
			w := httptest.NewRecorder()
			data := url.Values{
				"subject": {testSubject},
			}
			req, _ := http.NewRequest("POST", "/api/messages", bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			email := db.Email{}
			json.NewDecoder(w.Body).Decode(&email)
			assert.Equal(t, testSubject, email.Subject)
		})
	}
}

func TestMain(m *testing.M) {
	os.Setenv("DB_SUFFIX", "test")
	m.Run()
	os.Remove("./app-test.db")
}

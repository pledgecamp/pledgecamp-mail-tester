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

	"github.com/julienschmidt/httprouter"
	"github.com/pledgecamp/mail-tester/db"
	"github.com/stretchr/testify/assert"
)

func initResources() *httprouter.Router {
	db.InitDb(true)
	return setupRouter()
}

func initRecorder(router *httprouter.Router, method, url string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, nil)
	router.ServeHTTP(w, req)
	return w
}

func initFormRecorder(router *httprouter.Router, method, url string, data url.Values) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	return w
}

func initTestRouter(method, url string) *httptest.ResponseRecorder {
	router := initResources()
	return initRecorder(router, method, url)
}

func TestHome(t *testing.T) {
	w := initTestRouter("GET", "/")

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "<title>Mail Tester</title>")
}

func TestGetMail(t *testing.T) {
	router := initResources()

	testSubject := "Test subject"
	db.AddMail(&db.Email{
		Subject: testSubject,
	})
	t.Run("Get mail with id=1", func(t *testing.T) {
		w := initRecorder(router, "GET", "/mail/1")

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), fmt.Sprintf("<div class=\"r2\">%v</div>", testSubject))
	})

	t.Run("Get mail 404", func(t *testing.T) {
		w := initRecorder(router, "GET", "/mail/2")

		assert.Equal(t, 404, w.Code)
	})
}

func TestGetMailsAPI(t *testing.T) {
	router := initResources()

	t.Run("Get empty message list", func(t *testing.T) {
		w := initRecorder(router, "GET", "/api/messages")

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 0)
	})

	expectedSubjects := make([]string, 3)
	for i := 1; i <= 3; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		db.AddMail(&db.Email{
			Subject: testSubject,
		})
		expectedSubjects = append([]string{testSubject}, expectedSubjects...)
	}

	t.Run("Get the message list with 3 email messages", func(t *testing.T) {
		w := initRecorder(router, "GET", "/api/messages")

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 3)
		mailId := mailList[0].ID
		for i, email := range mailList {
			assert.GreaterOrEqual(t, mailId, email.ID)
			assert.Equal(t, expectedSubjects[i], email.Subject)
			mailId = email.ID
		}
	})
}

func TestGetMailAPI(t *testing.T) {
	router := initResources()

	for i := 1; i <= 3; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		db.AddMail(&db.Email{
			Subject: testSubject,
		})
	}

	t.Run("Get mail with id=1", func(t *testing.T) {
		w := initRecorder(router, "GET", "/api/messages/1")

		assert.Equal(t, 200, w.Code)
		email := db.Email{}
		json.NewDecoder(w.Body).Decode(&email)
		assert.Equal(t, fmt.Sprintf("Test subject 1"), email.Subject)
	})

	t.Run("Get mail with id=99 (404)", func(t *testing.T) {
		w := initRecorder(router, "GET", "/api/messages/99")

		assert.Equal(t, 404, w.Code)
	})
}

func TestPostMailAPI(t *testing.T) {
	router := initResources()

	testCases := []struct {
		subject string
	}{
		{"Test subject 1"},
		{"Test subject 2"},
		{"Test subject 3"},
	}

	for _, tc := range testCases {
		t.Run("Post mail", func(t *testing.T) {
			data := url.Values{
				"subject": {tc.subject},
			}
			w := initFormRecorder(router, "POST", "/api/messages", data)

			assert.Equal(t, 200, w.Code)
			email := db.Email{}
			json.NewDecoder(w.Body).Decode(&email)
			assert.Equal(t, tc.subject, email.Subject)
		})
	}
	t.Run("Get the message list with 3 email messages", func(t *testing.T) {
		w := initRecorder(router, "GET", "/api/messages")

		assert.Equal(t, 200, w.Code)
		mailList := make([]db.Email, 0)
		json.NewDecoder(w.Body).Decode(&mailList)
		assert.Len(t, mailList, 3)
	})
}

func TestPostMailAPIDbConnection(t *testing.T) {
	router := initResources()

	for i := 1; i <= 2<<7; i++ {
		testSubject := fmt.Sprintf("Test subject %v", i)
		t.Run("Post mail", func(t *testing.T) {
			data := url.Values{
				"subject": {testSubject},
			}
			w := initFormRecorder(router, "POST", "/api/messages", data)

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

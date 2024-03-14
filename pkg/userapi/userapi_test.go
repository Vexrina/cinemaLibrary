package userapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"

	"github.com/vexrina/cinemaLibrary/pkg/userapi"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

func TestRegisterHandler(t *testing.T) {
	// create test db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// create handler with test db
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userapi.RegisterHandler(w, r, db)
	})

	// create test request
	user := types.User{Username: "testuser", Email: "test@example.com", Password: "testpassword"}
	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Set header Content-Type for JSON
	req.Header.Set("Content-Type", "application/json")

	// Create expected request to db
	mock.ExpectQuery("SELECT COUNT(.+)").WithArgs(user.Username, user.Email).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO users").WithArgs(user.Username, user.Email, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	// Send test request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// check statuscode
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check that all request completed successfully 
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestLoginHandler(t *testing.T) {
	//create test db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// create handler with test db
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userapi.LoginHandler(w, r, db)
	})

	// create test request
	user := types.User{Email: "test@example.com", Password: "testpassword"}
	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Set header Content-Type for JSON
	req.Header.Set("Content-Type", "application/json")

	// Hash password for compare
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	// Create expected request to db
	mock.ExpectQuery("SELECT password").WithArgs(user.Email).WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(hashedPassword))

	// Send test request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// check statuscode
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
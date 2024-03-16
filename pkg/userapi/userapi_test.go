package userapi_test

import (
	"bytes"
	"encoding/json"

	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/userapi"
)

func TestRegisterHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userapi.RegisterHandler(w, r, orm)
	})


	tests := []struct {
		name         string
		requestBody  interface{}
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Successful registration",
			requestBody:  map[string]string{"username": "testuser", "email": "test@example.com", "password": "testpassword"},
			expectedCode: http.StatusCreated,
			expectedBody: "",
		},
		{
			name:         "Username or email already exists",
			requestBody:  map[string]string{"username": "existinguser", "email": "existing@example.com", "password": "testpassword"},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Username or email already exists\n",
		},
		{
			name:         "Bad request due to malformed JSON",
			requestBody:  map[string]string{},
			expectedCode: http.StatusBadRequest,
			expectedBody: "missing required fields\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest("POST", "/register", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}

			if tt.name != "Bad request due to malformed JSON" {
				if tt.name == "Username or email already exists" {
					mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				} else {
					mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
					mock.ExpectExec("INSERT INTO users").WithArgs("testuser", "test@example.com", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			if tt.name != "Bad request due to malformed JSON" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
func TestLoginHandler_SuccessfulLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userapi.LoginHandler(w, r, orm)
	})

	mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").WithArgs("test@example.com").WillReturnRows(sqlmock.NewRows([]string{"password", "adminflag"}).AddRow("$2a$10$jbRk/x7EcY7yM7jjLo/uYuCfJ48pJXQo2nFpPOJg.4LNmlvX3JPIG", false))

	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, response["token"])
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userapi.LoginHandler(w, r, orm)
    })

    mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").WithArgs("test@example.com").WillReturnError(errors.New("invalid credentials"))

    requestBody := map[string]string{
        "email":    "test@example.com",
        "password": "testpassword",
    }
    body, _ := json.Marshal(requestBody)

    req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLoginHandler_UnsuccessfulLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userapi.LoginHandler(w, r, orm)
	})

	mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").WithArgs("test@example.com").WillReturnRows(sqlmock.NewRows([]string{"password", "adminflag"}).AddRow("$2a$Y7yM7jjLo/uYuCfJ48pJXQo2nFpPOJg.4LNmlvX3JPIG", false))

	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRegisterHandler_InternalServerError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userapi.RegisterHandler(w, r, orm)
    })

    mock.ExpectQuery("SELECT count\\(\\*\\) FROM users WHERE username=? OR email=?").WithArgs("existinguser", "existing@example.com").WillReturnError(errors.New("database error"))

    requestBody := map[string]string{
        "username": "existinguser",
        "email":    "existing@example.com",
        "password": "testpassword",
    }
    body, _ := json.Marshal(requestBody)

    req, err := http.NewRequest("POST", "/register", bytes.NewReader(body))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

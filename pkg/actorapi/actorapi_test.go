package actorapi_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/vexrina/cinemaLibrary/pkg/actorapi"
	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

func TestCreateActorHandler_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectExec(`INSERT INTO actors \(name, gender, date_of_birth\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("John Doe", "male", "01.01.2000").
		WillReturnResult(sqlmock.NewResult(1, 1)).
		WillReturnError(nil)

	jsonData := []byte(`{"name": "John Doe", "gender":"male", "birthdate":"01.01.2000"}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorapi.CreateActorHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestUpdateActorHandler_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("UPDATE actors SET name = \\$1, gender = \\$2, date_of_birth = \\$3 WHERE id = \\$4").
        WithArgs("John Doe", "male", "2000-01-01", 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    jsonData := []byte(`{"ID": 1, "Name": "John Doe", "Gender": "male", "Birthdate": "2000-01-01"}`)
    req, err := http.NewRequest("PUT", "/", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.UpdateActorHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestDeleteActorHandler_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE actor_id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectExec("DELETE FROM actors WHERE id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    jsonData := []byte(`{"ID": 1}`)
    req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatal(err)
    }
	
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.DeleteActorHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestDeleteActorHandler_DecodingError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE actor_id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectExec("DELETE FROM actors WHERE id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    jsonData := []byte(`{"Name": "John Doe"}`)
    req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatal(err)
    }

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorapi.DeleteActorHandler(w, r, orm)
    })
	
	rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code, "status code is not BadRequest")
}

func TestGetActorsHandler_Success_AllActors(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    expectedActors := []types.ActorWithFilms{
        {ID: 1, Name: "John Doe", FilmTitles: []string{"Film 1", "Film 2"}},
        {ID: 2, Name: "Jane Smith", FilmTitles: []string{"Film 3", "Film 4"}},
    }

    mock.ExpectQuery("SELECT id, name FROM actors").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
            AddRow(expectedActors[0].ID, expectedActors[0].Name).
            AddRow(expectedActors[1].ID, expectedActors[1].Name))

    for _, actor := range expectedActors {
        mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = ?").
            WithArgs(actor.ID).
            WillReturnRows(sqlmock.NewRows([]string{"title"}).
                AddRow(actor.FilmTitles[0]).
                AddRow(actor.FilmTitles[1]))
    }

    req, err := http.NewRequest("GET", "/actors", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.GetActorsHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expectedResponse := `[{"id":1,"name":"John Doe","film_titles":["Film 1","Film 2"]},{"id":2,"name":"Jane Smith","film_titles":["Film 3","Film 4"]}]`
	if strings.TrimSpace(rr.Body.String()) != expectedResponse {
		fmt.Printf("expectedResponse: %v__\n", expectedResponse)
		// idk why, but rr.Body.String() return body with \n on end, so...
		fmt.Printf("gottedResponse:   %v__\n", strings.TrimSpace(rr.Body.String()))
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestGetActorsHandler_Success_ByFragment(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    expectedActors := []types.ActorWithFilms{
        {ID: 1, Name: "John Doe", FilmTitles: []string{"Film 1", "Film 2"}},
    }

    fragment := "Doe"

    mock.ExpectQuery("SELECT id, name FROM actors WHERE name LIKE ?").
        WithArgs("%"+fragment+"%").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
            AddRow(expectedActors[0].ID, expectedActors[0].Name))

    for _, actor := range expectedActors {
        mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = ?").
            WithArgs(actor.ID).
            WillReturnRows(sqlmock.NewRows([]string{"title"}).
                AddRow(actor.FilmTitles[0]).
                AddRow(actor.FilmTitles[1]))
    }

    req, err := http.NewRequest("GET", "/actors?fragment="+fragment, nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.GetActorsHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expectedResponse := `[{"id":1,"name":"John Doe","film_titles":["Film 1","Film 2"]}]`
    if strings.TrimSpace(rr.Body.String()) != expectedResponse {
        t.Errorf("handler returned unexpected body:\ngot %v\nwant %v", strings.TrimSpace(rr.Body.String()), expectedResponse)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestGetActorsHandler_InternalServerError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    expectedError := errors.New("database error")
    mock.ExpectQuery("SELECT id, name FROM actors").
        WillReturnError(expectedError)

    req, err := http.NewRequest("GET", "/actors", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.GetActorsHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusInternalServerError {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
    }

    expectedErrorResponse := "database error\n"
    if rr.Body.String() != expectedErrorResponse {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedErrorResponse)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestGetActorsHandler_FragmentInternalServerError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    expectedError := errors.New("database error")
    mock.ExpectQuery("SELECT id, name FROM actors").
        WillReturnError(expectedError)

    req, err := http.NewRequest("GET", "/actors?fragment=fragment", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        actorapi.GetActorsHandler(w, r, orm)
    })

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusInternalServerError {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
    }

    expectedErrorResponse := "database error\n"
    if rr.Body.String() != expectedErrorResponse {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedErrorResponse)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}
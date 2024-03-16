package filmapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/vexrina/cinemaLibrary/pkg/filmapi"
	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

func TestCreateFilmHandler_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	fakeFilm := types.Film{
		Title:       "Test Film",
		Description: "This is a test film",
		ReleaseDate: "2024-03-15",
		Rating:      8.5,
		Actors:      []int{1, 2, 3},
	}

	mock.ExpectQuery("INSERT INTO films").WithArgs(fakeFilm.Title, fakeFilm.Description, fakeFilm.ReleaseDate, fakeFilm.Rating).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	for _, actorID := range fakeFilm.Actors {
		mock.ExpectExec("INSERT INTO film_actors").WithArgs(1, actorID).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	body, err := json.Marshal(fakeFilm)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec("INSERT INTO films").
		WithArgs(fakeFilm.Title, fakeFilm.Description, fakeFilm.ReleaseDate, fakeFilm.Rating).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.CreateFilmHandler(w, r, orm)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestCreateFilmHandler_DatabaseError(t *testing.T) {
	fakeFilm := types.Film{
		Title:       "Test Film",
		Description: "This is a test film",
		ReleaseDate: "18.03.2023",
		Rating:      9.9,
		Actors:      []int{1, 2},
	}
	body, err := json.Marshal(fakeFilm)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	orm := orm.NewORM(db)
	mock.ExpectExec("INSERT INTO films").
		WithArgs(fakeFilm.Title, fakeFilm.Description, fakeFilm.ReleaseDate, fakeFilm.Rating, fakeFilm.Actors).
		WillReturnError(errors.New("database error"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.CreateFilmHandler(w, r, orm)
	})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUpdateFilmHandler_Success(t *testing.T) {
	fakeFilm := types.Film{
		ID:          1,
		Title:       "Updated Film Title",
		Description: "This is an updated film description",
		ReleaseDate: "2024-03-16",
		Rating:      9.0,
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	orm := orm.NewORM(db)
	mock.ExpectExec("UPDATE films").WithArgs(fakeFilm.ID, fakeFilm.Title, fakeFilm.Description, fakeFilm.ReleaseDate, fakeFilm.Rating).WillReturnResult(sqlmock.NewResult(1, 1))

	body, err := json.Marshal(fakeFilm)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PATCH", "/", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.UpdateFilmHandler(w, r, orm)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
func TestUpdateFilmHandler_Error(t *testing.T) {
	fakeFilm := types.Film{
		ID:          1,
		Title:       "Updated Film Title",
		Description: "This is an updated film description",
		ReleaseDate: "2024-03-16",
		Rating:      9.0,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	orm := orm.NewORM(db)

	mock.ExpectExec("UPDATE films").WithArgs(fakeFilm.ID, fakeFilm.Title, fakeFilm.Description, fakeFilm.ReleaseDate, fakeFilm.Rating).WillReturnError(errors.New("database error"))

	body, err := json.Marshal(fakeFilm)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PATCH", "/", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.UpdateFilmHandler(w, r, orm)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestIsValidEnumType(t *testing.T) {
	validValues := []filmapi.EnumType{filmapi.EnumValue1, filmapi.EnumValue2, filmapi.EnumValue3}
	for _, value := range validValues {
		if !filmapi.IsValidEnumType(value) {
			t.Errorf("Expected %s to be valid, but it was not", value)
		}
	}

	invalidValues := []filmapi.EnumType{"invalidValue", "", "random"}
	for _, value := range invalidValues {
		if filmapi.IsValidEnumType(value) {
			t.Errorf("Expected %s to be invalid, but it was not", value)
		}
	}
}

func TestValidateEnumType(t *testing.T) {
	validValues := []filmapi.EnumType{filmapi.EnumValue1, filmapi.EnumValue2, filmapi.EnumValue3}
	for _, value := range validValues {
		err := filmapi.ValidateEnumType(value)
		if err != nil {
			t.Errorf("Expected %s to be valid, but it was not: %v", value, err)
		}
	}

	invalidValues := []filmapi.EnumType{"invalidValue", "", "random"}
	for _, value := range invalidValues {
		err := filmapi.ValidateEnumType(value)
		if err == nil {
			t.Errorf("Expected %s to be invalid, but it was not", value)
		} else {
			expectedErrorMessage := "invalid enum value"
			if err.Error() != expectedErrorMessage {
				t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, err.Error())
			}
		}
	}
}

func TestReturnAnswer_Success(t *testing.T) {
	fakeResponse := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	filmapi.ReturnAnswer(fakeResponse, rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be application/json, got %s", contentType)
	}

	var films []types.Film
	err = json.Unmarshal(rr.Body.Bytes(), &films)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(films) != len(fakeResponse) {
		t.Errorf("Expected %d films, got %d", len(fakeResponse), len(films))
	}
	for i, film := range films {
		if film.ID != fakeResponse[i].ID || film.Title != fakeResponse[i].Title || film.Description != fakeResponse[i].Description || film.ReleaseDate != fakeResponse[i].ReleaseDate || film.Rating != fakeResponse[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", fakeResponse[i], film)
		}
	}
}

func TestGetFilmsHandler_Success_Default(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT \\* FROM films ORDER BY rating DESC").
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetFilmsHandler_Success_DefaultWithAsc(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT \\* FROM films ORDER BY rating ASC").
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/?asc=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetFilmsHandler_Success_AlternativeSort(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT \\* FROM films ORDER BY release_date DESC").
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/?sortby=release_date", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
func TestGetFilmsHandler_Success_AlternativeSortWithAsc(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT \\* FROM films ORDER BY title ASC").
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/?sortby=title&asc=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetFilmsHandler_Success_SearchByActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	query := `
	SELECT f.id, f.title, f.description, f.release_date, f.rating
        FROM films AS f
        JOIN film_actors AS fa ON f.id = fa.film_id
        JOIN actors AS a ON fa.actor_id = a.id
        WHERE a.name LIKE '%' || $1 || '%'
		`
	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
			AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
			AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0))

	mock.ExpectExec("INSERT INTO films_actor").WithArgs(1, "John Doe").WillReturnResult(sqlmock.NewResult(1, 1))

	req, err := http.NewRequest("GET", "/?actor=John%20Doe", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetFilmsHandler_Success_SearchByTitle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Matrix", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "John Wik", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT * FROM films WHERE title LIKE '%' || $1 || '%'").
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/?title=John", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetFilmsHandler_Success_SearchByBoth(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Matrix", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "John Wik", "Description 2", "2023-01-01", 8.0)
	
	
	queryByActor := `
	SELECT f.id, f.title, f.description, f.release_date, f.rating
	FROM films AS f
	JOIN film_actors AS fa ON f.id = fa.film_id
	JOIN actors AS a ON fa.actor_id = a.id
	WHERE a.name LIKE '%' || $1 || '%'
	`
	queryByTitle := `
	SELECT id, title, description, release_date, rating
	FROM films
	WHERE title LIKE '%' || $1 || '%'
	`
	query := queryByActor + " UNION ALL " + queryByTitle
		
	mock.ExpectQuery(query).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO films_actor").WithArgs(1, "John Doe").WillReturnResult(sqlmock.NewResult(1, 1))
	

	req, err := http.NewRequest("GET", "/?actor_title=John", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.GetFilmsHandler(w, r, orm)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}


func TestDeleteFilmHandler_SuccessfulDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM film_actors WHERE film_id = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("DELETE FROM films WHERE id = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	orm := orm.NewORM(db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.DeleteFilmHandler(w, r, orm)
	})

	jsonData := []byte(`{"ID": 1}`)

	req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "status code is not OK")
}

func TestDeleteFilmHandler_DecodingError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filmapi.DeleteFilmHandler(w, r, orm)
	})

	jsonData := []byte(`{"Title": "Film 1", "Description": "Description 1", "ReleaseDate": "2022-01-01", "Rating": 7.5}`)

	req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code, "status code is not BadRequest")
}
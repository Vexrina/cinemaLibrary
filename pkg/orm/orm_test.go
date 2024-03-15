package orm_test

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

// endpoint /actors

// post
func TestCreateActor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	actor := types.Actor{
		Name:      "John Doe",
		Gender:    "Male",
		Birthdate: "1990-01-01",
	}

	mock.ExpectExec("INSERT INTO actors").
		WithArgs(actor.Name, actor.Gender, actor.Birthdate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = orm.CreateActor(actor)
	if err != nil {
		t.Errorf("Error creating actor: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestCreateActor_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectExec("INSERT INTO actors").
		WillReturnError(fmt.Errorf("database error"))

	err = orm.CreateActor(types.Actor{})

	if err == nil {
		t.Errorf("Expected an error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// patch
func TestUpdateActor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	actor := types.Actor{
		ID:        1,
		Name:      "Jane Doe",
		Gender:    "Female",
		Birthdate: "1985-02-02",
	}

	mock.ExpectExec("UPDATE actors").
		WithArgs(actor.Name, actor.Gender, actor.Birthdate, actor.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = orm.UpdateActor(actor)
	if err != nil {
		t.Errorf("Error updating actor: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestUpdateActor_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectExec("UPDATE actors SET").
		WillReturnError(fmt.Errorf("database error"))

	err = orm.UpdateActor(types.Actor{})

	if err == nil {
		t.Errorf("Expected an error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// delete
func TestDeleteActorByID_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock database connection: %v", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE actor_id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 0))

    mock.ExpectExec("DELETE FROM actors WHERE id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 1))

    err = orm.DeleteActorByID(1)
    if err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

func TestDeleteActorByID_DeleteFilmActorsError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock database connection: %v", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE actor_id = \\$1").
        WithArgs(1).
        WillReturnError(errors.New("ошибка удаления"))

    err = orm.DeleteActorByID(1)
    if err == nil || err.Error() != "ошибка удаления" {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

// get
// utility for test:
func compareStringSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}
	return true
}

// utility func for get
func TestGetFilmsWithActor_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"title"}).AddRow("Film 1").AddRow("Film 2")

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	titles, err := orm.GetFilmsWithActor(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{"Film 1", "Film 2"}
	if !compareStringSlices(titles, expected) {
		t.Errorf("Expected titles %v, got %v", expected, titles)
	}
}

func TestGetFilmsWithActor_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	_, err = orm.GetFilmsWithActor(1)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// all actors
func TestGetActors_Success_NoFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Actor 1").
		AddRow(2, "Actor 2")

	mock.ExpectQuery("SELECT id, name FROM actors").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	actors, err := orm.GetActors()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(actors) != 2 {
		t.Errorf("Expected 2 actors, got %d", len(actors))
	}

	expected := []types.ActorWithFilms{
		{ID: 1, Name: "Actor 1", FilmTitles: nil},
		{ID: 2, Name: "Actor 2", FilmTitles: nil},
	}

	for i, actor := range actors {
		if actor.ID != expected[i].ID || actor.Name != expected[i].Name {
			t.Errorf("Mismatch in actor details. Expected %v, got %v", expected[i], actor)
		}
	}
}
func TestGetActors_Success_WithFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Actor 1").
		AddRow(2, "Actor 2")

	mock.ExpectQuery("SELECT id, name FROM actors").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("Film 1").AddRow("Film 2"))

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("Film 3"))

	actors, err := orm.GetActors()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(actors) != 2 {
		t.Errorf("Expected 2 actors, got %d", len(actors))
	}

	expected := []types.ActorWithFilms{
		{ID: 1, Name: "Actor 1", FilmTitles: []string{"Film 1", "Film 2"}},
		{ID: 2, Name: "Actor 2", FilmTitles: []string{"Film 3"}},
	}

	for i, actor := range actors {
		if actor.ID != expected[i].ID || actor.Name != expected[i].Name {
			t.Errorf("Mismatch in actor details. Expected %v, got %v", expected[i], actor)
		}
		if !reflect.DeepEqual(actor.FilmTitles, expected[i].FilmTitles) {
			t.Errorf("Mismatch in film titles for actor %d. Expected %v, got %v", actor.ID, expected[i].FilmTitles, actor.FilmTitles)
		}
	}
}
func TestGetActors_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT id, name FROM actors").
		WillReturnError(errors.New("database error"))

	_, err = orm.GetActors()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
func TestGetActors_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Actor 1").
		AddRow(2, "Actor 2")

	mock.ExpectQuery("SELECT id, name FROM actors").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnError(errors.New("scan error"))

	_, err = orm.GetActors()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
func TestGetActors_GetFilmsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Actor 1").
		AddRow(2, "Actor 2")

	mock.ExpectQuery("SELECT id, name FROM actors").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnError(errors.New("get films error"))

	_, err = orm.GetActors()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// fragment actors
func TestGetActorsWithFragment_Success_NoFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	actorFragment := "John"

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "John Doe").
		AddRow(2, "Johnny Walker")

	mock.ExpectQuery("SELECT id, name FROM actors WHERE name LIKE ?").
		WithArgs("%" + actorFragment + "%").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	actors, err := orm.GetActorsWithFragment(actorFragment)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(actors) != 2 {
		t.Errorf("Expected 2 actors, got %d", len(actors))
	}

	expected := []types.ActorWithFilms{
		{ID: 1, Name: "John Doe", FilmTitles: nil},
		{ID: 2, Name: "Johnny Walker", FilmTitles: nil},
	}

	for i, actor := range actors {
		if actor.ID != expected[i].ID || actor.Name != expected[i].Name {
			t.Errorf("Mismatch in actor details. Expected %v, got %v", expected[i], actor)
		}
	}
}

func TestGetActorsWithFragment_Success_WithFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	actorFragment := "John"

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "John Doe").
		AddRow(2, "Johnny Walker")

	mock.ExpectQuery("SELECT id, name FROM actors WHERE name LIKE ?").
		WithArgs("%" + actorFragment + "%").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("Film 1").AddRow("Film 2"))

	mock.ExpectQuery("SELECT f.title FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id WHERE fa.actor_id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("Film 3"))

	actors, err := orm.GetActorsWithFragment(actorFragment)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(actors) != 2 {
		t.Errorf("Expected 2 actors, got %d", len(actors))
	}

	expected := []types.ActorWithFilms{
		{ID: 1, Name: "John Doe", FilmTitles: []string{"Film 1", "Film 2"}},
		{ID: 2, Name: "Johnny Walker", FilmTitles: []string{"Film 3"}},
	}

	for i, actor := range actors {
		if actor.ID != expected[i].ID || actor.Name != expected[i].Name {
			t.Errorf("Mismatch in actor details. Expected %v, got %v", expected[i], actor)
		}
	}
}

func TestGetActorsWithFragment_Error_DB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	actorFragment := "John"

	mock.ExpectQuery("SELECT id, name FROM actors WHERE name LIKE ?").
		WithArgs("%" + actorFragment + "%").
		WillReturnError(errors.New("database error"))

	_, err = orm.GetActorsWithFragment(actorFragment)
	if err == nil {
		t.Error("Expected an error, but got nil")
	}
}

// endpoint /film
// post
func TestCreateFilm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mockFilm := types.Film{
		Title:       "Test Film",
		Description: "This is a test film",
		ReleaseDate: "2024-03-15",
		Rating:      8.5,
		Actors:      []int{1, 2, 3},
	}

	mock.ExpectQuery("INSERT INTO films").WithArgs(mockFilm.Title, mockFilm.Description, mockFilm.ReleaseDate, mockFilm.Rating).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	for _, actorID := range mockFilm.Actors {
		mock.ExpectExec("INSERT INTO film_actors").WithArgs(1, actorID).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	filmID, err := orm.CreateFilm(mockFilm)

	assert.NoError(t, err)
	assert.NotEqual(t, 0, filmID)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateFilm_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mockFilm := types.Film{
		Title:       "Test Film",
		Description: "This is a test film",
		ReleaseDate: "2024-03-15",
		Rating:      8.5,
		Actors:      []int{1, 2, 3},
	}

	mock.ExpectQuery("INSERT INTO films").WithArgs(mockFilm.Title, mockFilm.Description, mockFilm.ReleaseDate, mockFilm.Rating).WillReturnError(errors.New("database error"))

	filmID, err := orm.CreateFilm(mockFilm)

	assert.Error(t, err)
	assert.Equal(t, 0, filmID)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// patch
func TestUpdateFilm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mockFilm := types.Film{
		ID:          1,
		Title:       "Updated Film Title",
		Description: "This is an updated film description",
		ReleaseDate: "2024-03-16",
		Rating:      9.0,
	}

	mock.ExpectExec("UPDATE films").WithArgs(mockFilm.ID, mockFilm.Title, mockFilm.Description, mockFilm.ReleaseDate, mockFilm.Rating).WillReturnResult(sqlmock.NewResult(1, 1))

	err = orm.UpdateFilm(mockFilm)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUpdateFilm_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mockFilm := types.Film{
		ID:          1,
		Title:       "Updated Film Title",
		Description: "This is an updated film description",
		ReleaseDate: "2024-03-16",
		Rating:      9.0,
	}

	mock.ExpectExec("UPDATE films").WithArgs(mockFilm.ID, mockFilm.Title, mockFilm.Description, mockFilm.ReleaseDate, mockFilm.Rating).WillReturnError(errors.New("database error"))

	err = orm.UpdateFilm(mockFilm)
	assert.Error(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// get
func TestGetFilms_Success_DefaultSortAscending(t *testing.T) {
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

	films, err := orm.GetFilms("", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestGetFilms_Success_SortByTitleAscending(t *testing.T) {
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

	films, err := orm.GetFilms("title", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestGetFilms_Success_SortByReleaseDateDescending(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5)

	mock.ExpectQuery("SELECT \\* FROM films ORDER BY release_date DESC").
		WillReturnRows(rows)

	films, err := orm.GetFilms("release_date", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestSearchFilmsByFragment_Success_WithActorFragment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%' UNION ALL SELECT id, title, description, release_date, rating FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
		WithArgs("ActorFragment").
		WillReturnRows(rows)

	films, err := orm.SearchFilmsByFragment("ActorFragment")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestSearchFilmsByFragment_Success_WithTitleFragment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)
	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%' UNION ALL SELECT id, title, description, release_date, rating FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
		WithArgs("TitleFragment").
		WillReturnRows(rows)

	films, err := orm.SearchFilmsByFragment("TitleFragment")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestSearchFilmsByFragment_Success_WithBothFragment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
		AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
		AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

	mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%' UNION ALL SELECT id, title, description, release_date, rating FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
		WithArgs("Fragment").
		WillReturnRows(rows)

	films, err := orm.SearchFilmsByFragment("Fragment")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}

	expected := []types.Film{
		{ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
		{ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
	}

	for i, film := range films {
		if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
			t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
		}
	}
}

func TestSearchFilmsByActorFragment_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
        AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
        AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

    mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Actor").
        WillReturnRows(rows)

    films, err := orm.SearchFilmsByActorFragment("Actor")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if len(films) != 2 {
        t.Errorf("Expected 2 films, got %d", len(films))
    }

    expected := []types.Film{
        {ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
        {ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
    }

    for i, film := range films {
        if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
            t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
        }
    }
}

func TestSearchFilmsByActorFragment_EmptyResult(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"})

    mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Actor").
        WillReturnRows(rows)

    films, err := orm.SearchFilmsByActorFragment("Actor")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if len(films) != 0 {
        t.Errorf("Expected 0 films, got %d", len(films))
    }
}

func TestSearchFilmsByActorFragment_DatabaseError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectQuery("SELECT f.id, f.title, f.description, f.release_date, f.rating FROM films AS f JOIN film_actors AS fa ON f.id = fa.film_id JOIN actors AS a ON fa.actor_id = a.id WHERE a.name LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Actor").
        WillReturnError(errors.New("database error"))

    _, err = orm.SearchFilmsByActorFragment("Actor")
    if err == nil || err.Error() != "database error" {
        t.Errorf("Expected database error, got %v", err)
    }
}

func TestSearchFilmsByTitleFragment_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"}).
        AddRow(1, "Film 1", "Description 1", "2022-01-01", 7.5).
        AddRow(2, "Film 2", "Description 2", "2023-01-01", 8.0)

    mock.ExpectQuery("SELECT \\* FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Fragment").
        WillReturnRows(rows)

    films, err := orm.SearchFilmsByTitleFragment("Fragment")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if len(films) != 2 {
        t.Errorf("Expected 2 films, got %d", len(films))
    }

    expected := []types.Film{
        {ID: 1, Title: "Film 1", Description: "Description 1", ReleaseDate: "2022-01-01", Rating: 7.5},
        {ID: 2, Title: "Film 2", Description: "Description 2", ReleaseDate: "2023-01-01", Rating: 8.0},
    }

    for i, film := range films {
        if film.ID != expected[i].ID || film.Title != expected[i].Title || film.Description != expected[i].Description || film.ReleaseDate != expected[i].ReleaseDate || film.Rating != expected[i].Rating {
            t.Errorf("Mismatch in film details. Expected %v, got %v", expected[i], film)
        }
    }
}

func TestSearchFilmsByTitleFragment_EmptyResult(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    rows := sqlmock.NewRows([]string{"id", "title", "description", "release_date", "rating"})

    mock.ExpectQuery("SELECT \\* FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Fragment").
        WillReturnRows(rows)

    films, err := orm.SearchFilmsByTitleFragment("Fragment")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if len(films) != 0 {
        t.Errorf("Expected 0 films, got %d", len(films))
    }
}

func TestSearchFilmsByTitleFragment_DatabaseError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectQuery("SELECT \\* FROM films WHERE title LIKE '%' \\|\\| \\$1 \\|\\| '%'").
        WithArgs("Fragment").
        WillReturnError(errors.New("database error"))

    _, err = orm.SearchFilmsByTitleFragment("Fragment")
    if err == nil || err.Error() != "database error" {
        t.Errorf("Expected database error, got %v", err)
    }
}

func TestDeleteFilmByID_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE film_id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 0))

    mock.ExpectExec("DELETE FROM films WHERE id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 1))

    err = orm.DeleteFilmByID(1)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}

func TestDeleteFilmByID_DeleteFilmActorsError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE film_id = \\$1").
        WithArgs(1).
        WillReturnError(errors.New("delete error"))

    err = orm.DeleteFilmByID(1)
    if err == nil || err.Error() != "delete error" {
        t.Errorf("Expected delete error, got %v", err)
    }
}

func TestDeleteFilmByID_DeleteFilmError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    mock.ExpectExec("DELETE FROM film_actors WHERE film_id = \\$1").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 0))

    mock.ExpectExec("DELETE FROM films WHERE id = \\$1").
        WithArgs(1).
        WillReturnError(errors.New("delete error"))

    err = orm.DeleteFilmByID(1)
    if err == nil || err.Error() != "delete error" {
        t.Errorf("Expected delete error, got %v", err)
    }
}

// endpoint /users
// utility function
func TestCountUsersWithUsernameAndEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT COUNT").WithArgs("test_username", "test_email").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	count, err := orm.CountUsersWithUsernameAndEmail("test_username", "test_email")

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCountUsersWithUsernameAndEmail_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT COUNT").WithArgs("test_username", "test_email").WillReturnError(errors.New("database error"))

	count, err := orm.CountUsersWithUsernameAndEmail("test_username", "test_email")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// post
func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectExec("INSERT INTO users").WithArgs("test_username", "test_email", "hashed_password").WillReturnResult(sqlmock.NewResult(1, 1))

	err = orm.CreateUser("test_username", "test_email", "hashed_password")

	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateUser_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectExec("INSERT INTO users").WithArgs("test_username", "test_email", "hashed_password").WillReturnError(errors.New("database error"))

	err = orm.CreateUser("test_username", "test_email", "hashed_password")
	assert.Error(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// utility function
func TestGetUserPasswordByEmail_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Ошибка '%s' при инициализации mock базы данных", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    email := "test@example.com"
    expectedPassword := "hashedPassword"
    expectedAdminFlag := true

    mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").
        WithArgs(email).
        WillReturnRows(sqlmock.NewRows([]string{"password", "adminflag"}).AddRow(expectedPassword, expectedAdminFlag))

    storedPassword, adminflag, err := orm.GetUserPasswordByEmail(email)
    if err != nil {
        t.Errorf("Неожиданная ошибка: %v", err)
    }

    if storedPassword != expectedPassword || adminflag != expectedAdminFlag {
        t.Errorf("Полученные данные не соответствуют ожидаемым")
    }
}

func TestGetUserPasswordByEmail_UserNotFound(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Ошибка '%s' при инициализации mock базы данных", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    email := "nonexistent@example.com"

    mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, _, err = orm.GetUserPasswordByEmail(email)
    if err == nil || !errors.Is(err, sql.ErrNoRows) {
        t.Errorf("Ожидалась ошибка о отсутствии пользователя, получено %v", err)
    }
}

func TestGetUserPasswordByEmail_DatabaseError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Ошибка '%s' при инициализации mock базы данных", err)
    }
    defer db.Close()

    orm := orm.NewORM(db)

    email := "test@example.com"

    mock.ExpectQuery("SELECT password, adminflag FROM users WHERE email=?").
        WithArgs(email).
        WillReturnError(errors.New("ошибка базы данных"))

    _, _, err = orm.GetUserPasswordByEmail(email)
    if err == nil || err.Error() != "ошибка базы данных" {
        t.Errorf("Ожидалась ошибка базы данных, получено %v", err)
    }
}
package orm_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/vexrina/cinemaLibrary/pkg/orm"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

func TestCreateActor(t *testing.T) {
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

func TestUpdateActor(t *testing.T) {
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

func TestCreateFilm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// create new orm with mockDB
	orm := orm.NewORM(db)

	// Create film to insert
	mockFilm := types.Film{
		Title:       "Test Film",
		Description: "This is a test film",
		ReleaseDate: "2024-03-15",
		Rating:      8.5,
		Actors:      []int{1, 2, 3},
	}

	// Set expect to mock
	mock.ExpectQuery("INSERT INTO films").WithArgs(mockFilm.Title, mockFilm.Description, mockFilm.ReleaseDate, mockFilm.Rating).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	for _, actorID := range mockFilm.Actors {
		mock.ExpectExec("INSERT INTO film_actors").WithArgs(1, actorID).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Call func
	filmID, err := orm.CreateFilm(mockFilm)

	// assert not error
	assert.NoError(t, err)
	// check id!=0
	assert.NotEqual(t, 0, filmID)

	// Check mock expectation
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

/*
func TestCreateFilm_ActorInsertError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mockFilm := types.Film{
        Title:       "Test Film",
        Description: "This is a test film",
        ReleaseDate: "2024-03-15",
        Rating:      8.5,
        Actors:      []int{1, 2, 3},
    }

    mock.ExpectQuery("INSERT INTO films").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))
    for _, actorID := range mockFilm.Actors {
        mock.ExpectExec("INSERT INTO film_actors").WithArgs(123, actorID).WillReturnError(errors.New("actor insert error"))
        mock.ExpectExec("INSERT INTO film_actors").WithArgs(123, actorID).WillReturnError(errors.New("actor insert error"))
    }

    orm := orm.NewORM(db)

    filmID, err := orm.CreateFilm(mockFilm)

    assert.Error(t, err)
    assert.Equal(t, 0, filmID)
    err = mock.ExpectationsWereMet()
    assert.NoError(t, err)
}
*/

func TestUpdateFilm_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mockFilm := types.Film{
		ID:          "1",
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
		ID:          "1",
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

func TestGetUserPasswordByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT password FROM users").WithArgs("test_email").WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow("hashed_password"))

	password, err := orm.GetUserPasswordByEmail("test_email")

	assert.NoError(t, err)
	assert.Equal(t, "hashed_password", password)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserPasswordByEmail_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	orm := orm.NewORM(db)

	mock.ExpectQuery("SELECT password FROM users").WithArgs("test_email").WillReturnError(errors.New("database error"))

	password, err := orm.GetUserPasswordByEmail("test_email")

	assert.Error(t, err)
	assert.Empty(t, password)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

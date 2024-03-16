package database_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/vexrina/cinemaLibrary/pkg/database"
)

func TestTableExists_True(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "existing_table"
    mock.ExpectQuery("SELECT EXISTS (.+)").WithArgs(tableName).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

    exists, err := database.TableExists(db, tableName)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if !exists {
        t.Errorf("Expected table '%s' to exist, but it does not", tableName)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestTableExists_False(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "non_existing_table"
    mock.ExpectQuery("SELECT EXISTS (.+)").WithArgs(tableName).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

    exists, err := database.TableExists(db, tableName)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if exists {
        t.Errorf("Expected table '%s' not to exist, but it does", tableName)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestTableExists_Error(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "error_table"
    expectedError := errors.New("database error")
    mock.ExpectQuery("SELECT EXISTS (.+)").WithArgs(tableName).WillReturnError(expectedError)

    exists, err := database.TableExists(db, tableName)
    if err == nil {
        t.Fatalf("Expected error, but got nil")
    }
    if exists {
        t.Errorf("Expected table '%s' not to exist, but it does", tableName)
    }
    if err.Error() != expectedError.Error() {
        t.Errorf("Expected error '%v', but got '%v'", expectedError, err)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestColumnsExist_AllColumnsExist(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "existing_table"
    columnNames := []string{"column1", "column2", "column3"}
    mock.ExpectQuery("SELECT column_name FROM information_schema.columns WHERE table_name = (.+)").
        WithArgs(tableName).
        WillReturnRows(sqlmock.NewRows([]string{"column_name"}).
            AddRow(columnNames[0]).
            AddRow(columnNames[1]).
            AddRow(columnNames[2]))

    exists, err := database.ColumnsExist(db, tableName, columnNames)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if !exists {
        t.Errorf("Expected all columns to exist, but they do not")
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestColumnsExist_NotAllColumnsExist(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "existing_table"
    columnNames := []string{"column1", "column2", "column3"}
    mock.ExpectQuery("SELECT column_name FROM information_schema.columns WHERE table_name = (.+)").
        WithArgs(tableName).
        WillReturnRows(sqlmock.NewRows([]string{"column_name"}).
            AddRow(columnNames[0]).
            AddRow(columnNames[1]))

    exists, err := database.ColumnsExist(db, tableName, columnNames)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if exists {
        t.Errorf("Expected not all columns to exist, but they do")
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestColumnsExist_Error(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableName := "existing_table"
    expectedError := errors.New("database error")
    mock.ExpectQuery("SELECT column_name FROM information_schema.columns WHERE table_name = (.+)").
        WithArgs(tableName).
        WillReturnError(expectedError)

    exists, err := database.ColumnsExist(db, tableName, []string{"column1", "column2", "column3"})
    if err == nil {
        t.Fatalf("Expected error, but got nil")
    }
    if exists {
        t.Errorf("Expected not all columns to exist, but they do")
    }
    if err.Error() != expectedError.Error() {
        t.Errorf("Expected error '%v', but got '%v'", expectedError, err)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

// I used map for create tables, but the map in golang is not deterministic, so
// that testcase is not deterministic
func TestChecker_AllTablesAndColumnsExist(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    tableOrder := []string{"users", "films", "actors", "film_actors"}

	for _, table := range tableOrder {
        columns := database.TableColumn[table]
        exists := true

        rows := sqlmock.NewRows([]string{"exists"}).AddRow(exists)
        mock.ExpectQuery("SELECT EXISTS (.+)").WithArgs(table).WillReturnRows(rows)

        if !exists {
            mock.ExpectExec(database.TableQuery[table]).WillReturnResult(sqlmock.NewResult(0, 0))
        }
        mock.ExpectQuery("SELECT column_name FROM information_schema.columns WHERE table_name = (.+)").
            WithArgs(table).
            WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow(columns[0]).AddRow(columns[1]))
    }

    err = database.Checker(db)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}

func TestConnectToPG_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	mock.ExpectPing()

	db, err := database.ConnectToPG("host=172.20.0.2 port=5432 dbname=test_db user=root password=root sslmode=disable")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if db == nil {
		t.Fatalf("Expected a non-nil database connection")
	}
}

func TestConnectToPG_Failure_CheckerError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	mock.ExpectPing()

	mock.ExpectExec("SELECT EXISTS (.+)").WillReturnError(sql.ErrNoRows)

	db, err := database.ConnectToPG("host=172.20.0.251 port=5432 dbname=test_db user=root password=root sslmode=disable")

	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}
	
	if db != nil {
		t.Fatalf("Expected a nil database connection but got non-nil")
	}
}
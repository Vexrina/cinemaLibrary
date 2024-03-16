package database

import (
	"database/sql"
	"log"
	
	_ "github.com/lib/pq"
)


var TableQuery = map[string]string{
	"users": `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL,
		email VARCHAR(100) NOT NULL,
		password VARCHAR(100) NOT NULL,
		adminflag BOOLEAN NOT NULL DEFAULT false)`,
	"films": `CREATE TABLE films (
		id SERIAL PRIMARY KEY,
		title VARCHAR(150) NOT NULL,
		description TEXT,
		release_date DATE NOT NULL,
		rating DECIMAL(3,1) NOT NULL CHECK (rating >= 0 AND rating <= 10))`,
	"actors": `CREATE TABLE actors (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		gender VARCHAR(10) NOT NULL,
		date_of_birth DATE NOT NULL)`,
	"film_actors": `CREATE TABLE film_actors (
		film_id INTEGER REFERENCES films(id) ON DELETE CASCADE,
		actor_id INTEGER REFERENCES actors(id) ON DELETE CASCADE,
		PRIMARY KEY (film_id, actor_id))`,
}
var TableColumn = map[string][]string{
	"users":  {"id", "username", "email", "password", "adminflag"},
	"films":  {"id", "title", "description", "release_date", "rating"},
	"actors": {"id", "name", "gender", "date_of_birth"},
	"film_actors": {"film_id", "actor_id"},
}


func ConnectToPG(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Success connect")

	if Checker(db) != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func TableExists(db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func ColumnsExist(db *sql.DB, tableName string, columnNames []string) (bool, error) {
	query := "SELECT column_name FROM information_schema.columns WHERE table_name = $1"

	rows, err := db.Query(query, tableName)
	if err != nil {
		log.Println("Error querying database:", err)
		return false, err
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Println("Error scanning row:", err)
			return false, err
		}
		existingColumns[columnName] = true
	}

	for _, columnName := range columnNames {
		if !existingColumns[columnName] {
			return false, nil
		}
	}

	return true, nil
}

func Checker(db *sql.DB) error {
	for table, columns := range TableColumn {
		exists, err := TableExists(db, table)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !exists {
			db.Exec(TableQuery[table])
			continue
		}
		exists, err = ColumnsExist(db, table, columns)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !exists {
			db.Exec("DROP TABLE" + table)
			db.Exec(TableQuery[table])
			continue
		}
	}
	return nil
}

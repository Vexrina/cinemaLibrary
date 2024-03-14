package database

import (
	"database/sql"
	"log"
	
	_ "github.com/lib/pq"
)

var tableQuery = map[string]string{
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
var tableColumn = map[string][]string{
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

	if checker(db) != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func columnsExist(db *sql.DB, tableName string, columnNames []string) (bool, error) {
	// Строим запрос для проверки существования колонок
	query := "SELECT column_name FROM information_schema.columns WHERE table_name = $1"

	// Выполняем запрос к базе данных
	rows, err := db.Query(query, tableName)
	if err != nil {
		log.Println("Error querying database:", err)
		return false, err
	}
	defer rows.Close()

	// Создаем мапу для хранения существующих колонок
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Println("Error scanning row:", err)
			return false, err
		}
		existingColumns[columnName] = true
	}

	// Проверяем, что все колонки из списка присутствуют в таблице
	for _, columnName := range columnNames {
		if !existingColumns[columnName] {
			return false, nil
		}
	}

	return true, nil
}

func checker(db *sql.DB) error {
	for table, columns := range tableColumn {
		exists, err := tableExists(db, table)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !exists {
			db.Exec(tableQuery[table])
			continue
		}
		exists, err = columnsExist(db, table, columns)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !exists {
			db.Exec("DROP TABLE" + table)
			db.Exec(tableQuery[table])
			continue
		}
	}
	return nil
}

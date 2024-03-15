// pkg/orm/orm.go
package orm

import (
	"database/sql"

	"github.com/vexrina/cinemaLibrary/pkg/types"
)

/*
Not sure how to correctly divide ORM into different packages/modules for prod, for example filmORM, actorORM, etc..
Wrote all the ORMs in one file, although it doesn't seem to be quite correct
*/

type ORM struct {
	db *sql.DB
}

func NewORM(db *sql.DB) *ORM {
	return &ORM{db: db}
}

func (orm *ORM) CreateActor(actor types.Actor) error {
	// Add new actor to database
	_, err := orm.db.Exec("INSERT INTO actors (name, gender, date_of_birth) VALUES ($1, $2, $3)", actor.Name, actor.Gender, actor.Birthdate)
	if err != nil {
		return err
	}
	return nil
}

func (orm *ORM) UpdateActor(actor types.Actor) error {
	// Update actor with ID
	_, err := orm.db.Exec("UPDATE actors SET name = $1, gender = $2, date_of_birth = $3 WHERE id = $4", actor.Name, actor.Gender, actor.Birthdate, actor.ID)
	if err != nil {
		return err
	}
	return nil
}

func (orm *ORM) CreateFilm(film types.Film) (int, error) {
	// Add new film to database
	var filmID int
	err := orm.db.QueryRow("INSERT INTO films (title, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id", film.Title, film.Description, film.ReleaseDate, film.Rating).Scan(&filmID)
	if err != nil {
		return 0, err
	}

	// Add actors in film to table "film_actor"
	for _, actorID := range film.Actors {
		_, err := orm.db.Exec("INSERT INTO film_actors (film_id, actor_id) VALUES ($1, $2)", filmID, actorID)
		if err != nil {
			return 0, err
		}
	}

	return filmID, nil
}

func (orm *ORM) UpdateFilm(film types.Film) error {
	// Update film with ID
	query := "UPDATE films SET title = $2, description = $3, release_date = $4, rating = $5 WHERE id = $1"
	_, err := orm.db.Exec(query, film.ID, film.Title, film.Description, film.ReleaseDate, film.Rating)
	if err != nil {
		return err
	}

	return nil
}

func (orm *ORM) CountUsersWithUsernameAndEmail(username, email string) (int, error) {
	// Considers users with such a username and email
	var count int
	err := orm.db.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1 OR email=$2", username, email).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (orm *ORM) CreateUser(username, email, hashedPassword string) error {
	// Create new user
	_, err := orm.db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", username, email, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func (orm *ORM) GetUserPasswordByEmail(email string) (string, error) {
	var storedPassword string
	err := orm.db.QueryRow("SELECT password FROM users WHERE email=$1", email).Scan(&storedPassword)
	if err != nil {
		return "", err
	}
	return storedPassword, nil
}

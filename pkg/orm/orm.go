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

// endpoint: /actor
// post
func (orm *ORM) CreateActor(actor types.Actor) error {
	_, err := orm.db.Exec("INSERT INTO actors (name, gender, date_of_birth) VALUES ($1, $2, $3)", actor.Name, actor.Gender, actor.Birthdate)
	if err != nil {
		return err
	}
	return nil
}

// patch
func (orm *ORM) UpdateActor(actor types.Actor) error {
	_, err := orm.db.Exec("UPDATE actors SET name = $1, gender = $2, date_of_birth = $3 WHERE id = $4", actor.Name, actor.Gender, actor.Birthdate, actor.ID)
	if err != nil {
		return err
	}
	return nil
}

// delete
func (orm *ORM) DeleteActorByID(id int) error {
	_, err := orm.db.Exec("DELETE FROM film_actors WHERE actor_id = $1", id)
	if err != nil {
		return err
	}

	_, err = orm.db.Exec("DELETE FROM actors WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

// get
// utility (exported for tests)
func (orm *ORM) GetFilmsWithActor(actorId int) ([]string, error){
	filmsQuery := `
		SELECT f.title
		FROM films AS f
		JOIN film_actors AS fa ON f.id = fa.film_id
		WHERE fa.actor_id = $1
	`
	filmsRows, err := orm.db.Query(filmsQuery, actorId)
	if err != nil {
		return nil, err
	}
	defer filmsRows.Close()

	var filmTitles []string
	for filmsRows.Next() {
		var title string
		if err := filmsRows.Scan(&title); err != nil {
			return nil, err
		}
		filmTitles = append(filmTitles, title)
	}
	if err := filmsRows.Err(); err != nil {
		return nil, err
	}

	return filmTitles, nil
}

func (orm *ORM) GetActors() ([]types.ActorWithFilms, error) {
	query := `SELECT id, name FROM actors`

	rows, err := orm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actorsWithFilms []types.ActorWithFilms
	for rows.Next() {
		var actor types.ActorWithFilms
		if err := rows.Scan(&actor.ID, &actor.Name); err != nil {
			return nil, err
		}
		filmTitles, err := orm.GetFilmsWithActor(actor.ID)
		if err!=nil{
			return nil, err
		}
		actor.FilmTitles = filmTitles
		actorsWithFilms = append(actorsWithFilms, actor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return actorsWithFilms, nil
}

func (orm *ORM) GetActorsWithFragment(actorFragment string) ([]types.ActorWithFilms, error) {
	query := `SELECT id, name FROM actors WHERE name LIKE $1`
	actorFragment = "%" + actorFragment + "%"
	rows, err := orm.db.Query(query, actorFragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actorsWithFilms []types.ActorWithFilms
	for rows.Next() {
		var actor types.ActorWithFilms
		if err := rows.Scan(&actor.ID, &actor.Name); err != nil {
			return nil, err
		}
		filmTitles, err := orm.GetFilmsWithActor(actor.ID)
		if err!=nil{
			return nil, err
		}
		actor.FilmTitles = filmTitles
		actorsWithFilms = append(actorsWithFilms, actor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return actorsWithFilms, nil
}

// endpoint: /actor

// endpoint: /film
// post
func (orm *ORM) CreateFilm(film types.Film) (int, error) {
	var filmID int
	err := orm.db.QueryRow("INSERT INTO films (title, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id", film.Title, film.Description, film.ReleaseDate, film.Rating).Scan(&filmID)
	if err != nil {
		return 0, err
	}

	for _, actorID := range film.Actors {
		_, err := orm.db.Exec("INSERT INTO film_actors (film_id, actor_id) VALUES ($1, $2)", filmID, actorID)
		if err != nil {
			return 0, err
		}
	}

	return filmID, nil
}

// patch
func (orm *ORM) UpdateFilm(film types.Film) error {
	query := "UPDATE films SET title = $2, description = $3, release_date = $4, rating = $5 WHERE id = $1"
	_, err := orm.db.Exec(query, film.ID, film.Title, film.Description, film.ReleaseDate, film.Rating)
	if err != nil {
		return err
	}

	return nil
}

// get
func (orm *ORM) GetFilms(sortBy string, ascending bool) ([]types.Film, error) {
	// Default query
	orderBy := "rating"

	// another sort
	switch sortBy {
	case "title":
		orderBy = "title"
	case "release_date":
		orderBy = "release_date"
	}
	if ascending {
		orderBy = orderBy + " ASC"
	} else {
		orderBy = orderBy + " DESC"
	}

	query := "SELECT * FROM films ORDER BY " + orderBy
	rows, err := orm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// create list of films
	var films []types.Film
	for rows.Next() {
		var film types.Film
		if err := rows.Scan(&film.ID, &film.Title, &film.Description, &film.ReleaseDate, &film.Rating); err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return films, nil
}

func (orm *ORM) SearchFilmsByFragment(fragment string) ([]types.Film, error) {
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

	rows, err := orm.db.Query(query, fragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []types.Film
	for rows.Next() {
		var film types.Film
		if err := rows.Scan(&film.ID, &film.Title, &film.Description, &film.ReleaseDate, &film.Rating); err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return films, nil
}

func (orm *ORM) SearchFilmsByActorFragment(actorFragment string) ([]types.Film, error) {
	query := `
        SELECT f.id, f.title, f.description, f.release_date, f.rating
        FROM films AS f
        JOIN film_actors AS fa ON f.id = fa.film_id
        JOIN actors AS a ON fa.actor_id = a.id
        WHERE a.name LIKE '%' || $1 || '%'
    `

	rows, err := orm.db.Query(query, actorFragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []types.Film
	for rows.Next() {
		var film types.Film
		if err := rows.Scan(&film.ID, &film.Title, &film.Description, &film.ReleaseDate, &film.Rating); err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return films, nil
}

func (orm *ORM) SearchFilmsByTitleFragment(titleFragment string) ([]types.Film, error) {
	query := "SELECT * FROM films WHERE title LIKE '%' || $1 || '%'"

	rows, err := orm.db.Query(query, titleFragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []types.Film
	for rows.Next() {
		var film types.Film
		if err := rows.Scan(&film.ID, &film.Title, &film.Description, &film.ReleaseDate, &film.Rating); err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return films, nil
}

// delete
func (orm *ORM) DeleteFilmByID(filmID int) error {
	deleteFilmActorsQuery := "DELETE FROM film_actors WHERE film_id = $1"
	_, err := orm.db.Exec(deleteFilmActorsQuery, filmID)
	if err != nil {
		return err
	}

	deleteFilmQuery := "DELETE FROM films WHERE id = $1"
	_, err = orm.db.Exec(deleteFilmQuery, filmID)
	if err != nil {
		return err
	}

	return nil
}

// endpoint: /film

// endpoint: /user
// utility function
func (orm *ORM) CountUsersWithUsernameAndEmail(username, email string) (int, error) {
	var count int
	err := orm.db.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1 OR email=$2", username, email).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// post
func (orm *ORM) CreateUser(username, email, hashedPassword string) error {
	_, err := orm.db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", username, email, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func (orm *ORM) GetUserPasswordByEmail(email string) (string, bool, error) {
	var storedPassword string
	var adminflag bool
	err := orm.db.QueryRow("SELECT password, adminflag FROM users WHERE email=$1", email).Scan(&storedPassword, &adminflag)
	if err != nil {
		return "", false, err
	}
	return storedPassword, adminflag, err
}

// endpoint: /user
